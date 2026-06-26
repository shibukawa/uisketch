# uisketch

`uisketch` is a Go + Web workspace for rendering and editing `.uisketch.md` layouts. It uses the same source for the web editor, local project editing, CLI rendering, and the Wails desktop app.

## Web Editor

The hosted web editor is available on GitHub Pages:

https://shibukawa.github.io/uisketch/

## Setup

```bash
npm install
go test ./...
```

## Common Commands

### Start the Web Editor in Development Mode

```bash
npm run dev:web
```

This starts the Vite development server so you can edit sketches in a browser. To check the built web editor instead, use:

```bash
npm run build:web
npm run preview:web
```

### Render with the CLI

```bash
go run ./cmd/uisketch render examples/sample.uisketch.md
go run ./cmd/uisketch render --format ascii examples/sample.uisketch.md
```

Standard input is also supported:

```bash
cat examples/sample.uisketch.md | go run ./cmd/uisketch render
```

### Regenerate Markdown

```bash
go run ./cmd/uisketch markdown --output out.md examples/input.md
go run ./cmd/uisketch markdown --overwrite examples/input.md
```

### Edit a Local Project

```bash
go run ./cmd/uisketch edit --project . --no-open
go run ./cmd/uisketch edit --project . path/to/file.uisketch.md
```

`edit` only reads and writes files inside the selected project root. If you pass a file name, that file is opened as the initial editor state.

### Build the Desktop App

The Wails app uses the same web UI and Go file service.

```bash
wails dev
```

To build the desktop app:

```bash
wails build
```

If the `wails` command is not in your `PATH`, you can run the equivalent build with:

```bash
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build -nosyncgomod -nopackage
```

The Wails build runs `scripts/build-wails-frontend.sh`, which copies the `packages/web/dist` output into `desktop_assets`.
To update only the embedded web assets locally, use `npm run build:desktop-assets`.

## Additional Checks

```bash
npm run test
npm run typecheck
npm run build
```

## Modes

- Browser mode: named save, load, and delete operations use `localStorage`.
- Local project mode: `uisketch edit` or the Wails app directly reads and writes project files.
- Share URLs are for browser mode. They are not used in local project mode.

## Related Specifications

- [UI Sketch File Format](specs/ui-sketch-file-format.md)
- [Uisketch CLI](specs/uisketch-cli.md)
- [Web Visual Editor](specs/web-visual-editor.md)

## Agent Skill

This repository also provides an agent-facing Codex skill at [.agents/skills/uisketch-skill](.agents/skills/uisketch-skill). The skill helps agents understand, author, validate, revise, and use `.uisketch.md` files as implementation briefs.

## License

This project is licensed under the GNU Affero General Public License v3.0 or later. See [LICENSE](LICENSE).
