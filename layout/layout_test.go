package layout

import (
	"strings"
	"testing"
)

func TestParseFileExtractsUisketchLayout(t *testing.T) {
	file, err := ParseFile("sample.uisketch.md", sampleFile())
	if err != nil {
		t.Fatalf("ParseFile returned error: %v", err)
	}
	if file.Type != "uisketch" {
		t.Fatalf("Type = %q, want uisketch", file.Type)
	}
	if file.Layout.Type != "browser" {
		t.Fatalf("Layout.Type = %q, want browser", file.Layout.Type)
	}
	if got := file.Layout.Children[0].Children[0].Children[0].Action; got != "action.refresh-equipment" {
		t.Fatalf("first action = %q", got)
	}
	if got := strings.Join(file.Layout.Children[0].Children[1].Columns, ","); got != "Equipment,Status" {
		t.Fatalf("columns = %q", got)
	}
}

func TestParseFileIgnoresYAMLFenceUnderLayoutHeading(t *testing.T) {
	_, err := ParseFile("sample.uisketch.md", `---
id: screen.heading
type: uisketch
---

# Heading

## Layout: Editor

`+"```yaml"+`
browser:
  id: heading-browser
`+"```"+`
`)
	if err == nil {
		t.Fatal("ParseFile returned nil error for bare yaml fence")
	}
	if !strings.Contains(err.Error(), "no uisketch source fences or generated source comments found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseFileAtSelectsOneOriginLayoutSection(t *testing.T) {
	file, err := ParseFileAt("sample.uisketch.md", `---
id: screen.multi
type: uisketch
---

# Multi

## Layout: List

`+"```uisketch"+`
browser:
  id: first
  title: First
`+"```"+`

## Layout: Editor

`+"```uisketch"+`
window:
  id: second
  title: Second
`+"```"+`
`, 2)
	if err != nil {
		t.Fatalf("ParseFileAt returned error: %v", err)
	}
	if file.Layout.Type != "window" || file.Layout.ID != "second" {
		t.Fatalf("selected layout = (%q, %q), want (window, second)", file.Layout.Type, file.Layout.ID)
	}
}

func TestParseFileAtRejectsOutOfRangeLayoutIndex(t *testing.T) {
	_, err := ParseFileAt("sample.uisketch.md", sampleFile(), 2)
	if err == nil {
		t.Fatal("ParseFileAt returned nil error")
	}
	if !strings.Contains(err.Error(), "source index 2 out of range; found 1 renderable source(s)") {
		t.Fatalf("error %q does not contain out-of-range diagnostic", err)
	}
}

func TestParseLayoutNewSurfaceAndNavigationFields(t *testing.T) {
	node, err := ParseYAML(`dialog:
  id: confirm-delete
  title: Confirm Delete
  children:
    - label: Delete this item?
  buttons:
    - button:
        label: Cancel
        anchor: equipment-list
    - button:
        label: OK
        to: equipment-list.table
`)
	if err != nil {
		t.Fatalf("ParseYAML returned error: %v", err)
	}
	if node.Title != "Confirm Delete" {
		t.Fatalf("Title = %q, want Confirm Delete", node.Title)
	}
	if len(node.Buttons) != 2 {
		t.Fatalf("len(Buttons) = %d, want 2", len(node.Buttons))
	}
	if got := node.Buttons[0].Anchor; got != "equipment-list" {
		t.Fatalf("first button anchor = %q", got)
	}
	if got := node.Buttons[1].Anchor; got != "equipment-list.table" {
		t.Fatalf("legacy to anchor = %q", got)
	}
	if got := node.Buttons[1].LegacyTo; got != "equipment-list.table" {
		t.Fatalf("LegacyTo = %q", got)
	}
}

func TestParseYAMLAcceptsPromptOnAnyElement(t *testing.T) {
	node, err := ParseYAML(`browser:
  id: root
  prompt: |
    AI agents should keep this layout dense.
    Preserve the primary action.
  children:
    - button:
        label: Create Alert
        prompt: Prefer this as the primary action.
`)
	if err != nil {
		t.Fatalf("ParseYAML returned error: %v", err)
	}
	if !strings.Contains(node.Prompt, "keep this layout dense") {
		t.Fatalf("root Prompt = %q", node.Prompt)
	}
	if got := node.Children[0].Prompt; got != "Prefer this as the primary action." {
		t.Fatalf("child Prompt = %q", got)
	}
}

func TestParseYAMLRejectsNonStringPrompt(t *testing.T) {
	_, err := ParseYAML(`button:
  label: Save
  prompt:
    role: primary
`)
	if err == nil {
		t.Fatal("ParseYAML returned nil error")
	}
	if !strings.Contains(err.Error(), `prompt under button("Save") must be a string`) {
		t.Fatalf("error %q does not contain prompt diagnostic", err)
	}
}

func TestParseYAMLRejectsUnsupportedPropertyWithoutSuggestion(t *testing.T) {
	_, err := ParseYAML(`button:
  label: Save
  ai_note: Keep visible
`)
	if err == nil {
		t.Fatal("ParseYAML returned nil error")
	}
	if !strings.Contains(err.Error(), `unknown property "ai_note" under button`) {
		t.Fatalf("error %q does not contain unknown property diagnostic", err)
	}
}

func TestParseYAMLAllowsFlexibleYAMLIndentation(t *testing.T) {
	node, err := ParseYAML(`browser:
    id: flexible
    title: Flexible
    children:
      - button:
            id: save
            label: Save
`)
	if err != nil {
		t.Fatalf("ParseYAML returned error: %v", err)
	}
	if got := node.Children[0].ID; got != "save" {
		t.Fatalf("child ID = %q, want save", got)
	}
	if got := node.Children[0].Label; got != "Save" {
		t.Fatalf("child label = %q, want Save", got)
	}
}

func TestParseYAMLReadsTabsLabelsAndStackProportions(t *testing.T) {
	node, err := ParseYAML(`tabs:
  labels:
    - Home
    - [Settings]
    - About
  children:
    vstack:
      heights: [20, $, $]
      children:
        - label: Header
        - table:
            id: primary
        - table:
            id: secondary
`)
	if err != nil {
		t.Fatalf("ParseYAML returned error: %v", err)
	}
	if got := len(node.Labels); got != 3 {
		t.Fatalf("len(Labels) = %d, want 3", got)
	}
	if !node.Labels[1].Selected || node.Labels[1].Text != "Settings" {
		t.Fatalf("selected label = %#v, want Settings selected", node.Labels[1])
	}
	if got := len(node.Children); got != 1 {
		t.Fatalf("len(Children) = %d, want active body only", got)
	}
	body := node.Children[0]
	if body.Type != "vstack" {
		t.Fatalf("tab body type = %q, want vstack", body.Type)
	}
	if len(body.Heights) != 3 || !body.Heights[1].Star || !body.Heights[2].Star {
		t.Fatalf("body heights = %#v, want 20, $, $", body.Heights)
	}
}

func TestParseYAMLReadsMobileSpacerAndGridColumns(t *testing.T) {
	node, err := ParseYAML(`mobile:
  title: Catalog
  children:
    - hstack:
        children:
          - label: Catalog
          - spacer
          - button:
              label: Filter
    - grid:
        columns: 3
        children:
          - image:
              label: A
          - image:
              label: B
`)
	if err != nil {
		t.Fatalf("ParseYAML returned error: %v", err)
	}
	if node.Type != "mobile" {
		t.Fatalf("root type = %q, want mobile", node.Type)
	}
	row := node.Children[0]
	if got := row.Children[1].Type; got != "spacer" {
		t.Fatalf("hstack child type = %q, want spacer", got)
	}
	grid := node.Children[1]
	if got := grid.GridColumns; got != 3 {
		t.Fatalf("GridColumns = %d, want 3", got)
	}
}

func TestParseYAMLReadsScalarLabelShorthand(t *testing.T) {
	node, err := ParseYAML(`vstack:
  children:
    - button: Save
    - image: Hero
    - tree: Navigation
    - input: Search
    - checkbox: Enabled
    - table: Orders
    - custom-component: Chart
`)
	if err != nil {
		t.Fatalf("ParseYAML returned error: %v", err)
	}
	wants := []struct {
		typ   string
		label string
	}{
		{"button", "Save"},
		{"image", "Hero"},
		{"tree", "Navigation"},
		{"input", "Search"},
		{"checkbox", "Enabled"},
		{"table", "Orders"},
		{"custom-component", "Chart"},
	}
	if len(node.Children) != len(wants) {
		t.Fatalf("len(Children) = %d, want %d", len(node.Children), len(wants))
	}
	for i, want := range wants {
		child := node.Children[i]
		if child.Type != want.typ || child.Label != want.label {
			t.Fatalf("child %d = (%q, %q), want (%q, %q)", i, child.Type, child.Label, want.typ, want.label)
		}
	}
}

func TestParseYAMLReadsImageScalarShorthandLikeLabelProperty(t *testing.T) {
	withLabel, err := ParseYAML(`image:
  label: ラベル
`)
	if err != nil {
		t.Fatalf("ParseYAML with label property returned error: %v", err)
	}
	withShorthand, err := ParseYAML(`image: ラベル
`)
	if err != nil {
		t.Fatalf("ParseYAML with scalar shorthand returned error: %v", err)
	}
	if withLabel.Type != withShorthand.Type || withLabel.Label != withShorthand.Label {
		t.Fatalf("image shorthand = (%q, %q), want (%q, %q)", withShorthand.Type, withShorthand.Label, withLabel.Type, withLabel.Label)
	}
}

func TestParseYAMLReadsBareLeafChildrenInGridAndSection(t *testing.T) {
	node, err := ParseYAML(`section:
  label: Components
  children:
    - grid:
        children:
          - image
          - image
    - section:
        label: Inspector
        children:
          - tree
`)
	if err != nil {
		t.Fatalf("ParseYAML returned error: %v", err)
	}
	grid := node.Children[0]
	if grid.Type != "grid" {
		t.Fatalf("first child type = %q, want grid", grid.Type)
	}
	if len(grid.Children) != 2 || grid.Children[0].Type != "image" || grid.Children[1].Type != "image" {
		t.Fatalf("grid children = %#v, want two bare image nodes", grid.Children)
	}
	inspector := node.Children[1]
	if inspector.Type != "section" {
		t.Fatalf("second child type = %q, want section", inspector.Type)
	}
	if len(inspector.Children) != 1 || inspector.Children[0].Type != "tree" {
		t.Fatalf("section children = %#v, want one bare tree node", inspector.Children)
	}
}

func TestParseYAMLRejectsScalarShorthandForContainerNodes(t *testing.T) {
	_, err := ParseYAML(`vstack: Main`)
	if err == nil {
		t.Fatal("ParseYAML returned nil error")
	}
	if !strings.Contains(err.Error(), "node vstack under document root must use mapping properties") {
		t.Fatalf("error %q does not contain container scalar diagnostic", err)
	}
}

func TestParseYAMLRejectsInvalidGridColumns(t *testing.T) {
	_, err := ParseYAML(`grid:
  columns: 0
`)
	if err == nil {
		t.Fatal("ParseYAML returned nil error")
	}
	if !strings.Contains(err.Error(), "columns under grid must be a positive integer") {
		t.Fatalf("error %q does not contain grid columns diagnostic", err)
	}
}

func TestParseFileReportsActualLineParentAndTypoSuggestion(t *testing.T) {
	_, err := ParseFile("sample.uisketch.md", `---
id: screen.example
type: uisketch
---

# Example

## Layout

`+"```uisketch"+`
browser:
  id: root
  children:
    - button:
        id: save
        lable: Save
`+"```"+`
`)
	if err == nil {
		t.Fatal("ParseFile returned nil error")
	}
	message := err.Error()
	for _, want := range []string{
		"line 15:",
		`unknown property "lable" under button#save`,
		`did you mean "label"?`,
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error %q does not contain %q", message, want)
		}
	}
}

func TestParseFileReportsParentWhenChildNodeKeyIsMissing(t *testing.T) {
	_, err := ParseFile("sample.uisketch.md", `---
id: screen.example
type: uisketch
---

# Example

## Layout

`+"```uisketch"+`
browser:
  id: root
  children:
    - id: orphan
      label: Orphan
`+"```"+`
`)
	if err == nil {
		t.Fatal("ParseFile returned nil error")
	}
	message := err.Error()
	for _, want := range []string{
		"line 14:",
		"expected exactly one layout node key under children of browser#root",
		"found id, label",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error %q does not contain %q", message, want)
		}
	}
}

func TestParseFileOffsetsYAMLSyntaxErrorLine(t *testing.T) {
	_, err := ParseFile("sample.uisketch.md", `---
id: screen.example
type: uisketch
---

# Example

## Layout

`+"```uisketch"+`
browser:
  children: [
`+"```"+`
`)
	if err == nil {
		t.Fatal("ParseFile returned nil error")
	}
	if !strings.Contains(err.Error(), "[11:") {
		t.Fatalf("error %q does not contain actual file line 11", err)
	}
}

func TestParseFileCollectsMultipleLayoutErrors(t *testing.T) {
	_, err := ParseFile("sample.uisketch.md", `---
id: screen.example
type: uisketch
---

# Example

## Layout

`+"```uisketch"+`
browser:
  id: root
  children:
    - button:
        id: save
        lable: Save
    - id: orphan
      label: Orphan
    - table:
        columns:
          - Name
          - badge:
              label: Bad
  buttons: not-a-list
`+"```"+`
`)
	if err == nil {
		t.Fatal("ParseFile returned nil error")
	}
	message := err.Error()
	for _, want := range []string{
		"line 15:",
		`unknown property "lable" under button#save`,
		"line 17:",
		"expected exactly one layout node key under children of browser#root",
		"line 21:",
		"columns under table must contain scalar values",
		"line 23:",
		"buttons under browser#root must be a list",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error %q does not contain %q", message, want)
		}
	}
	if got := strings.Count(message, "\n") + 1; got < 4 {
		t.Fatalf("error %q contains %d diagnostics, want at least 4", message, got)
	}
}

func TestParseFileRejectsMissingLayoutBlock(t *testing.T) {
	_, err := ParseFile("sample.uisketch.md", `---
id: screen.example
type: uisketch
title: Example
---

# Example
`)
	if err == nil {
		t.Fatal("ParseFile returned nil error for missing layout")
	}
}

func sampleFile() string {
	return `---
id: screen.equipment-list
type: uisketch
title: Equipment List
screen:
  id: screen.equipment-list
---

# Equipment List

## Layout

` + "```uisketch" + `
browser:
  id: equipment-list
  title: Equipment List
  address: https://example.internal/equipment
  children:
    - vstack:
        children:
          - hstack:
              children:
                - button:
                    id: refresh
                    action: action.refresh-equipment
                    label: Refresh
                - button:
                    id: create-alert
                    action: action.create-alert
                    label: Create Alert
          - table:
              id: equipment-table
              columns:
                - Equipment
                - Status
          - label: Ready
` + "```" + `
`
}
