# Product

## Goal

Build a native-feeling desktop Markdown reader for folders of Markdown documents.

## MVP Scope

- Open a folder as a workspace.
- Show a persistent directory tree containing Markdown files.
- Keep the open workspace directory tree synchronized when Markdown files or folders are added, deleted, or renamed.
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
