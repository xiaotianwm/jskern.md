# Architecture

## Runtime

JS Kern.md uses Wails v2 with a Go backend and React + TypeScript frontend.

Electron is forbidden. The binary and process name must remain `jskernmd`.

## Ownership

Go owns:

- Workspace opening and directory tree scanning.
- Markdown reading and parsing.
- Local image and relative link resolution.
- Document outline extraction.
- App config, recent workspace state, theme, language, and persistence.
- i18n dictionaries and final dynamic user-visible messages.

React owns:

- Shell rendering.
- Directory tree UI.
- Markdown view rendering from Go-provided document data.
- Outline panel rendering.
- Short-lived UI interaction state such as hover, focus, pending buttons, and selection.

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
- `OpenDocument(path)`
- `OpenWorkspaceDocument(path)`

Planned APIs:

- `GetOutline(path)`
- `SearchWorkspace(query)`
- `SwitchLanguage(locale)`
- `SwitchTheme(mode)`

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

`settings.json` stores `storage_version` and `last_workspace`. Startup calls `RestoreWorkspace()` to rebuild the tree from the last valid folder. Child directories remain collapsed by default; restoring a workspace must not eagerly expand the whole tree in the UI.

## Local Resources

Markdown rendering rewrites local bitmap images to:

```text
/kern-asset?path=<workspace-relative-path>
```

The Wails asset server forwards this path to a Go handler. The handler resolves the path against the current workspace, validates the real path is still inside the workspace, and streams only supported bitmap image files.

Relative Markdown document links are rendered with `data-kern-document` and opened through `OpenWorkspaceDocument(path)`. React does not resolve filesystem paths; it only forwards the workspace-relative path back to Go.
