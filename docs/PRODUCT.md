# Product

## Goal

Build a native-feeling desktop Markdown reader for folders of Markdown documents.

## MVP Scope

- Open one or more folders as workspaces.
- Show a persistent multi-root directory tree containing Markdown files.
- Keep the open workspace directory trees synchronized when Markdown files or folders are added, deleted, or renamed.
- Persist the workspace list and top-level workspace order across launches.
- Let users remove a top-level workspace from JS Kern.md without deleting disk files.
- Let Windows users open Markdown files or add folders from Explorer context menus.
- Restore the last opened document and reading position inside the restored workspace.
- Restore the last open document tabs and active tab inside the restored workspace.
- Show all currently open documents in the upper left sidebar section, including workspace documents and documents opened through Explorer/CLI entry points.
- Keep the lower left sidebar section as the multi-root workspace directory tree, separated by a draggable horizontal divider.
- Preserve each tab's reading position when switching between open Markdown documents.
- Clear a document's saved reading position when its tab is closed, so reopening it later starts from the top.
- Provide desktop-style right-click menus for directory-tree items and open tabs.
- Show weak bottom-of-reader feedback for context-menu copy, reveal, rename, and workspace-removal results.
- Rename Markdown files and folders from the directory-tree context menu through Go-controlled validation.
- Render the selected Markdown document.
- Make `Ctrl/Cmd+A` select only the rendered Markdown body while preserving native select-all inside focused text inputs.
- Show a document outline from headings, keep the current section highlighted while reading, and keep long outlines synchronized with the reader position.
- Show a lightweight document reading-progress indicator without turning the reader surface into a dashboard.
- Resolve local images and relative links safely.
- Highlight code blocks with Shiki.
- Support Chinese and English UI through Go-managed i18n.
- Preserve a frameless, anti-web desktop shell.

## Explicit Non-Goals For MVP

- Markdown editing.
- Cloud sync.
- Plugin marketplace.
- Multi-window document management.
- Remote collaboration.

## Product Principle

The directory tree is not an enhancement. It is the core reading model. Single-file opening may be added as a convenience later, but the primary product must remain folder-first.
