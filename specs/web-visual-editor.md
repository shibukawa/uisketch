---
id: "req.web-visual-editor"
type: "requirement"
title: "Web Visual Editor"
aliases:
  - "Browser Editor"
  - "Drag And Drop Editor"
  - "VB-Style Editor"
tags:
  - "ui"
  - "editor"
  - "web"
  - "wasm"
facts:
  lifecycle.status: "draft"
---

# Web Visual Editor

## Summary

The web visual editor lets users create and edit UI sketch layouts in a browser by dragging semantic components onto a canvas, arranging them like classic Visual Basic style form editors, and saving the result as the same canonical `uisketch` source used by the Go library.

The editor is an authoring surface for [UI Layout DSL](ui-layout-dsl.md), not a separate drawing format. It must preserve semantic component structure so the same source can be validated and rendered by [Sketch Wireframe Renderer](sketch-wireframe-renderer.md).

## Product Intent

The editor should make the YAML-first model easier to author without turning the product into Figma or a pixel-perfect design tool.

Primary user workflows:

- Pick a root surface such as `browser` or `window`.
- Drag components such as `button`, `input`, `table`, `list`, `image`, `vstack`, and `hstack` from a palette.
- Drop components into valid containers or between sibling components.
- Reorder components by drag and drop.
- Select a component and edit semantic properties such as `id`, `label`, `action`, `hint`, `columns`, `children`, and `data`.
- Inspect and navigate the layout through a tree structure view.
- Preview the generated sketch SVG using the same Go renderer logic as the CLI.
- View or edit the generated `.uisketch.md` source, including direct editing of explicit `uisketch` source fence bodies.
- View the current visual editing result as formatted YAML without switching into edit mode.
- Save the canonical `.uisketch.md` file, not an editor-private scene graph.
- Share the current sketch through a URL that embeds a compressed layout payload.
- Copy source code, copy converted text output, and download the current layout, preview SVG, or share payload as local files.
- Save and restore recent editor work from browser `localStorage`.

## Architecture Recommendation

The browser editor should use the Go implementation as the single source of truth for parsing, validation, layout normalization, and SVG preview rendering.

Recommended architecture:

| Layer | Responsibility |
| --- | --- |
| Web UI | Component palette, canvas interaction, tree view, inspector, editable source view, read-only YAML viewer, and file/project shell. |
| Go Wasm module | Parse `.uisketch.md`, parse YAML, normalize layout tree, validate edits, calculate sketch layout, render preview SVG, and serialize canonical YAML. |
| Go CLI/library | Same packages used by the Wasm module for non-browser rendering and tests. |

Wasm is the preferred approach for the first browser editor because it reduces logic drift between the CLI and browser. The editor should not reimplement layout validation or rendering behavior in TypeScript except for immediate UI affordances such as hover targets and drag ghost placement.

The Wasm boundary should expose coarse operations rather than low-level internal structs:

```text
loadLayout(yaml) -> editor document, validation findings, preview SVG
loadSource(markdown) -> editor document, validation findings, preview SVG
insertComponent(parentId, index, componentType, defaults) -> updated YAML, findings, preview SVG
moveComponent(elementId, newParentId, index) -> updated YAML, findings, preview SVG
updateComponent(elementId, patch) -> updated YAML, findings, preview SVG
deleteComponent(elementId) -> updated YAML, findings, preview SVG
formatLayout(document) -> selected uisketch source fence body
formatSource(document, frontmatter, notes) -> canonical .uisketch.md source
encodeSharePayload(document) -> compact payload object or string
decodeSharePayload(payload) -> editor document, validation findings, preview SVG
```

## Editor Document Model

The editor document model should mirror the parsed UI layout tree and add only transient editor state.

Persisted source:

- Root semantic component.
- Child component tree.
- Component properties defined by [UI Component Catalog](ui-component-catalog.md).
- Optional `id`, `anchor`, and `data` metadata.
- Stack sizing properties such as `hstack.widths` and `vstack.heights`.

Transient editor state:

- Selected element ID.
- Hovered drop target.
- Drag source and proposed drop position.
- Inspector editing draft values before commit.
- Viewport zoom and pan.
- Source editor cursor position.
- Expanded/collapsed tree view nodes.
- Tree view filter text or search state.

Transient state must not be written into the canonical YAML layout.

## Tree Structure View

The editor should show a tree structure view of the current layout document alongside the canvas, inspector, and source view.

The tree view should present the semantic component hierarchy, not visual drawing primitives. Each row should identify:

- Component type, such as `browser`, `vstack`, `hstack`, `button`, or `table`.
- Element ID when available.
- Human-readable label, title, action, or hint when available.
- Validation severity marker when findings apply to that element.
- Stack proportion hints such as `widths` or `heights` when present.

Required behavior:

- Selecting a tree row selects the same element on the canvas and inspector.
- Selecting an element on the canvas highlights the same tree row.
- Reordering by drag and drop in the tree updates the same selected `uisketch` source tree as canvas drag and drop.
- Tree drops must use the same placement validation rules as canvas drops.
- Collapsing a tree branch affects only editor state and must not change canonical YAML.
- Deleting, duplicating, or wrapping a selected tree node should use the same editor commands as canvas operations when those commands exist.
- Validation findings should be discoverable from the tree so users can find errors in large layouts without scanning the canvas.
- The tree should handle generated or missing IDs by showing a stable display path derived from the root and child index until an ID is assigned.

The tree structure view must not become a second source of truth. It is a projection of the parsed editor document and should be rebuilt from canonical source after source edits, shared URL decode, file import, or local draft restore.

## YAML Source Editing And Viewer

The editor should provide both an editable source mode and a read-only YAML viewer.

Editable source mode:

- Allows direct editing of the canonical `.uisketch.md` source.
- Allows focused editing of only the selected `uisketch` source fence body when the user does not need to edit frontmatter or Markdown notes.
- Parses edits through the Go Wasm parser after debounce or explicit apply.
- Updates canvas, tree, inspector, validation findings, SVG preview, share payload, and local draft after a valid edit.
- Keeps invalid text in the source editor without overwriting the last valid parsed document.
- Shows parse and validation errors with source locations when available.
- Preserves non-sketch Markdown and frontmatter when only visual edits change a `uisketch` source fence body.

Read-only YAML viewer:

- Shows the current visual editing result as normalized YAML.
- Updates immediately after canvas, tree, or inspector edits.
- Is formatted by the same canonical serializer used for file save and download.
- Is copyable to the clipboard and downloadable as raw `.yaml`.
- Clearly indicates that it is a generated view of the current editor document, not a separate saved source.

The source editor and YAML viewer must not diverge. When source text is valid, both views should represent the same selected `uisketch` source tree after normalization. When source text is invalid, the viewer should continue to show the last valid visual document and the editor should make that stale relationship visible.

## Output Actions

The editor should make generated artifacts easy to copy or download without requiring the user to understand the internal file format.

Required copy actions:

- Copy canonical `.uisketch.md` source.
- Copy the selected `uisketch` source fence body as raw YAML.
- Copy normalized YAML from the read-only YAML viewer.
- Copy converted text output from the ASCII/text renderer.
- Copy the generated share URL.
- Copy validation findings as plain text or Markdown for issue reports.

Required download actions:

- Download canonical `.uisketch.md` source.
- Download the selected `uisketch` source fence body as `.yaml`.
- Download converted ASCII/text output as `.txt`.
- Download the current SVG preview as `.svg`.
- Download the compact share payload as `.uisketch` or `.json`.

Copy and download actions should use the last valid parsed document. If the source editor currently contains invalid text, actions must either operate on the last valid document with a visible stale-state warning or be disabled with a clear explanation.

The converted text output should be produced by the same Go ASCII/text renderer used by the CLI. It should not be a lossy text dump of the tree view.

## Helpful Follow-Up Features

These features are not required for the first runnable editor slice, but they are strong candidates after the core editor is usable:

| Feature | Value |
| --- | --- |
| Undo and redo | Makes drag-and-drop editing safe and expected for visual tools. |
| Keyboard shortcuts | Speeds up copy, paste, delete, duplicate, undo, redo, zoom, and save. |
| Component templates | Lets users start from common layouts such as CRUD list, detail page, dialog, and dashboard. |
| Validation panel with quick navigation | Helps users jump from errors to source, tree, inspector, or canvas. |
| Markdown embed snippet export | Copies a `uisketch:svg` or `uisketch:txt` fence for specs and README files. |
| PNG export | Useful for chat tools and documents that do not render SVG reliably. |
| Print/PDF export | Useful for reviews and offline sharing. |
| Side-by-side diff | Shows how visual edits changed the YAML source. |
| Sample gallery | Helps first-time users understand supported component patterns. |
| Accessibility labels checklist | Finds missing labels on interactive components such as `icon-button`. |

Follow-up features must preserve the canonical source rule: `.uisketch.md` and its explicit `uisketch` source fences remain the source of truth, while copied or exported outputs are generated artifacts.

## Static GitHub Pages Deployment

The browser editor should be deployable to GitHub Pages as a static web application.

Deployment requirements:

- The first hosted version must not require a server-side API.
- The app must work from a repository subpath such as `https://owner.github.io/repo-name/`.
- Routing should use hash routes or another GitHub Pages-compatible approach that survives refreshes without server rewrites.
- Wasm, JavaScript, CSS, schema, and sample assets should use relative URLs or a configurable base path.
- Share URLs should use the fragment form, such as `#s=<encoded-payload>`, so GitHub Pages can serve the same static file for all shared sketches.
- Import, export, localStorage, and URL sharing should work entirely in the browser.
- The deployed app should include a minimal sample sketch so first-time users can edit without importing a file.

Out of scope for GitHub Pages deployment:

- Server-side persistence.
- Account login.
- Private project storage.
- GitHub API writes from the editor.

## Shareable URL Encoding

The editor should support TypeScript Playground style sharing by embedding the current sketch in the URL. A shared URL must be enough to reconstruct the current selected `uisketch` source tree without requiring a server-side saved document.

The share payload should be generated from the canonical editor document, not from transient view state. Selection, hover, drag state, inspector drafts, zoom, pan, and source cursor position must not be included.

Recommended pipeline:

```text
canonical editor document
  -> compact share model
  -> short-key encoded object
  -> stable JSON
  -> lz-string compression
  -> URL-safe base64
  -> URL query or fragment parameter
```

The URL fragment is preferred for a static editor because browsers do not send fragments to servers. A query parameter may be supported when routing constraints require it.

The initial parameter name should be short, for example `s`, so a shared URL looks like:

```text
https://example/editor#s=<encoded-payload>
```

### Short-Key Encoding

Before compression, the editor should replace common type names and field names with stable short symbols to reduce payload size. Symbols should be one or two bytes where practical.

Example mapping:

| Canonical value | Encoded symbol |
| --- | --- |
| `browser` | `b` |
| `window` | `w` |
| `vstack` | `v` |
| `hstack` | `h` |
| `button` | `B` |
| `label` | `L` |
| `input` | `I` |
| `table` | `T` |
| `children` | `c` |
| `id` | `i` |
| `title` | `Tt` |
| `label` field | `l` |
| `action` | `a` |
| `anchor` | `k` |
| `buttons` | `bs` |
| `hint` | `n` |
| `data` | `d` |
| `columns` | `C` |
| `widths` | `ws` |
| `heights` | `hs` |

The symbol table must be versioned. A payload should include a compact version marker so future editors can decode older shared URLs.

Example compact payload shape:

```json
{
  "v": 1,
  "r": {
    "t": "b",
    "i": "equipment-list",
    "c": [
      {"t": "h", "c": [{"t": "B", "l": "Refresh", "a": "action.refresh", "k": "equipment-detail"}]},
      {"t": "T", "i": "equipment-table", "C": ["Equipment", "Status"]}
    ]
  }
}
```

The exact symbol table should be deterministic and shared by the Go Wasm module and the browser shell. If both sides implement encoding, tests must prove that Go and TypeScript produce equivalent canonical payloads for the same document.

Unknown future fields should either round-trip through an extension object or produce a clear unsupported-payload finding. The decoder must never execute script or treat decoded text as trusted HTML.

## Download And Browser Storage

The editor should support local persistence without requiring a backend.

Required download behavior:

- Download canonical `.uisketch.md` source.
- Download the selected `uisketch` source fence body as a `.yaml` file when a raw layout export is useful.
- Download the converted ASCII/text output as a `.txt` file.
- Download the current SVG preview as a `.svg` file.
- Download the compact share payload as a `.uisketch` or `.json` file for offline transfer.
- Use deterministic filenames based on the root ID or title when available.

Required browser storage behavior:

- Autosave the current canonical `.uisketch.md` source, selected `uisketch` source fence body, or compact editor document to `localStorage`.
- Restore the latest local draft when opening the editor without a shared URL or imported file.
- Keep a small list of recent drafts keyed by root ID or generated draft ID.
- Store schema or payload version with each local draft.
- Provide a clear way to discard local drafts.

Loading precedence should be deterministic:

| Source | Priority |
| --- | --- |
| Explicit imported file | Highest after user action. |
| Shared URL payload | Highest on initial page load. |
| `localStorage` draft restore | Used only when no shared URL payload is present. |
| Empty starter document | Used when no import, shared URL, or local draft is available. |

The editor must not silently overwrite a local draft with a shared URL payload until the user changes or saves the loaded sketch.

## Drag And Drop Behavior

The editor should behave like a form editor, but every action must produce valid semantic layout where possible.

Required interactions:

- Dragging from the palette creates a new component with safe defaults.
- Dragging an existing component moves it without changing its semantic type or ID.
- Containers expose visible insertion targets before, after, and inside child lists.
- Invalid drop targets are disabled with a validation message.
- Dropping a control onto an empty canvas prompts or creates a valid root surface when the project has no root.
- Dropping ordinary controls into `browser`, `window`, `dialog`, or `menu` appends them to the root `children` list.
- Dropping ordinary controls next to each other inside an axis container preserves the existing `vstack` or `hstack`.
- Dragging a component across incompatible containers should wrap or reject the move according to deterministic editor rules.

The editor should prefer semantic wrappers over coordinates. It may show bounding boxes and insertion lines, but it must not store absolute positions as the primary layout model.

## Automatic Padding And Spacing

The editor should provide automatic spacing defaults so users do not need to specify padding manually.

Padding is an editor and renderer policy, not a required author-facing DSL property in v0.1. The canonical YAML should not gain per-node `padding`, `margin`, `gap`, or CSS-like spacing unless a later requirement explicitly adds those properties.

Default spacing policy:

| Situation | Default |
| --- | --- |
| Root content inset | 16 px visual inset in the editor preview and SVG renderer. |
| `vstack` child gap | 8 px visual gap. |
| `hstack` child gap | 8 px visual gap. |
| Section/container inner inset | 12 px visual inset when a boundary is drawn. |
| Control text inset | Renderer-owned control padding based on control type. |

The editor may expose named density settings such as `compact`, `normal`, and `loose` later, but the first editor should use a single `normal` spacing profile.

When a user drops a component into a container, the editor should automatically place it using the current container spacing profile. It should not ask the user for pixel values during normal authoring.

## Layout Creation Rules

The editor should create structure according to these deterministic rules:

| User action | Resulting structure |
| --- | --- |
| Start new web sketch | Create a `browser` root with an empty `children` list. |
| Start new desktop sketch | Create a `window` root with an empty `children` list. |
| Drop component into empty root | Append component to root `children`. |
| Drop component before or after a sibling | Insert into the parent `children` at that index. |
| Drop two controls side by side | Create or use an `hstack`. |
| Drop control below another control | Create or use a `vstack`. |
| Drop major content next to sidebar-like content | Prefer `hstack` with `widths: [25, 75]` when the user chooses that placement. |
| Drop table/list/image as main content | Prefer the parent's default remaining-space behavior, or set `widths`/`heights` when the user asks for a specific proportion. |

The editor should show the generated structure in the YAML/source panel so users can learn and review the canonical model.

## Validation Feedback

The editor must surface validation results from [UI Validation Rules](ui-validation-rules.md) as the user edits.

Validation feedback should include:

- Parse errors when the source panel contains invalid YAML.
- Invalid child placement.
- Duplicate resolved element IDs.
- Unknown component types.
- Missing action references.
- Missing vocabulary references when strict vocabulary mode is enabled.
- Unsupported components that will render as generic fallback boxes.

Validation findings should be grouped by severity and linked to the relevant selected component when possible.

## First Runnable Editor Slice

The first browser editor slice should prove the high-risk assumption that the browser can use the same Go logic as the CLI.

In scope:

- Minimal web shell with palette, canvas preview, inspector, and `.uisketch.md` source view.
- Read-only YAML viewer that shows the current visual editing result as normalized YAML.
- Tree structure view synchronized with canvas selection, inspector edits, source edits, validation findings, and drag-and-drop moves.
- Go Wasm adapter that wraps parser, validator, YAML serializer, and SVG renderer operations.
- Static GitHub Pages-compatible build with relative asset paths and fragment-based share URLs.
- Share URL generation and loading using versioned short-key encoding plus lz-string/base64 compression.
- Copy current `.uisketch.md`, selected raw YAML source, converted text output, and share URL.
- Download current `.uisketch.md`, selected raw YAML source, converted text output, and SVG output.
- Autosave and restore one current draft from `localStorage`.
- New sketch starts from a `browser` root.
- Drag from palette for `button`, `label`, `input`, `table`, `vstack`, and `hstack`.
- Reorder within the same parent.
- Edit `id`, root `title`, `label`, `action`, `anchor`, and `hint` fields.
- Render SVG preview after each committed edit.
- Apply the default spacing profile in preview output.

Out of scope:

- Pixel-coordinate freeform layout.
- Multi-user collaboration.
- Figma import.
- Production design-system styling.
- Browser file-system synchronization beyond local import/export or simple project file writes.
- Server-side saved documents or account-backed cloud synchronization.
- Full recent-project management beyond the local draft needed for the first slice.
- Full keyboard accessibility pass, except basic focusable controls and form labels.
- Every catalog component in the first palette.

## Acceptance Criteria

- A user can create a browser sketch without writing YAML.
- Dragging controls updates an explicit `uisketch` source fence in a `.uisketch.md` source that validates with the Go validator.
- The preview SVG is produced by the same Go renderer package used by the CLI path.
- The same YAML renders identically through the CLI and browser Wasm renderer for the covered component set.
- Default padding and gaps appear in the visual preview without adding per-node spacing fields to YAML.
- Invalid drops are rejected or transformed by deterministic rules and produce visible validation feedback.
- The source view round-trips: editing `.uisketch.md` or a selected `uisketch` source fence body updates the canvas, and canvas edits update that fence body.
- The read-only YAML viewer reflects visual edits as normalized YAML and can be copied or downloaded.
- The tree structure view round-trips with canvas and source edits, including selection, reorder, and validation markers.
- The share button produces a URL that can restore the same canonical YAML in a fresh browser session.
- Share payloads use a versioned short-key model before lz-string/base64 compression.
- The user can copy source code, copy converted text output, copy the share URL, download canonical YAML, download converted text output, and download current SVG preview.
- The editor restores an autosaved local draft when no shared URL or imported file is present.
- The static build can be hosted from a GitHub Pages repository subpath and still load Wasm/assets, restore `#s=` share URLs, and run without a backend.

## Open Decisions

- Decide the frontend framework for the web shell.
- Decide whether local project file access should use a backend process, browser File System Access API, or import/export only for the first slice.
- Decide how Go Wasm errors should be encoded for stable TypeScript integration.
- Decide whether the editor should ship as a static web app, a local server launched by the CLI, or both.
- Decide the canonical short-key symbol table and version migration policy for share payloads.
- Decide whether `localStorage` should store canonical `.uisketch.md`, selected raw YAML source fence bodies, compact share payloads, or a combination.
- Decide whether tree drag-and-drop should support cross-container wrapping in the first slice or only same-parent reorder.
- Decide whether the source editor should default to full `.uisketch.md` editing, focused YAML layout editing, or a split toggle.
- Decide the GitHub Pages base-path strategy for local preview, repository Pages, and custom domains.

## Related Documents

- [UI Sketch Library Overview](ui-sketch-library-overview.md)
- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Layout DSL](ui-layout-dsl.md)
- [UI Component Catalog](ui-component-catalog.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [UI Validation Rules](ui-validation-rules.md)
- [Initial Implementation Slice](initial-implementation-slice.md)

## Native-Language Summary

ブラウザ版エディタは、VB のフォームエディタのように palette から semantic component をドラッグして canvas に置き、inspector で `id` や `label` などを編集する authoring surface とする。canvas だけでなく tree structure view でも component hierarchy を確認・選択・並べ替えでき、tree は `.uisketch.md` 内の明示的な `uisketch` source fence body から復元される projection として扱う。source editor では `.uisketch.md` 全体または選択中の `uisketch` source fence body を直接編集でき、ビジュアル編集結果は read-only YAML viewer で正規化 YAML として確認できる。保存される正は editor 独自形式ではなく `type: uisketch` の frontmatter と `uisketch` fence を持つ `.uisketch.md` であり、parse、validation、layout normalize、SVG preview、ASCII/text 変換は Go 実装を Wasm で呼び出して CLI と同じロジックを使う。出力操作として source code、raw YAML source、変換済み text、share URL、validation findings をコピーでき、`.uisketch.md`、raw YAML、text、SVG、compact payload をダウンロードできる。padding や gap は v0.1 では YAML の属性にせず、editor/renderer の既定 spacing policy として 16px root inset、8px stack gap、12px container inset を自動適用する。共有 URL は TypeScript Playground のように、canonical document を短い符号の compact payload に変換し、stable JSON、lz-string、URL-safe base64 を通して `#s=` に埋め込む。localStorage には version 付きの draft を autosave する。エディタは GitHub Pages に配置できる静的 Web アプリとして成立させ、Wasm と assets は相対 URL または base path 設定で読み込み、`#s=` 共有 URL はサーバーなしで復元する。追加候補として undo/redo、shortcuts、templates、validation panel、Markdown embed export、PNG/PDF export、diff、sample gallery、accessibility checklist を検討する。
