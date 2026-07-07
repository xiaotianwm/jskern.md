# Changelog

## 2026-07-07

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
