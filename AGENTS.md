# AGENTS.md

## Operating Rule

你作为一个 AI 工具，不要让我选择某个功能，应该给出我所需要功能最合理的方案。而不是根据我的需求，让我选择。

## Required Reading Before Changes

Before editing code, config, docs, build scripts, or GitHub workflow, read these files in order:

1. `docs/PROJECT_STATE.md`
2. `docs/CONSTRAINTS.md`
3. `docs/ARCHITECTURE.md`
4. `docs/DECISIONS.md`
5. `docs/CHANGELOG.md`

If the task touches product scope, also read `docs/PRODUCT.md`.

## Persistent Memory Rules

- The repository files are the source of truth for this project.
- Do not rely on chat memory when a repository document can answer the question.
- Every meaningful update must append a detailed entry to `docs/CHANGELOG.md`.
- Every development session must update `docs/PROJECT_STATE.md` with current phase, done work, next work, known issues, and validation.
- Architectural choices go into `docs/DECISIONS.md` only when they should prevent future re-litigation.
- Do not scatter project memory across random notes. Use the documented files only.

## Product Boundary

- JS Kern.md is a desktop Markdown reader.
- The MVP is directory-tree based, not single-file based.
- The directory tree is a core capability and must be preserved in every UI direction.
- Editing Markdown is out of MVP unless `docs/DECISIONS.md` is explicitly changed.

## Engineering Boundary

- Wails is the only desktop runtime.
- Electron is forbidden in this repository.
- Go owns filesystem access, Markdown parsing, path resolution, app config, i18n, persistence, and durable state.
- React owns rendering and short-lived UI interaction state only.
- Frontend source must not maintain translation dictionaries.
- User-visible UI text must come from Go bootstrap locale objects, except product logo typography and unavoidable platform labels.
- Use UTF-8 for all text files.
