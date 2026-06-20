package svg_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"uisketch/layout"
	md "uisketch/markdown"
	"uisketch/render/svg"
	"uisketch/sketch"
)

func TestRendererAcceptanceCases(t *testing.T) {
	root := filepath.Clean("../../testdata/renderer")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read acceptance testdata: %v", err)
	}

	caseName := regexp.MustCompile(`^\d{3}_[a-z0-9]+(?:_[a-z0-9]+)*$`)
	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !caseName.MatchString(name) {
			t.Fatalf("renderer acceptance case %q does not match NNN_name", name)
		}
		dirs = append(dirs, name)
	}
	sort.Strings(dirs)
	if len(dirs) == 0 {
		t.Fatal("no renderer acceptance cases found")
	}

	covered := map[string]bool{}
	for i, dir := range dirs {
		wantPrefix := fmt.Sprintf("%03d", i+1)
		if !strings.HasPrefix(dir, wantPrefix+"_") {
			t.Fatalf("renderer acceptance case %q is out of sequence, want prefix %s_", dir, wantPrefix)
		}
		t.Run(dir, func(t *testing.T) {
			caseDir := filepath.Join(root, dir)
			inputPath := filepath.Join(caseDir, "input.md")
			outputPath := filepath.Join(caseDir, "output.svg")
			input, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("read input.md: %v", err)
			}
			want, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("read output.svg: %v", err)
			}
			source, err := md.SelectSource(string(input), 1)
			if err != nil {
				t.Fatalf("parse input.md: %v", err)
			}
			node, err := layout.ParseYAML(source.Source)
			if err != nil {
				t.Fatalf("parse layout source: %v", err)
			}
			collectComponentTypes(node, covered)
			got := svg.Render(sketch.FromLayout("", node)) + "\n"
			if got != string(want) {
				t.Fatalf("SVG output mismatch\n--- got ---\n%s\n--- want ---\n%s", got, string(want))
			}
		})
	}

	for _, component := range supportedAcceptanceComponents {
		if !covered[component] {
			t.Fatalf("renderer acceptance cases do not cover component %q", component)
		}
	}
}

var supportedAcceptanceComponents = []string{
	"browser", "window", "dialog", "menu", "mobile",
	"vstack", "hstack", "grid", "table-layout", "split-pane", "tabs", "sidebar", "section", "menubar", "spacer",
	"button", "icon-button", "floating-action-button", "badge-button", "toggle-button", "link",
	"label", "hint", "note", "review", "image", "custom-component",
	"input", "checkbox", "switch", "slider",
	"table", "list", "tree", "calendar", "badge",
}

func collectComponentTypes(n *layout.Node, covered map[string]bool) {
	if n == nil {
		return
	}
	covered[n.Type] = true
	for _, child := range n.Children {
		collectComponentTypes(child, covered)
	}
	for _, button := range n.Buttons {
		collectComponentTypes(button, covered)
	}
}
