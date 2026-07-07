# Changelog

## 2026-07-07

### Added

- Added `goldmark` Markdown parsing and GFM support in the Go backend.
- Added `bluemonday` sanitization for rendered Markdown HTML.
- Added `OpenDocument(path)` Wails API.
- Added current-workspace path boundary validation before opening a document.
- Added symlink-aware real-path validation so files linked from inside a workspace cannot resolve outside the workspace root.
- Added document model fields for path, filename, title, sanitized HTML, and heading outline.
- Added heading extraction from the goldmark AST, including generated heading IDs for outline navigation.
- Added tests for Markdown rendering, sanitization, preserved heading IDs, and rejecting documents outside the workspace.
- Connected the frontend directory tree so clicking a Markdown file opens it through Go.
- Replaced the reader placeholder with a real Markdown reading view.
- Added document title and selectable path display.
- Added right-side outline rendering and heading scroll navigation.

### Changed

- The reader shell now clears the selected document when a new workspace is opened.
- The Markdown body and document path explicitly allow text selection while the rest of the shell remains anti-web.

### Validation

- `go test ./...` passed.
- `npm.cmd run build` passed.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

### Notes

- `go get @latest` could not reach `proxy.golang.org` or GitHub from this environment, so dependencies were added from the local module cache: `goldmark v1.7.4` and `bluemonday v1.0.27`.

---

### Added

- Initialized the Wails React + TypeScript project for `JS Kern.md`.
- Added durable project memory files:
  - `AGENTS.md`
  - `docs/PRODUCT.md`
  - `docs/ARCHITECTURE.md`
  - `docs/CONSTRAINTS.md`
  - `docs/PROJECT_STATE.md`
  - `docs/DECISIONS.md`
  - `docs/CHANGELOG.md`
- Added `product.manifest.json` as the current product identity source.
- Added Go-managed locale files for `zh-CN` and `en`.
- Added Go bootstrap API returning product info, shell locale, and business locale.
- Added Go workspace directory-tree scanning API.
- Replaced the generated Wails demo UI with a desktop Markdown reader shell:
  - frameless custom titlebar
  - workspace toolbar
  - left directory tree panel
  - center reader surface
  - right outline panel
- Added frontend desktop guards for context menu, refresh, find, zoom, F12, dragstart, and Ctrl-wheel behavior.
- Initialized the local Git repository on `main`.
- Created and pushed the public GitHub repository: `https://github.com/xiaotianwm/jskern.md`.

### Changed

- Set Wails output filename to `jskernmd`.
- Set app display title to `JS Kern.md`.
- Removed the default Wails demo interaction from the main UI.
- Removed the generated web font usage from active styles and switched to system fonts.
- Upgraded the frontend development toolchain to current Vite, TypeScript, and React plugin packages after npm audit found vulnerabilities in the Wails template defaults.
- Updated TypeScript config to modern `moduleResolution: "Bundler"` so the upgraded toolchain builds cleanly.

### Constraints Captured

- The MVP must be directory-tree based.
- Wails is the only allowed desktop runtime.
- Electron is forbidden.
- Go owns filesystem access, Markdown parsing, durable state, and i18n.
- React only renders Go-provided data and short-lived interaction state.

### Validation

- `go test ./...` passed.
- `npm.cmd run build` passed.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- GitHub push to `origin/main` completed.
