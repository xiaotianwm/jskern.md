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

## 2026-07-08: Keep JS Kern.md Independent From The Shared Control Plane

Reason: JS Kern.md is an independent Markdown reader, not part of the shared Cloudflare/control-plane software suite.

Consequence: Do not require `DEVELOPER_KEY`, Cloudflare Workers, D1, R2, or the shared control-plane publish/latest APIs for JS Kern.md releases or update checks. GitHub Releases remain the product's release and update source unless this product boundary is explicitly changed.

## 2026-07-08: Keep Workspace Refresh Go-Owned

Reason: The directory tree is the product's core reading model, and filesystem state should not drift into React during long-term development.

Consequence: Workspace structure refresh uses a Go Wails API and Go-owned structure signatures. React may weakly poll and render the returned tree, but it must not read directories, persist workspace paths, or decide filesystem truth. Active document content changes stay in the separate `StatDocument()` reminder flow.

## 2026-07-08: Keep Reading Memory Go-Owned

Reason: Reading progress is durable reader state, and long-term development would become confused if scroll positions and last-document memory lived in frontend browser storage.

Consequence: Reading memory is stored by Go under AppData `data/reading-memory.json`, scoped by workspace, and bounded per workspace. React may report transient scroll position and current heading, but it must not own the storage schema or write durable reading memory locally.
