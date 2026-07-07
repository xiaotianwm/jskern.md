# Constraints

## Hard Product Constraints

- The product is a desktop Markdown reader.
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

## Frontend

- React is only a rendering and interaction layer.
- Frontend must not read the filesystem directly.
- Frontend must not persist business data.
- Frontend must not maintain language dictionaries.
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
