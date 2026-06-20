---
id: "req.sketch-wireframe-renderer"
type: "requirement"
title: "Sketch Wireframe Renderer"
aliases:
  - "Wireframe Renderer"
  - "SVG Renderer"
  - "ASCII Renderer"
tags:
  - "ui"
  - "renderer"
facts:
  lifecycle.status: "draft"
---

# Sketch Wireframe Renderer

## Summary

The renderer converts `uisketch` source files and embedded `uisketch` Markdown sources into low-fidelity sketch wireframes. The first implementation must support SVG and ASCII-art output.

## Renderer Responsibilities

- Load explicit `uisketch` fenced sources from a Markdown file when rendering through the CLI.
- Ignore ordinary `yaml` fences and Markdown headings such as `## Layout` for render-source discovery.
- Resolve layout nodes into a simple structural drawing model.
- Resolve vocabulary-backed labels when a vocabulary provider is available.
- Apply deterministic default spacing for root content, stack gaps, container insets, `spacer` expansion, and stack child proportions when explicit `hstack.widths` or `vstack.heights` do not control the result.
- Render the drawing model to SVG.
- Render the drawing model to ASCII art.
- Render all `uisketch` fenced code blocks in ordinary Markdown to Markdown image references or `text` blocks as defined in [Markdown Embedding Workflow](markdown-embedding-workflow.md).
- Render a single selected embedded sketch from an ordinary Markdown input when `uisketch render` receives a 1-origin positional index, as defined in [Uisketch CLI](uisketch-cli.md).
- Preserve element `id` and `data` in the parsed model and generated metadata where the output format supports it, but do not render them as visible text by default.
- Apply `highlight` renderer hints without changing layout semantics.
- Emit non-fatal warnings for missing labels, unsupported nodes, or incomplete metadata when rendering can continue.

## SVG Output

SVG output should:

- Use monochrome or near-monochrome strokes by default.
- Use simple boxes, lines, and text.
- Prefer draw.io-like sketch rendering over polished UI components.
- Avoid color semantics unless explicitly needed for validation overlays in future versions.
- Include stable element IDs where useful for testing.

## Sketch Style

SVG rendering should imitate the sketch style familiar from draw.io diagrams while remaining implementation-owned and deterministic.

Required style characteristics:

- Slightly rough or hand-drawn outlines for boxes, containers, and controls.
- Low-fidelity strokes that make the output feel like a discussion sketch rather than a finished UI mockup.
- Monochrome or near-monochrome rendering by default.
- Simple text and simple structural shapes.
- Optional rough hatching or lightweight fill treatment for containers when it improves readability.

Root surface chrome should follow the draw.io wireframe shapes used for browser, window, mobile, and dialog examples:

- `browser` renders an outer window frame, a tab strip containing `browser.title`, top-right circular window controls, a toolbar row with back/forward/reload glyphs, and an address field containing `browser.address` when present.
- `window` renders an outer frame, a title bar containing `window.title`, and top-right circular window controls.
- `mobile` renders a tall rounded smartphone frame with a screen area, a top notch or speaker treatment, optional side buttons when practical, and a bottom home indicator or navigation bar. Its content must be inset inside the screen area.
- `dialog` renders an outer frame, a title bar containing `dialog.title`, and a top-right circular close/control mark.
- Browser, window, mobile, and dialog chrome should be visually distinct from the content area but remain monochrome or near-monochrome and deterministic.
- Dialog `buttons` render as a bottom action row with buttons right-aligned below the main content body.

Root titles are read from the root layout node, not from `.uisketch.md` frontmatter. Frontmatter title may be used only as a legacy migration fallback before a file is rewritten.

Tabbed containers must render the selected tab as connected to the content panel. For horizontal `tabs`, the active label should interrupt or omit the bottom border under that tab so the tab visually opens into the content area. Inactive labels should keep a bottom stroke and corner treatment so they appear behind or recessed relative to the active tab. SVG output should use the same structural treatment with monochrome sketch strokes: all labels are visible, the selected label is emphasized, and only the selected tab's `children` body is drawn inside the shared content panel.

Image placeholders should not render real image content. The SVG renderer should draw an `image` component as a darker filled rectangle with diagonal lines connecting opposite corners. The ASCII renderer should draw a boxed placeholder with crossing diagonal marks when possible, or a labeled box when diagonal marks would reduce readability.

`custom-component` should use the same fallback rendering as `image` unless a project-local renderer mapping exists.

Highlighted elements should stand out while remaining sketch-like. SVG may use a thicker rough outline, secondary callout stroke, or light emphasis fill. ASCII may use a stronger border or marker. Highlighting should not require color and must be deterministic.

The renderer must not depend on draw.io as a runtime or file-format dependency. draw.io is a visual reference for the default rendering style, not an input format or required output format.

The sketch effect should be deterministic for the same input, seed, and renderer options so generated SVG can be tested and reviewed in diffs. If randomized stroke perturbation is used, the renderer must expose a stable seed option.

## ASCII Output

ASCII output should:

- Render screen areas, controls, and tables as full ASCII-art UI layouts using box-drawing characters where available.
- Use layout structure such as `vstack`, `hstack`, `hstack.widths`, and `vstack.heights` to position controls instead of flattening every component into a text list.
- Fit within a configurable width.
- Be deterministic for tests and documentation diffs.
- Prefer readable labels over visual precision.
- Be covered by golden acceptance tests defined in [Renderer Acceptance Tests](renderer-acceptance-tests.md).

Renderers should resolve `hstack.widths` and `vstack.heights` before laying out direct children. Numeric entries reserve that percentage of the available stack axis. `*` entries divide the remaining size equally after numeric percentages. SVG and ASCII output should use the resolved proportions for child regions while still enforcing minimum readable sizes when a label or control would otherwise become unreadable.

For `grid`, renderers should use `grid.columns` when present to determine the number of columns. The default is `2` columns when omitted. The row count is derived from the child count and selected column count, rounding up. Children are placed in row-major order.

When an `hstack` has no usable `widths` declaration and contains one or more `spacer` children, the renderer must allocate content-based widths to ordinary visible children first, then split all remaining horizontal space evenly among the `spacer` children. `spacer` itself must not draw visible text, borders, fills, or placeholder marks. This lets a label remain left-aligned while trailing buttons are pushed to the right:

```yaml
hstack:
  children:
    - label: Title
    - spacer:
    - button:
        label: Cancel
    - button:
        label: Save
```

If an `hstack.widths` declaration is present and valid for the children, it takes precedence over automatic `spacer` expansion. In that case, a `spacer` receives the explicit slot allocated by `widths`.

Example:

```text
+------------------------------------------------+
| Equipment List                                  |
+------------------------------------------------+
| [Create Alert] [Refresh]                        |
+------------------------------------------------+
| Equipment              | Status                 |
| Pump A                 | Normal                 |
| Pump B                 | Warning                |
+------------------------------------------------+
| Ready                                           |
+------------------------------------------------+
```

Tabbed ASCII output should show the active tab label and connect it to the content panel by removing the separator directly below the active tab. Inactive tabs should retain a lower border, using box-drawing corners such as `─┘` and `└─` where width permits:

```text
┌─────────┐┌────────────┐┌─────────┐
│  Home   ││ *Settings* ││  About  │
─┘         └┴────────────┴┘         └──────────┐
│                                               │
│  [ General Settings ]                         │
│                                               │
└───────────────────────────────────────────────┘
```

The exact tab widths may vary with configured output width and label length, but the renderer must keep three invariants: every label is visible, the selected label is identifiable, and the content panel contains only the active tab body.

## Internal Rendering Pipeline

The first implementation should use this pipeline:

```text
Concept metadata
  -> optional screen reference
  -> Markdown source file
  -> explicit uisketch source fence
  -> Parsed layout tree
  -> Normalized sketch model
  -> SVG renderer
  -> ASCII renderer
```

The normalized sketch model should keep renderer-independent structure so future targets such as HTML, React, Markdown documentation, or PPTX can reuse the same semantics.

The CLI entry point renders from the `uisketch` source file itself, not from a `screen` concept wrapper. Screen concepts may reference `uisketch` sources for product modeling, but they are not required inputs for command-line rendering.

The Markdown embedding workflow adds two document-level pipelines:

```text
Plain Markdown source
  -> all uisketch or uisketch:svg fences
  -> Parsed layout trees
  -> SVG renderer
  -> Markdown image references plus uisketch:source comments
```

```text
Plain Markdown source
  -> all uisketch:txt/text/ascii fences
  -> Parsed layout trees
  -> ASCII renderer
  -> text fences plus uisketch:source comments
```

Generated Markdown can be rebuilt by reading the adjacent `uisketch:source` HTML comments, re-rendering the SVG or ASCII output, and overwriting the preceding generated Markdown block.

For CLI `render` with ordinary Markdown input, the renderer must select exactly one embedded source. When the document has multiple renderable embedded sources, the CLI supplies a 1-origin index and the renderer uses the Nth source in document order.

## Browser Editor Reuse

The browser visual editor should reuse the Go renderer through a Wasm adapter so preview behavior matches CLI rendering for the same YAML input.

The renderer API should support a stable preview mode that accepts the parsed layout tree and returns SVG plus validation findings. Browser-specific interaction state such as selection, hover targets, drag ghosts, zoom, and pan should stay outside the renderer model.

## Related Documents

- [UI Sketch Library Overview](ui-sketch-library-overview.md)
- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Layout DSL](ui-layout-dsl.md)
- [Uisketch CLI](uisketch-cli.md)
- [Markdown Embedding Workflow](markdown-embedding-workflow.md)
- [UI Validation Rules](ui-validation-rules.md)
- [Renderer Acceptance Tests](renderer-acceptance-tests.md)
- [Web Visual Editor](web-visual-editor.md)

## Native-Language Summary

Renderer は `.uisketch.md` 内の UI DSL YAML ブロックや Markdown 内の `uisketch` fence から低忠実度のワイヤーフレームを生成する。CLI の render は screen concept ではなく `.uisketch.md` を直接入力にする。通常 Markdown を render 入力にする場合、埋め込み図が 1 つならそれを描画し、複数ある場合は 1-origin index で指定された N 個目の図を描画する。Markdown build/rebuild では文書内のすべての埋め込み図を処理する。SVG は draw.io の sketch 風を参考にしたラフな線で描画するが、draw.io 自体には依存しない。browser はタブ、アドレスバー、戻る/進む/更新、右上ボタンを描き、window はタイトルバーと右上ボタン、mobile はスマートフォン枠、ノッチ、ホームインジケータを描き、dialog はタイトルバーと右上ボタンを描く。描画タイトルは frontmatter ではなく root layout node の title を使う。ASCII は単なる行リストではなく、罫線文字を使った UI アートとして画面領域やボタン配置を表現し、`vstack`、`hstack`、`spacer`、`grid.columns`、`hstack.widths`、`vstack.heights` などの構造を位置決めに反映する。hstack 内の spacer は見えない要素として残り幅を消費し、複数ある場合は按分する。grid は columns で列数を決め、省略時は 2 列で描画する。最初の出力先は SVG と ASCII で、将来の HTML、React、Markdown、PPTX 出力に備えて中間の sketch model を持つ。
