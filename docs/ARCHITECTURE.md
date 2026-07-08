# Architecture

## Runtime

JS Kern.md uses Wails v2 with a Go backend and React + TypeScript frontend.

Electron is forbidden. The binary and process name must remain `jskernmd`.

## Ownership

Go owns:

- Workspace opening and directory tree scanning.
- Workspace directory-tree refresh and structure change detection.
- Markdown reading and parsing.
- Local image and relative link resolution.
- Document outline extraction.
- App config, recent workspace state, reading memory, theme, language, and persistence.
- Update checks, installer download, checksum verification, ignored update version, and opening the downloaded installer.
- i18n dictionaries and final dynamic user-visible messages.

React owns:

- Shell rendering.
- Directory tree UI.
- Markdown view rendering from Go-provided document data.
- Outline panel rendering.
- Short-lived UI interaction state such as hover, focus, pending buttons, selection, and current-document find highlights.

React must not directly access the filesystem, maintain translation dictionaries, parse durable business state, or persist business data.

## Data Flow

```text
workspace folder
-> Go validates path
-> Go scans Markdown directory tree
-> React renders tree
-> user selects document
-> Go reads Markdown file
-> Go parses with goldmark
-> Go sanitizes HTML / structured document data
-> Go rewrites safe local images to controlled asset URLs
-> Go rewrites relative Markdown links to workspace-relative document actions
-> React renders reading view
-> Shiki highlights code blocks
-> React reports debounced reading position and nearest heading to Go
-> Go persists workspace-scoped reading memory under AppData
-> React asks Go for current document status while the document is open
-> Go validates the path and reports whether the file changed on disk
-> React shows a weak reload reminder without taking ownership of filesystem state
-> React weakly polls Go for workspace structure refresh while a workspace is open
-> Go re-scans the current workspace and compares a structure signature made from directory and Markdown file paths
-> React replaces the tree only when Go reports a structure change, preserving still-valid expanded directories
-> React may run transient in-document find highlighting over the already-rendered Markdown DOM
-> React may show an update reminder from Go-provided release metadata
-> Go downloads and verifies the installer into AppData temp storage when the user requests it
```

Go sanitization preserves `language-*` classes only on `pre` and `code` so Shiki can identify fenced code languages. React treats Shiki as a display pass over the current document DOM; unsupported languages remain plain code blocks.

## i18n Flow

```text
Go embedded locale JSON
-> GetBootstrap(locale)
-> shellLocale + businessLocale
-> React renders only returned strings
```

Supported initial locales: `zh-CN`, `en`.

## Initial Wails APIs

- `GetBootstrap(locale)`
- `OpenWorkspace()`
- `ScanWorkspace(path)`
- `RestoreWorkspace()`
- `RefreshWorkspace()`
- `OpenDocument(path)`
- `OpenWorkspaceDocument(path)`
- `StatDocument(path, knownModifiedAt, knownSize)`
- `SearchWorkspace(query)`
- `GetReadingMemory()`
- `GetReadingPosition(path)`
- `SaveReadingPosition(path, scrollTop, scrollRatio, headingID, modifiedAt, size)`
- `SwitchLanguage(locale)`
- `SwitchTheme(mode)`
- `CheckForUpdates()`
- `DismissUpdate(version)`
- `DownloadUpdate(downloadURL, sha256)`
- `OpenDownloadedUpdate(path)`

## AppData Storage

Go creates and owns the app data root for `jskernmd`.

On Windows the active path is:

```text
%APPDATA%\jskernmd
```

The required layout is:

```text
jskernmd/
  config/settings.json
  data/
  logs/
  cache/
  temp/
  runtime/
  crash/
```

`settings.json` stores `storage_version`, `last_workspace`, `locale`, `theme`, and `ignored_update_version`. Startup calls `GetBootstrap("")` to read Go-owned locale/theme preferences, then calls `RestoreWorkspace()` to rebuild the tree from the last valid folder. Child directories remain collapsed by default; restoring a workspace must not eagerly expand the whole tree in the UI.

`data/reading-memory.json` stores workspace-scoped reading memory with `storage_version`, the last document in each workspace, and bounded per-document reading positions. The saved position includes workspace-relative path, scroll offset, scroll ratio, nearest heading ID, document modified time, document size, and update time. Startup restores the last workspace first, then asks Go for reading memory; if the last document still exists, React opens it. If document metadata still matches, React restores the exact scroll offset. If the file changed, React falls back to the saved heading ID when it still exists, otherwise the document opens from the top.

While a workspace is open, React calls `RefreshWorkspace()` on a weak interval. Go owns the actual re-scan and compares only directory and Markdown-file structure; editing the currently open document does not refresh the tree and remains handled by `StatDocument()`. When the tree changes, React keeps expansion state only for directories that still exist and always keeps the workspace root expanded, so newly added child directories start collapsed.

## Update Flow

JS Kern.md is an independent desktop app. Its desktop update source is the GitHub Releases API for `xiaotianwm/jskern.md`, using only release assets that match the canonical installer name:

```text
JSKernMD-Setup-<version>-x64.exe
```

Go checks for newer non-draft releases, returns the latest usable installer metadata to React, downloads installers into AppData `temp/update/`, verifies SHA256 when GitHub provides a digest, and opens the local installer only after the user clicks install. React only renders the weak reminder, download/install buttons, release notes, and transient busy/error state.

Do not add Cloudflare Workers, D1, R2, `DEVELOPER_KEY`, or the shared control-plane latest/update APIs to this product unless the product boundary is explicitly changed later.

## Local Resources

Markdown rendering rewrites local bitmap images to:

```text
/kern-asset?path=<workspace-relative-path>
```

The Wails asset server forwards this path to a Go handler. The handler resolves the path against the current workspace, validates the real path is still inside the workspace, and streams only supported bitmap image files.

Relative Markdown document links are rendered with `data-kern-document` and opened through `OpenWorkspaceDocument(path)`. React does not resolve filesystem paths; it only forwards the workspace-relative path back to Go.

## Release Artifacts

Windows releases are packaged with Wails NSIS through `scripts/package-windows.ps1`.

The canonical GitHub Release installer name is:

```text
JSKernMD-Setup-<version>-x64.exe
```

The staged release directory must also contain `SHA256SUMS.txt`. Raw Wails executable builds remain local validation artifacts and are not the primary downloadable release package.

GitHub Release asset labels must match the uploaded filenames exactly so the public download list is self-explanatory.
