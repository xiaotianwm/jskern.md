# JS Kern.md

JS Kern.md is a lightweight desktop Markdown reader built with Go and Wails.

The product is folder-first: the core workflow is opening a directory, browsing a Markdown tree, reading documents, and navigating document outlines. Single-file opening can exist later, but it must not replace the directory-tree reading model.

## Stack

- Desktop runtime: Wails v2
- Backend: Go
- Frontend: React + TypeScript + Vite
- Markdown parser: goldmark
- Code highlighting: Shiki

## Development

```bash
wails dev
```

```bash
wails build
```

PowerShell may block `npm.ps1`; use `npm.cmd` for direct frontend commands on Windows.

## Project Memory

Before making changes, read:

- `AGENTS.md`
- `docs/PROJECT_STATE.md`
- `docs/CONSTRAINTS.md`
- `docs/ARCHITECTURE.md`
- `docs/DECISIONS.md`
- `docs/CHANGELOG.md`

Every meaningful update must end by updating `docs/PROJECT_STATE.md` and adding a detailed entry to `docs/CHANGELOG.md`.
