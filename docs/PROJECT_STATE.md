# Project State

## Current Phase

Initial Wails skeleton and durable project memory.

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

## Next

- Add `goldmark` Markdown parsing.
- Add sanitization policy.
- Add document selection and rendered Markdown view.
- Add heading extraction for the outline panel.
- Add local image and relative link resolver.
- Add Shiki code highlighting.
- Add theme and language switching APIs.

## Known Issues

- Directory tree currently scans eagerly with a depth cap of 8 and skips common heavy folders.
- Directory tree has no expand/collapse state yet.
- Reader area is a placeholder until Markdown rendering is implemented.
- Window controls need visual/manual UX verification even though the exe starts successfully.

## Validation

- `go test ./...` passed.
- `npm.cmd run build` passed.
- `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
- `wails build` passed and produced `build/bin/jskernmd.exe`.
- Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds.
- Initial code pushed to GitHub remote `origin`.
