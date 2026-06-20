package svg

import (
	"strings"
	"testing"

	"uisketch/sketch"
)

func TestRenderEquipmentList(t *testing.T) {
	doc := &sketch.Document{
		Title: "Equipment <List>",
		Root: &sketch.Node{
			Type:    "browser",
			Title:   "Equipment <List>",
			Address: "https://example.internal/equipment?a=1&b=2",
			Children: []*sketch.Node{
				{
					Type: "vstack",
					Children: []*sketch.Node{
						{
							Type: "hstack",
							Children: []*sketch.Node{
								{Type: "button", Label: "Refresh"},
								{Type: "button", Label: "Create Alert"},
							},
						},
						{Type: "table", Columns: []string{"Equipment", "Status"}},
						{Type: "label", Label: "Ready"},
					},
				},
			},
		},
	}

	got := Render(doc)
	for _, want := range []string{
		`<svg xmlns="http://www.w3.org/2000/svg"`,
		`filter id="wobble"`,
		`Equipment &lt;List&gt;`,
		`https://example.internal/equipment?a=1&amp;b=2`,
		`<polygon points="28,91 46,79 46,103" class="nav"/>`,
		`<polygon points="62,79 80,91 62,103" class="nav"/>`,
		`class="nav-stroke"`,
		`Refresh`,
		`Create Alert`,
		`Equipment | Status`,
		`Ready`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("SVG output does not contain %q:\n%s", want, got)
		}
	}
}

func TestRenderUsesRootTitleBeforeDocumentTitle(t *testing.T) {
	got := Render(&sketch.Document{
		Title: "Frontmatter Title",
		Root:  &sketch.Node{Type: "window", Title: "Window Title"},
	})
	if !strings.Contains(got, `Window Title`) {
		t.Fatalf("SVG output does not contain root title:\n%s", got)
	}
	if strings.Contains(got, `Frontmatter Title`) {
		t.Fatalf("SVG output used document title instead of root title:\n%s", got)
	}
}

func TestRenderWithOptions(t *testing.T) {
	got := RenderWithOptions(&sketch.Document{
		Root: &sketch.Node{Type: "window", Label: "Options"},
	}, Options{Width: 320, Height: 240})
	for _, want := range []string{`width="320"`, `height="240"`, `viewBox="0 0 320 240"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("SVG output does not contain %q:\n%s", want, got)
		}
	}
}
