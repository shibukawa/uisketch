---
id: "req.ui-concept-model"
type: "requirement"
title: "UI Concept Model"
aliases:
  - "Screen Action Flow Concepts"
tags:
  - "ui"
  - "concept"
facts:
  lifecycle.status: "draft"
---

# UI Concept Model

## Summary

UI artifacts are managed as concepts rather than images or views. The minimum concept set for the first implementation is `screen`, `action`, and `ui-flow`.

## Concept Types

| Concept | Example ID | Purpose |
| --- | --- | --- |
| Screen | `screen.equipment-list` | Defines a user-facing screen and its structural layout source. |
| Action | `action.create-alert` | Defines an operation a user can trigger or complete. |
| UI Flow | `flow.alert-investigation` | Defines a sequence or graph of screens and actions. |

## Screen Concept

A screen concept must include:

- Stable `id`.
- `type: screen`.
- Human-readable `title`.
- Source reference for its layout definition.
- Main purpose.
- Primary actions available on the screen.
- Related use cases or requirements when known.

Example:

```yaml
---
id: screen.equipment-list
type: screen
title: Equipment List
source:
  type: uisketch
  location: ui/equipment-list.uisketch.md
---
```

## Action Concept

An action concept must include:

- Stable `id`.
- `type: action`.
- Human-readable name.
- Triggering screens or flows.
- Expected permission or role requirement when known.
- Related requirement or business operation when known.

Actions are not merely button labels. They represent user-visible operations and must be reusable across screens and flows where the same business operation appears.

## UI Flow Concept

A UI flow concept must include:

- Stable `id`.
- `type: ui-flow`.
- Human-readable title.
- Ordered sequence or graph of screen and action references.
- Starting point.
- End states or outcomes.
- Related use case, requirement, or business workflow when known.

## Source Model

Concept documents can point to multiple source formats. The first implementation should support `uisketch` Markdown-like files with explicit `uisketch` source fences, as defined in [UI Sketch File Format](ui-sketch-file-format.md). Raw `ui-layout` YAML may be accepted as an import or migration convenience, and a dedicated indentation-based DSL may be added later.

Example:

```yaml
source:
  type: uisketch
  location: ui/equipment-list.uisketch.md
```

Figma may be linked as reference material but must not become the canonical source for the concept.

## Maturity Rules

Iteration Ready requires:

- Purpose.
- Major operations.
- Related use case or requirement.

Iteration Ready does not require:

- Detailed layout measurements.
- Color choices.
- Icon selections.
- Copy refinement.
- Font choices.

## Related Documents

- [UI Sketch Library Overview](ui-sketch-library-overview.md)
- [UI Flow Model](ui-flow-model.md)
- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Layout DSL](ui-layout-dsl.md)
- [UI Validation Rules](ui-validation-rules.md)

## Native-Language Summary

UI は画像や View ではなく Concept として管理する。最初の実装対象は screen、action、ui-flow の 3 種類であり、画面構造、操作、業務フローを議論できる粒度を優先する。screen concept は `source.type: uisketch` で `.uisketch.md` を参照し、そのファイル内の明示的な `uisketch` source fence をレンダリング対象とする。
