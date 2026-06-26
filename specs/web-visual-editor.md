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
  - "frontend"
facts:
  lifecycle.status: "draft"
---

# Web Visual Editor

## Summary

The web visual editor lets users create and edit UI sketch layouts in a browser by dragging semantic components onto a canvas, arranging them like classic Visual Basic style form editors, and saving the result as the same canonical `uisketch` source used by the Go library.

The editor is an authoring surface for [UI Layout DSL](ui-layout-dsl.md), not a separate drawing format. It must preserve semantic component structure so the same source can be validated and rendered by [Sketch Wireframe Renderer](sketch-wireframe-renderer.md).

The browser editor is the GitHub Pages entry point of [Frontend Workspace Architecture](frontend-workspace.md). It should be built as a static Vite application inside the npm workspace and should consume shared TypeScript packages and the Go Wasm adapter rather than keeping a separate one-off browser implementation.

## Product Intent

The editor should make the YAML-first model easier to author without turning the product into Figma or a pixel-perfect design tool.

Primary user workflows:

- Pick a root surface such as `browser` or `window`.
- Start a new browser, window, mobile, dialog, or menu sketch without losing unsaved work accidentally.
- Drag components such as `button`, `input`, `table`, `list`, `image`, `vstack`, and `hstack` from a palette.
- Drop components into valid containers or between sibling components.
- Reorder components by drag and drop.
- Select a component and edit semantic properties such as `id`, `label`, `action`, `hint`, `note`, `columns`, `children`, and `data`.
- Inspect and navigate the layout through a tree structure view.
- Preview the generated sketch SVG using the same Go renderer logic as the CLI.
- View or edit the generated `.uisketch.md` source, including direct editing of explicit `uisketch` source fence bodies.
- View the current visual editing result as formatted YAML without switching into edit mode.
- Save the canonical `.uisketch.md` file, not an editor-private scene graph.
- Share the current sketch through a URL that embeds a compressed layout payload.
- Copy source code, copy converted text output, and download the current layout, preview SVG, or share payload as local files.
- Save named browser-local documents to `localStorage`, load them later, and delete obsolete saved documents.
- In local project mode, open, create, and save `.uisketch.md` files through a local Go backend instead of browser storage.

## Architecture Recommendation

The browser editor should use the Go implementation as the single source of truth for parsing, validation, layout normalization, and SVG preview rendering.

Recommended architecture:

| Layer | Responsibility |
| --- | --- |
| Web UI | Component palette, canvas interaction, tree view, inspector, dedicated source mode, source-mode YAML viewer, and file/project shell. |
| Go Wasm module | Parse `.uisketch.md`, parse YAML, normalize layout tree, validate edits, calculate sketch layout, render preview SVG, and serialize canonical YAML. |
| Go CLI/library | Same packages used by the Wasm module for non-browser rendering and tests. |
| Optional local Go server | Project file listing, file reads, file writes, new file creation, and launch-time file selection for local project editing mode. |
| Shared frontend workspace packages | Host-neutral TypeScript document types, schema access, source-region identities, share-payload helpers, and typed Wasm wrapper shared with the VSCode extension. |

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

The browser package should import these operations through the shared `@uisketch/wasm` wrapper defined in [Frontend Workspace Architecture](frontend-workspace.md). Browser-only code may manage drag events, hit testing, localStorage, import/export, and GitHub Pages routing, but it should not bypass the shared adapter for document mutations that affect canonical YAML.

## Persistence Modes

The editor has two persistence modes with different capabilities.

| Mode | Entry point | Persistence | Sharing | Intended use |
| --- | --- | --- | --- | --- |
| Browser mode | Static web app such as GitHub Pages or local Vite dev server | Named records in browser `localStorage`, plus download/import | Enabled through `#s=<encoded-payload>` URLs | Quick sketches, demos, and shareable examples without a backend. |
| Local project mode | `uisketch edit` launched local Go server or Wails shell | Files inside the selected project root | Disabled by default | Editing repository files directly with explicit save semantics. |

The active mode must be visible in the file/project shell. The UI should not offer `localStorage` save/load/delete or share URL controls in local project mode, because those controls would create competing persistence models. Local project mode may still offer copy and download actions for generated outputs.

Switching between modes is an application relaunch or explicit open action, not an automatic background transition. A static GitHub Pages page must not attempt to write local files. A local project mode session must not silently sync edits to browser `localStorage`.

## Unsaved Change Protection

The editor must track whether the current valid canonical source differs from the last acknowledged save point.

Save points:

- Browser mode save point is the last explicit named `localStorage` save, imported file load that has not been edited, shared URL load that has not been edited, or newly created starter document before the first edit.
- Local project mode save point is the last successful backend file read or write for the current file path.
- Download, copy, preview render, and share URL generation are not save points.
- Autosave drafts may preserve recovery data, but autosave alone must not clear the unsaved indicator.

Operations that would replace or close dirty work must ask for confirmation before proceeding:

- New browser, New window, New mobile, New dialog, New menu, or any equivalent new-root shortcut.
- Loading another browser-local saved document.
- Importing a file or loading a shared URL into the active editor after the initial page load.
- Opening another local project file.
- Creating a new local project file in the active editor.
- Deleting the current browser-local saved document when it is the active dirty document.
- Closing or reloading the browser tab/window while dirty.
- Closing a Wails standalone window while dirty.

The confirmation must state that unsaved changes will be discarded. If the user cancels, the current editor document and dirty state remain unchanged. Browser `beforeunload` handling may use the browser-native message, but in-app navigation and new-document actions should use an editor-owned confirmation dialog.

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
- Undo/redo cursor and source snapshot history metadata.

Transient state must not be written into the canonical YAML layout.

## Palette And Root Surface Selection

The component palette should contain insertable canvas components only.

Required behavior:

- Root surface types such as `browser`, `window`, `mobile`, `dialog`, and `menu` should not appear as draggable palette items.
- Palette item types must use the baseline component names from [UI Component Catalog](ui-component-catalog.md). The editor must offer and emit `toggle` for on/off controls and `combobox` for closed-choice or suggested-choice inputs.
- The palette should include every baseline component that the visual editor can create, including structural components such as `splitter`, visible controls such as `radio`, `textarea`, and `combobox`, and data displays such as `tree`, `calendar`, and `badge` when supported by schema, renderer, and inspector defaults.
- The tree structure view should sit below the component palette in the left-side authoring rail so the properties inspector can use the right rail without competing with hierarchy navigation.
- Component keys outside the baseline catalog must not be offered as palette items. Source using unsupported component names should receive validation feedback until rewritten to supported baseline components or shared element properties such as `note`, `highlight`, `prompt`, `data`, `button.anchor`, or `button.badge`.
- The root surface type should be changed through the root component inspector selector.
- Changing the root surface type must preserve child layout content.
- Surface-specific hidden fields should be preserved while switching surfaces when doing so helps restore user input later. For example, a previous `browser.address` value and a previous titled-surface `title` value may be retained in editor state or hidden root metadata when the user switches to `menu`, then restored if the user switches back to a surface that uses that field.
- Hidden preserved fields must not be emitted as visible controls or canvas chrome for a surface that does not support them.
- The `menu` root surface does not have a visible `title` property in the inspector and does not render a title on the canvas.
- The root inspector should expose the `menu` property only for `window` and `mobile` roots. The editor should present it as an ordered list of text labels, serialize it as `menu: [File, Edit, View]` or equivalent YAML list syntax, and omit it when empty.
- `window.menu` and `mobile.menu` are root chrome attributes, not draggable palette components and not entries in `children`.

## Canvas Interaction Details

The visual editor canvas should keep drag, hover, selection, and delete affordances lightweight so the authored UI remains easy to inspect.

Drag insertion targets:

- Insertion targets should be compact and visually quiet when no drag operation is active.
- Starting a drag operation must not make every possible insertion target look selected or strongly highlighted.
- During a drag operation, only the insertion target currently hit by the pointer should expand, show insertion affordance text if needed, and use a strong highlight.
- Non-hit insertion targets may remain available for hit testing but should stay visually compact and low-contrast.
- The hovered insertion target should communicate the exact sibling position or container position where the component will be inserted.
- Horizontal containers such as `hstack` should show insertion targets as vertical separators, while vertical containers should show them as horizontal separators.

Canvas selection:

- The selected component outline must distinguish layout-only components from visible/content components.
- Layout-only components such as `vstack`, `hstack`, `grid`, `spacer`, and `splitter` should use a dashed selection outline or other dashed treatment consistent with their layout-only canvas boundary.
- Visible/content components should use a solid selection outline.
- Selection treatment should not add persistent labels or explanatory chrome that would be confused with the authored UI.

Hover and delete:

- Canvas components should react to hover as well as click.
- Non-root components should show a compact top-right delete affordance on hover.
- The delete affordance should be a small round button with an `×` glyph, not a text label such as `Delete`.
- Clicking the delete affordance removes the component through the same editor command and undo/redo history path as inspector deletion.
- The root component is never deletable from the canvas.

## Canvas Rendering Fidelity

The canvas should resemble the authored UI more than a debug tree visualization.

Required rendering guidance:

- Avoid persistent explanatory text inside canvas components unless that text is part of the authored component value.
- The component type name, debug labels, placeholder captions, and instructional chrome should be omitted from visible leaf components in normal canvas rendering.
- Layout-only components may keep subtle dashed boundaries because they otherwise have no visible authored representation.
- A `label` component should render primarily as its authored label text. It should not add a border, the word `label`, muted duplicate text, or extra debug text.
- A `button` component should render like a button whose visible text is its authored label. It should not add an additional component header inside the button body. Emoji-only labels are allowed and should size the button so the emoji is not clipped.
- An `input` component should render like an input field using its authored label and hint, without extra debug framing beyond what helps identify the field.
- A `table` component should render as a simple table-like surface with column headers and at least one empty or sample body row, rather than a pipe-delimited text line.
- `image` and `list` may use lightweight placeholders, but those placeholders should still look like the authored UI element rather than a debug node record.
- Containers may expose editor-only boundaries and insertion zones, but those affordances should be visually subordinate to authored content.

Root surface chrome:

- `browser` canvas chrome should resemble a browser frame with back, forward, reload, address display, and three window control buttons.
- `window` canvas chrome should show a title-bar style frame with three window control buttons. When `window.menu` is present, the canvas should show a fixed top menu row below the title bar and should keep ordinary `children` content below that row.
- `mobile` canvas chrome should show the device frame and title treatment when present. When `mobile.menu` is present, the canvas should show a fixed bottom navigation/command row inside the device frame and keep ordinary `children` content above that row.
- `dialog` canvas chrome should show a compact title-bar style frame with one close button. When `dialog.buttons` is present, the canvas should show a fixed bottom action row with buttons right-aligned.
- `menu` canvas chrome should not show a title.

Dialog buttons:

- The visual editor should support `dialog.buttons` as a first-class root property when the selected root is `dialog`.
- `dialog.buttons` should be edited as an ordered list of ordinary button nodes, with at least `label`, `action`, `anchor`, `badge`, `note`, `prompt`, and `data` fields available through the same inspector behavior used for normal `button` components.
- The visual canvas should render `button.badge` as a small marker on the button border using the authored value as plain text inside the marker.
- Canvas selection should make dialog action buttons selectable and editable, but their insertion surface should be visually separate from main `children` so authors understand they serialize under `buttons`, not under `children`.
- Dragging a button into the dialog action row should insert into `buttons`; dragging a non-button component there may be rejected or converted only through an explicit user action. The baseline authoring path should prefer `button` nodes.
- Empty `dialog.buttons` should be omitted from serialized YAML.

Spacer rendering:

- A `spacer` should communicate tension or stretch direction, similar to spacer widgets in Qt Creator.
- In a horizontal stack, a spacer should render as a horizontal stretch affordance and consume remaining horizontal space.
- In a vertical stack, a spacer should render as a vertical stretch affordance and consume remaining vertical space.
- Spacer visuals may use arrows, dashed tension lines, and small end handles, but should remain editor affordances rather than authored UI content.
- If a horizontal stack contains `button`, `spacer`, `button`, the second button should be pushed to the right edge because the spacer absorbs the remaining width.
- Spacer behavior should follow the parent layout direction. If a spacer is moved between horizontal and vertical containers, its visual orientation should update accordingly.

## Tree Structure View

The editor should show a tree structure view of the current layout document alongside the canvas and inspector in visual mode.

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

The editor should provide a top-level source editing mode matching the second layout in `web/webeditor.md`, plus a source-mode YAML viewer for normalized output.

Source mode layout:

- The primary editor tabs are `Visual Editor` and `Source`; selecting `Source` replaces the visual palette/canvas/inspector workspace with a source-focused split view.
- Visual mode must not include a lower source display or embedded source viewer. Source text, normalized YAML, diff, validation, and ASCII output belong in the dedicated `Source` mode or explicit output actions.
- The left side of source mode is an editable code editor for the canonical `.uisketch.md` document or the focused selected `uisketch` fence body.
- The right side of source mode is a preview area with `Preview(SVG)` and `Preview(ASCII)` tabs.
- `Preview(SVG)` must render the same SVG output as the Go CLI renderer for the current last valid source.
- `Preview(ASCII)` must render the same ASCII/text output as the Go CLI renderer for the current last valid source.
- Source mode preview tabs should use the same renderer adapter and validation state as the visual editor preview. They must not introduce browser-only renderer behavior.
- The source editor and visual editor operate on the same canonical source buffer and selected `uisketch` source region, so users can make a visual edit, refine the generated source by hand, then return to visual editing without importing or converting through another format.

Editable source mode:

- Allows direct editing of the canonical `.uisketch.md` source.
- Allows focused editing of only the selected `uisketch` source fence body when the user does not need to edit frontmatter or Markdown notes.
- Provides code-editor behavior suitable for structured source editing: line numbers, indentation support, bracket matching, search, and keyboard-friendly multiline editing.
- Provides syntax highlighting for Markdown frontmatter, fenced `uisketch` blocks, and UI Layout DSL YAML.
- Provides keyword and schema checks for known component names, root surface types, property keys, enum-like values, stack proportion syntax, and invalid nesting where the schema or validator can identify it.
- Provides completion candidates for component types, common properties, root types, tab/stack/grid fields, and locally known element IDs when editing references such as anchors or selection-related properties.
- Shows diagnostics inline or in a nearby findings list, using Go Wasm parser and validator findings as the source of truth. TypeScript-side editor checks may provide earlier hints, but they must not contradict accepted Go validation results.
- Parses edits through the Go Wasm parser after debounce or explicit apply.
- Updates canvas, tree, inspector, validation findings, SVG preview, share payload, and local draft after a valid edit.
- Keeps invalid text in the source editor without overwriting the last valid parsed document.
- Shows parse and validation errors with source locations when available.
- Preserves non-sketch Markdown and frontmatter when only visual edits change a `uisketch` source fence body.

Shared source behavior:

- A valid source commit replaces the current editor document and rebuilds the visual canvas, tree, inspector, YAML viewer, source mode previews, output actions, share payload, local draft, and undo/redo snapshot from that same canonical source.
- A visual edit serializes back into the selected `uisketch` source fence body and refreshes the source editor text. When the source editor is showing the full `.uisketch.md` document, non-sketch Markdown and frontmatter must be preserved around the updated fence body.
- If the user has an invalid unsaved source buffer, visual edits must not silently discard it. The editor should either require the user to accept/revert the invalid text before visual editing changes the same source region, or keep the invalid buffer clearly marked as stale while visual editing continues from the last valid document.
- Switching between `Visual Editor` and `Source` must preserve source cursor/selection state, selected visual element when it can be mapped, current preview tab, and validation navigation context as view state rather than canonical source.
- When Go Wasm exposes source ranges, clicking a validation finding or a selected visual/tree element should navigate the code editor to the best matching source location. Missing ranges must fall back to the selected source region or root fence.

Source-mode YAML viewer:

- Shows the current visual editing result as normalized YAML inside the dedicated `Source` mode, not as a lower panel under the visual editor.
- Updates immediately after canvas, tree, or inspector edits.
- Is formatted by the same canonical serializer used for file save and download.
- Is copyable to the clipboard and downloadable as raw `.yaml`.
- Clearly indicates that it is a generated view of the current editor document, not a separate saved source.

The source editor and YAML viewer must not diverge. When source text is valid, both views should represent the same selected `uisketch` source tree after normalization. When source text is invalid, the viewer should continue to show the last valid visual document and the editor should make that stale relationship visible.

## Inspector Editing Details

Inspector controls should optimize for ordinary editing, including temporary invalid values while the user is typing.

Tabs inspector:

- `tabs.labels` should not be edited as raw YAML syntax in a single text field when a structured UI is available.
- The inspector should present tab labels as an editable list of items.
- The selected tab should be edited separately from the label list, preferably with a combobox control populated from the current tab labels.
- Inspector changes should serialize back to the canonical YAML array representation used by [UI Layout DSL](ui-layout-dsl.md).
- If the selected tab is removed or renamed, the inspector should deterministically choose a valid selected tab or report a local inspector validation issue before committing.

Numeric inspector fields:

- Numeric fields such as `grid.columns` should allow a temporary empty string or syntactically invalid draft while the input is focused.
- Temporary invalid inspector drafts should be marked as local editor errors and must not immediately overwrite the last valid canonical YAML value.
- Canvas layout should use the last valid numeric value while the inspector draft is invalid.
- A user must be able to delete the final digit of a numeric field in order to replace it with another value.
- Committing a valid numeric draft updates canonical YAML, tree, canvas, source, undo/redo history, and validation state.
- `grid.columns: 2` should lay out children into two columns from left to right; children must not collapse into only the right column.

## Undo And Redo

The browser editor should support undo and redo across visual edits and valid source edits.

The history model should store snapshots of the selected canonical YAML source rather than only storing visual editor commands. YAML snapshots are preferred because users may directly edit source code, and visual command inversion cannot represent every source edit.

Required behavior:

- Record a history entry after each committed visual operation that changes canonical YAML.
- Record a history entry after a source edit is parsed successfully and accepted as the current valid document.
- Do not record invalid source text as a canonical history entry.
- Preserve the invalid source editing buffer separately from the last valid YAML snapshot.
- Undo restores the previous valid YAML snapshot, then rebuilds canvas, tree, inspector, source view, YAML viewer, validation findings, and preview from that snapshot.
- Redo reapplies the next valid YAML snapshot in the same way.
- Consecutive source editor keystrokes should be coalesced by debounce, explicit apply, blur, or another deterministic commit boundary so history does not record every character.
- Visual edits made after undo should discard redo entries, following ordinary editor behavior.

The browser source editor may also have its own text-level undo stack while focused. The product-level undo/redo history should be integrated carefully: text-level undo handles in-progress source typing, while product-level undo restores accepted YAML snapshots that affect the visual document. The exact keyboard shortcut routing can be implementation-defined, but the UI must make it clear whether the current invalid source buffer or the last valid visual document is being restored.

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
| Accessibility labels checklist | Finds missing labels on interactive components, including compact or emoji-only buttons. |

Follow-up features must preserve the canonical source rule: `.uisketch.md` and its explicit `uisketch` source fences remain the source of truth, while copied or exported outputs are generated artifacts.

## Static GitHub Pages Deployment

The browser editor should be deployable to GitHub Pages as a static web application.

Deployment requirements:

- The GitHub Pages app is built from the npm workspace browser package using Vite or an equivalent static bundler.
- The first hosted version must not require a server-side API.
- The app must work from a repository subpath such as `https://owner.github.io/repo-name/`.
- Routing should use hash routes or another GitHub Pages-compatible approach that survives refreshes without server rewrites.
- Wasm, JavaScript, CSS, schema, and sample assets should use relative URLs or a configurable base path.
- Share URLs should use the fragment form, such as `#s=<encoded-payload>`, so GitHub Pages can serve the same static file for all shared sketches.
- Import, export, localStorage, and URL sharing should work entirely in the browser.
- The deployed app should include a minimal sample sketch so first-time users can edit without importing a file.
- The first screen should follow the visual structure specified in `web/webeditor.md`, with primary tabs for visual editing and source editing, visual sub-tabs for edit/SVG preview/ASCII preview, a component palette, properties inspector, and tree inspector.

Out of scope for GitHub Pages deployment:

- Server-side persistence.
- Account login.
- Private project storage.
- GitHub API writes from the editor.

## Shareable URL Encoding

The editor should support TypeScript Playground style sharing by embedding the current sketch in the URL. A shared URL must be enough to reconstruct the current selected `uisketch` source tree without requiring a server-side saved document.

The share payload should be generated from the canonical editor document and associated with the current editor URL. Selection, hover, drag state, inspector drafts, zoom, pan, source cursor position, local file paths, and local project roots must not be included.

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

The share action should produce the current page URL plus the encoded content key in the fragment. In other words, it should preserve the current deployment base URL and route, replace any previous share key for the active document, and copy a URL whose `s` value alone is sufficient to reconstruct the shared sketch in browser mode.

Share URLs are a browser-mode feature. Local project mode must not expose a share URL action unless the user explicitly exports a browser-mode copy that removes local file identity and uses only canonical content.

### Short-Key Encoding

Before compression, the editor should replace common type names and field names with stable short symbols to reduce payload size. Symbols should be one or two bytes where practical.

Example mapping:

| Canonical value | Encoded symbol |
| --- | --- |
| `browser` | `b` |
| `window` | `w` |
| `vstack` | `v` |
| `hstack` | `h` |
| `splitter` | `s` |
| `button` | `B` |
| `label` | `L` |
| `input` | `I` |
| `table` | `T` |
| `children` | `c` |
| `id` | `i` |
| `title` | `Tt` |
| `label` field | `l` |
| `note` field | `n` |
| `action` | `a` |
| `anchor` | `k` |
| `buttons` | `bs` |
| `menu` root field | `m` |
| `input hint` field | `ih` |
| `data` | `d` |
| `columns` | `C` |
| `orientation` | `o` |
| `sizes` | `ss` |
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

Required browser-mode storage behavior:

- Let the user explicitly save the current canonical `.uisketch.md` source under a user-provided document name.
- Store named documents in `localStorage` with a stable generated document ID, display name, created time, updated time, schema or payload version, and canonical source content.
- Autosave recovery drafts separately from named saved documents so a crash or reload can recover work without pretending it was explicitly saved.
- Restore the latest local draft when opening the editor without a shared URL or imported file.
- Keep a list of saved browser-local documents that can be loaded later.
- Let the user rename a saved browser-local document without changing the canonical sketch source unless the user also edits source metadata.
- Let the user delete browser-local saved documents after confirmation.
- Provide a clear way to discard autosave recovery drafts.

Suggested `localStorage` keys:

| Key | Value |
| --- | --- |
| `uisketch.saved.index.v1` | Ordered list of saved document metadata and IDs. |
| `uisketch.saved.<id>.v1` | Canonical `.uisketch.md` source and save metadata for one named document. |
| `uisketch.draft.current.v1` | Autosave recovery state for the active unsaved browser-mode document. |

The named save dialog should require a non-empty name. If a save name already exists, the editor should ask whether to overwrite that saved browser-local document or choose a different name.

Loading precedence should be deterministic:

| Source | Priority |
| --- | --- |
| Explicit imported file | Highest after user action. |
| Shared URL payload | Highest on initial page load. |
| `localStorage` draft restore | Used only when no shared URL payload is present. |
| Empty starter document | Used when no import, shared URL, or local draft is available. |

The editor must not silently overwrite a local draft with a shared URL payload until the user changes or saves the loaded sketch.

## Local Project File Mode

Local project mode lets the editor read and write files in a repository through a local Go backend. This mode exists for users who want the visual editor to update actual `.uisketch.md` files rather than browser-local storage.

Required backend behavior:

- Start from an explicit project root, defaulting to the current working directory when launched by the CLI.
- Restrict all file listing, read, create, and write operations to the project root after path cleaning and symlink resolution.
- List editable files under the project root, prioritizing `.uisketch.md` and `.md` files that contain `type: uisketch` frontmatter or explicit `uisketch` source fences.
- Read a selected file and return content, normalized project-relative path, last modified time, size, and a revision token such as mtime plus size or a content hash.
- Write only when the client provides the last known revision token, so the backend can detect external file changes before overwriting.
- Create new `.uisketch.md` files from a safe starter document and reject paths that escape the project root.
- Return structured errors for permission denied, file not found, path outside project root, unsupported file type, invalid filename, and external modification conflict.

Required UI behavior:

- Show a project file list or file picker backed by the local server.
- Open a selected file into the same visual/source editor used in browser mode.
- Save writes the canonical `.uisketch.md` source back to the opened file path through the backend.
- Save As or New File creates a new project-relative file and then makes that path the active save target.
- If the active file changed on disk since it was opened or last saved, the editor must warn before overwriting and offer at least reload or cancel.
- Local project mode must not use named `localStorage` saves, autosave recovery as an authoritative save, or share URLs as persistence.

When a file path is supplied at launch, the local project editor should open that file immediately after startup if it is inside the project root. If the file does not exist, the editor may offer to create it as a new `.uisketch.md` file after confirming the project-relative path.

## Standalone App Option

The local project backend should be designed so the same file operations can be reused by a Wails standalone app.

Wails packaging is optional for the first local project slice, but the architecture should keep these constraints:

- The web UI should call a host file API interface rather than hard-coding HTTP-specific details throughout the editor.
- The Go file service should be reusable from both an HTTP local server and Wails bindings.
- Browser mode must remain static-hostable without Wails or a local server.
- Wails window close handling must use the same dirty-document confirmation rules as browser tab close handling.

## Drag And Drop Behavior

The editor should behave like a form editor, but every action must produce valid semantic layout where possible.

Required interactions:

- Dragging from the palette creates a new component with safe defaults.
- Dragging an existing component moves it without changing its semantic type or ID.
- Containers expose visible insertion targets before, after, and inside child lists.
- Insertion targets should be compact in the normal state, roughly a 3 px high or wide line depending on insertion direction.
- Insertion targets should not expand or show instructional text merely because the mouse pointer hovers over them without an active drag.
- During an active drag, insertion targets should become easier to see, and the hovered drop target should expand enough to communicate where the component will be inserted.
- The hovered active drop target should use a clear highlight treatment such as an accent line, soft fill, or stronger outline.
- Invalid drop targets are disabled with a validation message.
- Dropping a control onto an empty canvas prompts or creates a valid root surface when the project has no root.
- Dropping ordinary controls into `browser`, `window`, `dialog`, or `menu` appends them to the root `children` list.
- Dropping ordinary controls next to each other inside an axis container preserves the existing `vstack` or `hstack`.
- Dragging a component across incompatible containers should wrap or reject the move according to deterministic editor rules.

The editor should prefer semantic wrappers over coordinates. It may show bounding boxes and insertion lines, but it must not store absolute positions as the primary layout model.

## Canvas Hover And Component Actions

Canvas components should respond to pointer hover as well as click selection.

Required behavior:

- Hovering a component highlights the component boundary without changing canonical YAML.
- Hovering a non-root component shows a small delete affordance in the component's top-right corner.
- Clicking the hover delete affordance removes that component from its parent and selects the nearest valid parent or sibling.
- The root component must not show a delete affordance because a document always has exactly one root surface.
- Hover controls must not overlap essential component labels or make the canvas jump.
- Hover controls are editor UI only and must not be serialized into YAML.

Click selection remains required. Hover should make the editor feel responsive and discoverable, but editing commands that change YAML still require an explicit click or drop.

## Component Visual Classification

The visual editor should clearly distinguish layout-only components from components that represent visible UI content.

Layout-only components include:

- `vstack`
- `hstack`
- `grid`
- `spacer`
- `splitter`

Layout-only components should use a structural treatment such as a dashed outline, lighter fill, layout glyph, or muted chrome. This treatment communicates that they arrange children rather than represent visible product UI.

Visible or content-bearing components such as `button`, `label`, `input`, `table`, `list`, `image`, `section`, `tabs`, `dialog`, `menu`, and `custom` should use a stronger component treatment that resembles a concrete UI object or semantic surface.

The palette should separate layout components from visible/content components. A minimal grouping is:

| Palette Group | Examples |
| --- | --- |
| Layout | `vstack`, `hstack`, `grid`, `spacer`, `splitter` |
| Components | `button`, `label`, `input`, `table`, `list`, `image`, `section`, `tabs`, `custom` |

The palette grouping is an authoring affordance only. It must not change the canonical component vocabulary or generated YAML.

## Root Surface Editing

The editor document always has exactly one root component.

Required behavior:

- The root component cannot be deleted from the canvas, tree, hover menu, inspector, or keyboard commands.
- The root component type can be changed through the inspector using a combobox control.
- The first supported root type choices should include `browser`, `window`, `mobile`, `dialog`, and `menu`.
- Changing the root type preserves compatible fields such as `id`, `title`, and `children`.
- Changing the root type drops or hides fields that are not meaningful for the selected root type, such as `browser.address` when switching to `window`.
- Root type changes should be recorded in undo/redo history as ordinary accepted document changes.
- The generated YAML must continue to have exactly one root mapping after a root type change.

The editor may keep root creation shortcuts such as "New browser" and "New window", but those shortcuts reset the document and are separate from changing the current root type.

## Initial Palette Coverage

The first useful visual editor palette should include common layout and visible components beyond the earliest MVP.

Required palette entries:

- Root/surface components: `browser`, `window`, `mobile`, `dialog`, `menu`.
- Layout components: `vstack`, `hstack`, `grid`, `spacer`.
- Visible/content components: `button`, `label`, `input`, `table`, `list`, `image`, `section`, `tabs`.

Palette entries may create safe default children when the component is hard to understand empty. For example, `tabs` may start with two labels and one active child body, and `grid` may start with `columns: 2`.

The editor should reject or adapt invalid placements. For example, `spacer` should be allowed primarily inside `hstack`; dropping it elsewhere should either be rejected with a clear message or wrapped/transformed by a deterministic rule.

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
| Change root type in inspector | Replace only the root surface type while preserving compatible metadata and children. |
| Drop component into empty root | Append component to root `children`. |
| Drop component before or after a sibling | Insert into the parent `children` at that index. |
| Drop two controls side by side | Create or use an `hstack`. |
| Drop control below another control | Create or use a `vstack`. |
| Drop `grid` | Create `grid` with `columns: 2` unless the user chooses another column count. |
| Drop `tabs` | Create `tabs` with safe default labels and one active child body. |
| Drop `spacer` into `hstack` | Insert an invisible spacer child. |
| Drop major content next to side-panel-like content | Prefer `hstack` with `widths: [25, 75]` when the user chooses that placement. |
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

Finding payloads do not need to share one universal frontend format across Web and VSCode. The browser editor may adapt Go validation results into the shape needed for its validation panel, tree markers, inspector messages, and canvas overlays. Shared code should preserve source range, element ID, element path, severity, and message when available, but host-specific formatting is acceptable.

## Web Security

The browser editor must treat imported files, pasted source, shared URL payloads, and localStorage drafts as untrusted input.

Required rules:

- Decode share payloads with size limits and clear unsupported-version findings.
- Never execute scripts from Markdown, YAML, decoded payloads, or generated preview SVG.
- Render preview SVG through a controlled container and avoid inserting arbitrary unsanitized HTML.
- Keep localStorage data origin-local and provide a way to discard drafts.
- Do not upload source, generated SVG, validation findings, or drafts to a server in the GitHub Pages first slice.
- Clipboard and download actions should operate only after explicit user action.

The first supported browser set is current Chrome and Safari. Firefox support may be added later, but first-slice behavior should not depend on browser-only APIs that exclude Safari unless an import/export fallback exists.

## First Runnable Editor Slice

The first browser editor slice should prove the high-risk assumption that the browser can use the same Go logic as the CLI.

In scope:

- Minimal web shell with palette, canvas preview, inspector, and dedicated `.uisketch.md` source mode.
- Source editing mode matching the second `web/webeditor.md` layout, with a code editor on the left and SVG/ASCII preview tabs on the right.
- Vite-based npm workspace package for the GitHub Pages app as specified by [Frontend Workspace Architecture](frontend-workspace.md).
- Source-mode YAML viewer that shows the current visual editing result as normalized YAML without adding a lower source panel to visual mode.
- Tree structure view synchronized with canvas selection, inspector edits, source edits, validation findings, and drag-and-drop moves.
- Go Wasm adapter that wraps parser, validator, YAML serializer, and SVG renderer operations.
- Static GitHub Pages-compatible build with relative asset paths and fragment-based share URLs.
- Share URL generation and loading using versioned short-key encoding plus lz-string/base64 compression.
- Copy current `.uisketch.md`, selected raw YAML source, converted text output, and share URL.
- Download current `.uisketch.md`, selected raw YAML source, converted text output, and SVG output.
- Autosave and restore one current draft from `localStorage`.
- Unsaved-change confirmation for new root shortcuts, file loads, document loads, and browser/tab close.
- Undo and redo based on accepted YAML snapshots.
- New sketch starts from a `browser` root.
- Drag from palette for `button`, `label`, `input`, `table`, `list`, `image`, `section`, `tabs`, `vstack`, `hstack`, `grid`, and `spacer`.
- Reorder within the same parent.
- Separate palette groups for layout components and visible/content components.
- Distinct canvas styling for layout-only components and visible/content components.
- Compact insertion targets that expand and highlight only during active drag/drop operations.
- Hover highlight and top-right delete affordance for non-root components.
- Root component type selection in the inspector without allowing root deletion.
- Edit `id`, root `title`, root `menu` for `window` and `mobile`, `dialog.buttons`, `label`, `action`, `anchor`, `hint`, and `note` fields. The `note` field should be a single-line inspector input.
- Render SVG preview after each committed edit.
- Render SVG and ASCII previews from source mode through the same Go Wasm renderer path used by the CLI.
- Provide source editor highlighting, basic keyword/schema diagnostics, and completion for the first-slice component and property set.
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
- Performance tuning for very large sketches before real usage identifies bottlenecks.

## Acceptance Criteria

- A user can create a browser sketch without writing YAML.
- Dragging controls updates an explicit `uisketch` source fence in a `.uisketch.md` source that validates with the Go validator.
- The preview SVG is produced by the same Go renderer package used by the CLI path.
- The same YAML renders identically through the CLI and browser Wasm renderer for the covered component set.
- Default padding and gaps appear in the visual preview without adding per-node spacing fields to YAML.
- Invalid drops are rejected or transformed by deterministic rules and produce visible validation feedback.
- Drop insertion targets are compact at rest and expand/highlight only while dragging.
- Layout-only components are visually distinguishable from concrete UI/content components in both the palette and canvas.
- Hovering a non-root component reveals a delete affordance that removes that component without exposing delete on the root.
- The root component can be changed between supported root surface types through the inspector and cannot be deleted.
- The source view round-trips: editing `.uisketch.md` or a selected `uisketch` source fence body updates the canvas, and canvas edits update that fence body.
- Visual mode has no lower source display; it is limited to palette/tree, canvas preview, and inspector/output controls.
- Source mode follows the second `web/webeditor.md` layout with the active `Source` tab, editable code on the left, and `Preview(SVG)` / `Preview(ASCII)` tabs on the right.
- Source mode previews are generated from the last valid source by the same Go renderer logic used by the CLI for SVG and ASCII/text output.
- The source editor provides syntax highlighting, completion, and keyword/schema diagnostics for the supported UI Layout DSL component set.
- A user can move between visual editing and source editing without changing documents or losing the shared canonical source state.
- The source-mode YAML viewer reflects visual edits as normalized YAML and can be copied or downloaded.
- The tree structure view round-trips with canvas and source edits, including selection, reorder, and validation markers.
- The share button produces a URL that can restore the same canonical YAML in a fresh browser session.
- Share payloads use a versioned short-key model before lz-string/base64 compression.
- The user can copy source code, copy converted text output, copy the share URL, download canonical YAML, download converted text output, and download current SVG preview.
- The editor restores an autosaved local draft when no shared URL or imported file is present.
- Browser mode supports explicit named save, load, delete, and rename for `localStorage` documents.
- Dirty browser-mode work asks for confirmation before New browser, New window, load, import, shared URL replacement, tab close, or reload discards it.
- Local project mode can list project files, open one, save changes back to disk, create a new `.uisketch.md` file, and detect external modifications before overwrite.
- Local project mode hides browser-local saved document and share URL persistence controls.
- Undo and redo restore accepted YAML snapshots across visual edits and valid source edits.
- The static build can be hosted from a GitHub Pages repository subpath and still load Wasm/assets, restore `#s=` share URLs, and run without a backend.
- The web package can be built through the root npm workspace scripts and does not require manually copying Wasm or schema files.

## Open Decisions

- Decide the frontend framework for the web shell.
- Decide whether the Vite package should initially stay framework-light or use a component framework.
- Decide how Go Wasm errors should be encoded for stable TypeScript integration.
- Decide the canonical short-key symbol table and version migration policy for share payloads.
- Decide whether tree drag-and-drop should support cross-container wrapping in the first slice or only same-parent reorder.
- Decide whether the source editor should default to full `.uisketch.md` editing, focused YAML layout editing, or a split toggle.
- Decide the concrete browser code editor implementation for source mode, such as CodeMirror, Monaco, or a smaller custom editor wrapper.
- Decide the GitHub Pages base-path strategy for local preview, repository Pages, and custom domains.
- Decide whether Wails packaging should be part of the first local project release or a later distribution step.

## Related Documents

- [UI Sketch Library Overview](ui-sketch-library-overview.md)
- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Layout DSL](ui-layout-dsl.md)
- [UI Component Catalog](ui-component-catalog.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [UI Validation Rules](ui-validation-rules.md)
- [Initial Implementation Slice](initial-implementation-slice.md)
- [Frontend Workspace Architecture](frontend-workspace.md)

## Native-Language Summary

ブラウザ版エディタは、VB のフォームエディタのように palette から semantic component をドラッグして canvas に置き、inspector で `id` や `label` などを編集する authoring surface とする。canvas だけでなく tree structure view でも component hierarchy を確認・選択・並べ替えでき、tree は `.uisketch.md` 内の明示的な `uisketch` source fence body から復元される projection として扱う。drop target は通常時は 3px 程度の省スペースな挿入線として扱い、drag 中だけ展開・ハイライトして挿入位置を示す。layout component（vstack、hstack、grid、spacer など）と visible/content component は palette group と canvas styling の両方で区別し、layout component は破線や薄い chrome で構造要素であることを示す。component hover では非 root component の右上に削除 affordance を表示するが、root component は常に 1 つ必要なので削除不可とし、inspector の combobox で browser/window/mobile/dialog/menu などへ入れ替える。source editor では `.uisketch.md` 全体または選択中の `uisketch` source fence body を直接編集でき、ビジュアル編集結果は read-only YAML viewer で正規化 YAML として確認できる。保存される正は editor 独自形式ではなく `type: uisketch` の frontmatter と `uisketch` fence を持つ `.uisketch.md` であり、parse、validation、layout normalize、SVG preview、ASCII/text 変換は Go 実装を Wasm で呼び出して CLI と同じロジックを使う。New browser/New window などの新規作成、別 document/file の load、browser tab/window close は dirty な変更がある場合に破棄確認を出す。ブラウザモードでは名前付き document を localStorage に保存・読込・削除でき、共有 URL は現在の URL に canonical content key を `#s=` として付けて復元可能にする。ローカルプロジェクトモードでは Go backend が project root 内のファイル一覧、open、save、新規作成を担当し、localStorage 保存と共有 URL は使わない。CLI からファイル名付きで起動するとそのファイルを直接開く。将来的な Wails standalone app でも同じ Go file service と dirty close confirmation を再利用できるようにする。出力操作として source code、raw YAML source、変換済み text、share URL、validation findings をコピーでき、`.uisketch.md`、raw YAML、text、SVG、compact payload をダウンロードできる。padding や gap は v0.1 では YAML の属性にせず、editor/renderer の既定 spacing policy として 16px root inset、8px stack gap、12px container inset を自動適用する。エディタは GitHub Pages に配置できる静的 Web アプリとして成立させ、Wasm と assets は相対 URL または base path 設定で読み込む。追加候補として undo/redo、shortcuts、templates、validation panel、Markdown embed export、PNG/PDF export、diff、sample gallery、accessibility checklist を検討する。
