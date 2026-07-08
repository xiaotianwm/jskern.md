# Constraints

## Hard Product Constraints

- The product is a desktop Markdown reader.
- The product is an independent desktop app and does not use the shared Cloudflare/control-plane release system.
- The MVP must be folder-first and directory-tree based.
- Directory tree support is mandatory in the first usable version.
- Single-file mode is secondary and must not replace workspace mode.

## Naming

- Repository: `jskern.md`
- Display name: `JS Kern.md`
- Binary and process name: `jskernmd`
- Product app ID: `js.kern-md`
- App slug and data root: `jskernmd`

## Runtime

- Must use Wails + Go.
- Must not introduce Electron.
- Must support Windows 10/11 and macOS through Wails-supported native webviews.

## Backend

- Go is the authority for filesystem access, path validation, Markdown parsing, local asset resolution, config, persistence, i18n, and dynamic messages.
- Markdown parsing must use `goldmark`.
- Rendered Markdown must be sanitized before entering the frontend.
- Physical paths must use `path/filepath`; no string-concatenated paths.
- Large files and assets must use IDs, controlled URLs, or streaming paths instead of base64 JSON payloads.
- Local Markdown images must be served through a Go-controlled workspace asset endpoint, not embedded as base64 and not read by React.
- Relative Markdown document links must be resolved by Go against the current workspace before opening.
- App configuration and durable reader state must be stored by Go under the system app data root for `jskernmd`; on Windows this is `%APPDATA%\jskernmd`.
- The last opened workspace directory must be persisted and restored on startup so users do not need to reopen the same folder every launch.
- Reading memory is Go-owned durable reader state and must be stored under AppData `data/reading-memory.json`, not in frontend `localStorage`.
- Reading memory must remain workspace-scoped and directory-tree-first: restore the last document inside the restored workspace, and preserve per-document scroll position plus heading fallback when the document still exists.
- Open document tabs and the active tab are Go-owned durable reader session state and must be stored with reading memory under AppData, not in frontend `localStorage`.
- Tab switching must save the previous tab's current reading position before opening the target tab, then restore the target tab's own saved reading position instead of jumping to the top.
- Reloading a changed document from the weak disk-change reminder must preserve the current reader offset with the reloaded document metadata instead of reusing stale saved memory or jumping to the top.
- Directory tree refresh and workspace structure change detection are Go-owned responsibilities; React may ask for a refreshed tree but must not inspect the filesystem itself.
- Workspace structure refresh must distinguish directory/file structure changes from active document content changes. Content changes stay in the current-document status reminder flow.
- AppData storage must be versioned with `storage_version`, use a layered directory layout, and preserve bad JSON files with `.bad-*` backups instead of silently overwriting them.
- Update checking, ignored update versions, installer downloads, checksum verification, and opening downloaded installers are Go-owned responsibilities.
- Current update downloads are sourced from GitHub Releases and must only accept canonical `JSKernMD-Setup-<version>-x64.exe` assets from the official `xiaotianwm/jskern.md` repository.

## Frontend

- React is only a rendering and interaction layer.
- Frontend must not read the filesystem directly.
- Frontend must not persist business data.
- Frontend may report transient scroll position and current heading to Go, but must not own the reading-memory storage format or write durable reading state locally.
- Frontend may render the tab strip and call Go APIs to save the open-tab list, but must not own the durable tab-session storage format or persist open tabs locally.
- Frontend must not maintain language dictionaries.
- Frontend must not download update installers directly; it may only render Go-provided update metadata, busy/error state, and user actions.
- Frontend workspace auto-sync may use a weak polling loop against Go, but the loop must clean up timers, avoid overlapping requests, preserve valid expanded directories, and keep newly discovered directories collapsed by default.
- Code highlighting must use Shiki.
- Browser-default context menu, refresh, find, zoom, link/image drag, and page overscroll must be blocked.
- Text selection is disabled by default, then re-enabled only for Markdown body, code blocks, inputs, and explicit selectable data.

## Window And UI

- Wails window must be frameless.
- Custom titlebar must expose a drag region and exclude interactive controls from drag.
- Windows should enable Mica or a Wails-supported fallback.
- macOS should use transparent titlebar / vibrancy-capable configuration where supported.
- The first screen must be the actual reader shell, not a landing page.
- Layout must be dense and desktop-like: left workspace tree, central reader, right outline.

## i18n

- Source files should use stable English keys for user-visible UI text.
- Go owns locale dictionaries.
- Main bootstrap returns `shellLocale` and `businessLocale`.
- Shell text and business text stay separated.
- Dynamic errors and status messages are produced by Go in final visible language.

## Memory And Change Logging

- `docs/PROJECT_STATE.md` is the current short-term project memory.
- `docs/DECISIONS.md` is the long-term decision log.
- `docs/CHANGELOG.md` must receive a detailed entry for every meaningful update.
- Do not create scattered ad-hoc memory files unless the structure itself is being changed.

## Release Packaging

- JS Kern.md releases use GitHub Releases directly; do not require `DEVELOPER_KEY`, Cloudflare Workers, D1, R2, or the shared control-plane publish API.
- GitHub Releases must publish installer packages, not raw development executables.
- Windows release artifacts must use the user-facing naming pattern `JSKernMD-Setup-<version>-x64.exe`.
- GitHub Release asset labels must match their filenames exactly; do not use vague labels such as `Windows x64 installer`.
- Each release upload must include `SHA256SUMS.txt` for the published installer.
- The raw `build/bin/jskernmd.exe` is a local build output only and must not be the primary GitHub Release download.
- Windows installer staging is produced by `scripts/package-windows.ps1`, which wraps Wails NSIS packaging and copies the installer into `dist/releases/v<version>/`.
