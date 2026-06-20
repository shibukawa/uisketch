---
id: "req.markdown-embedding-workflow"
type: "requirement"
title: "Markdown Embedding Workflow"
aliases:
  - "Plain Markdown Embedding"
  - "Uisketch Fence Workflow"
  - "Markdown Build Output"
tags:
  - "ui"
  - "markdown"
  - "workflow"
facts:
  lifecycle.status: "draft"
---

# Markdown Embedding Workflow

## Summary

Plain Markdown documents may embed UI sketch sources directly in fenced code blocks. The build workflow renders those fences to reviewable Markdown output while preserving the original `uisketch` YAML source in an adjacent HTML comment so that generated output can be edited or rebuilt without a separate `.uisketch.md` file.

This workflow complements the canonical `.uisketch.md` source format defined in [UI Sketch File Format](ui-sketch-file-format.md). It is intended for design notes, product specs, README files, and other hand-edited Markdown where keeping the sketch close to the surrounding prose matters more than managing a separate source document.

VSCode may preview these same embedded sources without running the build workflow. That non-destructive editor workflow is defined in [VSCode Extension](vscode-extension.md); this document remains the source for the explicit bake and rebuild output shape.

## Source Fence Forms

The Markdown scanner recognizes `uisketch` fences with an optional output target suffix.

| Fence info string | Meaning |
| --- | --- |
| `uisketch` | Render the source as SVG output. |
| `uisketch:svg` | Render the source as SVG output. |
| `uisketch:txt` | Render the source as ASCII text output. |
| `uisketch:text` | Alias for `uisketch:txt`. |
| `uisketch:ascii` | Alias for `uisketch:txt`. |

The content of the fence is a [UI Layout DSL](ui-layout-dsl.md) YAML root node. The fence info string, not the surrounding Markdown heading, declares that the body is renderable UI sketch source. Plain `yaml` fences are ignored by the render and markdown commands even when they appear under headings such as `## Layout` or `## レイアウト`.

A single Markdown document may contain any number of embedded `uisketch` sources. The scanner must process every recognized source fence and every generated `uisketch:source` comment in document order. For example, a Markdown file with two `uisketch:svg` fences must produce two image references, two adjacent source comments, and two SVG asset files. A Markdown file that mixes SVG and ASCII fences must process both kinds in the same pass without requiring separate commands.

The command-line builder may receive an output-format override from `uisketch markdown --format`. When no override is provided, the source fence info string or generated comment metadata controls each sketch independently. When forced to SVG, every recognized sketch is emitted as SVG image output. When forced to ASCII, every recognized sketch is emitted as a rendered `text` fence. The generated `uisketch:source` comment metadata must match the emitted output kind, not the original fence suffix.

Example SVG source fence:

````markdown
```uisketch:svg
browser:
  id: equipment-list
  title: Equipment List
  children:
    - button:
        label: Refresh
```
````

Example ASCII source fence:

````markdown
```uisketch:txt
dialog:
  id: confirm-delete
  title: Confirm Delete
  children:
    - label: Delete this item?
  buttons:
    - button:
        label: Cancel
    - button:
        label: OK
```
````

## SVG Build Output

When building a `uisketch` or `uisketch:svg` fence, the builder must:

1. Parse the fence body as UI Layout DSL YAML.
2. Render the layout to an SVG asset file.
3. Replace the source fence with a Markdown image reference whose alt text is `uisketch:<id>`.
4. Write the source YAML immediately after the image reference in an HTML comment.

The `<id>` value should be the root layout node `id` when present. When the root has no `id`, the builder must generate a stable document-local ID from the file path and fence ordinal, and should preserve that ID in future rebuilds by reading the existing generated comment.

SVG asset filenames must be derived from the resolved sketch ID. The filename stem must be normalized for filesystem and URL stability:

- Convert letters to lowercase.
- Convert whitespace runs to a single underscore.
- Convert path separators, punctuation that is unsafe in filenames, and other non-identifier characters to `_`.
- Collapse repeated `_` characters.
- Trim leading and trailing `_`, `-`, and `.` characters.
- If the result is empty, use the generated document-local ID stem.
- If two sketches in the same output document normalize to the same stem, append a stable 1-origin numeric suffix such as `_2`, `_3`, and so on.

The asset directory is selected by the calling workflow and defaults to `assets`. It must be a relative path, not an absolute path. The cleaned asset directory must remain inside the destination document directory. For `uisketch markdown --overwrite`, the destination document directory is the input file's directory. For `uisketch markdown --output <file>`, the destination document directory is the output file's directory. Markdown image references should use the same relative asset path that is written beside the destination document.

Examples:

| Sketch ID | Asset path |
| --- | --- |
| `equipment-list` | `assets/equipment-list.svg` |
| `Equipment List` | `assets/equipment_list.svg` |
| `Dialog: Confirm Delete` | `assets/dialog_confirm_delete.svg` |

Example generated output:

````markdown
! [uisketch:equipment-list] (assets/equipment-list.svg)
<!-- uisketch:source id="equipment-list" format="svg"
```uisketch:svg
browser:
  id: equipment-list
  title: Equipment List
  children:
    - button:
        label: Refresh
```
-->
````

The comment is the editable source of truth for subsequent rebuilds of generated Markdown output. The image reference is generated output and may be overwritten.

## ASCII Build Output

When building a `uisketch:txt`, `uisketch:text`, or `uisketch:ascii` fence, the builder must:

1. Parse the fence body as UI Layout DSL YAML.
2. Render the layout to deterministic ASCII output.
3. Replace the source fence with a fenced `text` block containing the rendered ASCII output.
4. Write the source YAML immediately after the rendered `text` block in an HTML comment.

Example generated output:

````markdown
```text
+------------------------------+
| Confirm Delete               |
+------------------------------+
| Delete this item?             |
|                 [Cancel] [OK] |
+------------------------------+
```
<!-- uisketch:source id="confirm-delete" format="txt"
```uisketch:txt
dialog:
  id: confirm-delete
  title: Confirm Delete
  children:
    - label: Delete this item?
  buttons:
    - button:
        label: Cancel
    - button:
        label: OK
```
-->
````

The rendered `text` fence must always use the `text` info string, even when the source used `txt`, `ascii`, or `text`.

## Rebuild From Generated SVG Output

A generated Markdown file that already contains a `uisketch:source` HTML comment after a Markdown image reference can be rebuilt without the original source fence.

For SVG generated output, the rebuild workflow must:

1. Find image references whose alt text has the form `uisketch:<id>`.
2. Read the immediately following `uisketch:source` comment.
3. Parse the fenced source inside the comment.
4. Re-render the SVG asset file.
5. Recompute the asset path from the resolved ID and filename normalization policy.
6. Overwrite the image reference target when the recomputed path differs from the existing target.
7. Use stable suffixes for duplicate normalized names in the same document, based on document order.
8. Replace the adjacent source comment with the normalized current source.

The builder must treat the comment source as authoritative over the existing SVG image. Manual edits to the SVG file may be overwritten during rebuild.

## Rebuild From Generated ASCII Output

A generated Markdown file that already contains a `uisketch:source` HTML comment after a `text` fence can be rebuilt without the original source fence.

For ASCII generated output, the rebuild workflow must:

1. Find `text` fences immediately followed by a `uisketch:source` comment.
2. Read the fenced source inside the comment.
3. Parse the source as UI Layout DSL YAML.
4. Re-render the ASCII output.
5. Replace the preceding `text` fence content.
6. Replace the adjacent source comment with the normalized current source.

The builder must treat the comment source as authoritative over the existing `text` fence. Manual edits to rendered ASCII may be overwritten during rebuild.

## Comment Envelope

Generated source comments must use this envelope:

````markdown
<!-- uisketch:source id="<id>" format="<svg|txt>"
```uisketch[:svg|:txt]
<UI Layout DSL YAML>
```
-->
````

Rules:

- The opening marker must start with `<!-- uisketch:source`.
- The comment must contain exactly one fenced `uisketch` source block.
- `format` must be `svg` or `txt`.
- The source fence info string should match the normalized format: `uisketch:svg` for SVG and `uisketch:txt` for ASCII.
- Tools should preserve ordinary YAML comments inside the source fence, but must reject source text that cannot be represented safely inside an HTML comment.

## Validation Rules

Required validation findings:

| Finding | Severity |
| --- | --- |
| `uisketch` fence body is empty | Error |
| `uisketch` fence body is malformed UI Layout DSL YAML | Error |
| Multiple embedded sketches produce the same normalized asset path without a suffix policy | Error |
| Generated image alt text does not match `uisketch:<id>` | Warning |
| Generated output has no adjacent `uisketch:source` comment | Warning |
| Generated comment has malformed metadata | Error |
| Generated comment contains no fenced `uisketch` source block | Error |
| Generated comment source format conflicts with generated output kind | Error |
| Rebuild would overwrite an asset path outside the configured output directory | Error |

## Related Documents

- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Layout DSL](ui-layout-dsl.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [UI Validation Rules](ui-validation-rules.md)
- [Renderer Acceptance Tests](renderer-acceptance-tests.md)
- [VSCode Extension](vscode-extension.md)

## Native-Language Summary

通常の Markdown では、`uisketch` または `uisketch:svg` のコードフェンスを SVG 画像参照に、`uisketch:txt`、`uisketch:text`、`uisketch:ascii` のコードフェンスを `text` コードブロックに変換する。1 つの Markdown ファイルに複数の図がある場合は、文書順にすべて処理する。SVG asset のファイル名は root layout node の `id` を小文字化し、空白や危険な記号を `_` に正規化した stem から決める。名前が衝突する場合は文書順に `_2` などの安定 suffix を付ける。どちらの出力も直後に `<!-- uisketch:source ... -->` コメントとして元の YAML ソースを保持する。生成済み Markdown を再ビルドするときは、画像や `text` ブロックではなく隣接コメント内のソースを正として再解釈し、直前の生成結果を上書きする。
