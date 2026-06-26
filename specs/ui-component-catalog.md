---
id: "req.ui-component-catalog"
type: "requirement"
title: "UI Component Catalog"
aliases:
  - "UI Components"
  - "Component Vocabulary"
tags:
  - "ui"
  - "dsl"
  - "component"
facts:
  lifecycle.status: "draft"
---

# UI Component Catalog

## Summary

The UI DSL should provide a broad catalog of semantic UI components that covers common desktop, mobile, and web interfaces. Components describe purpose and interaction, not final visual design.

## Catalog Principles

- Components are semantic building blocks for sketches and validation.
- Component names should stay platform-neutral when the same concept appears across desktop, mobile, and web.
- A `screen` is a concept type, not a layout component.
- Platform-specific variants should be represented as hints, not separate component types.
- Direction should be represented with `orientation: vertical|horizontal` on components that support it.
- Components should support label references through vocabulary entries.
- Components should expose actions, inputs, data bindings, permissions, and state where relevant.
- Components must not require color, typography, spacing, or pixel-perfect styling.

## Component Metadata

Each component definition should be able to declare:

- Stable component type name.
- Short purpose.
- Supported platforms: `desktop`, `mobile`, `web`.
- Allowed child components.
- Supported stack sizing properties.
- Optional `id` for linking, editor selection, automation, and tests.
- Optional `highlight` renderer hint for review emphasis.
- Optional `anchor` for components that navigate to another UI definition or element.
- Key properties.
- Optional non-rendered `data` metadata.
- Related actions or events.
- Typical validation rules.
- Sketch rendering expectations for SVG and ASCII.

Example:

```yaml
component:
  type: button
  platforms: [desktop, mobile, web]
  action: action.create-alert
  label:
    vocabulary: vocab.create-alert
  variant: primary
```

The `variant` value is semantic priority, not final visual styling.

## Shared Element Identity

Every component instance may include an optional `id`. Root component IDs are optional, but must be project-unique when present. Child component IDs are resolved inside the root namespace, for example `window-id.child-id`.

`id` is not rendered as visible UI by default. It exists so links, walkthroughs, external systems, editor selection, validation, and generated tests can address elements reliably. Renderers should preserve `id` as output metadata where the format supports it.

Example:

```yaml
browser:
  id: equipment-list
  address: https://example.internal/equipment
  children:
    - button:
        id: refresh
        label: Refresh
```

The button resolves to `equipment-list.refresh`.

## Shared Element Data

Every component instance may include an optional `data` object. This object is JSON-like structured metadata for tooling and validation. It is not rendered as visible UI.

Example:

```yaml
input:
  label: Start Date
  hint: date
  data:
    vocabulary: vocab.start-date
    requirement: req.schedule-filter
    sourceField: equipment.startedAt
```

The catalog defines that `data` exists and is non-visual. Projects may define their own schema for keys inside `data`. Renderers should preserve `data` as output metadata where the format supports it.

## Shared Element Prompt

Every component instance may include an optional `prompt` string. `prompt` is free-form guidance for AI agents that generate, edit, review, or explain a sketch. It is not visible UI, not a user-facing hint, and not structured metadata.

Example:

```yaml
input:
  label: Search
  prompt: AI agents should keep this field prominent in filter-heavy layouts.
```

Renderers must not display `prompt` by default. Parsers and validators do not need to interpret the prompt text; they only need to recognize `prompt` as a supported key and validate that its value is a string. Structured, machine-readable metadata should continue to use `data`.

## Shared Element Note

Every component instance may include an optional `note` string. `note` is visible sketch annotation text used for design discussion, implementation reminders, or review comments. It is not product UI copy and must not change the component's semantic role, layout footprint, action, input behavior, or validation identity.

Example:

```yaml
button:
  label: Submit
  note: Confirm with legal before implementation.
```

Renderers show `note` differently from product UI content. SVG uses pale yellow annotation callout boxes arranged to the right of the root surface when practical, with yellow connector lines to the noted elements. ASCII marks the noted element with a numbered marker such as `[1]`; labels may render it as a text prefix such as `[1]Label`, while boxed controls may merge the marker into the border when useful. ASCII also appends the note text below the rendered UI with the same marker. Tools may hide notes for presentation output when a note-free rendering mode is added.

## Shared Highlighting

Every component instance may include an optional `highlight` field. `highlight` marks the element as important for review, walkthrough, or documentation without changing its semantic role.

Example:

```yaml
button:
  label: Create Alert
  highlight:
    reason: Discuss permission rule
```

Renderers should make highlighted elements stand out using low-fidelity sketch treatment, not production styling.

## Layout And Structure Components

| Component | Platforms | Purpose |
| --- | --- | --- |
| `window` | desktop | Desktop application window. |
| `browser` | web | Browser frame with address bar around web content. |
| `mobile` | mobile | Smartphone device frame around mobile content. |
| `vstack` | desktop, mobile, web | Children arranged vertically. |
| `hstack` | desktop, mobile, web | Children arranged horizontally. |
| `spacer` | desktop, mobile, web | Invisible horizontal filler that consumes remaining `hstack` space. |
| `grid` | desktop, mobile, web | Children arranged in rows and columns. |
| `splitter` | desktop, web | Two-pane layout with `orientation` and optional `sizes`. |
| `section` | desktop, mobile, web | Named content section. |
| `tabs` | desktop, mobile, web | Switchable content groups with `orientation`. |
| `dialog` | desktop, mobile, web | Root surface for a modal or transient interaction state. |
| `menu` | desktop, mobile, web | Command or navigation choices. May also be used as a root surface for contextual menus. |
| `custom` | desktop, mobile, web | Project-defined component with explicit name and fallback rendering. |

## Arrangement Components

Arrangement components describe how child components are placed. They are intentionally coarse and should not expose pixel-level layout.

| Component | Direction / Shape | Purpose |
| --- | --- | --- |
| `vstack` | vertical | Place children from top to bottom. |
| `hstack` | horizontal | Place children from left to right. |
| `spacer` | horizontal filler | Consume remaining horizontal space inside an `hstack`. |
| `grid` | rows and columns | Place children in a regular matrix. |
| `splitter` | vertical or horizontal split | Show two major regions separated by a conceptual divider. |

`layout` may remain as an internal parser or AST term, but author-facing DSL examples should prefer explicit arrangement components such as `vstack`, `hstack`, `grid`, and `splitter`.

Arrangement components may use coarse stack-level sizing properties from [UI Layout DSL](ui-layout-dsl.md): `hstack.widths` and `vstack.heights`. Content-based sizing is the default for ordinary controls and does not need a wrapper.

`hstack` may include a `widths` array and `vstack` may include a `heights` array to assign proportional space to direct children. These arrays contain percentages or `$`; `$` means the remaining space after numeric percentages, split evenly when multiple `$` entries are present. The arrays are coarse sketch proportions, not CSS sizing or pixel layout controls.

`spacer` is the preferred shorthand when an `hstack` needs content at both edges and no explicit percentages are needed. It is invisible and has no label. Ordinary siblings keep content-based widths, and all `spacer` siblings split the remaining horizontal space. This is useful for toolbar and action-row sketches:

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

Parsers should also accept scalar `- spacer` in child lists as shorthand for `- spacer:`, but generated YAML should prefer the explicit mapping form.

`grid` may include a numeric `columns` property. The default is `2` columns when omitted. `grid.columns` controls placement of child regions and must not be confused with `table.columns`, which names visible table fields.

`splitter` may include a `sizes` array to assign proportional space to its two direct children. `sizes` uses the same entries as stack sizing: numbers are percentages and `$` means the remaining space. When omitted, renderers should use deterministic default proportions, normally a smaller leading pane and a larger trailing pane.

Example:

```yaml
grid:
  columns: 3
  children:
    - image:
        label: Product A
    - image:
        label: Product B
    - image:
        label: Product C
```

Example:

```yaml
hstack:
  widths: [20, $, $]
  children:
    - section:
        title: Navigation
    - table:
        id: primary-list
    - vstack:
        children:
          - label: Details
          - input:
              label: Name
```

The catalog should avoid CSS-like pixel width, pixel height, margin, padding, and breakpoint controls.

Generic framed containers should normally use `vstack`, `hstack`, or `section` instead of a vague `panel` component. A renderer may draw a boundary around any container when helpful, but that boundary is a rendering choice rather than a separate semantic component.

`tabs` represents one visible tab state, not a dynamic widget that renders every possible tab body at once. It must separate tab labels from the active tab content:

| Property | Requirement |
| --- | --- |
| `labels` | Required non-empty list of tab labels. Each item is either a string label or a single-item list containing the selected label. Exactly one item should be selected. |
| `children` | Required active tab body. This is ordinary child layout content, normally a `vstack`, and it is the only tab body rendered. |
| `orientation` | Optional `horizontal` or `vertical`; default is `horizontal`. |

Example:

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
        - button:
            label: Close
```

The selected `labels` item identifies which tab label is visibly active. The `children` property contains the currently visible content for that active tab only. Authors should not put tab label nodes or multiple tab bodies inside `children`; doing so makes a static sketch show unrelated content that would not be visible in a single UI state.

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

When `splitter.sizes` is omitted, renderers use the default split ratio for the selected orientation.

Example:

```yaml
vstack:
  children:
    - hstack:
        widths: [25, 75]
        children:
          - button:
              action: action.refresh
              label: Refresh
    - hstack:
        widths: [25, 75]
        children:
          - section:
              title: Filters
          - table:
              id: equipment-list
```

## Custom Components

Projects may define components that are not in the baseline catalog by using `custom`.

Custom components must include:

- `name`, using a stable project-local component name.
- `purpose`, explaining the role in business or workflow terms.
- Optional `platforms`.
- Optional `children`.
- Optional `actions`, `inputs`, or `state` references.
- Optional renderer hint for fallback sketch output.

Example:

```yaml
custom:
  name: equipment-health-map
  purpose: Shows equipment status by location.
  platforms: [desktop, web]
  fallback:
    label: Equipment Health Map
    shape: box
```

Renderers should draw unsupported custom components as labeled boxes and emit a warning unless a project-local renderer mapping exists.

By default, `custom` should render like an `image` placeholder: a filled rectangle with diagonal crossing lines and a label. Project-local renderer mappings may override this fallback.

## Root Surface Components

Dynamic UI states should usually be separate root definitions rather than hidden states inside one picture. For example, an open dialog, an open menu, or an error dialog should be rendered from its own root layout.

Allowed root surface components:

| Root | Purpose |
| --- | --- |
| `window` | Desktop application frame. |
| `browser` | Web browser frame with address bar. |
| `mobile` | Smartphone device frame. |
| `dialog` | Modal or transient interaction state. |
| `menu` | Command, navigation, or contextual choices. |

Root surface components should use `children`, not `child`. A root surface's `children` list is interpreted as an implicit `vstack`, so authors do not need to wrap root contents in an explicit `vstack` unless they need a nested vertical group.

`browser`, `window`, `mobile`, and `dialog` support a visible `title` property. The title should be rendered in the browser tab, window title bar, mobile chrome when useful, or dialog title bar. This property is distinct from `.uisketch.md` frontmatter `title`, which is document metadata.

`window` and `mobile` support an optional `menu` property as root surface chrome. It is a list of visible string labels, not a child component list. `window.menu` is rendered as a fixed-height application menu bar at the top of the content area, below the title bar. `mobile.menu` is rendered as a fixed-height bottom navigation or command bar inside the device frame. A missing or empty `menu` property renders no bar.

`mobile` should render as a rounded smartphone frame with a screen content region. It may show a top notch or speaker and a bottom home indicator. The frame is platform-neutral and should not imply a specific iOS or Android version.

`dialog` also supports an optional `buttons` property for common bottom-right action rows. `buttons` is a list of ordinary child nodes, normally `button`, and is normalized as a bottom action row after the main `children` body.

Transient messages, empty/error states, and permission warnings should be represented with text, ordinary layout, or a separate `dialog` root when they need their own sketch.

Example browser frame:

```yaml
browser:
  title: Equipment List
  address: https://example.internal/equipment
  children:
    - hstack:
        children:
          - button:
              action: action.refresh
              label: Refresh
    - table:
        id: equipment-list
```

## Navigation Behavior

Navigation is behavior on an interactive component, not a separate baseline component. Use `button` with an optional `anchor` when clicking or pressing the element navigates to another view, dialog, root UI definition, or resolved element ID. Use `label` only for non-interactive display text.

```yaml
button:
  label: Admin
  action: action.open-admin
  anchor: admin-page
```

## Action Components

| Component | Platforms | Purpose |
| --- | --- | --- |
| `button` | desktop, mobile, web | Executes an action or navigation. May include `badge`. |
| `toggle` | desktop, mobile, web | Boolean on/off control. |

`toggle` is the baseline component for on/off state.

`button` is also the baseline for compact icon-like, floating, and badge-bearing actions. Authors may use a short text label, an emoji label, or an ordinary word label. When a button needs a small count or status indicator, set `badge` on the `button` instead of using a separate component type. Renderers and the visual editor should place the badge on the button's top-right border rather than appending it to the label text. If a badge shape is already drawn, the text inside it should be ordinary ASCII digits or authored text, not Unicode circled-number glyphs.

Example badge button:

```yaml
button:
  label: Notifications
  badge: 3
```

Example compact emoji button:

```yaml
button:
  label: 🔍
  action: action.search
```

Stepper and segmented-control patterns should be represented with ordinary layout instead of dedicated components.

Example stepper:

```yaml
hstack:
  children:
    - label: Step 1
    - label: Step 2
    - label: Step 3
```

Example segmented choice:

```yaml
hstack:
  children:
    - button:
        label: Daily
    - button:
        label: Weekly
    - button:
        label: Monthly
```

## Basic Content Components

| Component | Platforms | Purpose |
| --- | --- | --- |
| `label` | desktop, mobile, web | Ordinary display text. May use literal text or a vocabulary reference. |
| `image` | desktop, mobile, web | Image placeholder. Does not embed or require an actual picture. |

Example:

```yaml
label: Ready
```

```yaml
label:
  vocabulary: vocab.contract-holder
```

Example image placeholder:

```yaml
image:
  label: Product Photo
```

The renderer should draw `image` as a filled placeholder rather than showing a real image.

## Input Components

| Component | Platforms | Purpose |
| --- | --- | --- |
| `input` | desktop, mobile, web | Generic input. Label or hint text may describe date, number, password, search, or other expected content. |
| `textarea` | desktop, mobile, web | Multi-line free text input. |
| `combobox` | desktop, mobile, web | Single choice from a closed or suggested option list. |
| `checkbox` | desktop, mobile, web | Boolean or multi-select option. |
| `radio` | desktop, mobile, web | Single choice among visible mutually exclusive options. |
| `toggle` | desktop, mobile, web | Immediate on/off setting. |
| `slider` | desktop, mobile, web | Numeric range input. |

`combobox` is the baseline semantic name for a combo box or closed-choice input.

Input groups should be represented with ordinary layout components such as `vstack`, `hstack`, `label`, and `input` instead of dedicated `form` or `field` elements.

Example:

```yaml
vstack:
  children:
    - label: Login ID
    - input:
        hint: text
    - label: Password
    - input:
        hint: hidden text
```

Input specialization should be expressed with labels or small hints instead of new component types:

```yaml
input:
  label: Password
  hint: hidden text
```

```yaml
input:
  label: Start Date
  hint: date
```

Radio-style choices may use `radio` when the single-choice control itself matters. For sketches where each option needs rich layout or custom actions, use `vstack` or `hstack` with ordinary components instead.

## Data Display Components

| Component | Platforms | Purpose |
| --- | --- | --- |
| `table` | desktop, web | Tabular collection. |
| `list` | desktop, mobile, web | Repeated items. |
| `tree` | desktop, web | Hierarchical collection. |
| `calendar` | desktop, mobile, web | Date-oriented events or availability. |
| `badge` | desktop, mobile, web | Compact status or count indicator. |

Metrics, charts, timelines, and avatar-like visuals should normally use `label`, `table`, `list`, `image`, or `custom` instead of dedicated baseline components.

## Platform-Specific Hints

Platform hints may influence rendering while preserving semantic type.

Examples:

```yaml
button:
  action: action.save
  label: Save
  platforms:
    desktop:
      placement: action-row
    mobile:
      placement: action-row
    web:
      type: submit
```

The renderer may use these hints to choose sketch placement or labels. Validation may use them to detect missing accessible labels or mobile-specific navigation gaps.

## Minimal v0.1 Component Set

The first implementation does not need to render every catalog component. It should support these components end-to-end:

- `browser`
- `window`
- `mobile`
- `vstack`
- `hstack`
- `spacer`
- `grid`
- `splitter`
- `tabs`
- `dialog`
- `menu`
- `button`
- `toggle`
- `label`
- `image`
- `input`
- `textarea`
- `combobox`
- `checkbox`
- `radio`
- `table`
- `list`
- `custom`

Unsupported component names should produce clear validation warnings and render as generic labeled boxes when possible.

## Related Documents

- [UI Layout DSL](ui-layout-dsl.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [UI Validation Rules](ui-validation-rules.md)
- [Renderer Acceptance Tests](renderer-acceptance-tests.md)
- [Initial Implementation Slice](initial-implementation-slice.md)

## Native-Language Summary

UI コンポーネントは、デスクトップ、モバイル、Web で共通して使える意味ベースの catalog として定義するが、要素数は最小に保つ。screen は Concept でありコンポーネントではない。ルート要素は window、browser、mobile、dialog、menu を基本とする。mobile はスマートフォン枠を持つ root surface として扱う。browser/window/mobile/dialog は描画される title 属性を持てる。window と mobile は menu 属性で固定の chrome-level メニュー文字列を持て、window は上部、mobile は下部に描画する。dialog は buttons で下部右寄せのボタン列を表現できる。root id は任意だが参照時は推奨する。menu は通常メニューとコンテキストメニューを兼ねる。動的な状態は 1 枚の絵に隠さず別ルート定義として表す。普通の表示テキストは label、注釈は全コンポーネントに置ける note 属性で表す。note は製品 UI 文言ではなく、要素の意味、layout footprint、action、input behavior を変えない。実画像が必要ない画像領域と custom の既定 fallback は image プレースホルダーで表す。button は action と anchor で操作や遷移先 id を持てる。クリック可能な遷移は label ではなく button で表し、label は非インタラクティブな表示テキストに限定する。badge 付きの操作は button.badge で表す。button の label には短い文字列や絵文字を使える。toggle は on/off control の baseline component とする。combobox は閉じた選択肢または候補付き入力の baseline component とする。textarea と radio は baseline input component として扱う。id と data はレンダリング結果のメタデータとして保持する。prompt は非表示の AI 向け指示、note は表示される注釈、highlight は特定要素を議論・レビュー用に目立たせる renderer hint として扱う。配置は vstack、hstack、grid、splitter のように明示する。splitter は orientation と sizes で 2 pane の向きと比率を表す。grid は columns で列数を指定でき、省略時は 2 列である。入力グループは form や field ではなく vstack/hstack と label/input で表す。データ表現は table を基本とし、集計や詳細表示は既存の表示要素と layout で表す。toolbar は持たず、ボタン列は hstack または dialog.buttons で表現する。CSS のような細かい制御は扱わない。独自部品は custom として定義する。
