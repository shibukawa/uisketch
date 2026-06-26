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
		`>Equipment</text>`,
		`>Status</text>`,
		`Ready`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("SVG output does not contain %q:\n%s", want, got)
		}
	}
}

func TestDistributeHorizontalLeavesRemainingSpaceWithoutSpacer(t *testing.T) {
	children := []*sketch.Node{
		{Type: "label", Label: "Screen title"},
		{Type: "button", Label: "Primary action"},
	}
	got := distributeHorizontal(800, children, nil, false)
	want := []int{192, 163}
	if len(got) != len(want) {
		t.Fatalf("width count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("widths = %#v, want %#v", got, want)
		}
	}
}

func TestDistributeHorizontalSpacerAbsorbsRemainingSpace(t *testing.T) {
	children := []*sketch.Node{
		{Type: "label", Label: "Screen title"},
		{Type: "spacer"},
		{Type: "button", Label: "Primary action"},
	}
	got := distributeHorizontal(800, children, nil, false)
	want := []int{192, 445, 163}
	if len(got) != len(want) {
		t.Fatalf("width count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("widths = %#v, want %#v", got, want)
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

func TestRenderButtonBadgeAndNoteCallout(t *testing.T) {
	got := Render(&sketch.Document{Root: &sketch.Node{
		Type: "window",
		Children: []*sketch.Node{{
			Type:  "button",
			Label: "Inbox",
			Badge: "1",
			Note:  "Check notification count.",
		}},
	}})
	for _, want := range []string{">1<", ">Check notification<", ">count.<", "note-connector", `fill:#fff2cc`} {
		if !strings.Contains(got, want) {
			t.Fatalf("SVG output does not contain %q:\n%s", want, got)
		}
	}
}

func TestRenderWrapsLongNoteCallout(t *testing.T) {
	got := Render(&sketch.Document{Root: &sketch.Node{
		Type: "window",
		Children: []*sketch.Node{{
			Type:  "label",
			Label: "Status",
			Note:  "This note is intentionally long so it should wrap inside the yellow callout instead of overflowing past the note box.",
		}},
	}})
	for _, want := range []string{`height="136" rx="6" class="note"`, ">This note is<", ">intentionally long<", ">callout instead of<"} {
		if !strings.Contains(got, want) {
			t.Fatalf("SVG output does not contain wrapped note fragment %q:\n%s", want, got)
		}
	}
}

func TestRenderGrowsForLongForms(t *testing.T) {
	var children []*sketch.Node
	for i := 0; i < 12; i++ {
		children = append(children, &sketch.Node{Type: "textarea", Label: "Field"})
	}
	got := Render(&sketch.Document{Root: &sketch.Node{Type: "window", Children: children}})
	if !strings.Contains(got, `height="`) || strings.Contains(got, `height="640"`) {
		t.Fatalf("SVG output did not grow beyond default height:\n%s", got)
	}
}

func TestRenderDoesNotGrowForFlexibleDisplayOnlyRegions(t *testing.T) {
	got := Render(&sketch.Document{Root: &sketch.Node{
		Type: "browser",
		Children: []*sketch.Node{{
			Type: "vstack",
			Children: []*sketch.Node{
				{
					Type: "hstack",
					Children: []*sketch.Node{
						{Type: "label", Label: "UI Sketch Editor"},
						{Type: "spacer"},
						{Type: "button", Label: "Share"},
						{Type: "button", Label: "Export"},
					},
				},
				{
					Type:   "tabs",
					Labels: []sketch.TabLabel{{Text: "Visual Editor", Selected: true}, {Text: "Source"}},
					Children: []*sketch.Node{{
						Type:   "hstack",
						Widths: []sketch.SizeSlot{{Percent: 25}, {Percent: 60}, {Percent: 15}},
						Children: []*sketch.Node{
							{
								Type:  "section",
								Label: "Components",
								Children: []*sketch.Node{{
									Type: "grid",
									Children: []*sketch.Node{
										{Type: "image"}, {Type: "image"}, {Type: "image"}, {Type: "image"},
										{Type: "image"}, {Type: "image"}, {Type: "image"}, {Type: "image"},
									},
								}},
							},
							{
								Type:   "tabs",
								Labels: []sketch.TabLabel{{Text: "Edit", Selected: true}, {Text: "Preview(SVG)"}, {Text: "Preview(ASCII)"}},
								Children: []*sketch.Node{{
									Type:  "image",
									Label: "Preview Image",
								}},
							},
							{
								Type:    "vstack",
								Heights: []sketch.SizeSlot{{Percent: 50}, {Percent: 50}},
								Children: []*sketch.Node{
									{
										Type:  "section",
										Label: "Properties",
										Children: []*sketch.Node{{
											Type:    "table",
											Columns: []string{"Key", "Value"},
										}},
									},
									{
										Type:  "section",
										Label: "Inspector",
										Children: []*sketch.Node{{
											Type: "tree",
										}},
									},
								},
							},
						},
					}},
				},
			},
		}},
	}})
	if !strings.Contains(got, `height="640"`) {
		t.Fatalf("SVG output grew for flexible display-only regions:\n%s", got)
	}
}
