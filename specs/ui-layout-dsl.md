---
id: "req.ui-layout-dsl"
type: "requirement"
title: "UI Layout DSL"
aliases:
  - "UI DSL"
  - "Layout YAML"
tags:
  - "ui"
  - "dsl"
facts:
  lifecycle.status: "draft"
---

# UI Layout DSL

## Summary

The UI layout DSL describes structural UI layout in a constrained, renderable form. It is not a freeform drawing format like draw.io and does not model visual styling details.

## Design Principles

- Prefer semantic UI parts over arbitrary shapes.
- Preserve enough structure to render SVG and ASCII art.
- Keep authoring readable in code review.
- Make operations, inputs, labels, tables, and screen areas explicit.
- Avoid layout precision that invites visual-design debate.

## Supported Authoring Forms

The library should support YAML first because it is easy to parse and validate in Go. The preferred persisted input is a Markdown-like `uisketch` file defined in [UI Sketch File Format](ui-sketch-file-format.md); explicit `uisketch` fenced source blocks contain the YAML hierarchy described here.

Example:

```yaml
vstack:
  children:
    - hstack:
        children:
          - button:
              action: action.create-alert
              label:
                vocabulary: vocab.create
          - button:
              action: action.refresh-equipment
              label: Refresh
    - table:
        id: equipment-list
        columns:
          - vocabulary: vocab.equipment-name
          - vocabulary: vocab.status
    - label: Ready
```

A compact indentation-based DSL is a future candidate:

```text
browser EquipmentList
address https://example.internal/equipment
hstack
  button Create
  button Refresh
vstack
  table EquipmentTable
label Ready
```

A browser visual editor may generate and edit the YAML representation inside a `uisketch` file, and that YAML remains the canonical persisted layout source for the initial implementation. The editor must not introduce an editor-private persisted scene graph that bypasses validation or renderer behavior.

The same YAML root node may appear inside ordinary Markdown `uisketch` fences, as defined in [Markdown Embedding Workflow](markdown-embedding-workflow.md). In that embedding form, the fence info string supplies the source type and output target, so YAML frontmatter is not required.

## Scalar Label Shorthand

Leaf components whose only intended author-provided value is visible text may be written with a scalar value instead of a property object. The scalar value is interpreted exactly as the component's `label` field.

Canonical mapping form:

```yaml
button:
  label: ラベル
```

Equivalent shorthand:

```yaml
button: ラベル
```

The same rule applies to image placeholders. A placeholder written with a `label` property:

```yaml
image:
  label: ラベル
```

is equivalent to:

```yaml
image: ラベル
```

The shorthand is intended for label-bearing components that do not need nested `children`, `buttons`, `columns`, `labels`, sizing slots, metadata, or action/navigation fields at that location. It should be accepted for ordinary text-like, control-like, and placeholder-like leaf nodes including:

- `button` and `toggle`.
- `label` and `badge`.
- `input`, `textarea`, `combobox`, `checkbox`, `radio`, and `slider`.
- `image` and `custom`.
- `list`, `tree`, and `calendar` when the source only needs a visible placeholder label.

When a component needs any additional property such as `id`, `action`, `anchor`, `badge`, `columns`, `children`, `note`, `highlight`, or `data`, authors must use the mapping form:

```yaml
button:
  id: save
  label: Save
  action: action.save
```

Parsers must treat scalar shorthand as syntactic sugar only; it must not create a second semantic field. Rewriting tools may preserve the shorthand for hand-authored files, but normalizers and visual editors should prefer the canonical mapping form when they need stable round-trip output or when they add any additional property.

Inside `children`, a label-bearing leaf component may also be written as a bare component name when it has no properties at all. For example, `- image` is equivalent to `- image:` and `- tree` is equivalent to `- tree:`. This is useful for placeholder-heavy grids and inspector sections where the component type alone is enough.

Container and root components such as `browser`, `window`, `mobile`, `dialog`, `menu`, `vstack`, `hstack`, `grid`, `splitter`, `tabs`, and `section` should not use scalar label shorthand because their value is normally a structured object or child hierarchy. `table` should continue to use mapping form when it declares `columns`; a scalar `table: Orders` is only a placeholder label and does not define data columns.

## Core Layout Nodes

The canonical component vocabulary is defined in [UI Component Catalog](ui-component-catalog.md). This section lists only the minimum node set needed to understand the DSL shape.

| Node | Purpose |
| --- | --- |
| `browser` | Web browser frame with address bar. |
| `window` | Desktop application window. |
| `mobile` | Smartphone device frame. |
| `dialog` | Modal or transient interaction root. |
| `menu` | Command, navigation, or contextual menu root. |
| `vstack` | Arrange children vertically. |
| `hstack` | Arrange children horizontally. |
| `spacer` | Consume remaining horizontal space inside an `hstack`. |
| `grid` | Arrange children in rows and columns. |
| `splitter` | Arrange two regions with vertical or horizontal orientation and optional `sizes`. |
| `tabs` | Switch content groups with vertical or horizontal tab placement. |
| `custom` | Project-defined component with fallback rendering. |
| `button` | User-triggered operation. |
| `toggle` | Boolean on/off control. |
| `input` | Generic input whose label or hint describes expected content. |
| `textarea` | Multi-line free text input. |
| `combobox` | Single choice from a closed or suggested option list. |
| `checkbox` | Boolean or multi-select option. |
| `radio` | Single choice among visible mutually exclusive options. |
| `image` | Image placeholder drawn without embedding a real picture. |
| `table` | Tabular collection. |
| `list` | Repeated collection. |
| `label` | Text label, preferably vocabulary-backed. |

Root nodes such as `window`, `browser`, `mobile`, `dialog`, and `menu` should use `children`. Their `children` list is interpreted as an implicit `vstack`.

`browser`, `window`, `mobile`, and `dialog` may include a `title` field. This field is visible UI chrome and should be rendered in the browser tab, window title bar, mobile chrome when useful, or dialog title bar. A `.uisketch.md` frontmatter `title` is document metadata and must not be treated as the visible surface title except as a legacy migration fallback.

`window` and `mobile` may include an optional `menu` property for fixed chrome-level command or navigation labels. `menu` is a list of strings, not a child layout node and not a component type. `window.menu` renders as a fixed-height top bar below the window title bar and above content. `mobile.menu` renders as a fixed-height bottom bar below content and inside the device frame. When `menu` is absent or empty, no menu bar is drawn and the content region uses the space normally reserved for that bar.

Examples:

```yaml
window:
  title: UI Sketch Editor
  menu: [File, Edit, View]
  children:
    - section:
        title: Main content
```

```yaml
mobile:
  title: Catalog
  menu: [Home, Search, Settings]
  children:
    - list: Items
```

## Mobile Root

`mobile` represents a smartphone device frame. It is a root surface like `browser`, `window`, `dialog`, and `menu`, not an inner container. Use it when a sketch needs to discuss a mobile screen and the device boundary matters.

Authoring shape:

```yaml
mobile:
  id: mobile-settings
  title: Settings
  children:
    - vstack:
        children:
          - label: Account
          - button:
              label: Sign out
```

Rules:

- `mobile` must be accepted as a root layout node.
- `mobile.children` is interpreted as an implicit `vstack`, like other root surfaces.
- `mobile.title` is optional visible chrome. Renderers may show it in a status/title area when useful, but the mobile frame itself is the primary visual cue.
- Renderers should draw a rounded smartphone outline, a top notch or speaker area, side buttons when practical, and a bottom home indicator or navigation bar.
- The content region must be inset inside the phone screen area so content does not overlap the notch, rounded corners, or home indicator.
- `mobile` should not imply a specific platform such as iOS or Android unless a future platform hint is added.

## Tabs

`tabs` is a static representation of one selected tab state. Because the DSL does not model runtime interaction, a `tabs` node must render only one active tab body.

Authoring shape:

```yaml
tabs:
  labels:
    - Home
    - Settings
    - [About]
  children:
    vstack:
      children:
        - label: About this application
```

Rules:

- `labels` is a non-empty list of tab labels.
- A plain string label is inactive.
- A single-item list label, such as `[About]`, marks that label as selected.
- Exactly one label should be selected. Parsers may choose the first label as a migration fallback when no selected label exists, but validators should report the missing selection.
- `children` is the visible body for the selected tab only. It is a child layout node, not a list of per-tab bodies.
- `children` may be any ordinary layout subtree. Use `vstack` when the active tab contains several vertically arranged elements.
- Renderers must display every tab label and must visually distinguish the selected label from inactive labels.
- Renderers must not render inactive tab bodies because they are not present in this static source shape.

Legacy input that places labels and multiple content groups together in `children` may be accepted only for migration. New or rewritten files should use `labels` plus active `children`.

## Element Identity

Every layout element may include an optional `id`.

Root element IDs are optional. When present, root element IDs must be unique within the project. Root elements should have an `id` when they are referenced from links, flows, editor state, tests, or external systems, but the core DSL does not require one.

Child element IDs are scoped under the nearest root element. Authors may write child IDs as short relative IDs, and tools should resolve them to fully qualified IDs by prefixing the root ID.

Example:

```yaml
window:
  id: equipment
  children:
    - button:
        id: create-alert
        label: Create Alert
```

Resolved ID:

```text
equipment.create-alert
```

Authors may also provide a fully qualified child ID explicitly:

```yaml
button:
  id: equipment.create-alert
  label: Create Alert
```

Within a root element, resolved child IDs must be unique when present. IDs are for linking, system handling, editor selection, generated tests, and traceability. Renderers should preserve IDs in generated SVG metadata or element attributes, but IDs should not be visible in the sketch by default.

## Navigation Targets

Elements that can trigger navigation, especially `button` and other button-like controls, may include an optional `anchor` field. A clickable navigation element should be authored as `button`; `label` is reserved for non-interactive display text.

`anchor` references another UI definition by root ID, or a fully resolved element ID when a transition targets a specific element:

```yaml
button:
  label: Create Alert
  action: action.create-alert
  anchor: alert-create-dialog
```

The `anchor` field supports editor walkthroughs, clickable previews, and flow validation. It is optional and should not be rendered as visible text by default.

The legacy `to` field may be accepted when reading older files, but new or rewritten files should use `anchor`.

## Element Notes

Every layout element may include an optional `note` string. `note` is visible sketch annotation text, not product UI copy. It attaches annotation text to the element that owns the property without changing that element's semantic role.

Authoring shape:

```yaml
button:
  label: ボタン
  note: 出力するテキスト
```

Rules:

- `note` is optional and must be a non-empty string when present.
- `note` belongs to the component instance that declares it; it is not a component type and does not wrap another node.
- `note` must not determine layout footprint, child placement, action behavior, input behavior, or validation identity.
- Tools should expose `note` alongside other shared element properties such as `id`, `data`, `prompt`, and `highlight`.
- Use `label` or another visible component for text that is part of the product UI.

Renderers must show `note` annotations differently from product UI content. Tools may also hide notes for presentation output when a note-free rendering mode is added.

## Highlighting

Every layout element may include an optional `highlight` field. `highlight` is a renderer hint for making an element stand out during discussion, walkthrough, review, or documentation.

Simple highlight:

```yaml
button:
  label: Create Alert
  highlight: true
```

Reasoned highlight:

```yaml
table:
  id: equipment-list
  highlight:
    reason: Primary review target
```

`highlight` must not change layout semantics. Renderers may use a stronger outline, thicker sketch stroke, light emphasis fill, callout marker, or equivalent low-fidelity treatment. `highlight` should remain compatible with monochrome output and should not require color.

## Element Data

Every layout element may include an optional `data` object. `data` stores JSON-like structured metadata for validation, editor behavior, traceability, automation, or downstream tools.

Renderers must not display `data` by default. `data` is part of the source model, not the visible sketch.

Example:

```yaml
button:
  label: Create Alert
  action: action.create-alert
  data:
    requirement: req.alert-create
    permission: alert.create
    analytics:
      event: create_alert_clicked
```

`data` may contain objects, arrays, strings, numbers, booleans, or null values. The core DSL should not prescribe project-specific keys, but validation may enforce local conventions. Renderers should preserve `data` as metadata in outputs that support metadata.

## Element Prompt

Every layout element may include an optional `prompt` string. `prompt` is an authoring note for AI agents that edit, generate, review, or explain the sketch source. It is not visible UI text and must not affect parsing, layout, rendering, navigation, validation semantics, or generated screenshots.

Example:

```yaml
button:
  label: Create Alert
  action: action.create-alert
  prompt: Prefer this as the primary action when generating alert workflows.
```

Multiline YAML strings are allowed for longer guidance:

```yaml
section:
  title: Inspector
  prompt: |
    AI agents should keep this section dense.
    Prefer table-style property editing over prose controls.
```

Tools may preserve `prompt` while round-tripping source, expose it in editor inspectors, or pass it to AI-assisted workflows. Tools are not required to parse the prompt text or interpret it deterministically. Schema validation and key-name validation must recognize `prompt` as a supported element property so hand-authored files do not receive unknown-property diagnostics for it. If project-specific structured instructions are needed, authors should use `data`; `prompt` is intentionally free-form text.

## Sizing And Growth

Layout children should use stack-level proportional slots when authors need coarse sizing. The DSL should avoid wrapper-only sizing nodes because they make common layouts deeply nested and harder to edit.

The default sizing is content-based for ordinary controls and evenly shared for major regions when no explicit stack proportions are provided. Because the default is deterministic, it does not need a wrapper element.

`spacer` is a horizontal layout element for common left/right alignment patterns. Inside an `hstack`, ordinary content such as `label`, `button`, `input`, and other controls keep their content-based sketch width, while `spacer` consumes the remaining horizontal space. When more than one `spacer` appears in the same `hstack`, the remaining space is divided evenly among those spacers.

Example:

```yaml
hstack:
  children:
    - label: Search Results
    - spacer:
    - button:
        label: Export
    - button:
        label: Refresh
```

This places the label at the left edge and the two buttons at the right edge. `spacer` does not render visible text, borders, or controls. It only influences the allocation of horizontal space.

Parsers should accept `spacer` as either an ordinary mapping node or as a scalar shorthand inside `children`:

```yaml
hstack:
  children:
    - label: Search Results
    - spacer
    - button:
        label: Refresh
```

The scalar shorthand is equivalent to `- spacer:`. Rewriting tools should normalize to the mapping form `spacer:` when they output YAML.

`hstack` and `vstack` may declare coarse proportional slots for their direct children:

| Component | Property | Meaning |
| --- | --- | --- |
| `hstack` | `widths` | Width percentages for direct children, from left to right. |
| `vstack` | `heights` | Height percentages for direct children, from top to bottom. |

Each entry in `widths` or `heights` is either a number or `$`. A number is a percentage of the available size along the parent axis. `$` means "all remaining size after numeric percentages"; when more than one `$` appears, the remaining size is divided evenly among those `$` entries. Parsers may accept quoted `"*"` as a legacy spelling, but new or rewritten files should use `$` because bare `*` has YAML alias meaning.

When `hstack.widths` is present, it takes precedence over `spacer` for direct child sizing. In that case, a `spacer` child receives the width assigned by the corresponding `widths` entry. Authors should normally use either `spacer` for content-based left/right alignment or `hstack.widths` for explicit proportional layout, not both in the same `hstack`.

Examples:

```yaml
hstack:
  widths: [20, 30, 50]
  children:
    - section:
        title: Filters
    - list:
        id: equipment-list
    - table:
        id: equipment-detail
```

```yaml
vstack:
  heights: [20, $, $]
  children:
    - label: Header
    - table:
        id: upper-results
    - table:
        id: lower-results
```

In the second example, the first child receives 20 percent of the available height and the two `$` entries each receive 40 percent. `widths` and `heights` are sketch-level proportions, not pixel dimensions, CSS flex rules, or minimum/maximum constraints. The number of entries must match the number of direct `children`. Numeric percentages should total exactly `100` when no `$` is present and must not exceed `100` when `$` is present.

Numeric pixel width, pixel height, min, max, margin, padding, and CSS-like box constraints are intentionally out of scope for the core DSL because the tool is a simple sketch layout tool, not a detailed visual layout engine. Use `hstack.widths` and `vstack.heights` only when the sketch needs proportional space allocation among direct stack children.

## Grid Columns

`grid` may declare a numeric `columns` property to control how many columns are used when placing direct children.

Example:

```yaml
grid:
  columns: 3
  children:
    - button:
        label: One
    - button:
        label: Two
    - button:
        label: Three
    - button:
        label: Four
```

Rules:

- `grid.columns` is optional.
- When omitted, renderers should use the default column count of `2`.
- When present, `grid.columns` must be a positive integer.
- Values less than `1` are invalid.
- Renderers should fill cells in row-major order, left to right then top to bottom.
- The number of rows is derived from `len(children)` and `columns`, rounding up.
- `grid.columns` is a coarse sketch layout hint, not a responsive breakpoint or CSS grid definition.
- `grid.columns` is distinct from `table.columns`, which is a list of visible table column labels.

## Splitter

`splitter` represents a two-pane split layout. It should have exactly two direct `children`.

`splitter.orientation` controls the split direction:

- `horizontal` places panes left and right.
- `vertical` places panes top and bottom.

`splitter.sizes` is optional and assigns proportional space to the two panes in child order. It uses the same value rules as `hstack.widths` and `vstack.heights`: numeric entries are percentages, and `$` means the remaining space after numeric percentages. When `sizes` is omitted, renderers should allocate a smaller leading pane and a larger trailing pane by default.

Example:

```yaml
splitter:
  orientation: horizontal
  sizes: [25, 75]
  children:
    - section:
        title: Filters
    - table:
        id: equipment-list
```

The equivalent proportional structure when authored as a stack would be:

```yaml
hstack:
  widths: [25, 75]
  children:
    - section:
        title: Filters
    - table:
        id: equipment-list
```

## Dialog Buttons

`dialog` may include an optional `buttons` list for the common action-row pattern. `buttons` contains the same child node shapes that could appear in `children`, normally `button`.

Example:

```yaml
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
        action: action.confirm-delete
```

The `buttons` list is a layout shorthand. It is equivalent to a dialog body above a bottom action row with right-aligned buttons:

```yaml
dialog:
  children:
    - vstack:
        heights: [$, 12]
        children:
          - vstack:
              children:
                - label: Delete this item?
          - hstack:
              widths: [$, 15, 15]
              children:
                - label: ""
                - button:
                    label: Cancel
                - button:
                    label: OK
                    action: action.confirm-delete
```

Renderers and editors should present `buttons` as bottom-right dialog actions without requiring authors to write the spacer and action-row structure.

Visual spacing such as root padding, stack gaps, and container insets is renderer/editor policy. The browser editor may apply automatic spacing while previewing and inserting components, but it should not persist per-node `padding`, `margin`, or `gap` fields in the canonical YAML in v0.1.

## Vocabulary References

Labels may reference vocabulary entries instead of embedding literal text:

```yaml
label:
  vocabulary: vocab.contract-holder
```

Renderers must resolve the vocabulary reference to a locale-specific label when vocabulary data is available. If vocabulary data is unavailable, the renderer may show the vocabulary ID and emit a warning.

## Explicitly Unsupported in v0.1

- Pixel coordinates as the primary layout model.
- CSS-like detailed box constraints, margin, padding, and breakpoint control.
- Color, font, detailed spacing, shadow, and icon styling.
- Arbitrary vector paths as source UI elements.
- Freeform Figma import as source of truth.

## Machine-Readable YAML Schema

The project provides a JSON Schema for hand-editing and editor validation at `schemas/uisketch-layout.schema.json`.

The schema validates the YAML layout root node used in:

- Explicit `uisketch` source fences in `.uisketch.md` files.
- `uisketch`, `uisketch:svg`, `uisketch:txt`, `uisketch:text`, and `uisketch:ascii` fences in ordinary Markdown documents.
- Raw YAML layout imports when a tool supports them.

Schema rules:

- The document root must be a single UI component object such as `browser:`, `window:`, `mobile:`, `dialog:`, `menu:`, `vstack:`, or `hstack:`.
- Each node must contain exactly one component type key.
- Component keys must come from the baseline catalog in [UI Component Catalog](ui-component-catalog.md).
- Label-bearing leaf nodes may use scalar label shorthand, such as `button: Save`, which is equivalent to `button: { label: Save }` for parser semantics.
- `children`, `buttons`, and `child` recursively contain layout nodes.
- Inside `children`, bare leaf component values such as `image`, `tree`, and `spacer` are accepted as shorthand for empty component nodes such as `image:`, `tree:`, and `spacer:`.
- `grid.columns` may be a positive integer, while `table.columns` remains a list of visible table column labels.
- `label` and `title` may be literal strings or `{ vocabulary: <id> }` objects.
- `data` remains project-defined structured metadata.
- `prompt` may be a string on any element. It is non-rendered AI-agent guidance; schema validation recognizes the key and its string shape but does not interpret the text.
- `note` may be a non-empty string on any element. It is visible annotation text rendered as a callout or appended note.
- `splitter.orientation` may be `horizontal` or `vertical`; `splitter.sizes` is an optional two-entry proportional array whose entries follow the same rules as stack `widths` and `heights`.
- `window.menu` and `mobile.menu` may be lists of non-empty strings. The `menu` property is only valid on `window` and `mobile` roots and must not be parsed as a component node.
- `to` is accepted for legacy files but should be reported as deprecated by validators and rewritten as `anchor`.

For VS Code YAML validation, authors can associate the schema with standalone layout files by adding a workspace setting similar to:

```json
{
  "yaml.schemas": {
    "./schemas/uisketch-layout.schema.json": [
      "*.uisketch.yaml",
      "*.uisketch.yml"
    ]
  }
}
```

For fenced YAML inside Markdown, tools should extract the fence body and validate that body against the same schema rather than validating the entire Markdown document as YAML.

## Related Documents

- [UI Concept Model](ui-concept-model.md)
- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Component Catalog](ui-component-catalog.md)
- [Markdown Embedding Workflow](markdown-embedding-workflow.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [UI Validation Rules](ui-validation-rules.md)
- [Web Visual Editor](web-visual-editor.md)

## Native-Language Summary

UI DSL は自由図形ではなく、browser、window、mobile、dialog、menu、vstack、hstack、spacer、grid、splitter、tabs、button、toggle、label、input、textarea、combobox、checkbox、radio、image、table などの少数の意味要素で構造を記述する。永続化形式としては `type: uisketch` の frontmatter と明示的な `uisketch` source fence を持つ Markdown 風ファイルを優先し、その fence body にこの DSL の階層構造を書く。`## Layout` や `## レイアウト` は通常の Markdown 見出しであり、通常 `yaml` fence を render/markdown 入力として自動検出しない。frontmatter の title は文書メタデータで、描画される表題は browser/window/mobile/dialog の title 属性に置く。mobile はスマートフォン枠を描く root surface として扱う。root id は任意だが、参照される root には id を付けることを推奨する。root 要素は child ではなく children を持ち、children は暗黙の vstack として扱う。window.menu は window 上部、mobile.menu は mobile 下部に固定表示される文字列リストの chrome 属性であり、children 内の component ではない。button は action と anchor で操作や遷移先 id を持てる。クリック可能な遷移は button で表し、label は非インタラクティブな表示テキストとして扱う。badge 付き操作は button.badge で表す。button の label には短い文字列や絵文字を使える。hstack の中の spacer は残り横幅を消費し、複数ある場合は残り幅を按分する。hstack は widths、vstack は heights で直接の子要素に対する比率を指定でき、`$` は残り領域を表す。splitter は orientation で向きを決め、sizes で 2 pane の比率を指定できる。grid は columns で列数を指定でき、省略時は 2 列とする。dialog は buttons で下部右寄せのボタン列を表現できる。id と data はレンダリング結果のメタデータとして保持する。prompt は任意要素に置ける AI エージェント向けの非表示コメントで、note は任意要素に置ける表示注釈で、どちらも schema はキーと文字列型を検査する。highlight は特定要素を議論・レビュー用に目立たせる renderer hint として扱う。menu は通常メニューとコンテキストメニューを兼ね、ルート要素にもなれる。動的な状態は別ルート定義で表す。入力グループは form/field ではなく vstack/hstack と label/input で表現する。toggle は on/off control、combobox は閉じた選択肢または候補付き入力の baseline component として扱う。画像領域は実画像ではなく image プレースホルダーで表す。ボタン列は toolbar 専用要素ではなく hstack または dialog.buttons で表現する。サイズ指定は wrapper 要素を増やさず、hstack.widths、vstack.heights、splitter.sizes のような owner component property で表現する。通常の内容サイズはデフォルトなので専用要素にしない。CSS のような細かいレイアウト制御は扱わない。コンポーネント語彙は UI Component Catalog で管理する。
