package markdown

import (
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

	"uisketch/layout"
	"uisketch/render/ascii"
	"uisketch/render/svg"
	"uisketch/sketch"
)

type Options struct {
	AssetDir string
	Format   string
}

type Result struct {
	Markdown string
	Assets   map[string]string
}

type Source struct {
	ID     string
	Format string
	Source string
}

type sourceBlock struct {
	ID     string
	Format string
	Source string
}

var (
	fenceStartRE   = regexp.MustCompile("^```([^`]*)$")
	imageRE        = regexp.MustCompile(`^!\[uisketch:([A-Za-z0-9_.:-]+)\]\(([^)]+)\)$`)
	commentStartRE = regexp.MustCompile(`^<!--\s+uisketch:source\b(.*)$`)
	attrRE         = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_-]*)="([^"]*)"`)
)

func Build(content string, opts Options) (Result, error) {
	if opts.AssetDir == "" {
		opts.AssetDir = "assets"
	}
	assetDir, err := cleanAssetDir(opts.AssetDir)
	if err != nil {
		return Result{}, err
	}
	opts.AssetDir = assetDir
	forcedFormat, err := normalizeBuildFormat(opts.Format)
	if err != nil {
		return Result{}, err
	}
	lines := splitLines(content)
	out := make([]string, 0, len(lines))
	assets := map[string]string{}
	ordinal := 1
	for i := 0; i < len(lines); {
		if id, _, ok := parseGeneratedImage(lines[i]); ok && i+1 < len(lines) && isSourceCommentStart(lines[i+1]) {
			block, next, err := parseSourceComment(lines, i+1)
			if err != nil {
				return Result{}, err
			}
			if block.ID == "" {
				block.ID = id
			}
			if forcedFormat == "" && block.Format != "svg" {
				return Result{}, fmt.Errorf("uisketch image output for %q has source format %q, want svg", block.ID, block.Format)
			}
			applyFormatOverride(&block, forcedFormat)
			rendered, assetPath, err := renderSource(&block, opts.AssetDir, ordinal)
			if err != nil {
				return Result{}, err
			}
			appendRendered(&out, assets, block, rendered, assetPath)
			ordinal++
			i = next
			continue
		}
		if isFenceStart(lines[i], "text") {
			fenceEnd := findFenceEnd(lines, i+1)
			if fenceEnd < 0 {
				return Result{}, fmt.Errorf("unterminated text fence at line %d", i+1)
			}
			if fenceEnd+1 < len(lines) && isSourceCommentStart(lines[fenceEnd+1]) {
				block, next, err := parseSourceComment(lines, fenceEnd+1)
				if err != nil {
					return Result{}, err
				}
				if forcedFormat == "" && block.Format != "txt" {
					return Result{}, fmt.Errorf("uisketch text output for %q has source format %q, want txt", block.ID, block.Format)
				}
				applyFormatOverride(&block, forcedFormat)
				rendered, assetPath, err := renderSource(&block, opts.AssetDir, ordinal)
				if err != nil {
					return Result{}, err
				}
				appendRendered(&out, assets, block, rendered, assetPath)
				ordinal++
				i = next
				continue
			}
		}
		if info, ok := parseFenceInfo(lines[i]); ok {
			format, _, isUisketch := normalizeUisketchInfo(info)
			if isUisketch {
				end := findFenceEnd(lines, i+1)
				if end < 0 {
					return Result{}, fmt.Errorf("unterminated uisketch fence at line %d", i+1)
				}
				source := strings.Join(lines[i+1:end], "\n")
				block := sourceBlock{Format: format, Source: source}
				applyFormatOverride(&block, forcedFormat)
				rendered, assetPath, err := renderSource(&block, opts.AssetDir, ordinal)
				if err != nil {
					return Result{}, err
				}
				appendRendered(&out, assets, block, rendered, assetPath)
				ordinal++
				i = end + 1
				continue
			}
		}
		out = append(out, lines[i])
		i++
	}
	return Result{Markdown: strings.Join(out, "\n"), Assets: assets}, nil
}

func SelectSource(content string, index int) (Source, error) {
	if index < 1 {
		return Source{}, fmt.Errorf("source index must be 1 or greater")
	}
	lines := splitLines(content)
	ordinal := 0
	for i := 0; i < len(lines); {
		if id, _, ok := parseGeneratedImage(lines[i]); ok && i+1 < len(lines) && isSourceCommentStart(lines[i+1]) {
			block, next, err := parseSourceComment(lines, i+1)
			if err != nil {
				return Source{}, err
			}
			if block.ID == "" {
				block.ID = id
			}
			ordinal++
			if ordinal == index {
				return Source{ID: block.ID, Format: block.Format, Source: block.Source}, nil
			}
			i = next
			continue
		}
		if isFenceStart(lines[i], "text") {
			fenceEnd := findFenceEnd(lines, i+1)
			if fenceEnd < 0 {
				return Source{}, fmt.Errorf("unterminated text fence at line %d", i+1)
			}
			if fenceEnd+1 < len(lines) && isSourceCommentStart(lines[fenceEnd+1]) {
				block, next, err := parseSourceComment(lines, fenceEnd+1)
				if err != nil {
					return Source{}, err
				}
				ordinal++
				if ordinal == index {
					return Source{ID: block.ID, Format: block.Format, Source: block.Source}, nil
				}
				i = next
				continue
			}
		}
		if info, ok := parseFenceInfo(lines[i]); ok {
			format, _, isUisketch := normalizeUisketchInfo(info)
			if isUisketch {
				end := findFenceEnd(lines, i+1)
				if end < 0 {
					return Source{}, fmt.Errorf("unterminated uisketch fence at line %d", i+1)
				}
				ordinal++
				if ordinal == index {
					return Source{Format: format, Source: strings.Join(lines[i+1:end], "\n")}, nil
				}
				i = end + 1
				continue
			}
		}
		i++
	}
	if ordinal == 0 {
		return Source{}, fmt.Errorf("no uisketch source fences or generated source comments found")
	}
	return Source{}, fmt.Errorf("source index %d out of range; found %d renderable source(s)", index, ordinal)
}

func normalizeBuildFormat(format string) (string, error) {
	switch strings.TrimSpace(format) {
	case "", "source":
		return "", nil
	case "svg":
		return "svg", nil
	case "ascii", "txt", "text":
		return "txt", nil
	default:
		return "", fmt.Errorf("unsupported markdown format %q; supported formats: svg, ascii, source", format)
	}
}

func applyFormatOverride(block *sourceBlock, forcedFormat string) {
	if forcedFormat != "" {
		block.Format = forcedFormat
	}
}

func appendRendered(out *[]string, assets map[string]string, block sourceBlock, rendered, assetPath string) {
	if block.Format == "svg" {
		*out = append(*out, fmt.Sprintf("![uisketch:%s](%s)", block.ID, assetPath))
		assets[assetPath] = rendered
	} else {
		*out = append(*out, "```text")
		*out = append(*out, splitLines(strings.TrimRight(rendered, "\n"))...)
		*out = append(*out, "```")
	}
	*out = append(*out, sourceComment(block)...)
}

func AssetPaths(result Result) []string {
	paths := make([]string, 0, len(result.Assets))
	for p := range result.Assets {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}

func renderSource(block *sourceBlock, assetDir string, ordinal int) (string, string, error) {
	if strings.TrimSpace(block.Source) == "" {
		return "", "", fmt.Errorf("empty uisketch source")
	}
	if strings.Contains(block.Source, "--") {
		return "", "", fmt.Errorf("uisketch source cannot be embedded safely in an html comment because it contains --")
	}
	node, err := layout.ParseYAML(block.Source)
	if err != nil {
		return "", "", err
	}
	id := node.ID
	if id == "" {
		id = block.ID
	}
	if id == "" {
		id = fmt.Sprintf("uisketch-%03d", ordinal)
	}
	block.ID = id
	doc := sketch.FromLayout("", node)
	switch block.Format {
	case "svg":
		return svg.Render(doc), path.Join(assetDir, sanitizeID(id)+".svg"), nil
	case "txt":
		return ascii.Render(doc), "", nil
	default:
		return "", "", fmt.Errorf("unsupported uisketch format %q", block.Format)
	}
}

func sourceComment(block sourceBlock) []string {
	info := "uisketch:svg"
	if block.Format == "txt" {
		info = "uisketch:txt"
	}
	source := strings.TrimRight(block.Source, "\n")
	lines := []string{
		fmt.Sprintf(`<!-- uisketch:source id="%s" format="%s"`, block.ID, block.Format),
		"```" + info,
	}
	lines = append(lines, splitLines(source)...)
	lines = append(lines, "```", "-->")
	return lines
}

func parseSourceComment(lines []string, start int) (sourceBlock, int, error) {
	attrs := map[string]string{}
	for _, match := range attrRE.FindAllStringSubmatch(lines[start], -1) {
		attrs[match[1]] = match[2]
	}
	var fenceStart int = -1
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "-->" {
			return sourceBlock{}, i + 1, fmt.Errorf("uisketch source comment at line %d has no source fence", start+1)
		}
		if info, ok := parseFenceInfo(lines[i]); ok {
			_, _, isUisketch := normalizeUisketchInfo(info)
			if !isUisketch {
				return sourceBlock{}, i + 1, fmt.Errorf("uisketch source comment at line %d contains non-uisketch fence", start+1)
			}
			fenceStart = i
			break
		}
	}
	if fenceStart < 0 {
		return sourceBlock{}, start + 1, fmt.Errorf("unterminated uisketch source comment at line %d", start+1)
	}
	info, _ := parseFenceInfo(lines[fenceStart])
	format, _, _ := normalizeUisketchInfo(info)
	if attrs["format"] != "" {
		if attrs["format"] != "svg" && attrs["format"] != "txt" {
			return sourceBlock{}, start + 1, fmt.Errorf("unsupported uisketch source comment format %q", attrs["format"])
		}
		format = attrs["format"]
	}
	fenceEnd := findFenceEnd(lines, fenceStart+1)
	if fenceEnd < 0 {
		return sourceBlock{}, start + 1, fmt.Errorf("unterminated uisketch source fence at line %d", fenceStart+1)
	}
	if fenceEnd+1 >= len(lines) || strings.TrimSpace(lines[fenceEnd+1]) != "-->" {
		return sourceBlock{}, fenceEnd + 1, fmt.Errorf("uisketch source comment at line %d is not closed immediately after source fence", start+1)
	}
	return sourceBlock{
		ID:     attrs["id"],
		Format: format,
		Source: strings.Join(lines[fenceStart+1:fenceEnd], "\n"),
	}, fenceEnd + 2, nil
}

func parseGeneratedImage(line string) (string, string, bool) {
	match := imageRE.FindStringSubmatch(strings.TrimSpace(line))
	if match == nil {
		return "", "", false
	}
	return match[1], match[2], true
}

func isSourceCommentStart(line string) bool {
	return commentStartRE.MatchString(strings.TrimSpace(line))
}

func parseFenceInfo(line string) (string, bool) {
	match := fenceStartRE.FindStringSubmatch(strings.TrimSpace(line))
	if match == nil {
		return "", false
	}
	return strings.TrimSpace(match[1]), true
}

func isFenceStart(line, want string) bool {
	info, ok := parseFenceInfo(line)
	return ok && info == want
}

func findFenceEnd(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "```" {
			return i
		}
	}
	return -1
}

func normalizeUisketchInfo(info string) (format, normalized string, ok bool) {
	switch strings.TrimSpace(info) {
	case "uisketch", "uisketch:svg":
		return "svg", "uisketch:svg", true
	case "uisketch:txt", "uisketch:text", "uisketch:ascii":
		return "txt", "uisketch:txt", true
	default:
		return "", "", false
	}
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func sanitizeID(id string) string {
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), ".-")
	if out == "" {
		return "uisketch"
	}
	return out
}

func cleanAssetDir(assetDir string) (string, error) {
	cleaned := path.Clean(strings.ReplaceAll(assetDir, "\\", "/"))
	if cleaned == "." || cleaned == "/" || strings.HasPrefix(cleaned, "../") || cleaned == ".." || strings.HasPrefix(cleaned, "/") {
		return "", fmt.Errorf("asset dir %q must be a relative path inside the markdown output directory", assetDir)
	}
	return cleaned, nil
}
