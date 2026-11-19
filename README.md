# Trove

A local multimedia library and personal activity journal for **Windows x64**. Your reading history and personal records belong to the local library; external catalogs are optional.

**Español:** [Cómo empezar y usar Trove](docs/es/guia.md).

Trove is a personal project, publicly available for anyone who finds it useful or wants to contribute. The application currently uses a Spanish interface. Public documentation is maintained in English, with a short Spanish guide reviewed when preparing PRs to `main`.

## Available features

- **Books:** unified search across the local library and Hardcover, with Open Library providing complementary metadata for a safely identified edition.
- A shared local/catalog detail page with covers, descriptions, edition selection, and author/series navigation.
- Manual or catalog-based entry creation, linking to existing records, and saved bibliographic metadata that does not replace personal notes or reading history.
- **Light novels:** one library entry per novel, with independent volume records, editable completed/total progress, optional totals, and separate start/finish actions. Volumes retain their own edition details, synopsis, notes, ownership, favorites, and reading dates. Remote catalog integration is not available yet.
- Reading dashboards, notes, favorites, ownership, and quick status changes. Selecting a status when adding a book creates the entry directly.
- Provider health information, manual diagnostics, and automatic pauses after failures or rate limits.
- A per-profile **local-only mode** that blocks catalog requests and cover downloads while retaining local records, metadata, and cached covers.
- Success notifications that float above the page without shifting its content.
- Automatic SQLite + JSON backups and **Actual / Anterior** (current/previous) restoration with timestamps, validation, and a pre-restore safety copy.
- Separate development, test, and everyday-use profiles.

Hardcover requires a catalog access key; Open Library does not. Manual entry creation and personal library management work without a key or connection. See the [user guide](docs/usage.md).

Other media sections are under construction. External-service imports and ratings are not implemented. Lists, searches, series, and operations are scoped to the selected medium.

## Download and run

Verified releases are available from [GitHub Releases](https://github.com/Santiparra/Trove/releases). See the [release history](docs/releases/README.md#release-history) for user-facing changes, including the Books and Light Novels release, **v0.1.7**.

1. Download the **Windows x64** ZIP from a release.
2. Extract it into a writable folder.
3. Open `Trove.exe`.

Runtime requirement: [Microsoft WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/), normally available on modern Windows. Go, Node, a SQLite server, and a Trove account are not required to run the app. The portable package is not digitally signed.

To update, close Trove and preserve the existing `data/` folder while replacing the application files with the new release. Alternatively, use a new folder and restore a backup. Each release documents its changes and relevant limitations.

Releases use `data/habitual/` next to the executable. PR artifacts use a separate test profile. Copying only the executable does not transfer a library: preserve the data or restore a backup.

## Data and recovery

Choose a backup folder from **Respaldo y restauración**. Trove retains two complete SQLite and JSON generations. A Drive/Dropbox client can synchronize that folder; the backup indicator confirms local files, not cloud upload.

The restore selector shows **Actual** and **Anterior** with local timestamps. You can also select a backup file manually. Credentials, cached covers, and profile preferences, including local-only mode, are excluded from exports.

The documented JSON format preserves identities and relationships independently of SQLite. Restoration replaces the selected profile's library; it does not merge libraries or import another service's data. Database and backup files are not encrypted by Trove.

## Development and contributions

Install the tools listed in the [development guide](docs/development.md), then run:

```powershell
git clone https://github.com/Santiparra/Trove.git
cd Trove
go mod download
npm --prefix frontend ci
wails3 task dev
```

Work is integrated into `dev`. PRs targeting `main` run the Windows checks. Merging to `main` starts automatic publication once those checks pass.

English documentation is the day-to-day source of truth. Update the short [Spanish guide](docs/es/guia.md) when preparing a release PR, rather than duplicating every technical document. Published release notes retain their original language; new release notes are written in English.

## Documentation

| Topic | Reference |
| --- | --- |
| Usage and local storage | [User guide](docs/usage.md) |
| Getting started and everyday usage in Spanish | [Guía breve en español](docs/es/guia.md) |
| Tools and commands | [Development](docs/development.md) |
| Responsibilities and boundaries | [Architecture](docs/architecture.md) |
| Portable recovery | [Backup format](docs/backup-format.md) |
| Verifiable behavior | [Contracts and testing](docs/testing.md) |
| Publication and artifacts | [Releases](docs/releases/README.md) |
| Privacy boundaries | [Security](docs/security.md) |

## License

[MIT](LICENSE).
