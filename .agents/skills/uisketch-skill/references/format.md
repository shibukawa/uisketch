# Uisketch Format

## Canonical `.uisketch.md`

A canonical uisketch source file is Markdown with YAML frontmatter and one or more explicit renderable source fences.

Required frontmatter:

```yaml
---
id: screen.example
type: uisketch
title: Example
screen:
  id: screen.example
---
```

Only `id` and `type: uisketch` are required. `title`, `screen.id`, `status`, `source`, and `tags` are common metadata.

Renderable source must use one of these fence info strings:

- `uisketch` or `uisketch:svg`: SVG-oriented source.
- `uisketch:txt`, `uisketch:text`, or `uisketch:ascii`: ASCII-oriented source.

Example:

````markdown
```uisketch
browser:
  id: equipment-list
  title: Equipment List
  address: https://example.internal/equipment
  children:
    - hstack:
        children:
          - button:
              id: refresh
              action: action.refresh-equipment
              label: Refresh
              prompt: Fetch the latest equipment list from the backend before updating the table.
          - spacer:
          - button:
              id: create-alert
              action: action.create-alert
              label: Create Alert
    - table:
        id: equipment-table
        columns:
          - Equipment
          - Status
    - label: Ready
```
````

## Multiple Sketches

A document may contain multiple renderable definitions. Selection order is document order and 1-origin when a command accepts an index. Use separate headings for readability, but do not make the heading part of the semantics.

Use this pattern for related states:

````markdown
## Main Screen

```uisketch
browser:
  id: orders
  title: Orders
```

## Delete Confirmation

```uisketch
dialog:
  id: confirm-delete
  title: Confirm Delete
  children:
    - label: Delete this order?
  buttons:
    - button: Cancel
    - button:
        label: Delete
        action: action.delete-order
```
````

## Generated Output Comments

Rendered Markdown may replace a source fence with generated output plus an adjacent source comment. Treat the comment as the editable source of truth.

SVG form:

````markdown
![uisketch:orders](assets/orders.svg)
<!-- uisketch:source id="orders" format="svg"
```uisketch:svg
browser:
  id: orders
  title: Orders
```
-->
````

ASCII form:

````markdown
```text
+----------+
| Orders   |
+----------+
```
<!-- uisketch:source id="orders" format="txt"
```uisketch:txt
browser:
  id: orders
  title: Orders
```
-->
````

## Migration Notes

Older examples may show a `yaml` fence under `## Layout`. Current tools intentionally ignore bare `yaml` fences for rendering. When migrating, change the fence info string to `uisketch` after confirming the body is a UI Layout DSL root node.

## Validation Checklist

- Frontmatter exists for `.uisketch.md` and has `type: uisketch`.
- At least one renderable source fence or generated source comment exists.
- Each source body is one YAML document whose root is exactly one component key.
- Preferred root components are `browser`, `window`, `mobile`, `dialog`, or `menu`.
- `tabs.labels` selects exactly one active tab with a single-item list such as `[Settings]`.
- `hstack.widths` and `vstack.heights` match direct child count and use percentages plus `$`.
- `splitter.sizes` has two entries and uses percentages plus `$`.
- Element IDs are unique within the sketch once resolved by the root namespace.
- `prompt` values are strings, preserved during edits, and used only as non-rendered AI-agent guidance.
