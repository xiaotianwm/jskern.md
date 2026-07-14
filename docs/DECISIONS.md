# Decisions

## 2026-07-07: Use Wails As The Only Desktop Runtime

Reason: JS Kern.md targets a lightweight native-feeling desktop reader, and Wails keeps binary size and memory use lower than Electron for this product shape.

Consequence: Electron must not be introduced. Go and Wails APIs own native integration.

## 2026-07-07: Build A Folder-First Reader

Reason: The target product is a Markdown reading environment, not a single-file preview utility.

Consequence: Directory tree support is mandatory in the MVP. Single-file opening may be added later only as a convenience path.

## 2026-07-07: Keep Go As The Durable State And i18n Owner

Reason: Long-lived desktop development becomes confused when frontend state, translations, and filesystem logic are scattered across layers.

Consequence: Go owns filesystem, parsing, persistence, app config, and locale dictionaries. React only renders data returned by Go.

## 2026-07-07: Use Repository Documents As Project Memory

Reason: Chat memory drifts during long development. Repository files are versioned and visible to every future session.

Consequence: `PROJECT_STATE.md`, `DECISIONS.md`, and `CHANGELOG.md` must be updated as part of done.

## 2026-07-07: Persist Workspace State In AppData

Reason: A directory-tree reader should reopen the last reading workspace automatically instead of making the user choose the same folder every launch.

Consequence: Go owns `%APPDATA%\jskernmd` on Windows and the equivalent system config directory on other platforms. `settings.json` must remain versioned and store `last_workspace`; React must not use browser storage for this durable path state.

## 2026-07-07: Serve Local Markdown Assets Through Go

Reason: Folder-based Markdown libraries commonly use relative images and cross-document links, but React must not read local files or trust relative paths.

Consequence: Markdown rendering rewrites local bitmap images to a Go-controlled `/kern-asset` URL and rewrites relative Markdown links to workspace-relative document actions. Go validates every resolved path against the current workspace before serving or opening it.

## 2026-07-07: Keep Shiki As A Frontend Display Enhancement

Reason: Code blocks need TextMate-quality highlighting, but Markdown parsing, sanitization, and durable document semantics must remain Go-owned.

Consequence: Go preserves safe `language-*` classes on `pre` and `code`; React applies Shiki after rendering sanitized HTML. If Shiki fails or the language is unsupported, the original plain code block remains visible.

## 2026-07-08: Publish Installers To GitHub Releases

Reason: End users should download a normal desktop installer instead of a raw build executable, and artifact names need to stay stable across releases.

Consequence: Windows GitHub Release uploads use `JSKernMD-Setup-<version>-x64.exe` plus `SHA256SUMS.txt`. The release staging script is `scripts/package-windows.ps1`; `build/bin/jskernmd.exe` is kept as a local validation output only.
Every meaningful product update is treated as incomplete until the installer and checksum have been synchronized to GitHub Releases, unless the user explicitly pauses release work.

## 2026-07-08: Keep JS Kern.md Independent From The Shared Control Plane

Reason: JS Kern.md is an independent Markdown reader, not part of the shared Cloudflare/control-plane software suite.

Consequence: Do not require `DEVELOPER_KEY`, Cloudflare Workers, D1, R2, or the shared control-plane publish/latest APIs for JS Kern.md releases or update checks. GitHub Releases remain the product's release and update source unless this product boundary is explicitly changed.

## 2026-07-08: Keep Workspace Refresh Go-Owned

Reason: The directory tree is the product's core reading model, and filesystem state should not drift into React during long-term development.

Consequence: Workspace structure refresh uses a Go Wails API and Go-owned structure signatures. React may weakly poll and render the returned tree, but it must not read directories, persist workspace paths, or decide filesystem truth. Active document content changes stay in the separate `StatDocument()` reminder flow.

## 2026-07-08: Keep Reading Memory Go-Owned

Reason: Reading progress is durable reader state, and long-term development would become confused if scroll positions and last-document memory lived in frontend browser storage.

Consequence: Reading memory is stored by Go under AppData `data/reading-memory.json`, scoped by workspace, and bounded per workspace. React may report transient scroll position and current heading, but it must not own the storage schema or write durable reading memory locally.

## 2026-07-08: Keep Tab Sessions In Reading Memory

Reason: Multi-tab reading is part of the reader session, and users should not lose their open Markdown set or active tab after restarting the app.

Consequence: Open tabs and the active document are persisted by Go in the workspace-scoped reading memory file. React renders the tab strip and reports tab changes, but it must not use frontend storage for tab sessions. Switching tabs must save the outgoing document position and restore the incoming document position; reloading a changed document must preserve the current reader offset.

## 2026-07-09: Use Product Manifest For Go-Visible Version

Reason: The `v0.1.9` installer metadata was updated, but the Go update-check version constant remained `0.1.8`, causing a freshly installed `0.1.9` app to still report itself as old and prompt for the same update.

Consequence: Go embeds `product.manifest.json` and derives bootstrap product identity, update-check current version, update download metadata, AppData slug, and update User-Agent from that manifest data. A separate hardcoded Go app-version constant is forbidden.

## 2026-07-09: Treat Closing A Tab As Clearing Its Reading Position

Reason: Closing a document is an explicit signal that the document should leave the active reading session; reopening it later from the directory tree should feel like a fresh open, not a hidden restoration of an old closed-tab offset.

Consequence: `SaveOpenTabs()` is the Go-owned cleanup boundary for tab-session changes. When a document is no longer present in the reported open-tab list, its saved reading-position record is removed. `GetReadingPosition()` must also ignore stale records for documents outside the current open-tab list so old AppData state cannot resurrect closed-document positions.

## 2026-07-08: Keep Context-Menu File Actions Go-Validated

Reason: Directory-tree and tab context menus should feel native, but revealing or renaming local files and folders are still filesystem actions that must not drift into React.

Consequence: React may render transient app-owned context menus, inline rename inputs, copy Go-provided paths, and call existing reader actions. Any system file-manager reveal or directory-tree rename action must call a Go Wails API that validates the target is an existing file or directory inside the current workspace before launching the platform file manager or mutating the filesystem. React may remap transient tabs and expansion state from Go-returned old/new paths, but it must not decide filesystem truth or persist rename state locally.

## 2026-07-08: Make Windows Installer Upgrades Directory-Aware

Reason: Users expect an installer update to reuse the directory where the app is already installed, and the installer should speak the user's system language without adding another choice step.

Consequence: The Windows NSIS installer selects English or Simplified Chinese from the Windows UI language, writes `InstallLocation` to the uninstall registry entry, and reuses that path on future installs. Older installs that lack `InstallLocation` are handled by deriving the path from `UninstallString` when possible.

## 2026-07-09: Treat Workspaces As An Ordered Collection

Reason: A Markdown reader can cover several documentation roots or note libraries at once. Opening a new folder should not destroy the previous workspace tree.

Consequence: JS Kern.md supports multiple top-level workspace folders. Opening a folder adds or activates an entry instead of replacing all existing roots; nested duplicates are avoided by Go. The ordered workspace list, active workspace, and root metadata live in Go-managed AppData settings. React may render drag sorting, but Go persists the final order.

## 2026-07-09: Register Explorer Entry Points Through The Installer

Reason: Users should be able to open Markdown files or add folders from Windows Explorer, and those OS hooks must be removed with the app.

Consequence: Windows Explorer integration belongs to the installer/uninstaller lifecycle. The installer registers current-user context-menu keys for opening Markdown files with JS Kern.md and adding folders to the workspace list. The uninstaller deletes only those product-specific keys and must not remove user documents, AppData, or broad file associations.

## 2026-07-11: Load Workspace Trees One Directory Level At A Time

Reason: Recursively scanning every configured workspace during startup and every weak refresh makes the folder-first core scale with the entire library even when most directories remain collapsed.

Consequence: Workspace restore validates only each root. React requests `LoadDirectory(path)` when a directory is expanded, Go validates and reads only that directory's immediate children, and weak refresh scans only directory levels already loaded into the Go runtime tree. Search remains a separate bounded on-demand scan and no index database is introduced.

## 2026-07-14: Register Markdown Capability In The Installer Without Forcing The Default

Reason: JS Kern.md should appear as a normal Windows default-app candidate, but Windows protects each user's actual file-type choice through the hashed `UserChoice` key.

Consequence: The administrator-level NSIS installer registers the product ProgID, application capabilities, `RegisteredApplications`, supported Markdown extensions, and Explorer entry points under HKLM. The application may read association status and open the official Windows default-app settings page, but neither installer nor application may write `UserChoice`. Uninstall removes only JS Kern.md-owned keys and values.
