package layout

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml/ast"
	yamlparser "github.com/goccy/go-yaml/parser"
)

type File struct {
	ID          string
	Type        string
	Title       string
	ScreenID    string
	LayoutYAML  string
	Layout      *Node
	SourcePath  string
	Frontmatter string
}

type Node struct {
	Type        string
	ID          string
	Title       string
	Label       string
	Action      string
	Anchor      string
	LegacyTo    string
	Address     string
	Hint        string
	Name        string
	Purpose     string
	Badge       string
	Columns     []string
	GridColumns int
	Labels      []TabLabel
	Widths      []SizeSlot
	Heights     []SizeSlot
	Children    []*Node
	Buttons     []*Node
}

type TabLabel struct {
	Text     string
	Selected bool
}

type SizeSlot struct {
	Percent int
	Star    bool
}

func LoadFile(path string) (*File, error) {
	return LoadFileAt(path, 1)
}

func LoadFileAt(path string, layoutIndex int) (*File, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseFileAt(path, string(body), layoutIndex)
}

func ParseFile(path, content string) (*File, error) {
	return ParseFileAt(path, content, 1)
}

func ParseFileAt(path, content string, layoutIndex int) (*File, error) {
	if layoutIndex < 1 {
		return nil, fmt.Errorf("layout index must be 1 or greater")
	}
	fm, markdown, markdownStartLine, err := splitFrontmatter(content)
	if err != nil {
		return nil, err
	}
	file := &File{SourcePath: path, Frontmatter: fm}
	parseSketchFrontmatter(file, fm)
	if file.Type != "uisketch" {
		return nil, fmt.Errorf("frontmatter type is %q, want uisketch", file.Type)
	}
	layoutYAML, layoutStartLine, err := extractLayoutBlock(markdown, markdownStartLine, layoutIndex)
	if err != nil {
		return nil, err
	}
	node, err := ParseYAMLAt(layoutYAML, layoutStartLine-1)
	if err != nil {
		return nil, err
	}
	file.LayoutYAML = layoutYAML
	file.Layout = node
	return file, nil
}

func ParseYAML(src string) (*Node, error) {
	return ParseYAMLAt(src, 0)
}

func ParseYAMLAt(src string, lineOffset int) (*Node, error) {
	if strings.TrimSpace(src) == "" {
		return nil, fmt.Errorf("empty layout yaml")
	}
	file, err := yamlparser.ParseBytes([]byte(src), 0)
	if err != nil {
		return nil, offsetYAMLError(err, lineOffset)
	}
	if len(file.Docs) != 1 || file.Docs[0].Body == nil {
		return nil, fmt.Errorf("line %d: expected one layout YAML document", lineOffset+1)
	}
	ctx := &parseContext{lineOffset: lineOffset}
	node := parseLayoutNode(file.Docs[0].Body, ctx, "document root")
	if ctx.hasErrors() {
		return nil, ctx.err()
	}
	return node, nil
}

type parseContext struct {
	lineOffset  int
	diagnostics []string
}

func (ctx *parseContext) add(n ast.Node, format string, args ...any) {
	ctx.diagnostics = append(ctx.diagnostics, fmt.Sprintf("line %d: %s", actualLine(ctx, n), fmt.Sprintf(format, args...)))
}

func (ctx *parseContext) hasErrors() bool {
	return len(ctx.diagnostics) > 0
}

func (ctx *parseContext) err() error {
	return fmt.Errorf("%s", strings.Join(ctx.diagnostics, "\n"))
}

func parseLayoutNode(n ast.Node, ctx *parseContext, parent string) *Node {
	mv := asSingleMappingValue(n, ctx, parent)
	if mv == nil {
		return nil
	}
	nodeType := keyString(mv.Key)
	if nodeType == "" {
		ctx.add(mv.Key, "expected layout node key under %s", parent)
		return nil
	}
	out := &Node{Type: nodeType}
	if isScalarLabelNode(nodeType) && isScalarNode(mv.Value) {
		out.Label = scalarString(mv.Value)
		return out
	}
	if mv.Value == nil || mv.Value.Type() == ast.NullType {
		return out
	}
	props, ok := mv.Value.(*ast.MappingNode)
	if !ok {
		ctx.add(mv.Value, "node %s under %s must use mapping properties", describeNode(out), parent)
		return out
	}
	parseNodeProperties(out, props, ctx)
	return out
}

func asSingleMappingValue(n ast.Node, ctx *parseContext, parent string) *ast.MappingValueNode {
	switch node := n.(type) {
	case *ast.MappingValueNode:
		return node
	case *ast.MappingNode:
		if len(node.Values) == 0 {
			ctx.add(node, "expected layout node key under %s", parent)
			return nil
		}
		if len(node.Values) != 1 {
			ctx.add(node.Values[1].Key, "expected exactly one layout node key under %s; found %s", parent, joinMapKeys(node))
			return nil
		}
		return node.Values[0]
	default:
		ctx.add(n, "expected layout node mapping under %s, got %s", parent, n.Type().YAMLName())
		return nil
	}
}

func parseNodeProperties(out *Node, props *ast.MappingNode, ctx *parseContext) {
	for _, prop := range props.Values {
		name := keyString(prop.Key)
		if name == "" {
			ctx.add(prop.Key, "expected property key under %s", describeNode(out))
			continue
		}
		switch name {
		case "id":
			out.ID = scalarString(prop.Value)
		case "label":
			out.Label = labelLikeValue(prop.Value)
		case "title":
			out.Title = labelLikeValue(prop.Value)
		case "hint":
			out.Hint = scalarString(prop.Value)
		case "name":
			out.Name = scalarString(prop.Value)
		case "purpose":
			out.Purpose = scalarString(prop.Value)
		case "badge", "count":
			out.Badge = scalarString(prop.Value)
		case "action":
			out.Action = scalarString(prop.Value)
		case "anchor":
			out.Anchor = scalarString(prop.Value)
		case "to":
			out.Anchor = scalarString(prop.Value)
			out.LegacyTo = out.Anchor
		case "address":
			out.Address = scalarString(prop.Value)
		case "children":
			if out.Type == "tabs" {
				children, ok := parseTabChildren(prop.Value, ctx, "children", describeNode(out))
				if ok {
					out.Children = children
				}
			} else {
				children, ok := parseNodeList(prop.Value, ctx, "children", describeNode(out))
				if ok {
					out.Children = children
				}
			}
		case "buttons":
			buttons, ok := parseNodeList(prop.Value, ctx, "buttons", describeNode(out))
			if ok {
				out.Buttons = buttons
			}
		case "columns":
			if out.Type == "grid" {
				columns, ok := parsePositiveInt(prop.Value, ctx, "columns", describeNode(out))
				if ok {
					out.GridColumns = columns
				}
			} else {
				columns, ok := parseStringList(prop.Value, ctx, "columns", describeNode(out))
				if ok {
					out.Columns = columns
				}
			}
		case "labels":
			labels, ok := parseTabLabels(prop.Value, ctx, "labels", describeNode(out))
			if ok {
				out.Labels = labels
			}
		case "widths":
			widths, ok := parseSizeSlots(prop.Value, ctx, "widths", describeNode(out))
			if ok {
				out.Widths = widths
			}
		case "heights":
			heights, ok := parseSizeSlots(prop.Value, ctx, "heights", describeNode(out))
			if ok {
				out.Heights = heights
			}
		case "child":
			child := parseLayoutNode(prop.Value, ctx, "child of "+describeNode(out))
			if child != nil {
				out.Children = []*Node{child}
			}
		case "data", "highlight", "fallback":
			continue
		default:
			if suggestion := closestKnownProperty(name); suggestion != "" {
				ctx.add(prop.Key, "unknown property %q under %s; did you mean %q?", name, describeNode(out), suggestion)
			}
		}
	}
}

func parseTabChildren(n ast.Node, ctx *parseContext, prop, parent string) ([]*Node, bool) {
	if _, ok := n.(*ast.SequenceNode); ok {
		return parseNodeList(n, ctx, prop, parent)
	}
	child := parseLayoutNode(n, ctx, prop+" of "+parent)
	if child == nil {
		return nil, false
	}
	return []*Node{child}, true
}

func parseNodeList(n ast.Node, ctx *parseContext, prop, parent string) ([]*Node, bool) {
	seq, ok := n.(*ast.SequenceNode)
	if !ok {
		ctx.add(n, "%s under %s must be a list", prop, parent)
		return nil, false
	}
	var out []*Node
	for _, item := range seq.Values {
		if isScalarNode(item) {
			if nodeType := scalarString(item); isScalarChildNode(nodeType) {
				out = append(out, &Node{Type: nodeType})
				continue
			}
		}
		child := parseLayoutNode(item, ctx, prop+" of "+parent)
		if child != nil {
			out = append(out, child)
		}
	}
	return out, true
}

func parsePositiveInt(n ast.Node, ctx *parseContext, prop, parent string) (int, bool) {
	if !isScalarNode(n) {
		ctx.add(n, "%s under %s must be a positive integer", prop, parent)
		return 0, false
	}
	value, err := strconv.Atoi(scalarString(n))
	if err != nil || value <= 0 {
		ctx.add(n, "%s under %s must be a positive integer", prop, parent)
		return 0, false
	}
	return value, true
}

func parseTabLabels(n ast.Node, ctx *parseContext, prop, parent string) ([]TabLabel, bool) {
	seq, ok := n.(*ast.SequenceNode)
	if !ok {
		ctx.add(n, "%s under %s must be a list", prop, parent)
		return nil, false
	}
	var out []TabLabel
	for _, item := range seq.Values {
		if isScalarNode(item) {
			out = append(out, TabLabel{Text: scalarString(item)})
			continue
		}
		nested, ok := item.(*ast.SequenceNode)
		if !ok || len(nested.Values) != 1 || !isScalarNode(nested.Values[0]) {
			ctx.add(item, "%s under %s must contain string labels or single-item selected label lists", prop, parent)
			continue
		}
		out = append(out, TabLabel{Text: scalarString(nested.Values[0]), Selected: true})
	}
	return out, true
}

func parseSizeSlots(n ast.Node, ctx *parseContext, prop, parent string) ([]SizeSlot, bool) {
	seq, ok := n.(*ast.SequenceNode)
	if !ok {
		ctx.add(n, "%s under %s must be a list", prop, parent)
		return nil, false
	}
	var out []SizeSlot
	for _, item := range seq.Values {
		if !isScalarNode(item) {
			ctx.add(item, "%s under %s must contain percentages or \"*\"", prop, parent)
			continue
		}
		value := scalarString(item)
		if value == "*" {
			out = append(out, SizeSlot{Star: true})
			continue
		}
		percent, err := strconv.Atoi(value)
		if err != nil {
			ctx.add(item, "%s under %s must contain integer percentages or \"*\"", prop, parent)
			continue
		}
		out = append(out, SizeSlot{Percent: percent})
	}
	return out, true
}

func parseStringList(n ast.Node, ctx *parseContext, prop, parent string) ([]string, bool) {
	seq, ok := n.(*ast.SequenceNode)
	if !ok {
		ctx.add(n, "%s under %s must be a list", prop, parent)
		return nil, false
	}
	var out []string
	for _, item := range seq.Values {
		if !isScalarNode(item) {
			ctx.add(item, "%s under %s must contain scalar values", prop, parent)
			continue
		}
		out = append(out, scalarString(item))
	}
	return out, true
}

func keyString(n ast.Node) string {
	if n == nil {
		return ""
	}
	return strings.TrimSpace(n.String())
}

func scalarString(n ast.Node) string {
	if n == nil || n.Type() == ast.NullType {
		return ""
	}
	if s, ok := n.(ast.ScalarNode); ok {
		return fmt.Sprint(s.GetValue())
	}
	return strings.TrimSpace(n.String())
}

func labelLikeValue(n ast.Node) string {
	if m, ok := n.(*ast.MappingNode); ok {
		for _, prop := range m.Values {
			if keyString(prop.Key) == "vocabulary" {
				return scalarString(prop.Value)
			}
		}
		return ""
	}
	return scalarString(n)
}

func isScalarNode(n ast.Node) bool {
	if n == nil {
		return false
	}
	_, ok := n.(ast.ScalarNode)
	return ok
}

func actualLine(ctx *parseContext, n ast.Node) int {
	if n == nil || n.GetToken() == nil || n.GetToken().Position == nil {
		return ctx.lineOffset + 1
	}
	return ctx.lineOffset + n.GetToken().Position.Line - 1
}

func describeNode(n *Node) string {
	if n == nil {
		return "unknown parent"
	}
	if n.ID != "" {
		return fmt.Sprintf("%s#%s", n.Type, n.ID)
	}
	if n.Title != "" {
		return fmt.Sprintf("%s(%q)", n.Type, n.Title)
	}
	if n.Label != "" {
		return fmt.Sprintf("%s(%q)", n.Type, n.Label)
	}
	return n.Type
}

func joinMapKeys(n *ast.MappingNode) string {
	var keys []string
	for _, value := range n.Values {
		keys = append(keys, keyString(value.Key))
	}
	return strings.Join(keys, ", ")
}

func closestKnownProperty(name string) string {
	best := ""
	bestDistance := 3
	for _, candidate := range knownProperties {
		distance := editDistance(name, candidate)
		if distance < bestDistance {
			bestDistance = distance
			best = candidate
		}
	}
	return best
}

var knownProperties = []string{
	"id", "label", "title", "hint", "name", "purpose", "badge", "count",
	"action", "anchor", "to", "address", "children", "buttons", "columns",
	"labels", "widths", "heights", "child", "data", "highlight", "fallback",
}

func editDistance(a, b string) int {
	ar, br := []rune(a), []rune(b)
	prev := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur := make([]int, len(br)+1)
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 0
			if ar[i-1] != br[j-1] {
				cost = 1
			}
			cur[j] = minInt(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev = cur
	}
	return prev[len(br)]
}

func minInt(values ...int) int {
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
}

var yamlPositionRE = regexp.MustCompile(`\[(\d+):(\d+)\]`)

func offsetYAMLError(err error, lineOffset int) error {
	if lineOffset == 0 {
		return err
	}
	message := yamlPositionRE.ReplaceAllStringFunc(err.Error(), func(match string) string {
		parts := yamlPositionRE.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		var line, col int
		fmt.Sscanf(parts[1], "%d", &line)
		fmt.Sscanf(parts[2], "%d", &col)
		return fmt.Sprintf("[%d:%d]", line+lineOffset-1, col)
	})
	return fmt.Errorf("%s", message)
}

func isScalarLabelNode(t string) bool {
	switch t {
	case "button", "icon-button", "floating-action-button", "badge-button", "toggle-button", "link",
		"label", "hint", "note", "review", "badge",
		"input", "checkbox", "switch", "slider",
		"image", "custom-component",
		"table", "list", "tree", "calendar":
		return true
	default:
		return false
	}
}

func isScalarChildNode(t string) bool {
	return t == "spacer" || isScalarLabelNode(t)
}

func splitFrontmatter(content string) (string, string, int, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(content, "---\n") {
		return "", "", 0, fmt.Errorf("missing frontmatter")
	}
	rest := content[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", "", 0, fmt.Errorf("unterminated frontmatter")
	}
	fm := rest[:end]
	body := rest[end+len("\n---"):]
	bodyStartLine := strings.Count(fm, "\n") + 4
	if strings.HasPrefix(body, "\n") {
		body = body[1:]
	} else {
		bodyStartLine--
	}
	return fm, body, bodyStartLine, nil
}

func parseSketchFrontmatter(file *File, fm string) {
	var section string
	for _, raw := range strings.Split(fm, "\n") {
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := countIndent(line)
		trimmed := strings.TrimSpace(line)
		if indent == 0 {
			section = ""
		}
		key, value, ok := splitKeyValue(trimmed)
		if !ok {
			continue
		}
		if indent == 0 && value == "" {
			section = key
			continue
		}
		switch {
		case indent == 0 && key == "id":
			file.ID = value
		case indent == 0 && key == "type":
			file.Type = value
		case indent == 0 && key == "title":
			file.Title = value
		case section == "screen" && key == "id":
			file.ScreenID = value
		}
	}
}

func extractLayoutBlock(markdown string, baseLine, layoutIndex int) (string, int, error) {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	matchedSources := 0
	for i := 0; i < len(lines); {
		trimmed := strings.TrimSpace(lines[i])
		if isGeneratedOutputLine(trimmed) && i+1 < len(lines) && isUisketchSourceCommentStart(lines[i+1]) {
			source, sourceStartLine, next, err := extractSourceCommentBlock(lines, i+1, baseLine)
			if err != nil {
				return "", 0, err
			}
			matchedSources++
			if matchedSources == layoutIndex {
				return source, sourceStartLine, nil
			}
			i = next
			continue
		}

		if trimmed == "```text" {
			end := findFenceEnd(lines, i+1)
			if end < 0 {
				return "", 0, fmt.Errorf("unterminated text fence at line %d", baseLine+i)
			}
			if end+1 < len(lines) && isUisketchSourceCommentStart(lines[end+1]) {
				source, sourceStartLine, next, err := extractSourceCommentBlock(lines, end+1, baseLine)
				if err != nil {
					return "", 0, err
				}
				matchedSources++
				if matchedSources == layoutIndex {
					return source, sourceStartLine, nil
				}
				i = next
				continue
			}
			i = end + 1
			continue
		}

		if format, ok := normalizeUisketchFence(trimmed); ok {
			_ = format
			end := findFenceEnd(lines, i+1)
			if end < 0 {
				return "", 0, fmt.Errorf("unterminated uisketch fence at line %d", baseLine+i)
			}
			matchedSources++
			if matchedSources == layoutIndex {
				return strings.Join(lines[i+1:end], "\n"), baseLine + i + 1, nil
			}
			i = end + 1
			continue
		}
		i++
	}
	if matchedSources == 0 {
		return "", 0, fmt.Errorf("no uisketch source fences or generated source comments found")
	}
	return "", 0, fmt.Errorf("source index %d out of range; found %d renderable source(s)", layoutIndex, matchedSources)
}

func extractSourceCommentBlock(lines []string, start, baseLine int) (string, int, int, error) {
	fenceStart := -1
	for i := start + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "-->" {
			return "", 0, i + 1, fmt.Errorf("uisketch source comment at line %d has no source fence", baseLine+start)
		}
		if _, ok := normalizeUisketchFence(trimmed); ok {
			fenceStart = i
			break
		}
	}
	if fenceStart < 0 {
		return "", 0, start + 1, fmt.Errorf("unterminated uisketch source comment at line %d", baseLine+start)
	}
	fenceEnd := findFenceEnd(lines, fenceStart+1)
	if fenceEnd < 0 {
		return "", 0, start + 1, fmt.Errorf("unterminated uisketch source fence at line %d", baseLine+fenceStart)
	}
	if fenceEnd+1 >= len(lines) || strings.TrimSpace(lines[fenceEnd+1]) != "-->" {
		return "", 0, fenceEnd + 1, fmt.Errorf("uisketch source comment at line %d is not closed immediately after source fence", baseLine+start)
	}
	return strings.Join(lines[fenceStart+1:fenceEnd], "\n"), baseLine + fenceStart + 1, fenceEnd + 2, nil
}

func isGeneratedOutputLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "![uisketch:") || trimmed == "```text"
}

func isUisketchSourceCommentStart(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "<!--") && strings.Contains(line, "uisketch:source")
}

func normalizeUisketchFence(trimmed string) (string, bool) {
	switch trimmed {
	case "```uisketch", "```uisketch:svg":
		return "svg", true
	case "```uisketch:txt", "```uisketch:text", "```uisketch:ascii":
		return "txt", true
	default:
		return "", false
	}
}

func findFenceEnd(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "```" {
			return i
		}
	}
	return -1
}

func splitKeyValue(line string) (string, string, bool) {
	before, after, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(before), unquote(strings.TrimSpace(after)), true
}

func countIndent(s string) int {
	n := 0
	for _, r := range s {
		if r != ' ' {
			break
		}
		n++
	}
	return n
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
