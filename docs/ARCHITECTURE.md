# Architecture

## Runtime

JS Kern.md uses Wails v2 with a Go backend and React + TypeScript frontend.

Electron is forbidden. The binary and process name must remain `jskernmd`.

`product.manifest.json` is embedded into the Go binary and is the source for Go-visible product identity, including the version used by bootstrap and update checks.

## Ownership

Go owns:

- Workspace root validation and one-level-at-a-time directory tree loading.
- Workspace directory-tree refresh and structure change detection.
- Workspace path validation before system file-manager reveal actions.
- Markdown reading and parsing.
- Local image and relative link resolution.
- Document outline extraction.
- App config, recent workspace state, reading memory, theme, language, and persistence.
- Open document tab session state, including the active tab for each workspace.
- Update checks, installer download, checksum verification, ignored update version, and opening the downloaded installer.
- Windows Markdown association status and launching the official Windows default-app settings page through native `ShellExecuteW` protocol handling.
- i18n dictionaries and final dynamic user-visible messages.

React owns:

- Shell rendering.
- Directory tree UI.
- Open document tab strip rendering and tab switching interactions.
- Open-documents sidebar list rendering as a transient projection of the current open-tab state.
- App-owned context-menu rendering for directory tree and tab strip actions.
- Short-lived weak feedback for context-menu action results, with copy replaced by the next notice and all timers cleaned up on unmount.
- Markdown view rendering from Go-provided document data.
- App-level `Ctrl/Cmd+A` routing that selects only the rendered Markdown body unless a text input currently owns focus.
- Outline panel rendering, current-section highlighting, outline auto-follow, and transient reading-progress display.
- Settings dialog rendering, focus handling, and short-lived preference/association loading or error state.
- Short-lived UI interaction state such as hover, focus, pending buttons, selection, and current-document find highlights.

React must not directly access the filesystem, maintain translation dictionaries, parse durable business state, or persist business data.

## Data Flow

```text
workspace folder(s)
-> Go validates path
-> Go adds the folder to the AppData workspace collection
-> Go validates workspace roots without recursively scanning descendants
-> React renders multiple top-level workspace roots
-> user expands a directory
-> React calls `LoadDirectory(path)`
-> Go validates the directory and reads only its immediate Markdown files and child directories
-> React replaces only that directory node in the transient tree
-> user selects document
-> Go reads Markdown file
-> Go parses with goldmark
-> Go sanitizes HTML / structured document data
-> Go rewrites safe local images to controlled asset URLs
-> Go rewrites relative Markdown links to workspace-relative document actions
-> React renders reading view
-> Shiki highlights code blocks
-> React derives the visible heading and reading progress from the rendered document scroll container
-> React highlights and scrolls the Go-provided outline without persisting this transient navigation state
-> React reports debounced reading position and nearest heading to Go
-> Go persists workspace-scoped reading memory under AppData
-> React reports open tab order and active tab to Go
-> Go persists workspace-scoped tab session state with reading memory under AppData
-> React renders the same open-tab session in the upper left sidebar as an open-documents list
-> React keeps the lower left sidebar as the multi-root workspace directory tree
-> user right-clicks a directory-tree row or tab
-> React renders a transient app-owned context menu
-> file-manager reveal actions round-trip through Go workspace path validation
-> directory-tree rename actions round-trip through Go workspace path validation and OS rename
-> top-level workspace remove actions delete only AppData workspace membership, not disk files
-> React renders localized success/failure feedback from the Go-provided business locale at the bottom of the reader surface
-> top-level workspace drag sorting is sent to Go and persisted in settings
-> React remaps transient open-tab and expansion state from the Go returned old/new paths
-> React asks Go for current document status while the document is open
-> Go validates the path and reports whether the file changed on disk
-> React shows a weak reload reminder without taking ownership of filesystem state
-> React weakly polls Go for workspace structure refresh while workspaces are open
-> Go re-scans only directory levels already loaded into the runtime tree and compares structure signatures made from directory and Markdown file paths
-> React replaces the tree collection only when Go reports a structure change, preserving still-valid expanded directories
-> Explorer right-click or CLI args are parsed by Go and surfaced to React through a launch request
-> React may run transient in-document find highlighting over the already-rendered Markdown DOM
-> React may show an update reminder from Go-provided release metadata
-> Go downloads and verifies the installer into AppData temp storage when the user requests it
-> user opens the settings dialog from the toolbar
-> React changes language/theme only through Go preference APIs
-> React asks Go for the installed Windows Markdown association status
-> Go reads product registration and the current `.md` UserChoice without mutating either
-> React can ask Go to open the official Windows default-app settings page through `ShellExecuteW`, with the generic default-app page as an error fallback
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
- `GetMarkdownAssociationStatus()`
- `OpenWorkspace()`
- `AddWorkspace(path)`
- `ScanWorkspace(path)`
- `RestoreWorkspace()`
- `RestoreWorkspaces()`
- `LoadDirectory(path)`
- `RefreshWorkspace()`
- `RefreshWorkspaces()`
- `RemoveWorkspace(workspaceID)`
- `ReorderWorkspaces(workspaceIDs)`
- `ConsumeLaunchRequest()`
- `RevealPath(path)`
- `RenamePath(path, newName)`
- `OpenDocument(path)`
- `OpenWorkspaceDocument(path)`
- `StatDocument(path, knownModifiedAt, knownSize)`
- `SearchWorkspace(query)`
- `GetReadingMemory()`
- `GetReadingSession()`
- `GetReadingPosition(path)`
- `SaveOpenTabs(paths, activePath)`
- `SaveReadingPosition(path, scrollTop, scrollRatio, headingID, modifiedAt, size)`
- `SwitchLanguage(locale)`
- `SwitchTheme(mode)`
- `CheckForUpdates()`
- `DismissUpdate(version)`
- `DownloadUpdate(downloadURL, sha256)`
- `OpenDownloadedUpdate(path)`
- `OpenMarkdownDefaultAppsSettings()`

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

`settings.json` stores `storage_version`, legacy `last_workspace`, `active_workspace_id`, ordered `workspaces[]`, `locale`, `theme`, and `ignored_update_version`. Startup calls `GetBootstrap("")` to read Go-owned locale/theme preferences, then calls `RestoreWorkspaces()` to rebuild lightweight root nodes from valid workspace folders. Top-level workspace roots and child directories remain collapsed by default; `LoadDirectory(path)` reads one validated directory level only when the user expands it.

`data/reading-memory.json` stores workspace-scoped reading memory with `storage_version`, each workspace's open tab list, the active document, the legacy last document field, and bounded per-document reading positions. The saved position includes workspace-relative path, scroll offset, scroll ratio, nearest heading ID, document modified time, document size, and update time. Startup restores the workspace collection first, then asks Go for the reading session; if saved tabs still exist in any restored workspace, React restores the tab strip and opens the active tab. If there is only legacy last-document memory, Go normalizes it into a one-tab session. If document metadata still matches, React restores the exact scroll offset. If the file changed during ordinary document open, React falls back to the saved heading ID when it still exists, otherwise the document opens from the top. When React reports a reduced open-tab list, Go removes reading-position records for documents no longer in that list, and `GetReadingPosition()` ignores stale records for documents that are no longer open.

When switching tabs, React force-saves the previous tab's current scroll position through `SaveReadingPosition()` before opening the target tab and reading its own position through `GetReadingPosition()`. When the weak disk-change reminder reloads the current document, React captures the current reader offset before calling `OpenDocument()`, then applies that offset to the newly rendered document metadata so a reload does not jump to the top.

While workspaces are open, React calls `RefreshWorkspaces()` on a weak interval. Go refreshes only directory levels already loaded into its runtime tree and compares directory/Markdown-file structure signatures; unopened descendants are not scanned. Editing the current document remains handled by `StatDocument()`. When the tree collection changes, React keeps expansion state only for directories that still exist, so newly added directories start collapsed.

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
Every meaningful product update must end with a synchronized GitHub Release installer and checksum unless release work is explicitly paused by the user.

The Windows NSIS installer is localized through the native Modern UI language tables for English and Simplified Chinese. Installer startup reads the current Windows UI language and sets the installer/uninstaller language automatically, rather than showing a language picker.

The Windows installer writes `InstallLocation` to the HKLM uninstall registry entry and reads it during later installs so upgrades default to the user's existing install directory. For compatibility with older releases that did not write `InstallLocation`, installer startup also derives the prior directory from the quoted `UninstallString` path when possible.

Before packaging, `scripts/package-windows.ps1` synchronizes `wails.json.info` from `product.manifest.json`, keeping Wails installer metadata aligned with the manifest-driven product identity.

The administrator-level Windows installer registers machine-wide Explorer context-menu entries, the `JSKernMD.Markdown` ProgID, `Applications\jskernmd.exe`, application capabilities, `RegisteredApplications`, and OpenWith candidates for `.md`, `.markdown`, and `.mdown`. It does not write the protected per-user `UserChoice`; the settings dialog opens Windows Settings so the user can explicitly choose the default. The uninstaller removes only product-owned keys and values, including legacy current-user context-menu keys from older releases.

The installer also copies `build/windows/markdown-file.ico` to the install directory and uses it for both `JSKernMD.Markdown\DefaultIcon` and the `Applications\jskernmd.exe\DefaultIcon` fallback that Windows may store in `UserChoice`. Capabilities application identity, shortcuts, and Explorer command icons continue to reference `jskernmd.exe`, keeping document identity visually distinct from application identity.
