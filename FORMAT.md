# uisketch format

This document describes the canonical data structure for `uisketch` layout sources in this repository.

## Scope

The canonical persisted source is a `.uisketch.md` document with `type: uisketch` frontmatter and one or more explicit `uisketch` fenced blocks. The layout tree itself is YAML.

This format is shared by:

- browser mode saves,
- local project files,
- CLI rendering,
- Markdown embedding workflows,
- Wails desktop editing.

## Document Shape

At the layout level, a document is a single root node plus nested child nodes.

Typical root surfaces are:

- `browser`
- `window`
- `mobile`
- `dialog`
- `menu`

Example:

```yaml
browser:
  id: account-editor
  title: Account Editor
  address: https://example.local/accounts/42
  children:
    - vstack:
        children:
          - label: Account
          - input:
              id: name
              label: Name
          - hstack:
              children:
                - spacer:
                - button:
                    label: Save
                    action: save-account
```

The outer key is the root surface type. Everything below it is the semantic component tree.

## Common Fields

The parser preserves these fields on nodes when they are present:

| Field | Meaning |
| --- | --- |
| `id` | Stable element identifier used by selection, links, and tooling. |
| `title` | Visible title for titled root surfaces such as `browser`, `window`, `mobile`, and `dialog`. |
| `label` | Visible label text for controls and content nodes. |
| `action` | Action identifier attached to an interactive control. |
| `anchor` | Navigation target or reference to another definition. |
| `address` | Browser surface address or URL. |
| `hint` | Short helper text for an input-like control. |
| `children` | Nested semantic child nodes. |
| `labels` | Tab labels or other label collections where the component defines them. |
| `columns` | Column labels or grid columns, depending on the component type. |
| `widths` | Stack sizing for horizontal allocation. |
| `heights` | Stack sizing for vertical allocation. |
| `prompt` | Free-form AI guidance attached to the node. |
| `data` | Free-form metadata attached to the node. |

`prompt` is for human- or AI-authored guidance. It is not rendered as visible UI.

`data` is for structured or semi-structured metadata. It is also not rendered as visible UI.

Both fields round-trip through parsing and saving.

## Node Kinds

The repository uses a semantic layout tree rather than a pixel layout.

Container nodes typically use `children` to hold nested nodes:

- `browser`
- `window`
- `mobile`
- `dialog`
- `menu`
- `vstack`
- `hstack`
- `grid`
- `tabs`
- `section`
- `sidebar`
- `split-pane`

Leaf or content nodes usually carry visible content or behavior:

- `button`
- `input`
- `label`
- `table`
- `image`
- `list`
- `spacer`

The exact catalog is implemented in the Go parser, renderer, and editor inspector. The file format is intentionally permissive enough to preserve extra semantic fields as the model evolves.

## Tabs

`tabs` stores its tab labels separately from the active tab body.

The active body goes in `children`. The labels belong in `labels`.

## Prompt And Data

Every node may carry `prompt` and `data`.

- `prompt` is a textarea-friendly note for AI-oriented editing help.
- `data` is a freeform metadata field for project-specific information.

Neither field affects rendering by itself. They are saved with the node and restored on load.

## Round Trip Rules

- Preserve the root surface type.
- Preserve child order.
- Preserve `prompt` and `data`.
- Preserve unknown semantic fields when the editor or serializer supports them.
- Do not serialize transient editor state such as selection, drag state, zoom, or cursor position.

The canonical save result is the YAML source that the editor writes back into the `.uisketch.md` document or project file. Browser mode may wrap the same source in `localStorage`, and local project mode may store it on disk, but the YAML layout tree is the shared core representation.

## Related Specs

- [UI Sketch File Format](specs/ui-sketch-file-format.md)
- [Uisketch CLI](specs/uisketch-cli.md)
- [Web Visual Editor](specs/web-visual-editor.md)
- [UI Component Catalog](specs/ui-component-catalog.md)
- [UI Layout DSL](specs/ui-layout-dsl.md)

