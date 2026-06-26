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
- Apply deterministic default spacing for root content, stack gaps, container insets, `spacer` expansion, stack child proportions, and splitter proportions when explicit `hstack.widths`, `vstack.heights`, or `splitter.sizes` do not control the result.
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
- Avoid color semantics except for note callouts and explicitly needed validation overlays in future versions.
- Include stable element IDs where useful for testing.
- Grow the output height when the authored layout needs more vertical space than the default canvas so later form fields and controls are not clipped or omitted.

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
- `window` renders an outer frame, a title bar containing `window.title`, and top-right circular window controls. When `window.menu` is present, it renders a fixed-height top menu bar below the title bar and above the content region.
- `mobile` renders a tall rounded smartphone frame with a screen area, a top notch or speaker treatment, optional side buttons when practical, and a bottom home indicator or navigation bar. Its content must be inset inside the screen area. When `mobile.menu` is present, it renders a fixed-height bottom bar inside the device frame below the content region.
- `dialog` renders an outer frame, a title bar containing `dialog.title`, and a top-right circular close/control mark.
- Browser, window, mobile, and dialog chrome should be visually distinct from the content area but remain monochrome or near-monochrome and deterministic.
- Dialog `buttons` render as a bottom action row with buttons right-aligned below the main content body. The row is dialog chrome/layout policy, not an ordinary `children` entry.

Root titles are read from the root layout node, not from `.uisketch.md` frontmatter. Frontmatter title may be used only as a legacy migration fallback before a file is rewritten.

Tabbed containers must render the selected tab as connected to the content panel. For horizontal `tabs`, the active label should interrupt or omit the bottom border under that tab so the tab visually opens into the content area. Inactive labels should keep a bottom stroke and corner treatment so they appear behind or recessed relative to the active tab. SVG output should use the same structural treatment with monochrome sketch strokes: all labels are visible, the selected label is emphasized, and only the selected tab's `children` body is drawn inside the shared content panel.

Image placeholders should not render real image content. The SVG renderer should draw an `image` component as a darker filled rectangle with diagonal lines connecting opposite corners. The ASCII renderer should draw a compact boxed placeholder with crossing diagonal marks when possible, or a labeled box when diagonal marks would reduce readability. The default ASCII intrinsic height for image/custom placeholders should be about 5 rows so palettes and grids do not become unnecessarily tall.

`custom` should use the same fallback rendering as `image` unless a project-local renderer mapping exists.

Highlighted elements should stand out while remaining sketch-like. SVG may use a thicker rough outline, secondary callout stroke, or light emphasis fill. ASCII may use a stronger border or marker. Highlighting should not require color and must be deterministic.

## Button And Table Rendering

`button` is the only baseline button-like component. Renderers should draw ordinary, compact, emoji-labeled, badge-bearing, and navigation buttons with the same base button treatment.

Button rendering rules:

- The visible button text comes from `button.label`.
- Emoji-only labels are valid. SVG and ASCII renderers must allocate enough intrinsic width and height that the emoji is not clipped or visually collapsed.
- `button.badge` renders as a small count or status mark near the button's top-right corner. The badge is part of the button rendering, not a separate layout child. If a badge shape is already drawn, the text inside it should be ordinary ASCII digits or authored text, not Unicode circled-number glyphs.
- `button.anchor` is metadata for navigation and must not render as visible text by default.

SVG table rendering should include both horizontal row rules and vertical column rules so column boundaries are visually clear. ASCII table rendering may omit vertical detail when that would make the text output noisy, but it must preserve readable column labels and row separation.

Display-region components such as `image`, `custom`, `table`, `list`, `tree`, and `calendar` are flexible preview/data regions. Automatic SVG height estimation must not treat them as large fixed-height controls. They may expand to fill allocated layout space, but when a screen consists mostly of these flexible regions the renderer should keep the default viewport height unless fixed-height controls, form fields, note callouts, or explicit layout sizing require more space. Form controls such as `input`, `textarea`, `combobox`, buttons, tabs, and stacked control groups may still contribute their minimum heights so long forms are not clipped.

## Note Rendering

Any element may carry a `note` string. Renderers should lay out the element exactly as if `note` were absent. The annotation is an overlay or appended explanation, not a separate layout child that consumes stack/grid space.

SVG note rendering:

- Draw the noted component normally at its layout position.
- Draw note text in pale yellow callout boxes collected to the right of the root surface when practical.
- Wrap note text within the callout box width. If the wrapped text needs more vertical space, grow the note box height rather than clipping or overflowing the text.
- Draw a yellow connector line from the noted element bounds to the yellow callout box.
- The callout and connector may extend outside the ordinary root surface bounds when needed, but the SVG viewBox must include the full note callout so it is not clipped.
- SVG note callouts do not require an inline numbered marker when the connector line is sufficient to identify the target element.
- The yellow callout is the one baseline exception to the renderer's mostly monochrome style; it is annotation chrome, not product UI color semantics.

ASCII note rendering:

- Render the noted component normally at its layout position.
- Add a deterministic marker such as `[1]` at the noted element. Labels may prefix the visible text, for example `[1]Label`. Boxed controls may merge the marker into the top border when that improves legibility, but plain text elements must not get extra connector rules.
- Append note text after the rendered UI, below the main ASCII art, one note per line.
- Each appended line must start with the same marker used in the UI, followed by a space and the note text.
- Numbering must follow deterministic source order in ASCII output. The first note is `[1]`, then `[2]`, and so on.

Example appended ASCII notes:

```text
[1] 出力するテキスト
[2] Confirm with legal before implementation.
```

The renderer must not depend on draw.io as a runtime or file-format dependency. draw.io is a visual reference for the default rendering style, not an input format or required output format.

The sketch effect should be deterministic for the same input, seed, and renderer options so generated SVG can be tested and reviewed in diffs. If randomized stroke perturbation is used, the renderer must expose a stable seed option.

## ASCII Output

ASCII output should:

- Render screen areas, controls, and tables as full ASCII-art UI layouts using box-drawing characters where available.
- Use layout structure such as `vstack`, `hstack`, `splitter`, `hstack.widths`, `vstack.heights`, and `splitter.sizes` to position controls instead of flattening every component into a text list.
- Fit within a configurable width.
- Grow vertically when the rendered UI needs more lines than the default output height.
- Be deterministic for tests and documentation diffs.
- Prefer readable labels over visual precision.
- Be covered by golden acceptance tests defined in [Renderer Acceptance Tests](renderer-acceptance-tests.md).

Renderers should resolve `hstack.widths`, `vstack.heights`, and `splitter.sizes` before laying out direct children. Numeric entries reserve that percentage of the available layout axis. `$` entries divide the remaining size equally after numeric percentages. SVG and ASCII output should use the resolved proportions for child regions while still enforcing minimum readable sizes when a label or control would otherwise become unreadable.

For `splitter`, `orientation: horizontal` applies `sizes` as left-to-right widths and `orientation: vertical` applies `sizes` as top-to-bottom heights. When `sizes` is omitted, renderers should use deterministic default proportions, normally a smaller leading pane and a larger trailing pane.

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

Tabbed ASCII output should show the active tab label and connect it to the content panel by removing the separator directly below the active tab. The tab strip and the content panel top border share one seam row. On that seam row, the active tab's interior width must be blank space, not horizontal rule characters, so the active tab visibly opens into the content panel. Inactive tabs should retain a lower border, using box-drawing corners such as `─┘` and `└─` where width permits:

```text
┌─────────┐┌────────────┐┌─────────┐
│  Home   ││ *Settings* ││  About  │
─┘         └┴────────────┴┘         └──────────┐
│                                               │
│  [ General Settings ]                         │
│                                               │
└───────────────────────────────────────────────┘
```

For precise ASCII tab seams, describe each box-drawing cell by the directions it connects to: `t` = up, `l` = left, `b` = down, and `r` = right. Renderers should use equivalent Unicode box-drawing glyphs for these connections, for example `tb` = `│`, `lr` = `─`, `tr` = `└`, `tlr` = `┴`, and `tbr` = `├`. The expected horizontal tab seam rules are:

| Tab state | Left edge on seam row | Interior bottom on seam row | Right edge on seam row |
| --- | --- | --- | --- |
| First tab, active | `tb` | no stroke; fill with spaces | `tr` |
| First tab, inactive | `tbr` | `lr` | `tlr` |
| Non-first tab, active | `tr` | no stroke; fill with spaces | `tr` |
| Non-first tab, inactive | `tlr` | `lr` | `tlr` |

The panel top border starts or continues on the same seam row outside active-tab gaps. Adjacent tabs may share seam cells when their edges touch, but the resulting glyph must preserve the same connection semantics: inactive tab bottoms stay connected to the panel top line, while active tab bottoms remain open. A renderer must not draw a horizontal rule directly below the selected label.

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

Renderer は `.uisketch.md` 内の UI DSL YAML ブロックや Markdown 内の `uisketch` fence から低忠実度のワイヤーフレームを生成する。CLI の render は screen concept ではなく `.uisketch.md` を直接入力にする。通常 Markdown を render 入力にする場合、埋め込み図が 1 つならそれを描画し、複数ある場合は 1-origin index で指定された N 個目の図を描画する。Markdown build/rebuild では文書内のすべての埋め込み図を処理する。SVG は draw.io の sketch 風を参考にしたラフな線で描画するが、draw.io 自体には依存しない。browser はタブ、アドレスバー、戻る/進む/更新、右上ボタンを描き、window はタイトルバーと右上ボタンを描く。window.menu があれば上部固定メニューバーを描く。mobile はスマートフォン枠、ノッチ、ホームインジケータを描き、mobile.menu があれば下部固定バーを描く。dialog はタイトルバーと右上ボタンを描き、buttons があれば下部右寄せの固定ボタン行を描く。描画タイトルは frontmatter ではなく root layout node の title を使う。button は唯一の button-like component として描き、絵文字 label を切り詰めず、button.badge は右上の小さな count/status mark として描く。SVG の table は横罫線だけでなく縦罫線も描き、ASCII は読みやすさを優先して過度にうるさくしない。image、custom、table、list、tree、calendar は可変の表示領域として扱い、割り当て領域には伸びて描画できるが、自動高さ見積もりでは大きな固定高さとして積み上げず、固定高さが必要なフォームや操作要素が多い場合だけ縦に伸ばす。note 属性を持つ要素は通常要素として描画し、SVG では root surface の右側に黄色い注釈ボックスを並べて対象要素との黄色い接続線を出し、note 文はボックス幅で折り返して必要ならボックスの縦幅を伸ばす。ASCII では対象要素に `[1]` などの番号を付け、label では `[1]Label` のようにテキストへ prefix し、レンダリング結果下部に同じ番号付きの注釈文を列挙する。ASCII は単なる行リストではなく、罫線文字を使った UI アートとして画面領域やボタン配置を表現し、`vstack`、`hstack`、`splitter`、`spacer`、`grid.columns`、`hstack.widths`、`vstack.heights`、`splitter.sizes` などの構造を位置決めに反映する。ASCII の水平 tabs はタブ列とコンテンツパネル上端を同じ seam row に描き、アクティブタブの下部は罫線ではなく空白にしてパネルへ開いているように見せ、非アクティブタブの下部はパネル上端へ接続する。hstack 内の spacer は見えない要素として残り横幅を消費し、複数ある場合は按分する。splitter は orientation で向きを決め、sizes があれば 2 pane の比率に反映する。grid は columns で列数を決め、省略時は 2 列で描画する。最初の出力先は SVG と ASCII で、将来の HTML、React、Markdown、PPTX 出力に備えて中間の sketch model を持つ。
