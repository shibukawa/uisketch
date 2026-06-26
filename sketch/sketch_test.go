package sketch

import (
	"testing"

	"uisketch/layout"
)

func TestFromLayoutKeepsSplitterChildrenFlat(t *testing.T) {
	doc := FromLayout("", &layout.Node{
		Type: "window",
		Menu: []string{"File", "Edit"},
		Children: []*layout.Node{{
			Type:        "splitter",
			Orientation: "horizontal",
			Sizes:       []layout.SizeSlot{{Percent: 30}, {Percent: 70}},
			Children: []*layout.Node{
				{Type: "section", Title: "Filters"},
				{Type: "table", ID: "equipment-list"},
			},
		}},
	})
	children := doc.Root.Children[0].Children
	if got := children[0].Type; got != "section" {
		t.Fatalf("first split child type = %q, want section", got)
	}
	if got := children[1].Type; got != "table" {
		t.Fatalf("last split child type = %q, want table", got)
	}
	if got := doc.Root.Children[0].Sizes[0].Percent; got != 30 {
		t.Fatalf("first splitter size = %d, want 30", got)
	}
	if got := doc.Root.Menu; len(got) != 2 || got[0] != "File" || got[1] != "Edit" {
		t.Fatalf("root menu = %#v, want File/Edit", got)
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
