# Contracts and testing

Verification is organized around observable outcomes, not coverage percentages or function quotas. The table contains **35 core scenarios**, supplemented by catalog, provider, local-only, and interaction contracts below. A scenario may combine acceptance and rejection cases; it is not necessarily one file or test function, or the application's total test count.

## Acceptance contracts

| ID | Outcome / acceptance and boundary | Automated evidence |
| --- | --- | --- |
| C01 | Start an empty library without a provider or required repository | `TestAdmission_C01_EmptyDirectoryStartsOffline` |
| C02 | Separate library data and backup output by profile | `TestAdmission_C02_ProfilesDoNotShareLibraryOrBackupOutput` |
| C03 | Preserve DTO and version semantics across adapters | `TestAdmission_C03_WireDTOsPreserveMeaning` + bindings/TypeScript |
| C04 | Reject invalid input without partial records | `TestAdmission_C04_InvalidWriteDoesNotCreatePartialRecords` |
| C05 | Recover from standalone SQLite without JSON/manifest | `TestAdmission_C05_SQLiteRecoveryIsStandalone` |
| C06 | Recover from standalone JSON without SQLite/manifest | `TestAdmission_C06_JSONRecoveryIsStandalone` |
| C07 | Distinguish external imports from native recovery | `TestAdmission_C07_ExternalFileIsNotARecoveryBackup` |
| C08 | Allow local work when backup configuration fails | `TestAdmission_C08_BackupConfigurationFailureDoesNotRejectLocalWork` |
| C09 | Reject invalid profiles/paths without hidden fallback destinations | `TestAdmission_C09_InvalidProfileOrDataRootHasNoSilentFallback` |
| C10 | Save, edit, and reopen while preserving identity and series | `TestBooksSurviveReopenAndShareSeries` |
| C11 | Roll back failed transactions and preserve the previous edition | `TestSaveRollsBackAllTablesOnFailure`, `TestFailedEditPreservesSavedBookAndSeriesCanBeCleared` |
| C12 | Reject future schemas and migrate an older schema with a safety copy | `TestFutureSchemaIsRejected`, `TestVersionOneMigratesAndHasReadableSafetyCopy` |
| C13 | Respect manual/unknown dates and chronology | `TestValidationKeepsUnknownDatesAndRejectsInvalidInput`, `TestLiveReadingDates`, `TestSaveAppliesLiveDatesButHonorsManualEdits` |
| C14 | Change status without overwriting other fields | `TestQuickStatusChangesOnlyStatusAndUpdatesDashboard` |
| C15 | Order recent completions by known dates and limit previews | `TestDashboardUsesCompletionDatesAndLimitsRecentBooks` |
| C16 | Delete one entry without affecting another and roll back on failure | `TestDeletePreservesOtherBooksAndRollsBackOnFailure` |
| C17 | Keep notes, personal flags, and series ordering independent | `TestPersonalActionsAndSeriesOrdering` |
| C18 | Validate format, relationships, fields, and missing content | `TestValidationRejectsBrokenRelationshipsAndFutureFormats`, `TestJSONRejectsUnknownFieldsAndTrailingData` |
| C19 | Preserve all entities and dates in both backup representations | `TestArchiveRoundTripPreservesIDsAllCyclesAndUnknownDates` |
| C20 | Retain two valid generations and avoid unchanged duplicates | `TestAutomaticGenerationsRetentionFailureAndRestoreBothFormats` |
| C21 | Never replace a valid backup with an incomplete/altered generation | `TestIncompleteGenerationsAreIgnoredAndPreparedTamperingIsRejected` |
| C22 | Freeze prepared recovery and cancel without touching active data | `TestInvalidOrModifiedRestoreNeverTouchesActiveData` |
| C23 | Recover the previous library if replacement fails | `TestInstallRollbackLeavesPreviousLibraryReadable` |
| C24 | Keep destination profile configuration when restoring | `TestRestoreKeepsTargetProfileBackupConfiguration` |
| C25 | Back up after debounce, on shutdown, and after restart | `TestDebounceAndShutdownSavePendingChangesWithoutManualExport` |
| C26 | Wait for accepted writes and reject new ones during shutdown | `TestShutdownWaitsForAcceptedWritesAndBacksThemUp` |
| C27 | Navigate dashboard/states/series/search without implicit saves | `frontend/src/App.test.tsx` |
| C28 | Preserve drafts and saved data after UI errors | `frontend/src/App.test.tsx` |
| C29 | Require confirmation and distinguish recovery success/failure | `frontend/src/backup/BackupPanel.test.tsx` |
| C30 | Keep the generated Go/Wails/TypeScript contract consistent | `tools/verify.ps1 -CheckBindings` and TypeScript compilation |
| C31 | Build without Git and query identity from an empty directory | `tools/verify-without-git.ps1` against sources and portable |
| C32 | Distribute only explicit files with correct version/profile/checksums | `tools/build.ps1`, `tools/package.ps1` |
| C33 | Isolate reads, series, search, and all mutations across Books and Light Novels | `TestAdmission_C33_ReadingMediaAreIsolatedAcrossAllOperations` + UI |
| C34 | Recover both reading media from JSON or SQLite without losing identity | `TestAdmission_C34_MixedLibraryRestoresFromEitherRepresentation` |
| C35 | Accept older backups and migrate without inventing dates or merging libraries | `TestAdmission_C35_LegacyJSONRestoresAsBooksWithoutInventingDates`, `TestSchemaTwoMigrationKeepsBookAndSeriesIdentities`, `TestLegacyDocumentKeepsOriginalContentHash` |

## Running verification

Catalog tests in `internal/catalog`, `internal/providers`, `internal/credentials`, `internal/desktop/catalog_test.go`, and `internal/storage/sqlite/catalog_test.go` cover ISBN/equivalence validation, edition-specific cascading, credential protection, cover safety, preservation of personal data, match detection, and snapshot round trips. They also check exact-edition priority, informative initial selection, and creation with status/series without fabricated dates.

`frontend/src/App.test.tsx` checks local/remote search popovers, direct detail navigation, local recognition before creation, key settings, manual fallback, Light Novels isolation, and preservation of drafts across successful or failed refreshes. Provider tests use synthetic responses and require neither real keys nor external connectivity.

```powershell
powershell -NoProfile -File tools/verify.ps1
powershell -NoProfile -File tools/build.ps1 -Version development -Profile test -BuildID local
powershell -NoProfile -File tools/verify-without-git.ps1
```

The `tests/contracts` admission runner exercises adapters exposed by Wails using real SQLite, fresh directories, and synthetic data. It does not open a window, start external APIs, or use cloud services. UI tests use Vitest/Testing Library and a controlled bridge; generated bindings and consumers are checked against Go separately.

The no-Git build uses sources and generated frontend assets in a copy where Git cannot discover the repository. The portable queries identity with `--version` from an empty directory without opening a library. This is not full graphical automation.

## Remote verification

GitHub Actions runs **Windows x64** only for PRs targeting `main` or pushes to `main`. A passing suite enables integration but does not replace checking real interaction, WebView2, and native dialogs on a PC. PR artifacts allow that validation before merging.

The contract list grows with real capabilities. Ambiguous requirements are resolved before becoming assertions. Negative and recovery cases are added for the risks they protect, not to increase the test count.

## Focused contracts

- `frontend/src/books/BookSearch.test.tsx` uses a controlled clock to verify debounce, cancellation of pending timers, and local-before-provider ordering without real requests. Author/series navigation from local and remote pages and immediate creation on status selection are tested in `frontend/src/App.test.tsx`.
- `internal/providers/health_test.go` covers failure pauses, a single concurrent recovery attempt, queued requests, invalidation on key changes, cancellation, latency excluding local waits, 404/429, bounded checks, incompatible responses, and shutdown cancellation. It uses a controlled clock and synthetic HTTP servers. `frontend/src/catalog/ProviderHealth.test.tsx` covers summaries, distinct states, manual checks, and local-only status refresh.
- `internal/providers/network_test.go` covers profile persistence, blocking both catalogs and cover downloads, offline caching, and cancellation during an in-flight response. `internal/backup/manager_test.go` covers exact current/previous selection, tampering revalidation, and safety copies before replacement. UI tests cover timestamp selection, local-only personal management, and toast expiry/pausing.
- `internal/novels/service_test.go` covers historical 7/? progress, unknown read quantities, start/finish state transitions, known-total completion without N+1, preservation of volume IDs/data/dates, rejected writes, and isolation from Books and earlier standalone volumes. `internal/storage/sqlite/novels_test.go` checks aggregate/volume round trips through JSON and native backups and a nonempty restore preview for a novel with no known volumes. `internal/desktop/novels_test.go` rejects writes during shutdown.
- `frontend/src/novels/NovelWorkspace.test.tsx` checks one card per novel, the X-1 shortcut, unknown quantities, next-volume information without automatic starting, N/N completion, volume editing errors, and preservation of note drafts. App tests exercise the dedicated novel service rather than the older standalone-volume UI.
- `internal/providers/novel_providers_test.go` covers normal-name RanobeDB search, main-book hierarchy, original ordering/partial dates, rejection of manga/partial/compilation releases, exact-ISBN Hardcover supplementation, structured categories versus misleading community tags, and local-only policy. No real credentials or external APIs are required by these tests.
- `internal/novels/catalog_test.go` covers duplicate recognition, manual data preservation, multiple identifier schemes, safe refresh/unlink, and rejection of changed order/ISBN. `internal/desktop/novel_catalog_test.go` checks server-owned selection tokens and rejects background metadata writes across restoration. `internal/storage/sqlite/novels_test.go` also restores catalog snapshots from both JSON and native backups.
- `frontend/src/novels/NovelCatalogFlow.test.tsx` covers local/remote search, direct enriched profiles, creation/linking, existing records, publication selection, author identity navigation, next-volume metadata without starting it, and offline snapshot presentation.

Before tests, `tools/verify.ps1` checks script syntax, local Markdown links, and the release notes selected by `docs/releases/current.txt`. Publication uses those notes and the already-verified artifact; a PR does not publish or create tags. See [releases and artifacts](releases/README.md).
