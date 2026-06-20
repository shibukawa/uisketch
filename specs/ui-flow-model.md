---
id: "req.ui-flow-model"
type: "ui-flow"
title: "UI Flow Model"
aliases:
  - "Flow Concept"
tags:
  - "ui"
  - "flow"
facts:
  lifecycle.status: "draft"
---

# UI Flow Model

## Summary

UI flow is a first-class concept because the correctness of a UI often depends more on transitions and task sequence than on a single screen's layout.

## Flow Representation

A flow may be represented as a sequence when the path is linear:

```text
Login
  -> Dashboard
  -> Equipment List
  -> Equipment Detail
  -> Create Alert
```

A flow may be represented as a graph when it contains branches, loops, or role-specific paths.

## Flow Node Types

| Node Type | Purpose |
| --- | --- |
| `screen` | A screen concept that appears in the flow. |
| `action` | A user operation that advances or changes the flow. |
| `decision` | A branch point, usually based on user choice, validation, state, or permission. |
| `outcome` | A completed state or terminal result. |

## Flow Requirements

A UI flow must define:

- Start node.
- One or more screen references.
- Actions that cause transitions.
- End states or outcomes.
- Error, empty, or denied paths when they affect the business process.

## Validation Expectations

- Every screen referenced by a flow should exist as a screen concept.
- Every action referenced by a flow should exist as an action concept.
- Actions that claim to implement a requirement should link to that requirement.
- Permission-specific branches should reference known roles or authorization facts when available.

## Render Expectations

The first renderer may render flows as ASCII diagrams or simple SVG diagrams. Flow rendering should remain low-fidelity and structural.

## Related Documents

- [UI Concept Model](ui-concept-model.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [UI Validation Rules](ui-validation-rules.md)

## Native-Language Summary

UI Flow は Screen 単体よりも重要な Concept である。画面、操作、分岐、結果をつなぐことで、業務フローとして正しいかを議論できるようにする。
