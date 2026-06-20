package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderReadsPositionalUisketchInput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "sample.uisketch.md")
	writeFile(t, input, sampleUisketchSource())

	var out bytes.Buffer
	withIO(strings.NewReader(""), &out, &bytes.Buffer{}, func() {
		if err := run([]string{"render", input}); err != nil {
			t.Fatalf("run render: %v", err)
		}
	})

	if !strings.Contains(out.String(), "<svg") {
		t.Fatalf("render output does not contain svg:\n%s", out.String())
	}
	if !strings.Contains(out.String(), "Sample") {
		t.Fatalf("render output does not contain root title:\n%s", out.String())
	}
}

func TestRenderReadsStdinWhenInputOmitted(t *testing.T) {
	var out bytes.Buffer
	withIO(strings.NewReader(sampleUisketchSource()), &out, &bytes.Buffer{}, func() {
		if err := run([]string{"render", "--format", "ascii"}); err != nil {
			t.Fatalf("run render from stdin: %v", err)
		}
	})

	if !strings.Contains(out.String(), "Sample") {
		t.Fatalf("stdin render output does not contain title:\n%s", out.String())
	}
}

func TestRenderInfersSVGFormatFromOutputExtension(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "sample.uisketch.md")
	output := filepath.Join(dir, "sample.svg")
	writeFile(t, input, sampleUisketchSource())

	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		if err := run([]string{"render", "--output", output, input}); err != nil {
			t.Fatalf("run render: %v", err)
		}
	})

	body, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output svg: %v", err)
	}
	if !strings.Contains(string(body), "<svg") {
		t.Fatalf("output file does not contain svg:\n%s", body)
	}
}

func TestRenderSelectsLayoutByOneOriginPositionalIndex(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "multi.uisketch.md")
	writeFile(t, input, `---
id: screen.multi
type: uisketch
title: Multi
---

# Multi

## Layout: First

`+"```uisketch"+`
browser:
  id: first
  title: First Layout
`+"```"+`

## Layout: Second

`+"```uisketch"+`
window:
  id: second
  title: Second Layout
`+"```"+`
`)

	var out bytes.Buffer
	withIO(strings.NewReader(""), &out, &bytes.Buffer{}, func() {
		if err := run([]string{"render", input, "2"}); err != nil {
			t.Fatalf("run render: %v", err)
		}
	})

	if !strings.Contains(out.String(), "Second Layout") {
		t.Fatalf("render output does not contain selected layout title:\n%s", out.String())
	}
	if strings.Contains(out.String(), "First Layout") {
		t.Fatalf("render output contains first layout title:\n%s", out.String())
	}
}

func TestRenderRejectsScreenConceptInput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "screen.md")
	writeFile(t, input, `---
id: screen.sample
type: screen
title: Sample
---
`)

	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		err := run([]string{"render", input})
		if err == nil {
			t.Fatal("run render returned nil error for screen concept")
		}
		if !strings.Contains(err.Error(), "no uisketch source fences or generated source comments found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRenderIgnoresYAMLFenceUnderLayoutHeading(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "sample.md")
	writeFile(t, input, `# Sample

## Layout

`+"```yaml\n"+`browser:
  id: sample
  title: Sample
`+"```\n")

	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		err := run([]string{"render", input})
		if err == nil {
			t.Fatal("run render returned nil error for bare yaml layout fence")
		}
		if !strings.Contains(err.Error(), "no uisketch source fences or generated source comments found") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRenderRejectsLegacyScreenFlag(t *testing.T) {
	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		err := run([]string{"render", "--screen", "screen.md"})
		if err == nil {
			t.Fatal("run render returned nil error for --screen")
		}
		if !strings.Contains(err.Error(), "flag provided but not defined: -screen") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMarkdownReadsPositionalInputAndWritesAssets(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "doc.md")
	output := filepath.Join(dir, "out.md")
	writeFile(t, input, `# Doc

`+"```uisketch\n"+`browser:
  id: sample
  title: Sample
  children:
    - button:
        label: Save
`+"```\n")

	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		if err := run([]string{"markdown", "--output", output, input}); err != nil {
			t.Fatalf("run markdown: %v", err)
		}
	})

	body, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output markdown: %v", err)
	}
	if !strings.Contains(string(body), "![uisketch:sample](assets/sample.svg)") {
		t.Fatalf("output markdown did not contain generated image:\n%s", body)
	}
	asset, err := os.ReadFile(filepath.Join(dir, "assets", "sample.svg"))
	if err != nil {
		t.Fatalf("read generated asset: %v", err)
	}
	if !strings.Contains(string(asset), "<svg") {
		t.Fatalf("generated asset does not contain svg:\n%s", asset)
	}
}

func TestMarkdownWritesAsciiSourceFromStdinToOutput(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "out.md")
	input := `# Doc

` + "```uisketch:txt\n" + `dialog:
  id: confirm
  title: Confirm
  children:
    - label: Continue?
` + "```\n"

	withIO(strings.NewReader(input), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		if err := run([]string{"markdown", "--output", output}); err != nil {
			t.Fatalf("run markdown from stdin: %v", err)
		}
	})

	body, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output markdown: %v", err)
	}
	if !strings.Contains(string(body), "```text") || !strings.Contains(string(body), "Confirm") {
		t.Fatalf("output markdown did not contain rendered text:\n%s", body)
	}
}

func TestMarkdownRequiresDestinationMode(t *testing.T) {
	input := `# Doc

` + "```uisketch\n" + `browser:
  id: sample
  title: Sample
` + "```\n"

	withIO(strings.NewReader(input), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		err := run([]string{"markdown"})
		if err == nil {
			t.Fatal("run markdown returned nil error without destination")
		}
		if !strings.Contains(err.Error(), "specify exactly one of --output or --overwrite") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMarkdownRejectsOutputAndOverwriteTogether(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "doc.md")
	output := filepath.Join(dir, "out.md")
	writeFile(t, input, "# Doc\n")

	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		err := run([]string{"markdown", "--output", output, "--overwrite", input})
		if err == nil {
			t.Fatal("run markdown returned nil error for both destination modes")
		}
		if !strings.Contains(err.Error(), "specify exactly one of --output or --overwrite") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMarkdownRejectsOverwriteFromStdin(t *testing.T) {
	withIO(strings.NewReader("# Doc\n"), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		err := run([]string{"markdown", "--overwrite"})
		if err == nil {
			t.Fatal("run markdown returned nil error for overwrite stdin")
		}
		if !strings.Contains(err.Error(), "--overwrite requires a filesystem input path") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMarkdownOverwriteWritesAssetsBesideInput(t *testing.T) {
	dir := t.TempDir()
	inputDir := filepath.Join(dir, "input")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(inputDir, "doc.md")
	writeFile(t, input, `# Doc

`+"```uisketch\n"+`browser:
  id: sample
  title: Sample
`+"```\n")

	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		if err := run([]string{"markdown", "--overwrite", input}); err != nil {
			t.Fatalf("run markdown: %v", err)
		}
	})

	body, err := os.ReadFile(input)
	if err != nil {
		t.Fatalf("read overwritten input: %v", err)
	}
	if !strings.Contains(string(body), "![uisketch:sample](assets/sample.svg)") {
		t.Fatalf("overwritten markdown did not contain generated image:\n%s", body)
	}
	if _, err := os.ReadFile(filepath.Join(inputDir, "assets", "sample.svg")); err != nil {
		t.Fatalf("read input-relative asset: %v", err)
	}
}

func TestMarkdownOutputWritesAssetsBesideOutput(t *testing.T) {
	dir := t.TempDir()
	inputDir := filepath.Join(dir, "input")
	outputDir := filepath.Join(dir, "output")
	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(inputDir, "doc.md")
	output := filepath.Join(outputDir, "out.md")
	writeFile(t, input, `# Doc

`+"```uisketch\n"+`browser:
  id: sample
  title: Sample
`+"```\n")

	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		if err := run([]string{"markdown", "--output", output, input}); err != nil {
			t.Fatalf("run markdown: %v", err)
		}
	})

	if _, err := os.ReadFile(filepath.Join(outputDir, "assets", "sample.svg")); err != nil {
		t.Fatalf("read output-relative asset: %v", err)
	}
	if _, err := os.Stat(filepath.Join(inputDir, "assets", "sample.svg")); !os.IsNotExist(err) {
		t.Fatalf("input-relative asset exists or stat failed unexpectedly: %v", err)
	}
}

func TestMarkdownFormatAsciiForcesTextOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "doc.md")
	output := filepath.Join(dir, "out.md")
	writeFile(t, input, `# Doc

`+"```uisketch\n"+`browser:
  id: sample
  title: Sample
`+"```\n")

	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		if err := run([]string{"markdown", "--format", "ascii", "--output", output, input}); err != nil {
			t.Fatalf("run markdown: %v", err)
		}
	})

	body, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output markdown: %v", err)
	}
	if !strings.Contains(string(body), "```text") || !strings.Contains(string(body), `format="txt"`) {
		t.Fatalf("output markdown did not contain forced ascii output:\n%s", body)
	}
	if _, err := os.Stat(filepath.Join(dir, "assets", "sample.svg")); !os.IsNotExist(err) {
		t.Fatalf("unexpected svg asset for ascii output: %v", err)
	}
}

func TestMarkdownRejectsAbsoluteAssetDir(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "doc.md")
	output := filepath.Join(dir, "out.md")
	writeFile(t, input, "# Doc\n")

	withIO(strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, func() {
		err := run([]string{"markdown", "--output", output, "--asset-dir", filepath.Join(dir, "assets"), input})
		if err == nil {
			t.Fatal("run markdown returned nil error for absolute asset dir")
		}
		if !strings.Contains(err.Error(), "must be a relative path") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func withIO(in *strings.Reader, out, errOut *bytes.Buffer, fn func()) {
	oldIn, oldOut, oldErr := stdin, stdout, stderr
	stdin, stdout, stderr = in, out, errOut
	defer func() {
		stdin, stdout, stderr = oldIn, oldOut, oldErr
	}()
	fn()
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func sampleUisketchSource() string {
	return `---
id: screen.sample
type: uisketch
title: Frontmatter Title
---

# Sample

## Layout

` + "```uisketch\n" + `browser:
  id: sample
  title: Sample
  children:
    - button:
        label: Save
` + "```\n"
}
