package sketch

import (
	"testing"

	"uisketch/layout"
)

func TestFromLayoutKeepsSplitPaneChildrenFlat(t *testing.T) {
	doc := FromLayout("", &layout.Node{
		Type: "window",
		Children: []*layout.Node{{
			Type: "split-pane",
			Children: []*layout.Node{
				{Type: "sidebar", Title: "Filters"},
				{Type: "table", ID: "equipment-list"},
			},
		}},
	})
	children := doc.Root.Children[0].Children
	if got := children[0].Type; got != "sidebar" {
		t.Fatalf("first split child type = %q, want sidebar", got)
	}
	if got := children[1].Type; got != "table" {
		t.Fatalf("last split child type = %q, want table", got)
	}
}

func TestFromLayoutNormalizesDialogButtons(t *testing.T) {
	doc := FromLayout("", &layout.Node{
		Type:  "dialog",
		Title: "Confirm",
		Children: []*layout.Node{
			{Type: "label", Label: "Continue?"},
		},
		Buttons: []*layout.Node{
			{Type: "button", Label: "Cancel"},
			{Type: "button", Label: "OK"},
		},
	})
	if doc.Root.Title != "Confirm" {
		t.Fatalf("root title = %q, want Confirm", doc.Root.Title)
	}
	if len(doc.Root.Children) != 1 || doc.Root.Children[0].Type != "vstack" {
		t.Fatalf("dialog children were not normalized: %#v", doc.Root.Children)
	}
	row := doc.Root.Children[0].Children[1]
	if row.Type != "hstack" {
		t.Fatalf("button row type = %q, want hstack", row.Type)
	}
	if len(row.Children) != 3 {
		t.Fatalf("button row child count = %d, want spacer plus 2 buttons", len(row.Children))
	}
	if !row.Widths[0].Star {
		t.Fatalf("button row first width = %#v, want star spacer", row.Widths[0])
	}
}
