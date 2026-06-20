---
id: "req.vscode-extension"
type: "requirement"
title: "VSCode Extension"
aliases:
  - "VSCode Preview Extension"
  - "Uisketch VSCode Extension"
  - "Markdown Live Preview"
tags:
  - "ui"
  - "vscode"
  - "markdown"
  - "preview"
  - "schema"
facts:
  lifecycle.status: "draft"
---

# VSCode Extension

## Summary

The VSCode extension lets authors preview `uisketch` diagrams inside Markdown and `.uisketch.md` files without running a build that rewrites files. It should also make embedded YAML easier to edit by applying the UI Layout DSL schema to supported source regions, including fenced `uisketch` blocks and generated `uisketch:source` HTML comments.

The extension is an authoring and review surface for [Markdown Embedding Workflow](markdown-embedding-workflow.md), [UI Sketch File Format](ui-sketch-file-format.md), and [Sketch Wireframe Renderer](sketch-wireframe-renderer.md). It must not introduce a second persisted diagram format.

## Product Intent

Primary user workflows:

- Open a Markdown file that contains `uisketch` or `uisketch:svg` fenced blocks and see rendered SVG previews without changing the Markdown file.
- Open generated Markdown that contains an image plus a `uisketch:source` comment and preview the source from the comment without rebuilding the image asset.
- Open a `.uisketch.md` file and preview explicit `uisketch` source fences as SVG.
- Edit embedded UI Layout DSL YAML with completions, hover help, diagnostics, and schema validation.
- Run an explicit bake command when the author wants to replace source fences with generated Markdown image references and SVG assets.
- Run a rebuild command when the author wants to refresh existing baked image references from adjacent `uisketch:source` comments.

The default preview path should be non-destructive. Files are changed only by explicit save edits from the user or by explicit bake/rebuild commands.

## Preview Surfaces

Required preview surfaces:

| Surface | Behavior |
| --- | --- |
| Markdown preview contribution | Replace or augment `uisketch` source blocks with rendered SVG in VSCode's Markdown preview. |
| Editor decorations | Show an inline preview affordance near recognized `uisketch` sources when practical. |
| Side preview panel | Render the selected `uisketch` source fence or generated source comment in a dedicated webview. |
| Diagnostics panel | Show parse, schema, and renderer findings with file ranges when source locations are available. |

The Markdown preview contribution should not require the file to be baked first. It should render directly from the current editor buffer so unsaved edits can be previewed.

When live preview uses a source comment next to an existing image, the comment source is authoritative over the current SVG asset, matching the rebuild rule in [Markdown Embedding Workflow](markdown-embedding-workflow.md).

## Recognized Source Regions

The extension must recognize these source regions:

| Region | Source of truth |
| --- | --- |
| `uisketch` and `uisketch:svg` fences in ordinary Markdown | Fence body |
| `uisketch:txt`, `uisketch:text`, and `uisketch:ascii` fences | Fence body, rendered as text when the preview target supports it |
| `uisketch:source` HTML comments after generated image references | Fenced source inside the comment |
| `uisketch:source` HTML comments after generated `text` fences | Fenced source inside the comment |
| `.uisketch.md` files with `type: uisketch` frontmatter | Explicit `uisketch` source fences and generated `uisketch:source` comments |

The scanner should handle multiple sources in the same Markdown document. Each source region should receive a stable document-local identity derived from an explicit root `id`, comment `id`, or file path plus source ordinal.

The extension must not treat arbitrary YAML fences as renderable UI sketches. A fence is renderable only when it is explicitly marked with a `uisketch` info string or appears inside a generated `uisketch:source` comment.

## Schema-Aware Editing

The extension should provide YAML schema support for UI Layout DSL source regions, even when those regions live inside Markdown.

Required behavior:

- Apply the UI Layout DSL schema to `uisketch` fenced block bodies.
- Apply the same schema to the fenced source inside `uisketch:source` HTML comments.
- Apply the same schema to explicit `uisketch` source fences in `.uisketch.md` files.
- Provide completions for component types, common properties, and enum-like fields.
- Show diagnostics for malformed YAML, unknown component types, invalid child placement, duplicate IDs, and unsupported properties.
- Preserve ordinary YAML comments in source blocks and comments.
- Avoid applying UI Layout DSL diagnostics to unrelated Markdown YAML examples.

The implementation may use VSCode language features, a YAML language server schema association, virtual documents, or a custom language service. The observable behavior is that authors get schema feedback at the embedded source location, not only in a generated temporary file.

Schema validation should be backed by the same rule set used by [UI Validation Rules](ui-validation-rules.md) where practical. Editor-only schema checks may be faster or more local, but they must not contradict the Go validator.

## Live Rendering Pipeline

The preferred live rendering pipeline is:

```text
Current VSCode text buffer
  -> source region scanner
  -> UI Layout DSL parser and validator
  -> SVG renderer
  -> Markdown preview HTML or webview SVG
```

The renderer should be the same Go implementation used by the CLI, exposed to the extension through one of these integration modes:

| Mode | Use |
| --- | --- |
| Bundled CLI process | Preferred first implementation when startup and file watching are acceptable. |
| Go Wasm module in webview or extension host | Preferred later when the browser editor and VSCode extension can share the Wasm adapter. |
| Language server process | Useful when schema, diagnostics, preview, and bake commands need a long-lived service. |

The extension may cache parsed results and SVG output by document version and source region identity. Caches must be invalidated when the editor buffer changes, renderer options change, or related schema files change.

## Bake And Rebuild Commands

The extension should expose explicit commands for destructive or file-writing workflows.

Required commands:

| Command | Behavior |
| --- | --- |
| `uisketch.previewCurrentSource` | Open or refresh a non-destructive preview for the selected source region. |
| `uisketch.bakeMarkdownFile` | Replace recognized source fences with generated Markdown output and adjacent `uisketch:source` comments. |
| `uisketch.rebuildBakedMarkdownFile` | Re-render existing generated output from adjacent `uisketch:source` comments. |
| `uisketch.bakeCurrentSource` | Bake only the selected source region when it can be rewritten safely. |
| `uisketch.exportCurrentSvg` | Write the selected source region's SVG to a chosen file without rewriting Markdown. |
| `uisketch.validateCurrentFile` | Run parser, schema, and semantic validation for the current file. |

Bake behavior must follow [Markdown Embedding Workflow](markdown-embedding-workflow.md):

- `uisketch` and `uisketch:svg` fences become Markdown image references plus adjacent `uisketch:source` comments.
- `uisketch:txt`, `uisketch:text`, and `uisketch:ascii` fences become rendered `text` fences plus adjacent `uisketch:source` comments.
- The comment source remains the source of truth for rebuild.
- Asset paths must stay inside the configured output directory unless the user explicitly chooses a different safe location.

Bake commands must show a preview or summary of planned rewrites before applying changes when multiple regions or asset files will be modified.

## File Watching And Asset Policy

Live preview should not require generated SVG files on disk. It should prefer in-memory SVG for Markdown preview and webviews.

Baked SVG assets should use deterministic filenames based on source identity. When a root `id` exists, the default filename should include that ID. When no root `id` exists, the filename should use a stable document-local generated ID that survives rebuild by reading the existing `uisketch:source` comment metadata.

The extension should detect when a baked SVG asset is stale relative to its adjacent comment source and show a warning or code action. It should not silently rewrite stale assets during ordinary live preview.

## Security

The extension must treat Markdown and YAML source as untrusted input.

Required security rules:

- Generated preview SVG must be sanitized or produced by a trusted renderer that never emits executable script.
- Webviews must use strict content security policy.
- Decoded source comments must never be treated as trusted HTML.
- Bake and export commands must prevent path traversal outside the selected workspace or configured output directory.
- The extension must not execute arbitrary commands from Markdown frontmatter, YAML fields, comments, or workspace files.

## First Runnable Extension Slice

In scope:

- Recognize `uisketch` and `uisketch:svg` fenced blocks in Markdown.
- Render non-destructive SVG previews in Markdown preview or a side webview.
- Recognize explicit `uisketch` source fences in `.uisketch.md` files.
- Run validation diagnostics for malformed YAML and unknown component types.
- Provide schema-backed completions for root components, `children`, `id`, `title`, `label`, `action`, `anchor`, and `hint`.
- Provide `Bake Markdown File` for SVG fences, writing image references, SVG assets, and `uisketch:source` comments.
- Provide `Rebuild Baked Markdown File` for existing image-plus-comment pairs.

Out of scope:

- Full visual drag-and-drop editing inside VSCode.
- Multi-file project graph validation beyond the current file.
- PNG/PDF export.
- Remote workspace persistence or account features.
- Pixel-perfect preview parity with VSCode's built-in Markdown renderer themes.

## Acceptance Criteria

- A Markdown file with an unbaked `uisketch:svg` fence shows an SVG preview in VSCode without rewriting the file.
- Unsaved edits to a recognized source region can refresh the preview.
- A generated image followed by a `uisketch:source` comment previews from the comment source rather than the existing SVG asset.
- A `.uisketch.md` file previews explicit `uisketch` source fences.
- Embedded YAML source regions receive UI Layout DSL schema completions and diagnostics while unrelated YAML examples do not.
- Baking a Markdown file produces the output shape defined by [Markdown Embedding Workflow](markdown-embedding-workflow.md).
- Rebuilding a baked Markdown file updates the SVG asset from the adjacent `uisketch:source` comment.
- Bake and rebuild commands do not write outside the configured output directory without explicit user choice.
- Preview output and baked SVG output are generated by the same renderer behavior for the same YAML input and renderer options.

## Open Decisions

- Decide whether the first implementation invokes a bundled CLI, a language server, or a Wasm renderer.
- Decide how the UI Layout DSL schema is packaged for VSCode and shared with CLI validation.
- Decide the default baked asset folder for Markdown files.
- Decide whether code actions should offer "Bake this source", "Rebuild this output", and "Extract to .uisketch.md".
- Decide whether the extension should support ASCII previews inline or only through bake/export commands in the first slice.

## Related Documents

- [Markdown Embedding Workflow](markdown-embedding-workflow.md)
- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Layout DSL](ui-layout-dsl.md)
- [UI Validation Rules](ui-validation-rules.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [Web Visual Editor](web-visual-editor.md)

## Native-Language Summary

VSCode 拡張は、Markdown や `.uisketch.md` をビルドで書き換えなくても、編集中の buffer から明示的な `uisketch` fence または生成済み `uisketch:source` コメントを読み取って SVG preview できるようにする。`## Layout` や `## レイアウト` 見出し配下の通常 `yaml` fence は render/markdown と同じく自動検出しない。埋め込み YAML には UI Layout DSL schema を効かせ、補完、hover、diagnostics を出す。通常の preview は非破壊で、Markdown image と SVG asset に変換する bake、生成済み comment から再生成する rebuild、SVG export は明示 command として実行する。bake/rebuild の生成形は Markdown Embedding Workflow に従い、生成済み Markdown では隣接する `uisketch:source` コメント内の source を正とする。
