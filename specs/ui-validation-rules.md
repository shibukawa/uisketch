---
id: "req.ui-validation-rules"
type: "requirement"
title: "UI Validation Rules"
aliases:
  - "UI Traceability Validation"
tags:
  - "ui"
  - "validation"
facts:
  lifecycle.status: "draft"
---

# UI Validation Rules

## Summary

Validation checks consistency between UI concepts, requirements, vocabulary, and data-flow definitions. UI should not be validated only as isolated layout.

## Validation Scope

Validation should check:

- Concept metadata shape.
- `uisketch` frontmatter shape.
- Presence of explicit `uisketch` source fences or generated `uisketch:source` comments in `uisketch` source files.
- Layout source parseability.
- Screen-to-action consistency.
- Flow-to-screen consistency.
- Flow-to-action consistency.
- Vocabulary reference resolution.
- Requirement traceability.
- DFD operation traceability when DFD specifications exist.
- Root element ID uniqueness when root IDs are present.
- Child element ID uniqueness after resolving IDs into the root namespace.
- Unresolved `anchor` navigation targets.

## Required Findings

| Finding | Example |
| --- | --- |
| DFD operation missing from UI | A DFD includes `Update Equipment Status`, but no UI action references it. |
| UI action missing requirement | `action.create-alert` exists, but no requirement or use case references it. |
| Flow references missing screen | `flow.alert-investigation` references `screen.equipment-detail`, but the screen concept is absent. |
| Label missing vocabulary | A business term appears as a literal label where a vocabulary reference is expected. |
| Permission gap | A destructive action exists without role or permission metadata. |
| Duplicate root ID | Two root layouts both declare `id: equipment-list`. |
| Duplicate resolved child ID | `button id: refresh` appears twice under root `equipment-list`, producing duplicate `equipment-list.refresh`. |
| Missing referenced root ID | A root has no `id` but is referenced by a flow or interactive navigation anchor. |
| Unresolved navigation target | `button anchor: alert-create-dialog` references no known root or element ID. |
| Invalid uisketch source | A source file lacks `type: uisketch`, has no explicit `uisketch` source fence, or has no generated `uisketch:source` comment to rebuild from. |
| Invalid prompt property | A layout element uses `prompt` with a non-string value, such as an object or list. |
| Unsupported layout property | A layout element contains an unknown property key outside the supported DSL keys; project-specific structured values should be placed under `data`. |
| Unsupported component type | A layout uses a component key that is not in the baseline catalog. |
| Invalid note property | A layout element uses `note` with a non-string or empty value. |
| Invalid root menu property | A root uses `menu` outside `window` or `mobile`, or `menu` is not a list of non-empty strings. |
| Invalid dialog buttons property | A `dialog.buttons` value is not a list of layout nodes or contains unsupported non-button action-row content in a tool that only supports button actions. |
| Invalid tabs shape | A `tabs` node has no `labels`, has no selected label, has more than one selected label, or uses `children` as multiple inactive tab bodies instead of the one active tab body. |
| Invalid proportional sizes | An `hstack.widths`, `vstack.heights`, or `splitter.sizes` array has a different length than its target children, contains values other than numbers or `$`, uses numeric percentages that exceed `100` when `$` is present, or omits `$` while numeric percentages do not total `100`. |

The legacy `to` navigation field may be parsed for migration, but validators should prefer `anchor` in findings and should warn when newly authored or rewritten files still use `to`.

The `prompt` property is allowed on any layout element as AI-agent guidance. Validators and schema checks should recognize the key and require a string value, including YAML block scalar strings. They should not parse, execute, lint, or semantically validate the prompt text. Renderers should ignore it for visible output.

The `note` property is allowed on any layout element as visible annotation text. Validators and schema checks should recognize the key and require a non-empty string value, including YAML block scalar strings. A `note` key must not be interpreted as a component type.

Navigation validation should treat `button.anchor` as the baseline clickable navigation shape. `label` nodes are non-interactive display text and should not be used as clickable navigation targets. Badge-bearing buttons should use the `badge` property on `button`.

The `menu` property is allowed only on `window` and `mobile` root nodes. It must be a list of non-empty strings and must not be interpreted as a `menu` child component or as any separate menu-bar component. Validators should reject `menu` on `browser`, `dialog`, `menu`, or non-root layout nodes unless a future spec extends that surface.

The `buttons` property is allowed on `dialog` roots. It must contain ordinary layout node objects, normally `button`. Validators may permit any supported node shape for compatibility with the DSL, but visual editors may restrict authoring to `button` nodes for the fixed bottom action row.

For `tabs.labels`, validators should treat a string item as inactive and a single-item list item as selected. Any other nested value in `labels` is invalid. New or rewritten `tabs` nodes should have exactly one selected label and one ordinary `children` subtree for the active body.

For proportional sizes, validators should allow `hstack.widths` only on `hstack`, `vstack.heights` only on `vstack`, and `splitter.sizes` only on `splitter`. Numeric entries are percentages, not pixels. `$` entries are resolved by dividing the remaining percentage equally among all `$` entries. Parsers may accept quoted `"*"` as a legacy spelling, but writers should emit `$`.

`splitter` should have exactly two direct children. When `splitter.sizes` is present, it should contain exactly two entries.

## Vocabulary Principle

Business terms, UI terms, and system terms should be the same canonical vocabulary term whenever practical.

Example:

```yaml
button:
  vocabulary: vocab.contract-holder
```

Expected labels:

| Locale | Label |
| --- | --- |
| English | Contract Holder |
| Japanese | 契約者 |

## Severity Model

| Severity | Meaning |
| --- | --- |
| Error | The concept or layout cannot be parsed or references a missing required concept. |
| Warning | The model is usable but traceability or maturity is incomplete. |
| Info | Advisory finding for future refinement. |

## Related Documents

- [UI Concept Model](ui-concept-model.md)
- [UI Flow Model](ui-flow-model.md)
- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Layout DSL](ui-layout-dsl.md)

## Native-Language Summary

Validation は UI 単体ではなく、Requirement、DFD、Vocabulary との整合を見る。加えて `.uisketch.md` の frontmatter が `type: uisketch` を持つこと、明示的な `uisketch` fence または生成済み `uisketch:source` comment があることも検証する。`## Layout` や `## レイアウト` 見出し配下の通常 `yaml` fence は render/markdown の入力として扱わない。button の `anchor` は root ID または解決済み element ID として存在確認する。クリック可能な遷移は button で表し、label は非インタラクティブな表示テキストとして扱う。badge 付き操作は button.badge で表す。window/mobile の menu は root chrome 属性として非空文字列リストであることを検証し、component としては扱わない。dialog.buttons は dialog root の下部固定 action row として layout node list であることを検証する。note は任意要素に置ける表示注釈としてキーと非空文字列型を検証し、component type としては扱わない。prompt は任意要素に置ける AI エージェント向けコメントとしてキーと文字列型だけを検証し、内容は意味解釈しない。DFD にある操作が UI にない、UI にある操作が Requirement にない、Vocabulary と UI ラベルがずれている、といった問題を検知する。
