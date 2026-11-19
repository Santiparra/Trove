# Trove backup format

Backup format **version 3** is independent of the SQLite schema (currently 4) and the public application version. Reading and restoration of format 1 and 2 backups remain supported.

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

## Logical JSON, version 3

Encoding is UTF-8. Root fields:

| Field | Type / meaning |
| --- | --- |
| `format` | Constant string `trove-library` |
| `version` | Integer `3` in new generations |
| `generation` | Text generation identifier, not a work ID |
| `createdAt` | Generation timestamp in UTC, RFC 3339 |
| `catalog` | Logical content described below |

`catalog` contains four arrays. Empty collections are written as empty arrays. IDs are strings, not array positions or engine-specific auto-increment sequences.

| Collection | Fields |
| --- | --- |
| `series` | `id`, `name`, `medium` |
| `works` | `id`, `medium`, `title`, `authors` (string array), `notes`, `addedOn`, `favorite`; optional `seriesId` and `position` |
| `editions` | `id`, `workId`, `language`, `format`, `identifier`, `owned`; optional `identifiers` and `catalog` |
| `cycles` | `id`, `editionId`, `status`, `startedOn`, `finishedOn` |

Supported `medium` values are `books` and `light-novels`. Status values are `to-read`, `up-next`, `reading`, `read`, `paused`, and `did-not-finish`. Edition formats are an empty string (unknown), `paperback`, `hardcover`, `digital`, `audio`, and `other`.

Library/reading dates use `YYYY-MM-DD`, or an empty string when unknown; they have no time or time zone. `favorite` and `owned` are booleans independent of reading state. `position` is nonnegative and only meaningful with a series. `identifier` is optional edition text and is not assumed to always be an ISBN.

`identifiers` preserves scheme, normalized value, source, and a derived-equivalence flag. `catalog` holds the bibliographic snapshot: primary provider, RFC 3339 fetch time, remote work/edition, identifiers, description, cover URL, per-field sources, and warnings. It does not replace personal fields. Cover bytes and credentials are excluded; URLs and metadata are included.

Relationships: `works.seriesId` → `series.id`; `editions.workId` → `works.id`; `cycles.editionId` → `editions.id`. A work and its series must have the same medium. Every work must have an edition and reading record. All editions and cycles are preserved, not only those displayed by a frontend view. Normalized series names must be unique within a medium but may repeat across media.

Validation checks format identity, version, duplicates, references, dates, enums, and chronology without inventing missing dates. The catalog and all four collections must be present, even as empty arrays: omitted content is not interpreted as an empty library. Unknown fields or media are rejected instead of silently discarded. New data requires explicit format/compatibility and test updates.

Abbreviated **synthetic** example of an older work with unknown reading dates:

```json
{
  "format": "trove-library",
  "version": 3,
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

Format 1 contains only `works.medium = books`; series lack `medium` and are interpreted as Books. Format 2 adds both reading media without bibliographic snapshots. IDs, relationships, unknown dates, and the original logical hash are preserved: validating an older backup does not rewrite it as format 3 or invalidate its manifest.

New backups use format 3. Older restorations are prepared on the current schema without applying live reading-date behavior. Native snapshots identified as SQLite schemas 1, 2, 3, or 4 with a supported format are accepted. Older application versions may not accept schema 4 or format 3.

Restoring an earlier backup recovers exactly that state: format 1 contains no light novels, and format 2 contains no bibliographic snapshots. Later entries are not retained by merging. The pre-replacement safety copy preserves the complete current state.

## Native backup

`trove.sqlite` is generated from a consistent SQLite snapshot and includes a technical `trove_backup` table identifying its format, version, generation, and timestamp. It is not the active database or a copy of its files taken during writes. Reading it requires SQLite support.

JSON is derived from the completed snapshot, not from a later query against the active library. Both representations therefore describe the same state even if new writes occur while the backup is being prepared.

## Restoration and limits

Regular files up to 256 MiB per representation are accepted. The app prepares a new database, verifies recovered content, and shows the destination profile before replacement. It does not merge libraries. Manual file selection uses native dialogs.

Backup folder configuration lives in `backup-settings.json` alongside profile data and is **not** restored from a backup. Local-only mode lives in `provider-network.json`; it is neither exported nor changed during recovery. Machine paths, tokens, logs, WebView caches, and covers are excluded. Recovery does not query the Internet or apply live reading rules to historical dates.

The restore selector presents the two most recent complete valid generations as **Actual** (current) and **Anterior** (previous), with local dates and times. Selection retains the exact ID, revalidates the manifest and files before preparation, and requires the usual confirmation before replacement. A generation that disappeared, rotated out of the available pair, or was altered is not silently replaced by another. Manually selecting a SQLite or JSON file remains available through the native dialog.

`before-restore-*.sqlite` safety copies remain local for recovery; manually removing them is independent of automatic generation retention. Close other instances before restoring. If replacement fails, the app attempts to reopen the previous library and reports the safety-copy location if intervention is required.
