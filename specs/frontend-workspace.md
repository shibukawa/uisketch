---
id: "req.frontend-workspace"
type: "requirement"
title: "Frontend Workspace Architecture"
aliases:
  - "NPM Workspace"
  - "Web Frontend Packages"
  - "GitHub Pages And VSCode Frontend"
tags:
  - "frontend"
  - "npm"
  - "vite"
  - "vscode"
  - "wasm"
facts:
  lifecycle.status: "draft"
---

# Frontend Workspace Architecture

## Summary

The UI Sketch frontend should be reorganized as an npm workspace with two user-facing entry points:

- A static GitHub Pages web editor defined by [Web Visual Editor](web-visual-editor.md).
- A VSCode extension defined by [VSCode Extension](vscode-extension.md).

Both entry points should share TypeScript packages for UI Layout DSL types, schema access, source-region parsing, editor document operations, and Go Wasm integration. The Go implementation remains the source of truth for parsing, validation, normalization, SVG rendering, ASCII/text rendering, and canonical serialization.

## Workspace Goals

The workspace should make the browser editor and VSCode extension feel like two shells over the same UI Sketch authoring model, not two separate products.

Required goals:

- Build both frontends through npm workspace scripts from the repository root.
- Keep entry-point specific code in separate packages so GitHub Pages and VSCode packaging do not leak into each other.
- Put shared TypeScript models and adapters in internal workspace packages.
- Build the Go Wasm module once through a deterministic script and consume it from both frontends.
- Use Vite for browser-targeted bundles where practical.
- Keep the Go CLI/library usable without requiring npm.
- Avoid duplicating layout validation or rendering rules in TypeScript.

## Recommended Package Layout

The first implementation should use this structure unless implementation constraints reveal a better local pattern:

```text
package.json
packages/
  web/
    package.json
    src/
  vscode-extension/
    package.json
    src/
  core/
    package.json
    src/
  wasm/
    package.json
    src/
  schema/
    package.json
    src/
```

Recommended package responsibilities:

| Package | Responsibility |
| --- | --- |
| `@uisketch/web` | GitHub Pages web editor shell, Vite app, routes, browser-only storage, and visual editor chrome. |
| `@uisketch/vscode-extension` | VSCode extension activation, commands, Markdown preview contribution, diagnostics, webviews, and package metadata. |
| `@uisketch/core` | Shared TypeScript editor document types, source-region identities, view state types, command result shapes, and utility functions that contain no browser or VSCode dependency. |
| `@uisketch/wasm` | Typed wrapper around the Go Wasm bridge, asset loading helpers, result/error normalization, and runtime capability checks. |
| `@uisketch/schema` | Packaged UI Layout DSL schema, schema version metadata, completion snippets, and schema association helpers. |

The package names are internal defaults. They may be renamed before publish, but responsibilities should remain separated.

## Root Package Scripts

The root `package.json` should be a private workspace manifest.

Required root scripts:

| Script | Behavior |
| --- | --- |
| `npm run build` | Build Go Wasm and all workspace packages. |
| `npm run build:web` | Build only the GitHub Pages web app and its shared dependencies. |
| `npm run build:vscode` | Build only the VSCode extension and its shared dependencies. |
| `npm run build:wasm` | Compile `cmd/uisketch-wasm` into the workspace-consumable Wasm artifact and copy `wasm_exec.js` as needed. |
| `npm run dev:web` | Start the Vite dev server for the web editor. |
| `npm run typecheck` | Type-check every TypeScript package. |
| `npm run test` | Run frontend unit tests that do not require VSCode integration. |
| `npm run package:vscode` | Produce a VSCode extension package artifact. |

Workspace scripts should call project-local scripts such as `scripts/build-web-wasm.sh` when that is the established Go build path. A frontend build must not silently use a stale Wasm artifact.

## Go Wasm Integration

The shared Wasm package should expose coarse editor operations rather than raw Go internals.

Required operation families:

- Load and parse canonical `.uisketch.md` source.
- Load and parse focused YAML source fence bodies.
- Validate current editor document state.
- Insert, move, update, duplicate, wrap, and delete layout components.
- Serialize normalized YAML.
- Serialize canonical `.uisketch.md`.
- Render SVG preview.
- Render ASCII/text output.
- Encode and decode share payloads when Go owns the canonical payload implementation.

The TypeScript wrapper must provide stable Promise-based APIs with typed success and error results. The wrapper should normalize Go panic or bridge errors into explicit fatal error results so UI packages can show an actionable state instead of crashing.

Wasm artifacts are generated build outputs and should not be committed to source control. Local workspace builds and GitHub Actions builds must regenerate them from Go sources. `.gitignore` should exclude generated Wasm artifacts, generated JavaScript glue, and package-local copied Wasm outputs when those files live under build or distribution directories.

GitHub Actions should build Wasm after commit as part of frontend verification and deployment packaging. The repository source should contain the scripts needed to build Wasm, not the generated `.wasm` artifact itself.

## Existing Web Prototype Policy

The current `web/` implementation is a prototype and is not a maintenance target for the workspace migration.

Migration policy:

- Do not preserve the current `web/app.js`, `web/wasm_bridge.js`, checked-in `web/uisketch.wasm`, or prototype asset layout as implementation constraints.
- Preserve `web/webeditor.md` as the UI design reference for the first GitHub Pages screen.
- Treat any useful visual details in prototype SVGs or static assets as optional reference material, not canonical behavior.
- New browser implementation should live in the npm workspace web package.
- The old prototype may be deleted or archived after the workspace web package reaches equivalent first-screen coverage.

## GitHub Pages Entry Point

The GitHub Pages entry point is the browser app package.

Build requirements:

- Use Vite or equivalent static bundling that emits a deployable `dist` directory.
- Support a configurable base path for repository Pages, custom domains, and local preview.
- Load Wasm, JavaScript, CSS, schema, samples, and generated assets through relative URLs or the configured base path.
- Use hash routes or fragment parameters so refresh works without server rewrites.
- Preserve `#s=<encoded-payload>` share URLs.
- Include the screen shape specified in `web/webeditor.md` as the first visual target for the browser editor shell.
- Target current Chrome and Safari for the first supported browser set. Firefox compatibility is desirable but not required for the first slice.

The browser package owns localStorage, import/export, drag-and-drop UI, canvas/editor layout, and GitHub Pages deployment concerns. It should call shared packages for document semantics, source parsing, schema data, and Wasm operations.

## VSCode Extension Entry Point

The VSCode extension entry point is a separate workspace package.

Build requirements:

- Build the extension host bundle separately from any webview bundle.
- Package webview assets, schema assets, and Wasm assets in extension-safe locations.
- Use VSCode APIs only inside the extension package.
- Use the shared schema package for YAML completion and diagnostics.
- Use the shared Wasm package for rendering and validation when the selected implementation mode is Wasm.
- Keep any fallback CLI or language-server integration behind the same internal renderer/validator interface used by the extension.

The extension package should not depend on the GitHub Pages app. Shared UI elements may move to a future package only when both entry points genuinely need them.

## Shared Source Region Model

The workspace should share one TypeScript model for recognized source regions so browser import/export, VSCode preview, and Markdown bake/rebuild behavior use the same vocabulary.

Required fields:

- Source document URI or browser-local document identity.
- Region kind: `uisketch` fence, output-targeted `uisketch:*` fence, generated `uisketch:source` comment, focused YAML fence body, or complete `.uisketch.md` document.
- Document-local ordinal.
- Optional explicit root ID.
- Optional generated stable ID.
- Output target: SVG, ASCII/text, raw YAML, or source.
- Text range when the host editor can provide one.
- Source text and last parsed document version.

This model is host-neutral. Browser selection state and VSCode editor decorations should be stored outside the canonical source region object.

## Source Range Mapping

The Go parser and Wasm adapter should expose source range information where practical so host surfaces can connect YAML text, tree nodes, canvas selection, diagnostics, and preview output.

Required mapping concepts:

- Source range for each recognized `uisketch` source region.
- Source range for each parsed layout element when the YAML parser can provide it.
- Stable element path derived from root and child indexes when an explicit `id` is missing.
- Optional source range for individual properties such as `id`, `label`, `action`, `anchor`, `children`, `widths`, and `heights`.
- Mapping from validation findings to source range, element ID, element path, or source region identity when available.

Source range mapping should be best-effort. A missing range must not block rendering or visual editing, but surfaces should use ranges when present for source highlighting, tree navigation, and VSCode diagnostics.

## Schema Maintenance

The UI Layout DSL schema may be manually maintained.

Manual maintenance is acceptable because the DSL supports authoring conveniences that are hard to derive mechanically, including scalar shorthand, string-or-object values, selected tab labels encoded as one-item lists, and component-specific property rules. Generated schema fragments may be introduced later, but generation must not make the schema less accurate for embedded Markdown editing or VSCode completions.

## Build And Test Expectations

The workspace should include tests at the boundaries where drift is likely:

- `@uisketch/wasm` tests for successful load, parse, validation, SVG render, ASCII render, and bridge error normalization.
- `@uisketch/core` tests for source-region identity and editor command result shape.
- `@uisketch/schema` tests that packaged schema can be loaded by both browser and VSCode paths.
- Web tests for share URL encode/decode and local draft restore behavior.
- Extension tests or fixtures for Markdown source-region detection and schema association.
- Golden parity tests proving the same YAML input produces equivalent SVG through CLI and Wasm for the covered first-slice component set.

Frontend tests should not replace Go parser, validator, renderer, and serializer tests. They should prove integration behavior around the shared boundary.

GitHub Actions should run at least:

- Go tests.
- Spec compiler validation.
- Wasm build.
- npm workspace install.
- TypeScript typecheck.
- Web app build.
- VSCode extension build or package dry run.

The GitHub Pages deploy job should publish generated static artifacts from CI output. It should not require committing generated `dist` or Wasm files.

## First Runnable Workspace Slice

In scope:

- Root private npm workspace manifest.
- Vite-based `@uisketch/web` package that can render the current visual editor shell from [Web Visual Editor](web-visual-editor.md).
- `@uisketch/vscode-extension` package scaffold with activation, recognized commands, and build output.
- Shared `@uisketch/core`, `@uisketch/wasm`, and `@uisketch/schema` packages.
- Root scripts for `build`, `build:web`, `build:vscode`, `build:wasm`, `dev:web`, `typecheck`, and `test`.
- Go Wasm build integrated into the frontend build.
- GitHub Actions workflow that regenerates Wasm and builds both frontend entry points.
- Static GitHub Pages output that can run without a backend.
- Extension package output that includes the schema and renderer integration path selected for the first VSCode slice.

Out of scope:

- Publishing npm packages publicly.
- Publishing the VSCode extension to the marketplace.
- Monorepo task runners beyond npm workspace scripts unless build time or dependency graph complexity requires them.
- Server-side persistence for the web editor.
- A shared visual component library before concrete duplication exists.
- Maintaining the old prototype `web/` implementation after the workspace web package exists.

## Acceptance Criteria

- Running `npm run build` from the repository root builds the Go Wasm artifact, shared packages, GitHub Pages app, and VSCode extension package without manually copying files.
- Running `npm run dev:web` starts a Vite development server for the browser editor.
- The browser build can be served from a repository subpath and still load Wasm and restore a `#s=` share URL.
- The VSCode extension build can load the packaged schema and selected renderer/validator integration path.
- GitHub Actions can regenerate Wasm and build the web app without committed Wasm artifacts.
- Shared TypeScript packages contain no direct dependency on browser DOM APIs or VSCode APIs unless that package is explicitly host-specific.
- The same source-region identity model is usable by browser import/export and VSCode Markdown preview workflows.
- The build fails when the Wasm artifact is missing or older than the Go sources it depends on, or the build script regenerates it deterministically.

## Open Decisions

- Decide the TypeScript framework for `@uisketch/web`.
- Decide whether the VSCode first slice uses Wasm in the extension/webview, a bundled CLI process, or a language server behind the shared renderer interface.
- Decide whether to use npm alone or add a monorepo task runner after the first slice.
- Decide the exact distribution path for GitHub Pages deployment artifacts.
- Decide whether `@uisketch/schema` should be generated from `schemas/uisketch-layout.schema.json` during build or copy that file directly.

## Related Documents

- [Web Visual Editor](web-visual-editor.md)
- [VSCode Extension](vscode-extension.md)
- [UI Layout DSL](ui-layout-dsl.md)
- [UI Validation Rules](ui-validation-rules.md)
- [Sketch Wireframe Renderer](sketch-wireframe-renderer.md)
- [Markdown Embedding Workflow](markdown-embedding-workflow.md)

## Native-Language Summary

フロントエンドは npm workspace のマルチパッケージ構成に再編する。利用者向け entry point は GitHub Pages 向け Web editor と VSCode extension の 2 つで、共通の TypeScript package と Go Wasm adapter を共有する。既存 `web/` は prototype として扱い、実装継承は不要だが `web/webeditor.md` は初期画面の UI 設計資料として保存する。Web は Vite の静的 app として `#s=` share URL、relative/base path、localStorage、drag-and-drop UI を担当する。VSCode extension は Markdown preview、schema 補完/diagnostics、bake/rebuild commands、webview/extension host bundle を担当する。Go 実装は parse、validation、normalize、SVG/ASCII render、canonical serialize の正であり、TypeScript は UI と host integration を担う。Wasm artifact は source control に入れず、local build と GitHub Actions で生成する。最初の実装では root `package.json` に npm workspace scripts を置き、`packages/web`、`packages/vscode-extension`、`packages/core`、`packages/wasm`、`packages/schema` を作り、`npm run build` で Wasm と両 entry point を一括ビルドできる状態を目指す。
