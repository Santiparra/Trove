# Releases and artifacts

## Release history

| Release | User-facing changes |
| --- | --- |
| Next release; version assigned by Actions | [English documentation and a Spanish quick guide](english-documentation.md) |
| [v0.1.7](https://github.com/Santiparra/Trove/releases/tag/v0.1.7) | [Books catalog and Light Novels library](catalogo-libros-y-novelas.md) — original published notes in Spanish |
| [v0.1.4](https://github.com/Santiparra/Trove/releases/tag/v0.1.4) | Earlier release; original notes remain on GitHub |

Each release document describes user-visible changes rather than listing commits. Documents remain in this folder as history. See [GitHub Releases](https://github.com/Santiparra/Trove/releases) for exact versions and publication dates. Published notes retain their original content and language; new release notes are written in English.

## Preparing release notes

1. Create a new English document, such as `docs/releases/english-documentation.md`, describing features, improvements, and relevant limitations. Preserve documents from earlier releases.
2. Put its filename in `docs/releases/current.txt`, without a path or directory prefix. That file selects the release body and does not contain a version number.
3. Add the document to this history and update the README summary and affected English guides. When preparing the following release, identify the actual published version in the previous row and add a new row.
4. Review the short [Spanish guide](../es/guia.md) for the release's setup and usage changes. It is maintained at PR preparation time, not translated on every development edit. Avoid a second technical documentation tree or parallel release history.
5. Include these files in the PR to `main`. Verification rejects missing, empty, or untitled notes. Before merging, review the Windows check and artifact for the PR's latest commit.

The publication job uses the selected document as the release body through `--notes-file`, including when resuming a draft. It does not replace notes for an already-published valid release. It does not use GitHub-generated notes or depend on copying the PR body or merge message.

PR titles and bodies describe changes for review; merge titles and bodies become Git history. They may summarize the same changes, but **the selected document is the source of public release notes**. The release title remains `Trove v<version>`.

## Events

| Event | Result |
| --- | --- |
| Push to `dev` | No remote automation |
| PR targeting `main` | Windows validation and portable artifact with the `test` profile |
| Push to `main` after integration | Validation, `habitual` artifact, and automatic publication on success |

Merging to `main` approves publication. Required branch checks prevent integration if a required contract fails. The latest PR artifact lets you check the distribution contract that will also be verified during publication.

Integration uses merge commits to preserve the relationship between the long-lived branches; `dev` is not deleted after merging. Before starting the next work block, update `dev` from `main` so checks use an up-to-date base.

Updating `dev` while it has an open PR to `main` validates that PR again through `pull_request`. There is no push-to-`dev` workflow and no publication from that update.

## Downloading a PR build

An artifact is the output of a GitHub Actions run, not an installation or public release:

1. Open the PR and its **Windows x64** check.
2. Open the Actions run summary.
3. Under **Artifacts**, download **Trove-windows-x64**.
4. Extract the entire ZIP and open `Trove.exe`.

This artifact uses `data/test/`. It includes `build-info.json`, queryable version/revision, instructions, and checksums. WebView2 Runtime is required to open the window; Go, Node, and an installed SQLite server are not. Downloading artifacts may require signing in to GitHub.

## Version and contents

Integrations into `main` receive `0.1.<run number>` and tag `v0.1.<number>`. PRs receive `0.0.0-pr.<PR>.<run>` without publication. Numbers may skip and do not describe schema changes. The PR and subsequent push to `main` are separate runs, so the next release is not necessarily the previous patch number plus one. Do not predict a version in the notes.

Version and SHA are embedded before testing the executable. The publication job downloads that artifact, verifies identity/checksums, and packages its files without rebuilding. Releases contain the Windows x64 ZIP and `SHA256SUMS.txt`.

The workflow creates a draft release, uploads assets, then publishes it. Retrying the same run reuses its version: an already-published release for that commit is not overwritten; a tag collision with another commit fails. Personal PATs and automatic versioning commits are not required; `GITHUB_TOKEN` has permissions limited to the publication job.

## Compatibility

Application version, SQLite schema, and JSON format are independent. An older app may not open a database migrated by a newer one; recovery requires compatible data. Release notes should describe compatibility changes when applicable.

Packages are built from explicit files. They exclude databases, backups, personal preferences, credentials, and internal documentation. Actions configuration or rules may be maintained in later PRs; normal runs do not require new manual publication authorizations.
