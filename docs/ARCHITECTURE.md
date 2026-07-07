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
-> React renders reading view
-> Shiki highlights code blocks
```

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

Planned APIs:

- `ResolveAsset(path)`
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
