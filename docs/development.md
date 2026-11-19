# Development on Windows x64

## Tools

| Tool | Verified reference |
| --- | --- |
| Go | 1.27.4; the module requires 1.26 or later |
| Node.js / npm | Node 26.13.0 and its bundled npm |
| Wails CLI | 3.0.0 |
| Git | A stable release |
| WebView2 Runtime | Required to run the native window |

Official downloads: [Go](https://go.dev/dl/), [Node](https://nodejs.org/), and [WebView2](https://developer.microsoft.com/microsoft-edge/webview2/).

After installing Go, run from any directory:

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0
node --version
npm --version
wails3 doctor
```

Default PATH entries: `C:\Program Files\Go\bin` and `%USERPROFILE%\go\bin`. Open a new terminal after changing PATH; restart the editor if its terminal still has the old environment.

SQLite is embedded through a pure-Go driver without CGo. React, Vite, TypeScript, and frontend tooling are local dependencies, not global installations.

## Setup and execution

```powershell
go mod download
npm --prefix frontend ci
wails3 task dev
```

Vite listens on `127.0.0.1:9245`. The frontend supports hot reload, and Go changes restart the backend. Stop with Ctrl+C. Development data is separate from test and everyday-use profiles.

## Verification

```powershell
powershell -NoProfile -File tools/verify.ps1
```

This runs Go and frontend tests. `-CheckBindings` regenerates the Wails contract and requires it to match committed files; use it after committing binding changes. This is not an end-to-end test of a browser or native window.

```powershell
wails3 task bindings
npm --prefix frontend run format
wails3 task test
```

The script keeps temporary directories under `.local/`. Tests use synthetic data, not personal libraries or real provider APIs.

## Building

Quick local test build:

```powershell
wails3 task build
.\bin\Trove-test.exe
```

An identified package with license, instructions, metadata, and checksums:

```powershell
powershell -NoProfile -File tools/build.ps1 -Version development -Profile test -BuildID local
powershell -NoProfile -File tools/verify-without-git.ps1
```

Output is written to `artifacts/portable/`. Building does not require the directory to be a Git repository. Revision is injected explicitly when available; builds use `-buildvcs=false`. The no-Git contract rebuilds the executable from a source copy with generated frontend assets and tests both executable identities from empty directories.

Application version, SQLite schema version, and JSON backup version are independent. Scripts do not increment version files or create commits.

## Integration and documentation

`dev` is the working integration branch. Only PRs targeting `main` and pushes to `main` run GitHub Actions. A push to `dev` alone does not start a workflow.

The Windows check validates contracts, frontend, bindings, and the portable package. PR artifacts use the `test` profile. After merging, verification runs again on the integrated commit; the `habitual` package is published from that run's artifact without rebuilding in the publication job.

Maintain public documentation in English during everyday development. The interface remains in Spanish; preserve actual UI labels in user instructions. The short [Spanish guide](es/guia.md) is reviewed when preparing a PR to `main`, rather than mirrored on every development change. It covers getting started and everyday use, not every technical detail. Internal planning, notes, and development discussions remain in Spanish.

Each release gets a new English notes document in `docs/releases/`, selected by `current.txt`. Verification checks that the document exists and has content; publication uses it as the release body. Actions assigns the actual version, which is not predicted in the document. Previously published notes retain their original content and language. See [preparing release notes](releases/README.md#preparing-release-notes).

See [contracts](testing.md), [architecture](architecture.md), and [releases](releases/README.md).
