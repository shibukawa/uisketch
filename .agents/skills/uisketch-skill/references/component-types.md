# Component Types

Use uisketch components as semantic UI intent, not final visual design.

## Root Surfaces

| Component | Use |
| --- | --- |
| `browser` | Web page in browser chrome; supports `title`, `address`, `children`. |
| `window` | Desktop application window; supports `title`, `children`. |
| `mobile` | Smartphone frame; use for mobile screen sketches. |
| `dialog` | Modal or transient state; supports `title`, `children`, `buttons`. |
| `menu` | Menu or context menu; may be a root surface. |

## Layout

| Component | Use |
| --- | --- |
| `vstack` | Vertical arrangement. May use `heights`. |
| `hstack` | Horizontal arrangement. May use `widths`. |
| `spacer` | Invisible horizontal filler inside `hstack`. Prefer `spacer:` in generated YAML. |
| `grid` | Regular matrix; `columns` is a positive integer. |
| `splitter` | Two-pane layout with `orientation: horizontal|vertical` and optional `sizes`. |
| `section` | Named content region. |
| `tabs` | One selected tab state; requires `labels` and one active `children` body. |
| `custom` | Project-specific semantic display or control; include `name` or `purpose` when useful. |

Prefer canonical names in new YAML. Treat older names such as `split-pane`, `custom-component`, and button variants as legacy input to normalize when practical.

## Actions And Navigation

| Component | Use |
| --- | --- |
| `button` | Action or navigation trigger. Use `badge` for count/status. |
| `toggle` | On/off control. |

Use `action` for operation IDs and `anchor` for navigation targets. Use `to` only when preserving legacy input. Use `button.anchor` for clickable navigation; `label` is non-interactive text.

## Content And Annotations

| Component | Use |
| --- | --- |
| `label` | Visible text. |
| `note` | Visible sketch annotation for design discussion or implementation reminders. |
| `review` | Open question or review prompt. |
| `image` | Placeholder for an image region; does not embed a real bitmap. |
| `badge` | Compact status or count. |

Scalar shorthand is allowed for leaf labels, such as `label: Ready` or `button: Save`. Use mapping form when the element has `id`, `action`, `anchor`, `data`, `prompt`, `highlight`, or children.

## Inputs

| Component | Use |
| --- | --- |
| `input` | Generic input. Use `label` or `hint` for date, number, password, search, etc. |
| `textarea` | Multi-line text input. |
| `combobox` | Closed-choice or suggested-choice input. |
| `checkbox` | Boolean or multi-select option. |
| `radio` | Single visible choice among mutually exclusive options. |
| `slider` | Numeric range input. |

Represent forms with `vstack`/`hstack` and ordinary labels/inputs. Avoid inventing `form` or `field` components.

## Data Display

| Component | Use |
| --- | --- |
| `table` | Tabular collection; `columns` is a list of visible column labels. |
| `list` | Repeated items. |
| `tree` | Hierarchical collection. |
| `calendar` | Date-oriented events or availability. |

Charts, maps, timelines, avatars, and metrics should usually be `custom-component`, `image`, `table`, `list`, or `label` depending on the implementation intent.

## Shared Properties

Common properties include:

- `id`: Stable element identity for links, tests, editor selection, and code generation context.
- `title`, `label`, `hint`: Visible or annotation text.
- `prompt`: Non-rendered AI-agent instruction. Use for implementation guidance that cannot be represented visually, such as external API loading, state synchronization, permissions, validation rules, or data flow.
- `data`: Structured non-rendered metadata.
- `highlight`: Renderer hint for review emphasis.
- `children`: Child layout nodes.
- `buttons`: Dialog action row.
- `columns`: Integer for `grid`, list of labels for `table`.
- `labels`: Tab labels; selected item is a one-element list.
- `widths`, `heights`, `sizes`: Direct child size slots using integer percentages and `$`.

Prefer `prompt` for natural-language agent instructions and `data` for machine-readable facts. Do not use `prompt` as user-facing helper text.
