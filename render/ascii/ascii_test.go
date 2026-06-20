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
