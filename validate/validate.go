package validate

import (
	"fmt"

	"uisketch/concept"
	"uisketch/layout"
)

type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
)

type Finding struct {
	Severity Severity
	Message  string
}

func LoadScreenSketch(screenPath string) (*concept.Screen, *layout.File, []Finding) {
	var findings []Finding
	screen, err := concept.LoadScreen(screenPath)
	if err != nil {
		return nil, nil, []Finding{{Severity: Error, Message: err.Error()}}
	}
	if screen.Source.Location == "" {
		findings = append(findings, Finding{Severity: Error, Message: "screen source.location is required"})
		return screen, nil, findings
	}
	if screen.Source.Type != "uisketch" {
		findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("screen source.type is %q, want uisketch", screen.Source.Type)})
		return screen, nil, findings
	}
	file, err := layout.LoadFile(screen.SourcePath())
	if err != nil {
		findings = append(findings, Finding{Severity: Error, Message: err.Error()})
		return screen, nil, findings
	}
	if file.ScreenID != "" && file.ScreenID != screen.ID {
		findings = append(findings, Finding{Severity: Warning, Message: fmt.Sprintf("uisketch screen.id %q does not match screen concept %q", file.ScreenID, screen.ID)})
	}
	findings = append(findings, ValidateLayout(file.Layout, screen.Actions)...)
	return screen, file, findings
}

func ValidateLayout(root *layout.Node, actions []string) []Finding {
	var findings []Finding
	if root == nil {
		return []Finding{{Severity: Error, Message: "layout root is required"}}
	}
	if !isRoot(root.Type) {
		findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("layout root %q is not a supported root node", root.Type)})
	}
	seen := map[string]bool{}
	actionSet := map[string]bool{}
	for _, action := range actions {
		actionSet[action] = true
	}
	walk(root, func(n *layout.Node) {
		if !isSupported(n.Type) {
			findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("unsupported layout node %q", n.Type)})
		}
		if n.ID != "" {
			if seen[n.ID] {
				findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("duplicate element id %q", n.ID)})
			}
			seen[n.ID] = true
		}
		if n.Action != "" && len(actionSet) > 0 && !actionSet[n.Action] {
			findings = append(findings, Finding{Severity: Warning, Message: fmt.Sprintf("unresolved action reference %q", n.Action)})
		}
		if n.LegacyTo != "" {
			findings = append(findings, Finding{Severity: Warning, Message: fmt.Sprintf("legacy navigation field to is deprecated; use anchor for %q", n.LegacyTo)})
		}
		findings = append(findings, validateNodeShape(n)...)
	})
	return findings
}

func validateNodeShape(n *layout.Node) []Finding {
	var findings []Finding
	if n.Type == "expanded" || n.Type == "fixed-size" {
		findings = append(findings, Finding{Severity: Warning, Message: fmt.Sprintf("legacy sizing wrapper %q should be replaced with hstack.widths or vstack.heights", n.Type)})
	}
	if n.Type == "tabs" {
		selected := 0
		for _, label := range n.Labels {
			if label.Selected {
				selected++
			}
		}
		if len(n.Labels) == 0 {
			findings = append(findings, Finding{Severity: Warning, Message: "tabs should define labels and one active children body"})
		} else if selected != 1 {
			findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("tabs labels must select exactly one label; found %d", selected)})
		}
		if len(n.Children) == 0 {
			findings = append(findings, Finding{Severity: Error, Message: "tabs children must define the active tab body"})
		}
	}
	if len(n.Widths) > 0 {
		if n.Type != "hstack" && n.Type != "tabs" {
			findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("widths is only valid on hstack, not %s", n.Type)})
		}
		if n.Type == "hstack" {
			findings = append(findings, validateStackSlots("widths", len(n.Children), n.Widths)...)
		} else if n.Type == "tabs" {
			findings = append(findings, validateTabLabelSlots("widths", n.Widths)...)
		}
	}
	if len(n.Heights) > 0 {
		if n.Type != "vstack" {
			findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("heights is only valid on vstack, not %s", n.Type)})
		}
		findings = append(findings, validateStackSlots("heights", len(n.Children), n.Heights)...)
	}
	return findings
}

func validateTabLabelSlots(name string, slots []layout.SizeSlot) []Finding {
	var findings []Finding
	total := 0
	stars := 0
	for _, slot := range slots {
		if slot.Star {
			stars++
			continue
		}
		if slot.Percent < 0 {
			findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("%s percentages must be non-negative", name)})
		}
		total += slot.Percent
	}
	if stars > 0 && total > 100 {
		findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("%s numeric percentages total %d; must not exceed 100 when * is present", name, total)})
	}
	if stars == 0 && total > 100 {
		findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("%s numeric percentages total %d; must not exceed 100", name, total)})
	}
	return findings
}

func validateStackSlots(name string, childCount int, slots []layout.SizeSlot) []Finding {
	var findings []Finding
	if len(slots) != childCount {
		findings = append(findings, Finding{Severity: Warning, Message: fmt.Sprintf("%s length %d does not match children length %d; renderer will ignore mismatched sizing hints", name, len(slots), childCount)})
		return findings
	}
	total := 0
	stars := 0
	for _, slot := range slots {
		if slot.Star {
			stars++
			continue
		}
		if slot.Percent < 0 {
			findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("%s percentages must be non-negative", name)})
		}
		total += slot.Percent
	}
	if stars > 0 && total > 100 {
		findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("%s numeric percentages total %d; must not exceed 100 when * is present", name, total)})
	}
	if stars == 0 && total != 100 {
		findings = append(findings, Finding{Severity: Error, Message: fmt.Sprintf("%s numeric percentages total %d; must equal 100 when no * is present", name, total)})
	}
	return findings
}

func HasErrors(findings []Finding) bool {
	for _, finding := range findings {
		if finding.Severity == Error {
			return true
		}
	}
	return false
}

func walk(n *layout.Node, fn func(*layout.Node)) {
	fn(n)
	for _, child := range n.Children {
		walk(child, fn)
	}
	for _, button := range n.Buttons {
		walk(button, fn)
	}
}

func isRoot(t string) bool {
	switch t {
	case "browser", "window", "dialog", "menu", "mobile":
		return true
	default:
		return false
	}
}

func isSupported(t string) bool {
	switch t {
	case "browser", "window", "dialog", "menu", "mobile",
		"vstack", "hstack", "grid", "table-layout", "split-pane", "tabs", "sidebar", "section", "menubar", "spacer",
		"expanded", "fixed-size",
		"button", "icon-button", "floating-action-button", "badge-button", "toggle-button", "link",
		"label", "hint", "note", "review", "image", "custom-component",
		"input", "checkbox", "switch", "slider",
		"table", "list", "tree", "calendar", "badge":
		return true
	default:
		return false
	}
}
