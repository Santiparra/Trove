# Trove backup format

Backup format **version 6** is independent of the SQLite schema (currently 6) and the public application version. Earlier supported formats retain their own data; standalone volume entries are not converted into grouped novels.

## File layout

```text
Chosen folder/
  test/                       # or development / habitual
    generation-<id>/
      trove.sqlite
      trove.json
      complete.json
    generation-<previous-id>/
      trove.sqlite
      trove.json
      complete.json
```

The two most recent complete, valid generations are retained. `complete.json` contains a manifest format version, generation ID, UTC timestamp, logical content hash, and SHA-256 hashes of both representations. It detects incomplete or altered pairs before they qualify for retention. It is neither a signature nor encryption.

A generation is published after its files are finished. External synchronization may download them separately: **a valid Trove JSON or SQLite file works independently**, without the other representation or manifest. An incomplete synchronization copy is not a completed generation.

## Logical JSON, version 6

Encoding is UTF-8. Root fields:

| Field | Type / meaning |
| --- | --- |
| `format` | Constant string `trove-library` |
| `version` | Integer `6` in new generations |
| `generation` | Text generation identifier, not a work ID |
| `createdAt` | Generation timestamp in UTC, RFC 3339 |
| `catalog` | Logical content described below |

`catalog` always contains the four base arrays below, with empty arrays for empty collections. Format 4 adds an optional `novels` array, omitted when empty. IDs are strings, not array positions or engine-specific auto-increment sequences.

| Collection | Fields |
| --- | --- |
| `series` | `id`, `name`, `medium` |
| `works` | `id`, `medium`, `title`, `authors` (string array), `notes`, `addedOn`, `favorite`; optional `seriesId` and `position` |
| `editions` | `id`, `workId`, `language`, `format`, `identifier`, `owned`; optional `identifiers` and `catalog` |
| `cycles` | `id`, `editionId`, `status`, `startedOn`, `finishedOn` |
| `novels` | `id`, `title`, `authors`, `status`, nullable `total`/`completed`, `notes`, `favorite`, `addedOn`, `volumes`, and optional `catalog` |

Format 6 also preserves optional novel `statusChangedAt` (a UTC RFC 3339 activity timestamp) and volume `label` (display text). Catalog main-book references and publications can carry their own original display labels. The integer `number` is a local reading position, not a printed label: `6.5` counts as one publication without becoming a fractional completed count. Status activity supports dashboard ordering and does not replace or fabricate reading dates.

A grouped novel has its own identity and personal state. Each `volumes` entry contains `id`, positive integer `number`, optional-title text `title`, `description`, `edition`, `cycle`, `notes`, `favorite`, and optional `catalog`. Editions preserve ID, language, format, raw/normalized identifiers, and ownership; cycles preserve ID, status, and known start/completion dates. These nested entities are independent of the base Books collections.

Format 5 adds novel bibliographic snapshots. Novel `catalog` preserves the RanobeDB series ID, names, author references, ordered main-book IDs, original source ordering/titles, fetch time, and language/format preference. Volume `catalog` preserves series/book/release IDs, local reading position, publication identity, verified ISBN, edition metadata, description, cover URL, field provenance, warnings, and optional verified Hardcover category. A book-level snapshot may have no selected release and cannot claim a publication ISBN/language/format. The source hierarchy, type evidence, order, dates, identifiers, and source references are validated without querying providers. Personal fields remain separate; cover bytes and credentials are still excluded.

`total = null` means an unknown total, and `completed = null` means an unknown completed quantity. Zero is a known count, not unknown. Known completion history includes records numbered 1 through n; unknown quantity does not generate records. Known totals, progress, unique volume numbers/identities, dates, and relationships are validated. The `volumes` collection must be present, even as an empty array. A read novel with a known total has N/N progress; a read novel with unknown quantity can have no volumes at all.

Supported `medium` values are `books` and `light-novels`. Status values are `to-read`, `up-next`, `reading`, `read`, `paused`, and `did-not-finish`. Edition formats are an empty string (unknown), `paperback`, `hardcover`, `digital`, `audio`, and `other`.

Library/reading dates use `YYYY-MM-DD`, or an empty string when unknown; they have no time or time zone. `favorite` and `owned` are booleans independent of reading state. `position` is nonnegative and only meaningful with a series. `identifier` is optional edition text and is not assumed to always be an ISBN.

`identifiers` preserves scheme, normalized value, source, and a derived-equivalence flag. `catalog` holds the bibliographic snapshot: primary provider, RFC 3339 fetch time, remote work/edition, identifiers, description, cover URL, per-field sources, and warnings. It does not replace personal fields. Cover bytes and credentials are excluded; URLs and metadata are included.

Relationships: `works.seriesId` → `series.id`; `editions.workId` → `works.id`; `cycles.editionId` → `editions.id`. A work and its series must have the same medium. Every work must have an edition and reading record. All editions and cycles are preserved, not only those displayed by a frontend view. Normalized series names must be unique within a medium but may repeat across media.

Validation checks format identity, version, duplicates, references, dates, enums, and chronology without inventing missing dates. The catalog and all four collections must be present, even as empty arrays: omitted content is not interpreted as an empty library. Unknown fields or media are rejected instead of silently discarded. New data requires explicit format/compatibility and test updates.

Abbreviated **synthetic** example of an older work with unknown reading dates:

```json
{
  "format": "trove-library",
  "version": 6,
  "generation": "synthetic-generation",
  "createdAt": "2026-01-01T00:00:00Z",
  "catalog": {
    "series": [],
    "works": [{"id":"work-example","medium":"books","title":"An example story","authors":["Fictional author"],"notes":"","addedOn":"2020-01-01","favorite":false}],
    "editions": [{"id":"edition-example","workId":"work-example","language":"en","format":"paperback","identifier":"","owned":true}],
    "cycles": [{"id":"cycle-example","editionId":"edition-example","status":"read","startedOn":"","finishedOn":""}]
  }
}
```

## Earlier format support

Format 1 contains only `works.medium = books`; series lack `medium` and are interpreted as Books. Format 2 adds both reading media without bibliographic snapshots; format 3 adds book catalog metadata. Grouped novels are introduced in format 4, their catalog snapshots in format 5, and display labels/status-activity data in format 6. IDs, relationships, unknown dates, and the original logical hash of existing supported formats are preserved. Restoring earlier standalone volumes does not group or convert them.

New backups use format 6. Existing supported older restorations are prepared on the current schema without applying live reading-date behavior. Native snapshots identified as SQLite schemas 1 through 6 with a supported format are accepted. Older application versions may not accept schema 6 or format 6.

Restoring an earlier backup recovers exactly that state: format 1 contains no light novels, and format 2 contains no bibliographic snapshots. Later entries are not retained by merging. The pre-replacement safety copy preserves the complete current state.

## Native backup

`trove.sqlite` is generated from a consistent SQLite snapshot and includes a technical `trove_backup` table identifying its format, version, generation, and timestamp. It is not the active database or a copy of its files taken during writes. Reading it requires SQLite support.

JSON is derived from the completed snapshot, not from a later query against the active library. Both representations therefore describe the same state even if new writes occur while the backup is being prepared.

## Restoration and limits

Regular files up to 256 MiB per representation are accepted. The app prepares a new database, verifies recovered content, and shows the destination profile before replacement. It does not merge libraries. Manual file selection uses native dialogs.

Backup folder configuration lives in `backup-settings.json` alongside profile data and is **not** restored from a backup. Local-only mode lives in `provider-network.json`; it is neither exported nor changed during recovery. Machine paths, tokens, logs, WebView caches, and covers are excluded. Recovery does not query the Internet or apply live reading rules to historical dates.

The restore selector presents the two most recent complete valid generations as **Actual** (current) and **Anterior** (previous), with local dates and times. Selection retains the exact ID, revalidates the manifest and files before preparation, and requires the usual confirmation before replacement. A generation that disappeared, rotated out of the available pair, or was altered is not silently replaced by another. Manually selecting a SQLite or JSON file remains available through the native dialog.

`before-restore-*.sqlite` safety copies remain local for recovery; manually removing them is independent of automatic generation retention. Close other instances before restoring. If replacement fails, the app attempts to reopen the previous library and reports the safety-copy location if intervention is required.
