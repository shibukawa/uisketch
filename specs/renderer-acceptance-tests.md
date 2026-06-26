---
id: "req.renderer-acceptance-tests"
type: "requirement"
title: "Renderer Acceptance Tests"
aliases:
  - "Renderer Golden Tests"
  - "ASCII Renderer Acceptance Tests"
tags:
  - "go"
  - "renderer"
  - "test"
facts:
  lifecycle.status: "draft"
---

# Renderer Acceptance Tests

## Summary

Renderer acceptance tests verify that real `uisketch` source files render to deterministic ASCII-art UI output. The tests use golden files under `testdata/renderer/` so component combinations can be reviewed as examples, not only as unit-level assertions.

## Directory Layout

Renderer acceptance cases must use this directory structure:

```text
testdata/
  renderer/
    001_browser_dashboard/
      input.md
      output.txt
    002_desktop_master_detail/
      input.md
      output.txt
```

Each `[case]` directory name must match:

```text
NNN_name
```

Rules:

- `NNN` is a three-digit sequence number starting at `001`.
- `name` is a short lowercase snake_case description.
- Sequence numbers should be stable once committed so failures and reviews can refer to cases by number.
- New cases should normally append a new number instead of renumbering existing cases.

## Case Files

Each case directory must contain:

| File | Requirement |
| --- | --- |
| `input.md` | A complete Markdown source file containing an explicit `uisketch` source fence. |
| `output.txt` | Expected ASCII-art renderer output for `input.md`, including the final trailing newline. |

The `input.md` file should be a realistic layout sketch, not an isolated component sample, unless the component is hard to place in a natural screen. Cases should combine components into common UI patterns such as browser dashboards, desktop master-detail views, dialogs, menus, tabbed forms, mobile-like settings, and data review screens.

The `output.txt` file should preserve the intended UI composition. It should show containers, controls, and major content regions with box-drawing characters when practical, and should reflect structural positioning from layout components such as `vstack`, `hstack`, `hstack.widths`, and `vstack.heights`. A renderer that only emits one text row per component does not satisfy these acceptance tests.

## Component Coverage

Acceptance cases must cover every component in the supported renderer component set at least once.

The initial renderer acceptance set should include these components:

- Root and structural components: `browser`, `window`, `mobile`, `dialog`, `menu`, `vstack`, `hstack`, `grid`, `splitter`, `tabs`, `section`.
- Root chrome and action-row attributes: `window.menu`, `mobile.menu`, and `dialog.buttons`.
- Action and navigation components: `button`, `button.badge`, `button.anchor`, emoji-labeled `button`, and `toggle`.
- Content components and annotation attributes: `label`, `image`, `custom`, and `note` on at least one visible component.
- Input components: `input`, `textarea`, `combobox`, `checkbox`, `radio`, `slider`.
- Data display components: `table`, `list`, `tree`, `calendar`, `badge`.

When a new component becomes part of the supported renderer set, the acceptance tests must add or update at least one case that exercises it before the implementation is considered complete.

## Golden Test Behavior

`go test ./...` must include a test that:

1. Walks `testdata/renderer/`.
2. Selects directories whose base name matches `NNN_name`.
3. Reads `input.md`.
4. Parses it as a `uisketch` file.
5. Converts the parsed layout to the normalized sketch model.
6. Renders ASCII output.
7. Compares the rendered output plus trailing newline with `output.txt`.

The test must fail when:

- A case directory has no `input.md`.
- A case directory has no `output.txt`.
- `input.md` cannot be parsed.
- Rendering returns output different from `output.txt`.
- A case directory name does not match `NNN_name`.
- Sequence numbers are duplicated or not strictly increasing.

The test should report the case name in failure messages so maintainers can update or inspect the relevant fixture quickly.

## Golden Update Policy

Golden files should not be updated silently by default. If an update flag is added later, it must be explicit, for example:

```bash
go test ./... -update
```

Until such a flag exists, maintainers should inspect renderer changes and update `output.txt` intentionally.

## Related Documents

- [UI Sketch File Format](ui-sketch-file-format.md)
- [UI Component Catalog](ui-component-catalog.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [Initial Implementation Slice](initial-implementation-slice.md)

## Native-Language Summary

Renderer の受け入れテストは `testdata/renderer/NNN_name/` に配置する。各ケースは `input.md` に明示的な `uisketch` source fence を含む Markdown 入力、`output.txt` に期待する ASCII 出力を持つ。`go test` はこのディレクトリを巡回し、入力を parse、sketch model に変換、ASCII render して `output.txt` と完全一致比較する。ケース名は 3 桁通し番号と snake_case 名で固定し、対応済みコンポーネントは少なくとも 1 回ずつ現実的なレイアウト内で使う。
