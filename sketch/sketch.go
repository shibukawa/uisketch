package sketch

import "uisketch/layout"

type Document struct {
	Title string
	Root  *Node
}

type Node struct {
	Type        string
	ID          string
	Title       string
	Label       string
	Action      string
	Anchor      string
	Address     string
	Hint        string
	Note        string
	Name        string
	Purpose     string
	Badge       string
	Orientation string
	Menu        []string
	Columns     []string
	Options     []string
	GridColumns int
	Labels      []TabLabel
	Widths      []SizeSlot
	Heights     []SizeSlot
	Sizes       []SizeSlot
	Children    []*Node
}

type TabLabel struct {
	Text     string
	Selected bool
}

type SizeSlot struct {
	Percent int
	Star    bool
}

func FromLayout(title string, root *layout.Node) *Document {
	return &Document{Title: title, Root: convert(root)}
}

func convert(n *layout.Node) *Node {
	if n == nil {
		return nil
	}
	if (n.Type == "expanded" || n.Type == "fixed-size") && len(n.Children) > 0 {
		return convert(n.Children[0])
	}
	out := &Node{
		Type:        n.Type,
		ID:          n.ID,
		Title:       n.Title,
		Label:       n.Label,
		Action:      n.Action,
		Anchor:      n.Anchor,
		Address:     n.Address,
		Hint:        n.Hint,
		Note:        n.Note,
		Name:        n.Name,
		Purpose:     n.Purpose,
		Badge:       n.Badge,
		Orientation: n.Orientation,
		Menu:        append([]string(nil), n.Menu...),
		Columns:     append([]string(nil), n.Columns...),
		Options:     append([]string(nil), n.Options...),
		GridColumns: n.GridColumns,
		Labels:      convertLabels(n.Labels),
		Widths:      convertSizeSlots(n.Widths),
		Heights:     convertSizeSlots(n.Heights),
		Sizes:       convertSizeSlots(n.Sizes),
	}
	children := n.Children
	if n.Type == "dialog" && len(n.Buttons) > 0 {
		children = normalizeDialogChildren(n.Children, n.Buttons)
	}
	for _, child := range children {
		out.Children = append(out.Children, convert(child))
	}
	return out
}

func convertLabels(labels []layout.TabLabel) []TabLabel {
	out := make([]TabLabel, len(labels))
	for i, label := range labels {
		out[i] = TabLabel{Text: label.Text, Selected: label.Selected}
	}
	return out
}

func convertSizeSlots(slots []layout.SizeSlot) []SizeSlot {
	out := make([]SizeSlot, len(slots))
	for i, slot := range slots {
		out[i] = SizeSlot{Percent: slot.Percent, Star: slot.Star}
	}
	return out
}

func normalizeDialogChildren(children, buttons []*layout.Node) []*layout.Node {
	body := &layout.Node{
		Type:     "vstack",
		Children: children,
	}
	actionRow := &layout.Node{
		Type:   "hstack",
		Widths: append([]layout.SizeSlot{{Star: true}}, fixedButtonSlots(len(buttons))...),
		Children: append([]*layout.Node{
			{Type: "label", Label: " "},
		}, buttons...),
	}
	return []*layout.Node{{
		Type:     "vstack",
		Heights:  []layout.SizeSlot{{Star: true}, {Percent: 12}},
		Children: []*layout.Node{body, actionRow},
	}}
}

func fixedButtonSlots(count int) []layout.SizeSlot {
	out := make([]layout.SizeSlot, count)
	for i := range out {
		out[i] = layout.SizeSlot{Percent: 15}
	}
	return out
}
