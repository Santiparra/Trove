# Security and privacy

Trove keeps your library local. It has no Trove accounts, telemetry, or integrated AI. Books queries Hardcover with optional Open Library supplementation. Light Novels queries RanobeDB with verified Hardcover ISBN supplementation. The packaged frontend does not load arbitrary remote scripts or images.

SQLite and backups are not encrypted by Trove: anyone with access to those files can read them. Device and storage-account protection depend on Windows and the configured storage service. Credentials, libraries, and personal paths are not bundled in the executable.

## Files and recovery

- JSON and SQLite backups are identified and validated before restoration. Unknown versions, broken references, and incomplete content are rejected.
- Restoration prepares a separate database and requires confirmation of the destination profile.
- A safety copy is kept, and recovery is attempted if replacement fails.
- Backup paths, logs, caches, and credentials are excluded from the portable data model.
- Local backup status does not confirm that a cloud synchronization client has uploaded the files.

## Optional catalogs

Searches send the query or ISBN to Hardcover. Selecting or refreshing an edition queries its references and, where applicable, its ISBN in Open Library. Cover downloads expose the IP address to the image host. Trove does not query or synchronize the user's personal Hardcover library.

Novel searches send title/author text or ISBN to RanobeDB. Author navigation uses its staff filter when an identity is known. Novel lookups send series, book, and release IDs; Hardcover only receives a verified publication ISBN and public catalog queries, never approximate novel-title fallback searches. Personal notes, favorites, ownership, and reading dates are not sent. Open Library is not part of this path.

At startup, the app checks Hardcover's API, only when a key is configured, using a single catalog reference query. Open Library uses a fixed public diagnostic ISBN and RanobeDB uses a fixed public series search. These checks also expose the IP address, but do not send library titles, notes, or progress. They run once in the background and may be repeated manually; there is no periodic remote polling. Status, generic errors, and timings are kept only in memory. Normal queries update the same diagnostics; messages do not include remote response bodies or credentials.

Local-only mode is loaded before startup checks and blocks the HTTP transport shared by catalogs and covers, including redirects. Enabling it cancels in-flight requests, including response-body reads. Database and cached-cover reads do not use the network. This preference belongs to the profile, is excluded from backups, and does not remove credentials.

The key is entered in a password field and stored per profile in `catalog-credentials.bin`, protected with DPAPI for the Windows account. It is not returned to the frontend or included in backups. Saving a key does not verify it until a query is made. Removing it preserves the library and snapshots.

Go restricts endpoints, HTTPS, redirects, response sizes, and timeouts; it does not return raw error bodies containing credentials. Covers are accepted only from `assets.hardcover.app`, `images.hardcover.app`, `covers.openlibrary.org`, and `images.ranobedb.org`. They are validated as raster images and delivered to React as data URLs through a per-profile local cache. The key is not sent to image hosts, RanobeDB, or Open Library. Metadata is rendered as text, without executing external HTML or SQL.

## Automation

PR validation uses read permissions. Only the publication job for a push to `main` requests content-write permissions. Third-party Actions are pinned to specific revisions. PRs and pushes to `dev` do not publish releases.

Portable releases are not digitally signed. Checksums verify package integrity; they are not encryption or a substitute for a publisher signature.
