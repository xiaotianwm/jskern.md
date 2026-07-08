# Changelog

## 2026-07-08

### Added

- Added Go-owned update checking against the official GitHub Releases feed for `xiaotianwm/jskern.md`.
- Added installer asset filtering so update prompts only accept canonical `JSKernMD-Setup-<version>-x64.exe` release assets.
- Added Go-owned update installer download into AppData `temp/update/`.
- Added SHA256 verification for downloaded update installers when release metadata provides a digest.
- Added `DismissUpdate(version)` persistence through AppData `ignored_update_version`.
- Added `OpenDownloadedUpdate(path)` so the app can open the downloaded installer only after the user explicitly clicks install.
- Added localized toolbar update reminder UI with release notes, download, install, and ignore actions.
- Added Go tests for update release parsing, installer asset filtering, checksum validation, and ignored update persistence.
- Added app-owned current-document find for the active Markdown document.
- Added a `Ctrl/Cmd+F` handoff from the desktop guard layer to the reader find UI, keeping browser default find blocked.
- Added rendered Markdown match highlighting with current-match emphasis and previous/next navigation.
- Added localized current-document find labels for Chinese and English through Go-owned locale dictionaries.
- Added Go-managed persistent `locale` and `theme` settings under AppData `config/settings.json`.
- Added `SwitchLanguage(locale)` and `SwitchTheme(theme)` Wails APIs.
- Added toolbar language and theme controls that call Go APIs and consume Go-owned locale strings.
- Added system/light/dark theme support using existing CSS variables and a `prefers-color-scheme` listener for system mode.
- Added localized shell labels for system, light, and dark theme options.
- Added Go tests for persisted language/theme switching and normalization.
- Added `scripts/package-windows.ps1` to build a Wails NSIS installer, stage it under `dist/releases/v<version>/`, and generate `SHA256SUMS.txt`.

### Changed

- Advanced the product version to `0.1.2` for the update-check and find-focus release.
- `Ctrl/Cmd+F` now focuses and selects the current-document find input after the find bar is mounted, fixing the previous timing race.
- `settings.json` storage advanced to version 2 for the ignored update version field.
- Advanced the product version to `0.1.1` for the current-document find release.
- Switching documents now clears transient current-document find state and removes match highlights.
- Closing the find bar now removes all current-document highlights from the rendered Markdown DOM.
- Architecture notes now record current-document find as transient React-owned UI state and move language/theme switching into the implemented API list.
- Startup bootstrap now calls `GetBootstrap("")` so Go settings choose the current locale and theme instead of hardcoding `zh-CN` in React.
- `settings.json` now preserves locale/theme defaults while keeping existing workspace persistence behavior.
- GitHub Release packaging policy now treats installers as the primary user-facing artifact and reserves raw `jskernmd.exe` for local validation.
- Windows installer naming was corrected to the user-facing `JSKernMD-Setup-<version>-x64.exe` pattern instead of the internal binary-style name.
- GitHub Release asset labels now match their filenames exactly so the download list is readable.

### Release Packaging

- Windows installer artifact name: `JSKernMD-Setup-<version>-x64.exe`.
- Checksum artifact: `SHA256SUMS.txt`.
- Published release target: `v0.1.2`.
- Windows installer artifact: `JSKernMD-Setup-0.1.2-x64.exe`.
- Installer SHA256: `449f99550137ea0d36457860b18b456ed3224825cd14a7ac273661f9985b4574`.
- Published release target: `v0.1.1`.
- Windows installer artifact: `JSKernMD-Setup-0.1.1-x64.exe`.
- Installer SHA256: `83513e2681d3a753136a60c6d777f3722ea67d4169a6dd022bd85565bae910a7`.

### Validation

- `v0.1.2` update-check and find-focus release:
  - Product version sources were updated to `0.1.2`.
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed from `frontend/`.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - `scripts/package-windows.ps1` passed with process-local `-ExecutionPolicy Bypass` and produced `dist/releases/v0.1.2/JSKernMD-Setup-0.1.2-x64.exe`.
  - `SHA256SUMS.txt` was generated with SHA256 `449f99550137ea0d36457860b18b456ed3224825cd14a7ac273661f9985b4574`.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- `v0.1.1` installer release:
  - `go test ./...` passed.
  - `npm.cmd run build` passed from `frontend/`.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - Direct `scripts\package-windows.ps1` execution was blocked by PowerShell policy; rerunning with process-local `-ExecutionPolicy Bypass` succeeded.
  - `scripts/package-windows.ps1` passed and produced `dist/releases/v0.1.1/JSKernMD-Setup-0.1.1-x64.exe`.
  - `SHA256SUMS.txt` was generated with SHA256 `83513e2681d3a753136a60c6d777f3722ea67d4169a6dd022bd85565bae910a7`.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- Current-document find:
  - `npm.cmd run build` passed from `frontend/`.
  - `go test ./...` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- Language/theme settings and installer staging:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - Initial `scripts/package-windows.ps1` run built `build/bin/jskernmd.exe` but could not create the NSIS installer because `makensis` was missing.
  - Installed `NSIS.NSIS` through `winget` and updated `scripts/package-windows.ps1` to detect common NSIS install paths when PATH is not refreshed.
  - `scripts/package-windows.ps1` passed and produced `dist/releases/v0.1.0/jskernmd-v0.1.0-windows-amd64-setup.exe`.
  - `SHA256SUMS.txt` was generated with SHA256 `3cbbca75ffbbf8561f12599ab575a031c2e79e5530746af42801be8544ddf2c0`.
  - GitHub Release `v0.1.0` now contains the installer and checksum file; the previous raw exe asset was removed.
  - Windows launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- Installer rename correction:
  - `scripts/package-windows.ps1` passed and produced `dist/releases/v0.1.0/JSKernMD-Setup-0.1.0-x64.exe`.
  - `SHA256SUMS.txt` was regenerated with SHA256 `f591d4b676e4cb5b05184e4c9c71ccbab5c869f7029f43225be51d8a898d0bfb`.
  - GitHub Release `v0.1.0` was updated to use `JSKernMD-Setup-0.1.0-x64.exe`.
- Release asset label correction:
  - `JSKernMD-Setup-0.1.0-x64.exe` label now matches the filename.
  - `SHA256SUMS.txt` label now matches the filename.

---

## 2026-07-07

### Added

- Added `SearchWorkspace(query)` as a Go-owned Wails API for workspace Markdown search.
- Added bounded on-demand Markdown search across the current workspace:
  - matches Markdown file names and workspace-relative paths
  - matches document body text and returns a compact snippet
  - skips hidden entries and heavy folders such as `node_modules`, `dist`, `build`, and `vendor`
  - keeps search results capped at 50 items
- Added Go tests for file-name hits, content hits, skipped folders, and searching without an open workspace.
- Added a toolbar search input with debounced Wails calls, stale-response protection, keyboard Enter/Escape handling, and click-to-open results.
- Added localized search UI text in `zh-CN` and `en`.
- Regenerated Wails frontend bindings for the new `SearchResult` model and `SearchWorkspace` API.

### Changed

- Moved `SearchWorkspace(query)` from planned architecture work into the implemented Wails API surface.

### Validation

- Workspace search:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### Changed

- Reader position changes are now direct offset assignments instead of browser scroll animations.
- Removed smooth scrolling from the reader container so newly opened documents appear at the top immediately.
- Moved the external document-change reminder out of the Markdown document flow and into a bottom overlay inside the center reader area.
- Split the reader surface into a fixed shell plus an internal scroll container so status reminders stay visible regardless of document scroll position.
- Opening or reloading a document from the workspace tree now resets the center reader scroll position to the top instead of inheriting the previous document's scroll offset.
- Workspace-relative Markdown links with heading fragments still navigate to their requested heading after the new document renders.

### Validation

- Instant reader positioning:
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - Initial `npm.cmd audit --audit-level=moderate` hit an npm registry `ECONNRESET`; retrying through `127.0.0.1:10808` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
- Reader status banner and scroll reset:
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### Added

- Added Go-provided document freshness metadata:
  - `modifiedAt`
  - `size`
- Added `StatDocument(path, knownModifiedAt, knownSize)` as the Go-owned document status check API.
- Added Go tests for unchanged, changed, deleted, and outside-workspace document status checks.
- Added localized reader-surface error copy for document open failures.
- Added localized weak external-change reminder copy with reload and dismiss actions.

### Changed

- Failed document opens now clear stale reader content and show a visible error panel instead of silently leaving the previous document onscreen.
- The current document now polls Go for disk freshness and shows a non-modal reminder when the file changes externally.
- Reloading from the reminder reuses the existing document open path, so workspace boundary validation and Markdown rendering stay Go-owned.
- Dismissing an external-change reminder suppresses only that exact changed snapshot; a later file change can surface a new reminder.
- Wails frontend bindings were regenerated for the new `DocumentStatus` model and `StatDocument` API.

### Validation

- Document status notices:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### Added

- Added `markdown-reader-icon.svg` as the product app icon source artwork.
- Added a converted 1024x1024 alpha PNG app icon at `build/appicon.png`.
- Added a regenerated Windows ICO at `build/windows/icon.ico` so Wails embeds the new icon into `jskernmd.exe`.

### Changed

- Replaced the default Wails application icon with the JS Kern.md Markdown reader icon.

### Validation

- App icon integration:
  - Rendered `markdown-reader-icon.svg` to `build/appicon.png` with transparent corners.
  - Regenerated `build/windows/icon.ico` from the new PNG through Wails.
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### Added

- Added Shiki-based syntax highlighting for rendered Markdown code blocks.
- Added a focused frontend highlighter module that scans Go-sanitized Markdown HTML after document render.
- Added explicit language alias handling for common Markdown fence labels such as `js`, `ts`, `sh`, `ps1`, and `yml`.
- Added a Go sanitizer allowance for `class` attributes on `pre` and `code` elements so fenced code language markers survive into the renderer.
- Added a Go test proving fenced code blocks preserve `language-*` classes for the Shiki handoff.

### Changed

- Code highlighting now remains a frontend display enhancement while Markdown parsing, HTML rendering, and sanitization stay in Go.
- Shiki now uses a fine-grained bundled language/theme set instead of importing the full Shiki language catalog.
- Unsupported or unlabeled code blocks intentionally fall back to the existing plain code-block rendering.

### Validation

- Shiki code highlighting:
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### Added

- Added a Go-controlled `/kern-asset` endpoint through the Wails asset server for local Markdown images.
- Added Markdown AST rewriting for workspace-local bitmap image references.
- Added Markdown AST rewriting for relative Markdown document links.
- Added `OpenWorkspaceDocument(path)` Wails API for opening workspace-relative Markdown links.
- Added frontend Markdown click handling for links marked with `data-kern-document`.
- Added image sizing styles for rendered Markdown images.
- Added Go tests for local image rewriting, relative Markdown link rewriting, workspace-relative document opening, and asset endpoint path rejection.

### Changed

- `OpenDocument(path)` now renders Markdown through the App instance so it can resolve workspace-local resources.
- Workspace-local image serving now streams files through Go instead of embedding image bytes in JSON.
- Workspace-relative document links now round-trip through Go path validation instead of letting the WebView resolve local paths.
- SVG images are intentionally not served in this first local asset pass; bitmap formats are supported first.

### Validation

- Local image and relative Markdown link support:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.

---

### Added

- Added Go-managed AppData initialization for the `jskernmd` data root.
- Added the required local storage layout:
  - `config/`
  - `data/`
  - `logs/`
  - `cache/`
  - `temp/`
  - `runtime/`
  - `crash/`
- Added versioned `config/settings.json` with `storage_version` and `last_workspace`.
- Added atomic settings writes through a temporary file and rename.
- Added `.bad-*` backup behavior for invalid JSON settings files before falling back to defaults.
- Added `RestoreWorkspace()` Wails API.
- Added startup restore in the frontend so the last valid workspace tree reappears automatically.
- Added Go tests for AppData layout creation, settings persistence, workspace restore, and bad settings backup.

### Changed

- `ScanWorkspace(path)` now persists the successfully opened workspace directory.
- Startup workspace restoration keeps the root directory expanded while child directories remain collapsed by default.
- Project constraints, architecture notes, and decision log now record that directory-tree workspace state belongs in Go-managed AppData, not frontend storage.

### Validation

- AppData workspace persistence:
  - `go test ./...` passed.
  - `wails generate module` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
  - AppData smoke check passed: `C:\Users\cool\AppData\Roaming\jskernmd\config\settings.json` was created with `storage_version: 1`.

---

### Added

- Added expand/collapse behavior to the workspace tree.
- Added root-only default expansion: the workspace root opens, child directories start collapsed.
- Added an internal scroll region for the left workspace tree panel.
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

- Directory rows now act as toggles instead of disabled labels.
- The reader shell now clears the selected document when a new workspace is opened.
- The Markdown body and document path explicitly allow text selection while the rest of the shell remains anti-web.

### Validation

- Directory tree collapse/scroll update:
  - `go test ./...` passed.
  - `npm.cmd run build` passed.
  - `npm.cmd audit --audit-level=moderate` passed with 0 vulnerabilities.
  - `wails build` passed and produced `build/bin/jskernmd.exe`.
  - Launch smoke test passed: `jskernmd.exe` started and remained alive after 4 seconds before being stopped.
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
