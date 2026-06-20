---
id: "req.ui-sketch-library-overview"
type: "requirement"
title: "UI Sketch Library Overview"
aliases:
  - "UI Sketch"
  - "Sketch Wireframe Library"
tags:
  - "ui"
  - "go"
  - "wireframe"
facts:
  lifecycle.status: "draft"
---

# UI Sketch Library Overview

## Summary

The library provides a Go implementation for managing UI as structured concepts and rendering those concepts as low-fidelity UI sketches. The purpose is to support discussion of structure, flow, and responsibility, not to replace Figma or produce production-quality visual design.

## Goals

- Represent UI as first-class concepts such as [Screen Concept](ui-concept-model.md), [Action Concept](ui-concept-model.md), and [UI Flow Concept](ui-flow-model.md).
- Generate sketch-style wireframes from definition files.
- Support at least SVG output and ASCII-art output.
- Support a browser-based visual editor that writes the same canonical UI layout source rather than an editor-private scene graph.
- Support a VSCode extension that previews embedded sketches without requiring a file-rewriting build step.
- Keep design review focused on screens, transitions, operations, input fields, permissions, and business flow.
- Avoid encouraging discussion about colors, spacing, fonts, icon style, and pixel-level layout.

## Non-Goals

- The library is not a Figma alternative.
- The library does not ingest Figma as a primary source of truth.
- The first version does not need high-fidelity visual rendering, responsive production layouts, or design-system styling.
- The first version does not need React, HTML, Markdown, or PPTX output, though the model should not prevent those render targets later.

## Primary Users

- Product and business stakeholders who need to discuss workflow and responsibility.
- Engineers who need implementation-oriented UI structure before visual design is finalized.
- Analysts who need traceability between UI operations, requirements, vocabulary, and data-flow diagrams.

## Required Inputs

- Concept metadata files that declare screens, actions, and flows.
- UI sketch source files written as Markdown-like `.uisketch.md` documents with `type: uisketch` frontmatter and explicit `uisketch` source fences.
- Optional vocabulary references for labels.
- Optional references to requirements and DFD operations for validation.

## Required Outputs

- Sketch wireframe as SVG using a draw.io-like sketch style.
- Sketch wireframe as ASCII art.
- Validation findings that explain missing or inconsistent relationships.
- Browser editor preview output that is generated through the same Go parser, validator, and renderer logic when the editor is available.
- VSCode preview output for Markdown and `.uisketch.md` sources when the extension is available.
- Human-readable documentation generated from the same concepts in a future version.

## Related Documents

- [UI Concept Model](ui-concept-model.md)
- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Layout DSL](ui-layout-dsl.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [UI Validation Rules](ui-validation-rules.md)
- [Web Visual Editor](web-visual-editor.md)
- [VSCode Extension](vscode-extension.md)

## Native-Language Summary

この Go ライブラリは、UI を画像ではなく Screen / Action / UI Flow という Concept として管理し、`type: uisketch` の frontmatter と明示的な `uisketch` source fence を持つ Markdown 風定義ファイルから draw.io の sketch 風 SVG または ASCII のスケッチ風ワイヤーフレームを生成する。目的は見た目の完成度ではなく、画面遷移、操作、入力項目、権限、業務フローの議論を促進することである。
