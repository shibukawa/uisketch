package markdown

import (
	"strings"
	"testing"
)

func TestBuildConvertsSVGFenceAndRebuildsFromComment(t *testing.T) {
	input := strings.Join([]string{
		"# Sketch",
		"",
		"```uisketch",
		"browser:",
		"  id: equipment-list",
		"  title: Equipment List",
		"  children:",
		"    - button:",
		"        label: Refresh",
		"```",
	}, "\n")
	result, err := Build(input, Options{AssetDir: "assets"})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(result.Markdown, "![uisketch:equipment-list](assets/equipment-list.svg)") {
		t.Fatalf("markdown did not contain generated image reference:\n%s", result.Markdown)
	}
	if !strings.Contains(result.Markdown, "<!-- uisketch:source id=\"equipment-list\" format=\"svg\"") {
		t.Fatalf("markdown did not contain source comment:\n%s", result.Markdown)
	}
	if got := result.Assets["assets/equipment-list.svg"]; !strings.Contains(got, "<svg") || !strings.Contains(got, "Equipment List") {
		t.Fatalf("svg asset was not generated correctly:\n%s", got)
	}

	rebuilt, err := Build(result.Markdown, Options{AssetDir: "assets"})
	if err != nil {
		t.Fatalf("rebuild returned error: %v", err)
	}
	if rebuilt.Markdown != result.Markdown {
		t.Fatalf("rebuild changed markdown\n--- got ---\n%s\n--- want ---\n%s", rebuilt.Markdown, result.Markdown)
	}
	if _, ok := rebuilt.Assets["assets/equipment-list.svg"]; !ok {
		t.Fatal("rebuild did not regenerate svg asset")
	}
}

func TestBuildConvertsASCIIFenceAndRebuildsFromComment(t *testing.T) {
	input := strings.Join([]string{
		"# Sketch",
		"",
		"```uisketch:ascii",
		"dialog:",
		"  id: confirm-delete",
		"  title: Confirm Delete",
		"  children:",
		"    - label: Delete this item?",
		"  buttons:",
		"    - button:",
		"        label: Cancel",
		"    - button:",
		"        label: OK",
		"```",
	}, "\n")
	result, err := Build(input, Options{})
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	if !strings.Contains(result.Markdown, "```text\n") {
		t.Fatalf("markdown did not contain text fence:\n%s", result.Markdown)
	}
	if strings.Contains(result.Markdown, "```uisketch:ascii") {
		t.Fatalf("source fence was not normalized in comment:\n%s", result.Markdown)
	}
	if !strings.Contains(result.Markdown, "```uisketch:txt") {
		t.Fatalf("source comment did not use normalized txt fence:\n%s", result.Markdown)
	}
	if len(result.Assets) != 0 {
		t.Fatalf("ASCII build generated unexpected assets: %#v", result.Assets)
	}

	edited := strings.Replace(result.Markdown, "Delete this item?", "Delete this item permanently?", 1)
	rebuilt, err := Build(edited, Options{})
	if err != nil {
		t.Fatalf("rebuild returned error: %v", err)
	}
	if strings.Contains(rebuilt.Markdown, "Delete this item permanently?") {
		t.Fatalf("rebuild did not overwrite generated text fence:\n%s", rebuilt.Markdown)
	}
}

func TestBuildFormatOverrideForcesSVGAndASCII(t *testing.T) {
	asciiInput := strings.Join([]string{
		"```uisketch:ascii",
		"label: Hello",
		"```",
	}, "\n")
	svgResult, err := Build(asciiInput, Options{Format: "svg"})
	if err != nil {
		t.Fatalf("Build svg override returned error: %v", err)
	}
	if !strings.Contains(svgResult.Markdown, "![uisketch:") || !strings.Contains(svgResult.Markdown, `format="svg"`) {
		t.Fatalf("svg override did not produce image output:\n%s", svgResult.Markdown)
	}
	if len(svgResult.Assets) != 1 {
		t.Fatalf("svg override assets = %#v, want one asset", svgResult.Assets)
	}

	svgInput := strings.Join([]string{
		"```uisketch:svg",
		"label: Hello",
		"```",
	}, "\n")
	asciiResult, err := Build(svgInput, Options{Format: "ascii"})
	if err != nil {
		t.Fatalf("Build ascii override returned error: %v", err)
	}
	if !strings.Contains(asciiResult.Markdown, "```text") || !strings.Contains(asciiResult.Markdown, `format="txt"`) {
		t.Fatalf("ascii override did not produce text output:\n%s", asciiResult.Markdown)
	}
	if len(asciiResult.Assets) != 0 {
		t.Fatalf("ascii override generated assets: %#v", asciiResult.Assets)
	}
}

func TestBuildRejectsEscapingAssetDir(t *testing.T) {
	_, err := Build("```uisketch\nlabel: Hello\n```", Options{AssetDir: "../assets"})
	if err == nil {
		t.Fatal("Build returned nil error for escaping asset dir")
	}
}

func TestBuildRejectsGeneratedOutputFormatConflict(t *testing.T) {
	input := strings.Join([]string{
		"```text",
		"old output",
		"```",
		"<!-- uisketch:source id=\"x\" format=\"svg\"",
		"```uisketch:svg",
		"label: Hello",
		"```",
		"-->",
	}, "\n")
	_, err := Build(input, Options{})
	if err == nil {
		t.Fatal("Build returned nil error for text output with svg source comment")
	}
}

func TestSelectSourceSelectsUisketchFenceByOneOriginIndex(t *testing.T) {
	input := strings.Join([]string{
		"# Sketch",
		"",
		"```uisketch",
		"browser:",
		"  id: first",
		"  title: First",
		"```",
		"",
		"```uisketch:txt",
		"dialog:",
		"  id: second",
		"  title: Second",
		"```",
	}, "\n")
	source, err := SelectSource(input, 2)
	if err != nil {
		t.Fatalf("SelectSource returned error: %v", err)
	}
	if source.Format != "txt" || !strings.Contains(source.Source, "Second") {
		t.Fatalf("unexpected selected source: %#v", source)
	}
}

func TestSelectSourceSelectsGeneratedImageComment(t *testing.T) {
	input := strings.Join([]string{
		"![uisketch:sample](assets/sample.svg)",
		"<!-- uisketch:source id=\"sample\" format=\"svg\"",
		"```uisketch:svg",
		"browser:",
		"  id: sample",
		"  title: Sample",
		"```",
		"-->",
	}, "\n")
	source, err := SelectSource(input, 1)
	if err != nil {
		t.Fatalf("SelectSource returned error: %v", err)
	}
	if source.ID != "sample" || source.Format != "svg" || !strings.Contains(source.Source, "Sample") {
		t.Fatalf("unexpected selected source: %#v", source)
	}
}

func TestSelectSourceIgnoresYAMLFenceUnderLayoutHeading(t *testing.T) {
	input := strings.Join([]string{
		"# Sketch",
		"",
		"## Layout",
		"",
		"```yaml",
		"browser:",
		"  id: sample",
		"  title: Sample",
		"```",
	}, "\n")
	_, err := SelectSource(input, 1)
	if err == nil {
		t.Fatal("SelectSource returned nil error for bare yaml fence")
	}
	if !strings.Contains(err.Error(), "no uisketch source fences or generated source comments found") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSelectSourceRejectsOutOfRangeIndex(t *testing.T) {
	input := "```uisketch\nlabel: Hello\n```"
	_, err := SelectSource(input, 2)
	if err == nil {
		t.Fatal("SelectSource returned nil error for out-of-range index")
	}
	if !strings.Contains(err.Error(), "source index 2 out of range; found 1 renderable source(s)") {
		t.Fatalf("unexpected error: %v", err)
	}
}
