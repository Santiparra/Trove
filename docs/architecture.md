# Architecture

Trove is a modular local application: a React/TypeScript interface hosted by Wails v3, Go use cases, and embedded SQLite. It does not require a database server or a Trove account.

## Responsibilities

| Component | Responsibility |
| --- | --- |
| `frontend/src/books` | Presentation, forms, navigation, and interaction state |
| `frontend/src/backup` | Visible backup settings, status, and recovery confirmation |
| `frontend/src/catalog` | Catalog search, edition review, linking, and local cover presentation |
| `frontend/src/novels` | Novel-level library, progress editing, and volume detail forms |
| `internal/desktop` | Wails adapters, input/output contracts, and native dialogs |
| `internal/books` | Library identities, validation, and use cases |
| `internal/novels` | Novel aggregates, numbered volumes, progress, and reading actions |
| `internal/storage/sqlite` | SQL, transactions, migrations, and snapshot adaptation |
| `internal/backup` | Logical format, validation, generations, and restoration through a backend contract |
| `internal/config` | Profile and data-location resolution |
| `internal/catalog` | Provider-independent bibliographic contracts, ISBN handling, cascading, and provenance |
| `internal/novelcatalog` | Series/main-book/release identity, publication snapshots, and safe Hardcover supplementation for novels |
| `internal/providers` | Hardcover/Open Library adapters and local cover caching |
| `internal/credentials` | Per-profile catalog key protection using Windows DPAPI |
| `main.go` | Dependency composition and native lifecycle |

The frontend does not access SQL. External provider types and driver-specific types do not define the domain. Bindings are generated from Go, committed, and checked in CI alongside their TypeScript consumers.

## Local model

A work has its own identity and may belong to a series at a given position. An edition holds format, language, an optional identifier, and ownership. Reading activity is stored in separate cycles with status and known dates. Favorites and notes belong to the personal record.

Books and Light Novels share the reading model, with an explicit `books` / `light-novels` discriminator. Services are scoped per request; the backend has no mutable global "current medium." The adapter requires a medium for queries and mutations. An ID from another section does not grant access to read, edit, or delete it in the wrong context. Editing an edition does not change the entry's medium. Series also have medium-specific identities, even when their names match.

Books works with one edition/cycle per displayed reading entry. Light Novels uses one entry per novel with independent numbered volume entities, editions/cycles, notes, and personal flags. Its completed count and optional total describe aggregate progress; live start/finish actions operate on the next numbered volume. Historical progress does not fabricate dates, and unknown quantities do not materialize volumes. Backups preserve all stored entities, not just a list projection.

The novel model has its own repository and Wails service rather than reinterpreting a Books series as a novel. SQLite stores separate `novels` and `novel_volumes` rows, with a foreign key and unique per-novel volume numbers. Typed Go records are serialized in row payloads, and row IDs/numbers are checked against them on reads. Aggregate writes replace parent/volume data within one transaction and notify the existing backup mechanism. The shared library gate serializes novel mutations against restoration and shutdown.

No conversion of the earlier standalone-volume model is performed. Existing Books and standalone volume rows remain intact; the novel-level interface reads the new aggregate storage.

## Persistence and recovery

SQLite is encapsulated behind concrete operations. Transactions prevent partial updates. The schema is versioned, and existing migrations create a safety copy where applicable. An unrecognized future schema is rejected.

Backups produce SQLite and an engine-independent JSON model from the same consistent snapshot. Generations are staged and published only when complete; failed generations do not replace the last valid one.

Restoration validates the source, prepares a new database, preserves a safety copy of the active library, and serializes replacement against library operations. Machine-specific configuration is not part of restored data. Changing database engines would require another adapter and a migration, but reading the JSON does not require SQLite.

## Media and provider boundaries

Each medium has its own identity and rules. Sharing mechanisms does not require every medium to have the same structure. Importing external files, linking bibliographic metadata, and restoring backups are separate operations.

Books uses Hardcover as its primary catalog. Open Library supplements missing fields for the selected edition using a safe ISBN match. It is not automatically promoted to primary and is not queried for every search candidate. Snapshots keep remote references, sources, and warnings separate from personal data. Refreshing does not overwrite title, authors, series, notes, status, or dates; a failed refresh preserves the previous snapshot.

Go prepares a selection and confirms it through a temporary token, rather than accepting a frontend-supplied snapshot. Local matches require review before linking; an already-linked entity opens without creating a duplicate. External-service importers are not implemented yet.

Light Novels uses RanobeDB series identities, explicitly ordered main-book references, and single complete releases. Original titles and source ordering are preserved separately from local reading positions. Hardcover supplements only verified ISBN/language/format matches and checks its structured book category; community genre tags are not used as medium classification. The novel path does not invoke Open Library or approximate Hardcover title searches.

Novel and volume snapshots remain separate from personal fields. Refreshes preserve manually supplied text and dates, reject changed identities/order, and never infer totals or completed history from catalog counts. Metadata for the next volume is loaded without starting it. SQL indexes enforce unique linked RanobeDB series/book identities. Novel refreshes capture the library's restore epoch and cannot apply data after a successful restoration or an intervening switch to local-only mode.

`BookSearch` presents local matches first and then remote candidates in the same popover. `BookProfile` shares local and remote detail presentation; available management actions depend on whether a personal record exists. Go recognizes matches before creation and prefers a local or more informative edition when opening a work. Edition changes are handled within the detail page. Key configuration lives in the header, outside bibliographic content.

`internal/providers/health.go` integrates diagnostics and circuit breakers with actual adapter calls. It holds temporary per-provider state, distinguishes access errors, rate limits, missing records, and network/response failures, and stops queued requests when a circuit opens. Startup launches two bounded checks in parallel; shutdown cancels their contexts and waits for completion. The frontend refreshes state through local IPC without periodically probing remote services. Key changes reset Hardcover's circuit and invalidate older in-flight results.

`internal/providers/network.go` maintains the profile's network policy and protects catalog and cover-download HTTP transports. Local-only mode blocks new requests and cancels existing request contexts; intentional interruptions do not count as provider failures. Future adapters must share this policy. Restore selection uses generation IDs validated by Go, not frontend-constructed paths.

## Distribution

The frontend is embedded in the executable. Build identity includes version, revision, and profile and can be read without opening SQLite. Packages contain the verified executable, not a library or personal configuration. See [development](development.md), the [backup format](backup-format.md), and [releases](releases/README.md).

Release notes are kept per release in `docs/releases/`. `current.txt` selects the public document used as the release body, independently of the PR title or merge message. The version comes from the identity calculated by Actions.
