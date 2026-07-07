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
