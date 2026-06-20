---
id: "req.uisketch-cli"
type: "requirement"
title: "Uisketch CLI"
aliases:
  - "Command Line Interface"
  - "CLI"
tags:
  - "cli"
  - "renderer"
  - "markdown"
facts:
  lifecycle.status: "draft"
---

# Uisketch CLI

## Summary

The `uisketch` command line tool operates on explicit `uisketch` Markdown fences and generated `uisketch:source` comments. It must not require a `screen` concept input for rendering, and it must not infer renderable source from ordinary Markdown headings.

The CLI is a local conversion and validation tool, not the canonical manager for product-level screen concepts, actions, or flows.

## Commands

The first CLI surface should provide these commands:

| Command | Purpose |
| --- | --- |
| `uisketch render [options] [input] [index]` | Render one selected `uisketch` fence or generated `uisketch:source` comment from a Markdown file to SVG or ASCII. |
| `uisketch markdown [options] [input]` | Build or rebuild ordinary Markdown documents that contain embedded `uisketch` fences or generated `uisketch:source` comments. |

Both commands treat `input` as a positional argument. When `input` is omitted or is `-`, the command reads from standard input.

## Input Arguments

`input` is a required command argument in the command contract, with standard input as the default source when no filesystem path is supplied.

Rules:

- Do not expose the primary input as `--screen`, `--input`, or another named option.
- `uisketch render` accepts Markdown files that contain embedded `uisketch` sources. It renders one selected embedded sketch rather than rebuilding the whole Markdown document.
- `uisketch render` and `uisketch markdown` must not scan `## Layout`, `## Layout: ...`, `## レイアウト`, or similar headings for bare YAML blocks. A YAML block is renderable only when its fence info string is `uisketch`, `uisketch:svg`, `uisketch:txt`, `uisketch:text`, or `uisketch:ascii`, or when it appears inside a generated `uisketch:source` comment.
- `uisketch render` does not accept `type: screen` concept files.
- `uisketch markdown` accepts ordinary Markdown content and processes only explicit `uisketch` source fences or generated `uisketch:source` comments.
- A filesystem input path gives parsers a source location for diagnostics and stable generated IDs.
- Standard input should still be usable, but diagnostics and generated IDs may use a conventional virtual path such as `<stdin>`.

## Render Command

Usage:

```text
uisketch render [--format svg|ascii] [--output <file>] [input] [index]
```

Behavior:

1. Read source content from `input`, or from standard input when `input` is omitted or `-`.
2. Scan embedded `uisketch` source fences and generated `uisketch:source` comments as defined in [Markdown Embedding Workflow](markdown-embedding-workflow.md).
3. Ignore ordinary YAML fences, including YAML fences under `## Layout` or `## レイアウト` headings.
4. When `index` is omitted, use `1`.
5. `index` is 1-origin and selects the Nth renderable embedded sketch in document order.
6. Reject `index` values less than 1, non-integer values, and values greater than the number of renderable embedded sketches.
7. Validate the selected layout with the same rules used by editor and renderer previews.
8. Render the selected layout to the selected output format.
9. Write to `--output` when supplied, otherwise write to standard output.

Examples:

```text
uisketch render --format svg spec.md 1
uisketch render --format ascii spec.md 2
```

If `spec.md` contains two embedded sketches and no `index` is supplied, `uisketch render spec.md` renders the first sketch. Authors can pass `2` as the second positional argument to render the second sketch.

The default `--format` should be `svg`, because SVG is the primary reviewable artifact for visual sketches. ASCII remains available for terminal previews, text documentation, and golden tests.

When `--format` is omitted and `--output` has the `.svg` extension, the command must behave as if `--format svg` was specified. An explicit `--format` value always takes precedence over the output file extension.

The rendered surface title comes from the selected root layout node, not from a screen concept and not from frontmatter `title`.

Generated or rebuilt Markdown output is not written by `render`. The command only selects one embedded sketch and renders that sketch to SVG or ASCII. Use `uisketch markdown` to process every embedded sketch and write Markdown plus SVG assets.

## Markdown Command

Usage:

```text
uisketch markdown [--output <file> | --overwrite] [--format svg|ascii|source] [--asset-dir assets] [input]
```

Behavior:

1. Require exactly one destination mode: either `--output <file>` or `--overwrite`.
2. Reject a command that provides neither destination mode, because the Markdown workflow rewrites Markdown and may write SVG assets.
3. Reject a command that provides both `--output` and `--overwrite`.
4. Read ordinary Markdown from `input`, or from standard input when `input` is omitted or `-`.
5. Reject `--overwrite` when the input is omitted, is `-`, or otherwise has no filesystem path to overwrite.
6. Process explicit `uisketch`, `uisketch:svg`, `uisketch:txt`, `uisketch:text`, and `uisketch:ascii` fences as defined in [Markdown Embedding Workflow](markdown-embedding-workflow.md).
7. Rebuild generated Markdown output from adjacent `uisketch:source` comments when present.
8. Write rebuilt Markdown to `--output` when supplied.
9. When `--overwrite` is supplied, write rebuilt Markdown back to the input file.
10. Write generated SVG assets using `--asset-dir` relative to the destination document directory.

Destination rules:

- `--output <file>` writes the rebuilt Markdown to that path. SVG assets are written under `--asset-dir` relative to the output file's directory.
- `--overwrite` writes the rebuilt Markdown back to the input file. SVG assets are written under `--asset-dir` relative to the input file's directory.
- `--asset-dir` defaults to `assets`.
- `--asset-dir` must be a relative path. Absolute paths are invalid.
- `--asset-dir` must not escape the destination document directory after cleaning, so values such as `../assets` are invalid unless a future explicit unsafe override is added.

Output format selection:

- By default, each source fence's info string chooses the output kind: `uisketch` and `uisketch:svg` produce SVG image output; `uisketch:txt`, `uisketch:text`, and `uisketch:ascii` produce ASCII `text` output.
- `--format svg` forces all recognized source fences and rebuilt source comments to SVG output.
- `--format ascii` forces all recognized source fences and rebuilt source comments to ASCII `text` output.
- `--format source` preserves each source block's declared or comment metadata format. This is the default behavior.
- For forced SVG output, the builder writes SVG assets and replaces each sketch with a Markdown image reference plus a `uisketch:source` comment whose format metadata is `svg`.
- For forced ASCII output, the builder writes no SVG assets and replaces each sketch with a fenced `text` block plus a `uisketch:source` comment whose format metadata is `txt`.

## Diagnostics

CLI diagnostics should be concise and source-oriented:

- Validation findings go to standard error.
- Rendered or converted document output goes to standard output only for commands whose contract is stdout-oriented. `markdown` writes to `--output` or `--overwrite` and should not emit the rebuilt Markdown to stdout.
- Invalid input type should identify the detected frontmatter type and expected type `uisketch`.
- Missing path with standard input unavailable should produce usage text that shows the positional input form.

## Non-Goals

- The CLI does not require or render from `screen` concept files.
- The CLI does not resolve product-level action catalogs, flow graphs, or requirement metadata as part of `render`.
- The CLI does not persist editor-only state such as selection, hover, zoom, or pan.

## Related Documents

- [UI Sketch File Format](ui-sketch-file-format.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [Markdown Embedding Workflow](markdown-embedding-workflow.md)
- [UI Validation Rules](ui-validation-rules.md)

## Native-Language Summary

`uisketch` CLI は Markdown 内の明示的な `uisketch` fence と生成済み `uisketch:source` comment だけを描画・変換対象にする。`## Layout` や `## レイアウト` のような見出し配下にある通常の `yaml` fence は、render と markdown のどちらでも自動検出しない。`render` は `screen` concept を入力にせず、`uisketch render [input] [index]` の positional input から N 個目の埋め込み図を読む。`index` は 1-origin で省略時は 1 とする。`input` を省略した場合または `-` の場合は標準入力を読む。主入力は `--screen` や `--input` のようなオプションにしない。`--format` 省略時に `--output` が `.svg` 拡張子なら `--format svg` と同じ扱いにし、明示された `--format` は拡張子より優先する。`markdown` は通常 Markdown 内の明示的な `uisketch` fence と `uisketch:source` comment だけを処理し、`--output` または `--overwrite` のどちらかを必須にする。`--overwrite` は入力ファイルを上書きし、SVG assets は入力ファイル基準の相対 `--asset-dir` に書く。`--output` は出力先ファイルへ書き、SVG assets は出力先ファイル基準の相対 `--asset-dir` に書く。`--asset-dir` は相対パスだけを許可し、絶対パスや宛先ディレクトリ外へ出るパスはエラーにする。`markdown --format svg|ascii|source` で全 sketch の出力種別を強制または source 指定どおりにできる。
