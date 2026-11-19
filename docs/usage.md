# Usage and local data

The application currently uses Spanish UI labels, which are preserved below so you can find the actual controls. For a shorter introduction in Spanish, see [Cómo empezar y usar Trove](es/guia.md).

## Books and Light Novels libraries

Each section opens a dashboard with reading states, current reads, recently finished entries with known dates, and a visual distribution. Both sections share navigation patterns, but their lists, searches, and series are independent. Previews fit the available space; complete lists remain accessible.

Each light novel volume is registered separately and may belong to a series at a given position. Matching series names in Books and Light Novels are not merged automatically. Switching sections does not reclassify or change existing records.

- Search first shows local matches by title, author, series, or identifier. In Books, it adds Hardcover candidates in the same popover when a key is configured, without blocking local matches. Enter or **Ver todos los resultados** opens the full list.
- Selecting a reading state opens its entries. The buttons above the list change the filter.
- The series appears above the title. Local series keep their entries ordered by position, with unnumbered entries last; in Books, they also show catalog candidates searched by series name. A catalog series opens the shared search by name.
- Clicking an author on a local or remote detail page opens the shared search for their books. Remote results require a configured key; local matches remain available without one.
- The detail page allows saving notes and changing favorites, ownership, and status without opening the full editor.
- Only the title is required when creating a book. ISBN, authors, and dates can be added later.
- Light Novels uses **Añadir volumen**, with the same minimal edition and personal fields. Its detail page supports editing, favorites/ownership, notes, and all reading states, including **A continuación**.
- Deletion requires confirmation and removes the entry's associated data, not other entries in its series.

## Optional Books catalog

1. Configure a Hardcover catalog key from **Claves**, next to the backup indicator. Saving it does not query or synchronize your personal provider library.
2. Use the main search by title, author, series, or ISBN. Selecting a result opens the shared detail page with cover, information, and management actions. Opening a work prefers a linked local edition; otherwise, the most informative edition among the provider's returned results is selected. A specific ISBN respects that edition.
3. Use **Elegir otra edición** or **Cambiar edición vinculada** within the detail page to view up to 20 editions and search for another by ISBN. Searching and reviewing do not create library entries.
4. **Agregar · Por leer** adds an entry with one click. Choosing a different status in its dropdown creates it immediately; existing entries also change state when selected. A single bibliographic series is preselected and can be removed; multiple associations require a choice. Adding an already-read book does not invent reading dates.
5. An already-linked edition opens its local record and management actions instead of offering creation. ISBN matches or another edition of the same work require review before linking. Different editions are not merged automatically.

The detail page uses three columns for cover, information, and management when space allows, preserving the image's proportions. Long descriptions expand with **Leer más**. Identifiers and sources are in a collapsible section at the bottom.

Popover search waits for a 140 ms pause before searching locally, then another 350 ms after local results arrive before querying Hardcover. Continued typing cancels pending timers; stale responses do not replace the current query. Typing alone does not cancel requests already sent over the network.

From a detail page, you can link or change an edition, refresh metadata, and unlink it. These actions preserve personal data; a failed refresh keeps the previous snapshot. Saved descriptions and metadata can be read offline. Downloaded covers are reused locally, while unavailable ones display a placeholder. Open Library only supplements missing fields for the selected edition using a safe ISBN match.

Manual entry creation remains available without a key or connection. Light Novels keeps its manual workflow. Keys and cached image bytes are not exported: restoring elsewhere may require configuring the key and downloading covers again.

## Provider health

### Local-only mode

In **Proveedores**, **Modo solo local** disables Hardcover and Open Library requests, health checks, and cover downloads. Go enforces the policy, cancels in-flight requests, and saves the preference per profile so it is respected on restart. It neither removes keys nor treats the intentional block as a provider outage.

The library, saved snapshots, cached covers, and personal management remain available. Undownloaded covers display a placeholder; remote metadata refresh and edition search are disabled. Re-enabling the network allows requests again, subject to the circuit-breaker rules.

To test restoration offline, enable the mode, restore the selected generation, and browse local detail pages. Network settings are excluded from backups and remain unchanged by restoration.

### Indicator and diagnostics

The **Proveedores** indicator, beside keys and backups, shows the last known state of the Hardcover and Open Library APIs. Open it for availability, last request duration/time, and temporary pauses. Missing or rejected credentials are distinguished from service failures. It does not check cover hosts or guarantee that every provider operation works.

- Startup runs a brief background check per API with an 8-second budget, without blocking the library. Hardcover is skipped without a key. Normal requests update the same state.
- **Comprobar ahora** repeats diagnostics at most once every 15 seconds and respects each provider's pause. There is no periodic remote polling. While the window is visible, the indicator refreshes local state every 3 seconds.
- Three consecutive observed failures pause new requests for one minute. Afterwards, one recovery request is admitted: success reopens access; failure starts another pause. Queued requests also respect the circuit.
- HTTP 429 pauses immediately, for at least one minute and respecting `Retry-After` up to 24 hours. Hardcover access rejection blocks further requests until the key is saved again or replaced. Removing it shows **Sin configurar**.
- A missing record (404) does not mark a provider as down. Local cancellations do not count as failures. Measured duration excludes locally imposed pacing waits.

Diagnostics do not change the library or its snapshots. Health state is temporary and excluded from exports. An Open Library problem does not promote it to primary or prevent using saved data.

## Reading states and dates

The available states are **Por leer** (to read), **A continuación** (up next), **Leyendo** (reading), **Leído** (read), **Pausado** (paused), and **Sin terminar** (did not finish). Up next is assigned manually.

Entering Reading records a start date if missing. Entering Read from another state records completion only when a start exists and completion is missing. Pausing and resuming preserve existing dates. Marking an older book as read without a start does not create an artificial completion date.

Dates are editable. Completion cannot precede the start, but both may be the same day. Explicitly setting or clearing a date takes precedence over automatic behavior for that edit. Restoration does not apply live reading-date rules.

Recently finished entries need a completion date for chronological ordering. Entries without one remain in the Read list.

## Profiles

| Profile | Default location |
| --- | --- |
| `habitual` | `data/habitual/` next to the release executable |
| `test` | `data/test/` next to the test executable |
| `development` | `.local/development/` under the execution directory |

The header indicator shows profile, path, version, and revision. Build identity can also be queried without opening data:

```powershell
.\Trove.exe --version
```

An alternate root always receives a profile subdirectory:

```powershell
.\Trove.exe --profile test --data-dir "C:\TroveTests"
```

This example uses `C:\TroveTests\test\`. Do not share an active directory between instances during restoration.

## Operation notifications

Success messages float above the page without shifting its content. They close after 5 seconds, pause while focused or hovered, and can be dismissed manually. Errors remain in the context of the operation.

## Automatic backups

1. Open **Respaldo y restauración**.
2. Choose a folder separate from active data. The profile name is appended.
3. An initial backup is created. Later changes are grouped for two seconds and backed up automatically.

Two complete generations are retained, each containing `trove.sqlite`, `trove.json`, and `complete.json`. Unchanged content does not create a duplicate generation. Normal shutdown waits for accepted writes and completes pending backups. Forced termination may leave changes only in the active database; content is checked again on startup.

Output failures preserve local changes and the last valid copy. Errors are reported and retried. **Al día** confirms local files, not an external synchronization client's state. Backups exclude cover bytes, caches, logs, credentials, and configuration paths.

## Restoration

1. Open **Respaldo y restauración**. The selector lists complete valid generations from the configured folder as **Actual** (current) and **Anterior** (previous), with local dates and times.
2. Choose a generation and click **Revisar respaldo seleccionado**. Preparation revalidates that exact generation. If it changed or is unavailable, the app reports the issue instead of selecting another automatically.
3. Review date, contents, and destination profile, approve the confirmation, and click **Restaurar y reemplazar biblioteca**. **Cambiar selección** discards preparation without replacing data.

You can also use **Seleccionar respaldo para restaurar** to choose a Trove-generated SQLite or JSON file through the native dialog. Either valid representation works independently of the other file and manifest when selected manually. This is not an import from an external service.

Preparation is validated before replacement and expires after ten minutes. A `before-restore-*.sqlite` safety copy is kept in the profile. If replacement fails, reopening the previous library is attempted. Backup output configuration remains tied to the destination profile.

If a damaged database prevents startup, open a new data root and restore there without changing the original:

```powershell
.\Trove.exe --profile test --data-dir "C:\TroveRecovery"
```

Supported formats, limits, and relationships are documented in the [backup format](backup-format.md).

The current local schema is 4. Migration creates a `before-schema-4-*.sqlite` safety copy before adding normalized identifiers and bibliographic snapshots, preserving raw identifiers, identities, and unknown dates. New backups use format 3 and contain both reading media and linked metadata. Formats 1 and 2 can be recovered; format 1 is interpreted as Books. Restoration replaces the entire profile library without merging later entries.
