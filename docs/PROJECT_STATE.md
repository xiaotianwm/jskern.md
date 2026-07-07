# Project State

## Current Phase

Readable Markdown document MVP loop.

## Done

- Wails React + TypeScript project initialized.
- Product identity set to `JS Kern.md` / `jskernmd`.
- Wails window configured as frameless.
- Windows Mica-capable and macOS transparent titlebar options added where Wails supports them.
- Go bootstrap API added.
- Go-managed `zh-CN` and `en` locale dictionaries added.
- Go directory-tree workspace scan API added.
- Frontend replaced default demo with a desktop reader shell: titlebar, toolbar, workspace tree, reader surface, outline panel.
- Desktop browser-behavior guards added.
- Project memory documents added.
- Frontend toolchain upgraded from the old Wails template versions to current Vite / TypeScript / React plugin dependencies.
- npm audit now reports 0 vulnerabilities.
- Git repository initialized on `main`.
- Public GitHub repository created: `https://github.com/xiaotianwm/jskern.md`.
- Added goldmark-based Markdown rendering.
- Added bluemonday HTML sanitization for rendered Markdown.
- Added `OpenDocument(path)` Go API with current-workspace path validation.
- Added symlink-aware real-path validation so workspace links cannot open files outside the workspace root.
- Added document heading extraction and outline data.
- Added Go tests for Markdown rendering, sanitization, heading IDs, and workspace boundary rejection.
- Frontend directory-tree file clicks now open and render Markdown documents.
- Reader surface now shows document title, path, rendered Markdown body, and outline navigation.

## Next

- Add local image and relative link resolver.
- Add Shiki code highlighting.
- Add theme and language switching APIs.
- Add directory expand/collapse state.
- Add user-visible error display for failed document loads.

## Known Issues

- Directory tree currently scans eagerly with a depth cap of 8 and skips common heavy folders.
- Directory tree has no expand/collapse state yet.
- Local relative images and links are not resolved through a Go asset endpoint yet.
- Code blocks render as plain code until Shiki is integrated.
- Failed document opens currently keep the previous document without a visible error panel.
- Window controls need visual/manual UX verification even though the exe starts successfully.

## Validation

- `go test ./...` passed.
- `npm.cmd run build` passed.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds.
- Initial code pushed to GitHub remote `origin`.
- Latest validation after Markdown reading loop:
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds.
