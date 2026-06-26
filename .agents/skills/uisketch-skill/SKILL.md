---
name: uisketch-skill
description: Understand, author, extract, validate, revise, and use uisketch UI sketch Markdown files with embedded semantic UI YAML. Use when Codex needs to create or edit `.uisketch.md` files, interpret uisketch fences, explain UI layout DSL intent, generate renderable sketches, validate sketch definitions, or derive implementation briefs from uisketch structure, including AI-agent guidance stored in `prompt` fields.
---

# Uisketch Skill

Use this skill to work with uisketch source: Markdown documents containing semantic UI layout YAML in explicit `uisketch` fences.

## Workflow

1. Identify the input kind:
   - Canonical sketch file: `.uisketch.md` with YAML frontmatter `type: uisketch`.
   - Embedded sketch: ordinary Markdown containing one or more `uisketch`, `uisketch:svg`, `uisketch:txt`, `uisketch:text`, or `uisketch:ascii` fences.
   - Generated output: Markdown image or `text` fence followed by `<!-- uisketch:source ... -->`.
2. Extract definitions before reasoning about the UI when the document may contain multiple sketches:
   - Run `scripts/extract_uisketch.py <file>` for a YAML summary.
   - Add `--body-only` when another tool needs raw layout YAML documents.
   - Add `--include-yaml-fences` only for migration from older examples; do not generate new bare `yaml` layout fences.
3. Validate before handing source to a renderer or code generator:
   - Run `scripts/validate_uisketch.py <file> --schema references/uisketch-layout.schema.json`.
   - The validator uses bundled `pyyaml-pure` for YAML parsing and bundled `fastjsonschema` for schema validation.
   - Treat errors as blocking. Treat warnings as authoring guidance unless the user asks for a strict pass.
4. Generate or revise source using semantic components, not visual drawing primitives.
5. If the task is code generation, treat the sketch as an implementation brief. Extract root surface, major regions, actions, inputs, data displays, navigation anchors, IDs, `data`, `note`, `review`, and every `prompt` field before writing code.

## Authoring Rules

- Prefer `.uisketch.md` for persisted source.
- Put document metadata in frontmatter. Keep the renderable hierarchy inside explicit `uisketch` fences.
- Use `uisketch` fences for new source. Do not rely on headings such as `## Layout` or `## レイアウト` to make a `yaml` fence renderable.
- Use root surfaces `browser`, `window`, `mobile`, `dialog`, or `menu` for renderable screens and states.
- Use `children`, not `child`, in new root and container definitions.
- Put visible chrome titles on `browser.title`, `window.title`, `mobile.title`, or `dialog.title`; frontmatter `title` is document metadata.
- Use `anchor` for navigation. Accept legacy `to` only when reading old files.
- Use `prompt` for non-rendered AI-agent instructions and `data` for structured non-rendered metadata.
- Preserve `prompt` and `data` when revising source. Do not turn `prompt` into visible UI copy, helper text, or comments in generated code unless the user asks.
- Avoid CSS-like properties such as pixel width, margin, padding, color, font, breakpoint, or arbitrary coordinates.

## References

- Read `references/format.md` when creating or editing `.uisketch.md` files, generated output comments, or multiple sketches.
- Read `references/component-types.md` when choosing component types or translating product intent into layout YAML.
- Use `references/uisketch-layout.schema.json` for schema-based validation and editor integration.
- Read `README.md` when checking bundled library provenance, licenses, runtime dependency behavior, or update procedure.

## Scripts

- `scripts/extract_uisketch.py`: Extract frontmatter and every renderable uisketch definition from Markdown in document order. Supports multiple sketches and generated source comments.
- `scripts/validate_uisketch.py`: Validate extracted definitions for frontmatter, fence selection, root shape, common DSL rules, duplicate IDs, and schema compatibility using bundled YAML parsing plus bundled `fastjsonschema`.

## Code Generation Context

When using uisketch as input for app or feature implementation, convert the sketch into a concise implementation brief:

- Root: surface type, ID, title, platform implication.
- Structure: sections, stacks, grids, splitters, tabs, menus, dialogs.
- Operations: buttons, action IDs, anchors, dialog buttons.
- Inputs: labels, hints, expected data type implied by hint text.
- Data display: tables, lists, trees, calendars, badges, image placeholders.
- Traceability: element IDs, `data` metadata, `prompt` instructions, `note` and `review` annotations.

Preserve sketch semantics in generated code, but do not copy low-fidelity renderer choices as production visual design unless the user explicitly asks for a wireframe-like UI.

`prompt` is the main bridge from low-fidelity sketch to implementation. Read it as direct guidance to the coding agent, for example "load select options from the external equipment API" or "keep this filter state in the URL." Apply that guidance when designing data flow, API calls, state, validation, and component behavior. Keep visible UI grounded in `label`, `title`, `hint`, and component semantics.
