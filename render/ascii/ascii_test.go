package ascii

import (
	"strings"
	"testing"

	"uisketch/sketch"
)

func TestRenderEquipmentList(t *testing.T) {
	doc := &sketch.Document{
		Title: "Equipment List",
		Root: &sketch.Node{
			Type:    "browser",
			Address: "https://example.internal/equipment",
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
		"Equipment List",
		"https://example.internal/equipment",
		"Refresh",
		"Create Alert",
		"Equipment │ Status",
		"Ready",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("ASCII output does not contain %q:\n%s", want, got)
		}
	}
}

func TestRenderButtonBadgeAndNoteMarkers(t *testing.T) {
	got := Render(&sketch.Document{Root: &sketch.Node{
		Type: "window",
		Children: []*sketch.Node{{
			Type:  "button",
			Label: "Inbox",
			Badge: "1",
			Note:  "Check notification count.",
		}},
	}})
	for _, want := range []string{"[1]──", "[1] Check notification count."} {
		if !strings.Contains(got, want) {
			t.Fatalf("ASCII output does not contain %q:\n%s", want, got)
		}
	}
}

func TestRenderGrowsForLongForms(t *testing.T) {
	var children []*sketch.Node
	for i := 0; i < 16; i++ {
		children = append(children, &sketch.Node{Type: "input", Label: "Field " + string(rune('A'+i))})
	}
	got := Render(&sketch.Document{Root: &sketch.Node{Type: "window", Children: children}})
	if lines := strings.Count(got, "\n") + 1; lines <= defaultHeight {
		t.Fatalf("rendered height = %d, want greater than default %d:\n%s", lines, defaultHeight, got)
	}
	if !strings.Contains(got, "Field P") {
		t.Fatalf("ASCII output does not contain final field:\n%s", got)
	}
}

func TestDistributeHorizontalLeavesRemainingSpaceWithoutSpacer(t *testing.T) {
	children := []*sketch.Node{
		{Type: "label", Label: "Screen title"},
		{Type: "button", Label: "Primary action"},
	}
	got := distributeHorizontal(70, children, nil, false)
	want := []int{20, 18}
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
	got := distributeHorizontal(70, children, nil, false)
	want := []int{20, 32, 18}
	if len(got) != len(want) {
		t.Fatalf("width count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("widths = %#v, want %#v", got, want)
		}
	}
}

func TestRenderHStackAlignsLabelWithButtonText(t *testing.T) {
	got := Render(&sketch.Document{Root: &sketch.Node{
		Type: "window",
		Children: []*sketch.Node{{
			Type: "hstack",
			Children: []*sketch.Node{
				{Type: "label", Label: "UI Sketch Editor"},
				{Type: "spacer"},
				{Type: "button", Label: "Share"},
			},
		}},
	}})
	for _, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "UI Sketch Editor") && strings.Contains(line, "Share") {
			return
		}
	}
	t.Fatalf("label and button text are not on the same row:\n%s", got)
}

func TestRenderTabsShowsOnlyActiveBody(t *testing.T) {
	got := Render(&sketch.Document{Root: &sketch.Node{
		Type: "window",
		Children: []*sketch.Node{{
			Type: "tabs",
			Labels: []sketch.TabLabel{
				{Text: "Home"},
				{Text: "Settings", Selected: true},
			},
			Children: []*sketch.Node{{
				Type: "vstack",
				Children: []*sketch.Node{
					{Type: "label", Label: "Visible settings"},
				},
			}},
		}},
	}})
	for _, want := range []string{"Home", "Settings", "Visible settings"} {
		if !strings.Contains(got, want) {
			t.Fatalf("ASCII output does not contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Hidden home") {
		t.Fatalf("ASCII output contains inactive body:\n%s", got)
	}
}

func TestRenderTabsActiveSeamIsOpen(t *testing.T) {
	tests := []struct {
		name     string
		labels   []sketch.TabLabel
		wantLine string
	}{
		{
			name: "first active",
			labels: []sketch.TabLabel{
				{Text: "Visual Editor", Selected: true},
				{Text: "Source"},
			},
			wantLine: "││                 └┴────────┴",
		},
		{
			name: "second active",
			labels: []sketch.TabLabel{
				{Text: "Visual Editor"},
				{Text: "Source", Selected: true},
			},
			wantLine: "│├───────────────┴┘          └",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(&sketch.Document{Root: &sketch.Node{
				Type: "window",
				Children: []*sketch.Node{{
					Type:   "tabs",
					Labels: tt.labels,
					Children: []*sketch.Node{{
						Type: "label", Label: "Active body",
					}},
				}},
			}})
			if !strings.Contains(got, tt.wantLine) {
				t.Fatalf("ASCII tab seam does not contain %q:\n%s", tt.wantLine, got)
			}
		})
	}
}
