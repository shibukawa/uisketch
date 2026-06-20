---
id: "req.initial-implementation-slice"
type: "requirement"
title: "Initial Implementation Slice"
aliases:
  - "v0.1 Scope"
tags:
  - "go"
  - "implementation"
facts:
  lifecycle.status: "draft"
---

# Initial Implementation Slice

## Summary

The first runnable slice should prove the end-to-end path from concept metadata and a `uisketch` source file to SVG and ASCII-art sketch output.

## In Scope

- Go module for parsing concept metadata.
- Go module for parsing Markdown-like `uisketch` files and their explicit `uisketch` source fences.
- Basic UI component catalog support for the reduced v0.1 component set.
- Internal normalized sketch model.
- SVG renderer.
- ASCII renderer.
- Basic validation for missing concept IDs, missing layout source, invalid `uisketch` frontmatter, missing renderable `uisketch` source fence, unresolved action references, and malformed YAML.
- CLI entry point for rendering one selected `uisketch` source fence.

## Out of Scope

- Figma import.
- Dedicated visual editor. The browser editor is specified separately in [Web Visual Editor](web-visual-editor.md) and should follow after the Go parser, validator, serializer, and renderer are reusable.
- React code generation.
- HTML generation.
- PPTX generation.
- Full requirement, DFD, and vocabulary repository integration.
- Pixel-perfect rendering.

## Suggested Package Boundaries

| Package | Responsibility |
| --- | --- |
| `concept` | Screen, action, and UI flow metadata types. |
| `layout` | `uisketch` file parser, YAML layout parser, and layout node model. |
| `sketch` | Renderer-independent normalized sketch model. |
| `render/svg` | SVG output. |
| `render/ascii` | ASCII-art output. |
| `validate` | Cross-concept and source validation. |
| `cmd/uisketch` | CLI wrapper. |

## Minimum CLI Behavior

```bash
uisketch render --format svg --output equipment-list.svg ui/equipment-list.uisketch.md
uisketch render --format ascii ui/equipment-list.uisketch.md
```

The CLI may accept project-root configuration later. The first slice takes a positional input path or standard input as defined in [Uisketch CLI](uisketch-cli.md).

## Acceptance Criteria

- A sample `.uisketch.md` file using `browser`, `vstack`, `hstack`, and `table` renders to deterministic SVG.
- The same sample renders to deterministic ASCII art.
- Missing layout source produces a validation error.
- Unknown action reference produces a validation warning or error according to strictness mode.
- Unit tests cover parsing, validation, and both renderers.
- ASCII renderer acceptance tests use `testdata/renderer/NNN_name/input.md` and `output.txt` golden files as defined in [Renderer Acceptance Tests](renderer-acceptance-tests.md).

## Related Documents

- [UI Sketch Library Overview](ui-sketch-library-overview.md)
- [UI Concept Model](ui-concept-model.md)
- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Layout DSL](ui-layout-dsl.md)
- [UI Component Catalog](ui-component-catalog.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [Renderer Acceptance Tests](renderer-acceptance-tests.md)
- [Web Visual Editor](web-visual-editor.md)

## Native-Language Summary

最初の実装では、Concept メタデータと `.uisketch.md` 内の明示的な `uisketch` source fence から SVG と ASCII を生成する最小の一気通貫を作る。UI コンポーネント catalog は v0.1 の基本部品から対応する。Figma 連携や React 生成ではなく、screen / action / ui-flow と renderer の成立を優先する。
