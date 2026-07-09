# Product

## Goal

Build a native-feeling desktop Markdown reader for folders of Markdown documents.

## MVP Scope

- Open a folder as a workspace.
- Show a persistent directory tree containing Markdown files.
- Keep the open workspace directory tree synchronized when Markdown files or folders are added, deleted, or renamed.
- Restore the last opened document and reading position inside the restored workspace.
- Restore the last open document tabs and active tab inside the restored workspace.
- Preserve each tab's reading position when switching between open Markdown documents.
- Clear a document's saved reading position when its tab is closed, so reopening it later starts from the top.
- Provide desktop-style right-click menus for directory-tree items and open tabs.
- Rename Markdown files and folders from the directory-tree context menu through Go-controlled validation.
- Render the selected Markdown document.
- Show a document outline from headings.
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
