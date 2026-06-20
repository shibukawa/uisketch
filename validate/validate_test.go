package validate

import (
	"os"
	"path/filepath"
	"testing"

	"uisketch/layout"
)

func TestValidateLayoutWarnsOnUnknownAction(t *testing.T) {
	root := &layout.Node{
		Type: "browser",
		Children: []*layout.Node{
			{Type: "button", Action: "action.unknown", Label: "Unknown"},
		},
	}
	findings := ValidateLayout(root, []string{"action.known"})
	if len(findings) != 1 {
		t.Fatalf("len(findings) = %d, want 1", len(findings))
	}
	if findings[0].Severity != Warning {
		t.Fatalf("severity = %s, want warning", findings[0].Severity)
	}
}

func TestValidateLayoutErrorsOnUnsupportedNode(t *testing.T) {
	root := &layout.Node{Type: "browser", Children: []*layout.Node{{Type: "unknown-widget"}}}
	findings := ValidateLayout(root, nil)
	if !HasErrors(findings) {
		t.Fatalf("HasErrors = false, findings = %#v", findings)
	}
}

func TestValidateLayoutAllowsMobileAndSpacer(t *testing.T) {
	root := &layout.Node{
		Type: "mobile",
		Children: []*layout.Node{{
			Type: "hstack",
			Children: []*layout.Node{
				{Type: "label", Label: "Title"},
				{Type: "spacer"},
				{Type: "button", Label: "Save"},
			},
		}},
	}
	findings := ValidateLayout(root, nil)
	if HasErrors(findings) {
		t.Fatalf("HasErrors = true, findings = %#v", findings)
	}
}

func TestValidateLayoutWalksDialogButtonsAndWarnsOnLegacyTo(t *testing.T) {
	root := &layout.Node{
		Type: "dialog",
		Buttons: []*layout.Node{
			{Type: "button", Label: "OK", LegacyTo: "admin-page", Anchor: "admin-page"},
		},
	}
	findings := ValidateLayout(root, nil)
	if len(findings) != 1 {
		t.Fatalf("len(findings) = %d, want 1: %#v", len(findings), findings)
	}
	if findings[0].Severity != Warning {
		t.Fatalf("severity = %s, want warning", findings[0].Severity)
	}
}

func TestValidateLayoutChecksTabsAndStackProportions(t *testing.T) {
	root := &layout.Node{
		Type: "browser",
		Children: []*layout.Node{{
			Type: "hstack",
			Widths: []layout.SizeSlot{
				{Percent: 60},
				{Percent: 60},
				{Star: true},
			},
			Children: []*layout.Node{
				{Type: "label", Label: "A"},
				{Type: "label", Label: "B"},
				{Type: "label", Label: "C"},
			},
		}, {
			Type: "tabs",
			Labels: []layout.TabLabel{
				{Text: "A", Selected: true},
				{Text: "B", Selected: true},
			},
			Children: []*layout.Node{{Type: "label", Label: "Active"}},
		}, {
			Type:     "expanded",
			Children: []*layout.Node{{Type: "label", Label: "Legacy"}},
		}},
	}
	findings := ValidateLayout(root, nil)
	var sawWidths, sawTabs, sawLegacy bool
	for _, finding := range findings {
		if finding.Message == "widths numeric percentages total 120; must not exceed 100 when * is present" {
			sawWidths = true
		}
		if finding.Message == "tabs labels must select exactly one label; found 2" {
			sawTabs = true
		}
		if finding.Message == `legacy sizing wrapper "expanded" should be replaced with hstack.widths or vstack.heights` {
			sawLegacy = true
		}
	}
	if !sawWidths || !sawTabs || !sawLegacy {
		t.Fatalf("findings = %#v, want widths, tabs, and legacy warnings/errors", findings)
	}
}

func TestValidateLayoutWarnsOnStackSlotLengthMismatch(t *testing.T) {
	root := &layout.Node{
		Type: "browser",
		Children: []*layout.Node{{
			Type: "hstack",
			Widths: []layout.SizeSlot{
				{Percent: 25},
				{Percent: 60},
				{Percent: 15},
			},
			Children: []*layout.Node{
				{Type: "label", Label: "A"},
				{Type: "label", Label: "B"},
			},
		}},
	}
	findings := ValidateLayout(root, nil)
	if HasErrors(findings) {
		t.Fatalf("HasErrors = true, findings = %#v", findings)
	}
	if len(findings) != 1 || findings[0].Severity != Warning || findings[0].Message != "widths length 3 does not match children length 2; renderer will ignore mismatched sizing hints" {
		t.Fatalf("findings = %#v, want one mismatch warning", findings)
	}
}

func TestValidateLayoutAllowsTabsWidthsIndependentFromChildren(t *testing.T) {
	root := &layout.Node{
		Type: "browser",
		Children: []*layout.Node{{
			Type: "tabs",
			Labels: []layout.TabLabel{
				{Text: "Source"},
				{Text: "Preview", Selected: true},
				{Text: "Settings"},
			},
			Widths: []layout.SizeSlot{
				{Percent: 30},
				{Percent: 40},
				{Percent: 30},
			},
			Children: []*layout.Node{{
				Type: "hstack",
				Children: []*layout.Node{
					{Type: "label", Label: "Editor"},
					{Type: "label", Label: "Preview"},
				},
			}},
		}},
	}
	findings := ValidateLayout(root, nil)
	if HasErrors(findings) {
		t.Fatalf("HasErrors = true, findings = %#v", findings)
	}
}

func TestLoadScreenSketchErrorsOnMissingSourceLocation(t *testing.T) {
	dir := t.TempDir()
	screenPath := filepath.Join(dir, "screen.md")
	if err := os.WriteFile(screenPath, []byte(`---
id: screen.missing-source
type: screen
title: Missing Source
source:
  type: uisketch
---
`), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, findings := LoadScreenSketch(screenPath)
	if !HasErrors(findings) {
		t.Fatalf("HasErrors = false, findings = %#v", findings)
	}
}
