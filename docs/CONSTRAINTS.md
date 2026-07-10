# Constraints

## Hard Product Constraints

- The product is a desktop Markdown reader.
- The product is an independent desktop app and does not use the shared Cloudflare/control-plane release system.
- The MVP must be folder-first and directory-tree based.
- Directory tree support is mandatory in the first usable version.
- The workspace model supports multiple top-level folders; opening a new folder adds it to the workspace list instead of replacing the existing roots.
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
- Go-visible product identity and update-check current version must come from `product.manifest.json`; do not maintain a second hardcoded Go version constant.
- Markdown parsing must use `goldmark`.
- Rendered Markdown must be sanitized before entering the frontend.
- Physical paths must use `path/filepath`; no string-concatenated paths.
- Large files and assets must use IDs, controlled URLs, or streaming paths instead of base64 JSON payloads.
- Local Markdown images must be served through a Go-controlled workspace asset endpoint, not embedded as base64 and not read by React.
- Relative Markdown document links must be resolved by Go against the current workspace before opening.
- App configuration and durable reader state must be stored by Go under the system app data root for `jskernmd`; on Windows this is `%APPDATA%\jskernmd`.
- The workspace list, active workspace, and top-level workspace order must be persisted and restored on startup so users do not need to reopen the same folders every launch.
- Reading memory is Go-owned durable reader state and must be stored under AppData `data/reading-memory.json`, not in frontend `localStorage`.
- Reading memory must remain workspace-scoped and directory-tree-first: restore the last document inside the restored workspace, and preserve per-document scroll position plus heading fallback when the document still exists.
- Open document tabs and the active tab are Go-owned durable reader session state and must be stored with reading memory under AppData, not in frontend `localStorage`.
- Tab switching must save the previous tab's current reading position before opening the target tab, then restore the target tab's own saved reading position instead of jumping to the top.
- Closing a document tab must remove that document's saved reading position from Go-owned reading memory; reopening a closed document from the directory tree must start at the top unless it is still part of the current open-tab session.
- Reloading a changed document from the weak disk-change reminder must preserve the current reader offset with the reloaded document metadata instead of reusing stale saved memory or jumping to the top.
- Directory tree refresh and workspace structure change detection are Go-owned responsibilities; React may ask for a refreshed tree but must not inspect the filesystem itself.
- Workspace structure refresh must distinguish directory/file structure changes from active document content changes. Content changes stay in the current-document status reminder flow.
- Context-menu actions that reveal, rename, or remove workspace roots must go through Go path validation. Removing a workspace only removes it from JS Kern.md and must not delete disk files.
- Explorer right-click entry points must route through Go-owned CLI argument handling: Markdown files open with JS Kern.md, and folders join the workspace list.
- AppData storage must be versioned with `storage_version`, use a layered directory layout, and preserve bad JSON files with `.bad-*` backups instead of silently overwriting them.
- Update checking, ignored update versions, installer downloads, checksum verification, and opening downloaded installers are Go-owned responsibilities.
- Current update downloads are sourced from GitHub Releases and must only accept canonical `JSKernMD-Setup-<version>-x64.exe` assets from the official `xiaotianwm/jskern.md` repository.

## Frontend

- React is only a rendering and interaction layer.
- Frontend must not read the filesystem directly.
- Frontend must not persist business data.
- Frontend may report transient scroll position and current heading to Go, but must not own the reading-memory storage format or write durable reading state locally.
- Frontend may derive current-section highlighting, outline auto-follow, and visual reading progress from the rendered document scroll container; this navigation state is transient and must not become a second durable reading-memory source.
- Frontend may render the tab strip and call Go APIs to save the open-tab list, but must not own the durable tab-session storage format or persist open tabs locally.
- Frontend must not maintain language dictionaries.
- Frontend must not download update installers directly; it may only render Go-provided update metadata, busy/error state, and user actions.
- Frontend may render app-owned context menus for the directory tree and tab strip, including inline rename editing, but menu and rename-edit state are transient UI state only and must not become durable workspace/session state.
- Context-menu action feedback must remain transient, use Go-provided locale strings, replace rather than stack repeated notices, and clean up its dismissal timer.
- Frontend workspace auto-sync may use a weak polling loop against Go, but the loop must clean up timers, avoid overlapping requests, preserve valid expanded directories, and keep newly discovered directories collapsed by default.
- Frontend may render top-level workspace drag sorting, but the persisted order is Go-owned AppData state.
- Code highlighting must use Shiki.
- Browser-default context menu, refresh, find, zoom, link/image drag, and page overscroll must be blocked.
- Text selection is disabled by default, then re-enabled only for Markdown body, code blocks, inputs, and explicit selectable data.
- App-level `Ctrl/Cmd+A` outside text inputs must select only the Markdown body; document headers, paths, toolbars, search UI, sidebars, and outline rows must stay outside that selection range.

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
- Every meaningful product update must be packaged and synchronized to GitHub Releases in the same development session unless the user explicitly pauses release work.
- GitHub Releases must publish installer packages, not raw development executables.
- Windows release artifacts must use the user-facing naming pattern `JSKernMD-Setup-<version>-x64.exe`.
- Windows installer UI must support English and Simplified Chinese and choose the visible installer language from the current Windows UI language without asking the user to pick.
- Windows NSIS project templates must stay ASCII unless the build command explicitly configures a UTF-8 input charset; rely on NSIS built-in language tables for localized wizard text.
- Windows installer upgrades must default to the previously installed directory by reading the app uninstall registry entry; new installs must write `InstallLocation`, and upgrades from older installers must fall back to deriving the directory from the previous `UninstallString`.
- Windows installer metadata in `wails.json.info` must be synchronized from `product.manifest.json` before packaging so the installer and uninstall entry stay aligned with the product identity.
- Windows installer must register Explorer context-menu entries for opening Markdown files and adding folders to the JS Kern.md workspace list, and the uninstaller must delete only the `JSKernMD.*` registry keys it created.
- GitHub Release asset labels must match their filenames exactly; do not use vague labels such as `Windows x64 installer`.
- Each release upload must include `SHA256SUMS.txt` for the published installer.
- The raw `build/bin/jskernmd.exe` is a local build output only and must not be the primary GitHub Release download.
- Windows installer staging is produced by `scripts/package-windows.ps1`, which wraps Wails NSIS packaging and copies the installer into `dist/releases/v<version>/`.
