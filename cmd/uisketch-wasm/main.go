//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"syscall/js"

	"uisketch/layout"
	md "uisketch/markdown"
	"uisketch/render/ascii"
	"uisketch/render/svg"
	"uisketch/sketch"
	"uisketch/validate"
)

type wasmNode struct {
	Type     string     `json:"type"`
	ID       string     `json:"id,omitempty"`
	Title    string     `json:"title,omitempty"`
	Label    string     `json:"label,omitempty"`
	Action   string     `json:"action,omitempty"`
	Anchor   string     `json:"anchor,omitempty"`
	Address  string     `json:"address,omitempty"`
	Hint     string     `json:"hint,omitempty"`
	Prompt   string     `json:"prompt,omitempty"`
	Data     any        `json:"data,omitempty"`
	Name     string     `json:"name,omitempty"`
	Purpose  string     `json:"purpose,omitempty"`
	Badge    string     `json:"badge,omitempty"`
	Size     string     `json:"size,omitempty"`
	Weight   string     `json:"weight,omitempty"`
	Columns  []string   `json:"columns,omitempty"`
	Labels   []tabLabel `json:"labels,omitempty"`
	Widths   []sizeSlot `json:"widths,omitempty"`
	Heights  []sizeSlot `json:"heights,omitempty"`
	Children []wasmNode `json:"children,omitempty"`
	Buttons  []wasmNode `json:"buttons,omitempty"`
}

type tabLabel struct {
	Text     string `json:"text"`
	Selected bool   `json:"selected,omitempty"`
}

type sizeSlot struct {
	Percent int  `json:"percent,omitempty"`
	Star    bool `json:"star,omitempty"`
}

type wasmFinding struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type wasmResponse struct {
	OK             bool          `json:"ok"`
	Error          string        `json:"error,omitempty"`
	Frontmatter    string        `json:"frontmatter,omitempty"`
	LayoutYAML     string        `json:"layoutYaml,omitempty"`
	NormalizedYAML string        `json:"normalizedYaml,omitempty"`
	Root           *wasmNode     `json:"root,omitempty"`
	SVG            string        `json:"svg,omitempty"`
	Text           string        `json:"text,omitempty"`
	Findings       []wasmFinding `json:"findings,omitempty"`
}

func main() {
	if input := js.Global().Get("uisketchWasmInput"); input.Type() == js.TypeString {
		js.Global().Set("uisketchWasmOutput", encode(responseForLayout(input.String())))
		return
	}
	js.Global().Set("uisketchLoadSource", js.FuncOf(loadSource))
	js.Global().Set("uisketchLoadLayout", js.FuncOf(loadLayout))
	js.Global().Set("uisketchRenderDocument", js.FuncOf(renderDocument))
	js.Global().Set("uisketchWasmReady", true)
	select {}
}

func loadSource(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return encodeError(fmt.Errorf("loadSource requires source text"))
	}
	source, err := md.SelectSource(args[0].String(), 1)
	if err != nil {
		return encodeError(err)
	}
	node, err := layout.ParseYAML(source.Source)
	if err != nil {
		return encodeError(err)
	}
	root := fromLayout(node)
	resp := responseForRoot(root)
	resp.LayoutYAML = source.Source
	return encode(resp)
}

func loadLayout(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return encodeError(fmt.Errorf("loadLayout requires YAML text"))
	}
	return encode(responseForLayout(args[0].String()))
}

func responseForLayout(source string) wasmResponse {
	node, err := layout.ParseYAML(source)
	if err != nil {
		return wasmResponse{OK: false, Error: err.Error(), Findings: []wasmFinding{{Severity: "error", Message: err.Error()}}}
	}
	root := fromLayout(node)
	resp := responseForRoot(root)
	resp.LayoutYAML = source
	return resp
}

func renderDocument(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return encodeError(fmt.Errorf("renderDocument requires a JSON node"))
	}
	var root wasmNode
	if err := json.Unmarshal([]byte(args[0].String()), &root); err != nil {
		return encodeError(err)
	}
	return encode(responseForRoot(&root))
}

func responseForRoot(root *wasmNode) wasmResponse {
	if root == nil {
		return wasmResponse{OK: false, Error: "layout root is required"}
	}
	layoutRoot := toLayout(root)
	doc := sketch.FromLayout("", layoutRoot)
	findings := validate.ValidateLayout(layoutRoot, nil)
	return wasmResponse{
		OK:             !validate.HasErrors(findings),
		Root:           root,
		NormalizedYAML: formatYAML(root),
		SVG:            svg.Render(doc),
		Text:           ascii.Render(doc),
		Findings:       fromFindings(findings),
	}
}

func encode(resp wasmResponse) string {
	data, err := json.Marshal(resp)
	if err != nil {
		return encodeError(err)
	}
	return string(data)
}

func encodeError(err error) string {
	data, _ := json.Marshal(wasmResponse{OK: false, Error: err.Error(), Findings: []wasmFinding{{Severity: "error", Message: err.Error()}}})
	return string(data)
}

func fromFindings(findings []validate.Finding) []wasmFinding {
	out := make([]wasmFinding, 0, len(findings))
	for _, finding := range findings {
		out = append(out, wasmFinding{Severity: string(finding.Severity), Message: finding.Message})
	}
	return out
}

func fromLayout(n *layout.Node) *wasmNode {
	if n == nil {
		return nil
	}
	if (n.Type == "expanded" || n.Type == "fixed-size") && len(n.Children) > 0 {
		return fromLayout(n.Children[0])
	}
	out := &wasmNode{
		Type:    n.Type,
		ID:      n.ID,
		Title:   n.Title,
		Label:   n.Label,
		Action:  n.Action,
		Anchor:  n.Anchor,
		Address: n.Address,
		Hint:    n.Hint,
		Prompt:  n.Prompt,
		Data:    n.Data,
		Name:    n.Name,
		Purpose: n.Purpose,
		Badge:   n.Badge,
		Columns: append([]string(nil), n.Columns...),
		Labels:  fromLayoutLabels(n.Labels),
		Widths:  fromLayoutSizeSlots(n.Widths),
		Heights: fromLayoutSizeSlots(n.Heights),
	}
	for _, child := range n.Children {
		if converted := fromLayout(child); converted != nil {
			out.Children = append(out.Children, *converted)
		}
	}
	for _, button := range n.Buttons {
		if converted := fromLayout(button); converted != nil {
			out.Buttons = append(out.Buttons, *converted)
		}
	}
	return out
}

func toLayout(n *wasmNode) *layout.Node {
	if n == nil {
		return nil
	}
	out := &layout.Node{
		Type:    n.Type,
		ID:      n.ID,
		Title:   n.Title,
		Label:   n.Label,
		Action:  n.Action,
		Anchor:  n.Anchor,
		Address: n.Address,
		Hint:    n.Hint,
		Prompt:  n.Prompt,
		Data:    n.Data,
		Name:    n.Name,
		Purpose: n.Purpose,
		Badge:   n.Badge,
		Columns: append([]string(nil), n.Columns...),
		Labels:  toLayoutLabels(n.Labels),
		Widths:  toLayoutSizeSlots(n.Widths),
		Heights: toLayoutSizeSlots(n.Heights),
	}
	for i := range n.Children {
		out.Children = append(out.Children, toLayout(&n.Children[i]))
	}
	for i := range n.Buttons {
		out.Buttons = append(out.Buttons, toLayout(&n.Buttons[i]))
	}
	return out
}

func fromLayoutLabels(labels []layout.TabLabel) []tabLabel {
	out := make([]tabLabel, len(labels))
	for i, label := range labels {
		out[i] = tabLabel{Text: label.Text, Selected: label.Selected}
	}
	return out
}

func toLayoutLabels(labels []tabLabel) []layout.TabLabel {
	out := make([]layout.TabLabel, len(labels))
	for i, label := range labels {
		out[i] = layout.TabLabel{Text: label.Text, Selected: label.Selected}
	}
	return out
}

func fromLayoutSizeSlots(slots []layout.SizeSlot) []sizeSlot {
	out := make([]sizeSlot, len(slots))
	for i, slot := range slots {
		out[i] = sizeSlot{Percent: slot.Percent, Star: slot.Star}
	}
	return out
}

func toLayoutSizeSlots(slots []sizeSlot) []layout.SizeSlot {
	out := make([]layout.SizeSlot, len(slots))
	for i, slot := range slots {
		out[i] = layout.SizeSlot{Percent: slot.Percent, Star: slot.Star}
	}
	return out
}

func formatYAML(root *wasmNode) string {
	if root == nil {
		return ""
	}
	var lines []string
	formatNode(&lines, root, 0, false)
	return joinLines(lines) + "\n"
}

func formatNode(lines *[]string, n *wasmNode, indent int, listItem bool) {
	pad := spaces(indent)
	if listItem {
		*lines = append(*lines, fmt.Sprintf("%s- %s:", pad, n.Type))
		indent += 4
	} else {
		*lines = append(*lines, fmt.Sprintf("%s%s:", pad, n.Type))
		indent += 2
	}
	props := []struct {
		key   string
		value string
	}{
		{"id", n.ID},
		{"title", n.Title},
		{"label", n.Label},
		{"action", n.Action},
		{"anchor", n.Anchor},
		{"address", n.Address},
		{"hint", n.Hint},
		{"prompt", n.Prompt},
		{"name", n.Name},
		{"purpose", n.Purpose},
		{"badge", n.Badge},
		{"size", n.Size},
		{"weight", n.Weight},
	}
	for _, prop := range props {
		if prop.value != "" {
			*lines = append(*lines, fmt.Sprintf("%s%s: %s", spaces(indent), prop.key, yamlScalar(prop.value)))
		}
	}
	if n.Data != nil {
		*lines = append(*lines, fmt.Sprintf("%sdata:", spaces(indent)))
		appendYAMLValue(lines, n.Data, indent+2)
	}
	if len(n.Columns) > 0 {
		*lines = append(*lines, spaces(indent)+"columns:")
		for _, column := range n.Columns {
			*lines = append(*lines, fmt.Sprintf("%s- %s", spaces(indent+2), yamlScalar(column)))
		}
	}
	if len(n.Labels) > 0 {
		*lines = append(*lines, spaces(indent)+"labels:")
		for _, label := range n.Labels {
			if label.Selected {
				*lines = append(*lines, fmt.Sprintf("%s- [%s]", spaces(indent+2), yamlScalar(label.Text)))
			} else {
				*lines = append(*lines, fmt.Sprintf("%s- %s", spaces(indent+2), yamlScalar(label.Text)))
			}
		}
	}
	if len(n.Widths) > 0 {
		*lines = append(*lines, fmt.Sprintf("%swidths: %s", spaces(indent), yamlSlots(n.Widths)))
	}
	if len(n.Heights) > 0 {
		*lines = append(*lines, fmt.Sprintf("%sheights: %s", spaces(indent), yamlSlots(n.Heights)))
	}
	if len(n.Children) > 0 {
		*lines = append(*lines, spaces(indent)+"children:")
		if n.Type == "tabs" && len(n.Labels) > 0 && len(n.Children) == 1 {
			formatNode(lines, &n.Children[0], indent+2, false)
		} else {
			for i := range n.Children {
				formatNode(lines, &n.Children[i], indent+2, true)
			}
		}
	}
	if len(n.Buttons) > 0 {
		*lines = append(*lines, spaces(indent)+"buttons:")
		for i := range n.Buttons {
			formatNode(lines, &n.Buttons[i], indent+2, true)
		}
	}
}

func yamlSlots(slots []sizeSlot) string {
	values := make([]string, len(slots))
	for i, slot := range slots {
		if slot.Star {
			values[i] = "$"
		} else {
			values[i] = fmt.Sprintf("%d", slot.Percent)
		}
	}
	return "[" + strings.Join(values, ", ") + "]"
}

func yamlScalar(s string) string {
	for _, r := range s {
		if !(r == ' ' || r == '-' || r == '_' || r == '.' || r == ':' || r == '/' || r == '@' ||
			(r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			data, _ := json.Marshal(s)
			return string(data)
		}
	}
	if s == "" {
		return `""`
	}
	return s
}

func appendYAMLValue(lines *[]string, value any, indent int) {
	switch v := value.(type) {
	case map[string]any:
		if len(v) == 0 {
			*lines = append(*lines, spaces(indent)+"{}")
			return
		}
		for _, key := range sortedKeys(v) {
			item := v[key]
			if yamlStructured(item) {
				*lines = append(*lines, fmt.Sprintf("%s%s:", spaces(indent), key))
				appendYAMLValue(lines, item, indent+2)
			} else {
				*lines = append(*lines, fmt.Sprintf("%s%s: %s", spaces(indent), key, yamlAnyScalar(item)))
			}
		}
	case []any:
		if len(v) == 0 {
			*lines = append(*lines, spaces(indent)+"[]")
			return
		}
		for _, item := range v {
			if yamlStructured(item) {
				*lines = append(*lines, spaces(indent)+"-")
				appendYAMLValue(lines, item, indent+2)
			} else {
				*lines = append(*lines, fmt.Sprintf("%s- %s", spaces(indent), yamlAnyScalar(item)))
			}
		}
	default:
		*lines = append(*lines, spaces(indent)+yamlAnyScalar(value))
	}
}

func yamlStructured(value any) bool {
	switch value.(type) {
	case map[string]any, []any:
		return true
	default:
		return false
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func yamlAnyScalar(value any) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case string:
		return yamlScalar(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprint(v)
	case float32:
		return fmt.Sprint(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(v)
	default:
		return yamlScalar(fmt.Sprint(v))
	}
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	out := lines[0]
	for _, line := range lines[1:] {
		out += "\n" + line
	}
	return out
}
