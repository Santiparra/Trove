# Light novels, enriched and organized

Trove now gives each light novel its own library entry, with independent volumes and a complete catalog workflow powered by RanobeDB. Reading progress stays personal and local, while the detail page brings together the information needed to choose and manage your next volume.

## A library built around novels

- Track completed volumes and optional totals without creating separate library cards for every volume.
- Record historical progress without inventing reading dates. Unknown quantities and totals remain supported.
- Start and finish volumes explicitly; finishing presents the next volume without starting it automatically.
- Keep each volume's own notes, edition details, ownership, favorites, synopsis, and reading dates.
- Treat labels such as **6.5** and parts separately from completed counts. Every volume in the reading sequence counts once, and catalog shortcuts use its identity instead of arithmetic on its printed label.

## Find, review, and link

- Search your local novels and RanobeDB from the same search box, then open a shared detail page directly.
- Browse authors, volumes, and publications, selecting language and format inside the profile.
- Add a new novel with its reading state, or link an existing manual record without replacing personal progress or notes.
- Recognize already-linked novels and review publication changes instead of creating duplicates.
- Use Hardcover to fill missing information only when a publication's ISBN, language, format, and classification are coherent. Open Library remains part of Books, not the novel workflow.
- Disambiguate novels and manga behind the scenes: no `(Light Novel)` suffix is required in your search. Community tags are not treated as definitive medium classification.

## A clearer everyday view

- A compact novel dashboard brings together library access, current reads, up-next covers, and recent volume completions.
- Lists include cover thumbnails; volume galleries retain future catalog entries as you browse.
- Quick progress arrows update completion states without changing reading dates. Deletion and completion use application dialogs.
- Management actions are grouped, with catalog operations in **Identificadores y fuentes** above the volume gallery.
- Long descriptions expand only when needed. Web sources can be opened or copied directly.
- Missing Hardcover setup is highlighted, with a link to the official key page.

## Local data and considerate loading

- Metadata is saved for offline use; local-only mode also blocks RanobeDB requests and remote cover downloads.
- Profiles load progressively with skeletons. Requests are paced, repeated book details are briefly reused, and images load lazily where appropriate.
- Backups retain novel/volume identities, progress, visible labels, and catalog snapshots. A background novel refresh cannot overwrite a restoration completed while it was loading.

## Documentation and scope

Public documentation is now in English, with a short [Spanish getting-started and usage guide](https://github.com/Santiparra/Trove/blob/main/docs/es/guia.md). The application interface remains in Spanish. Previously published notes retain their original language.

**Upgrading from v0.1.7:** the earlier standalone light-novel volume records are not converted or deleted. They remain stored and backed up, but do not appear as grouped novels in the new view; create new novel entries for this model. Books data remains separate. Backups and local storage have advanced, so older application versions may not understand newer data.

Provider keys, cached cover bytes, and profile preferences are still excluded from backups. Ratings, classification-based reading filters, advanced copies/reread management, and external-service imports are not part of this release.

[Complete user guide](https://github.com/Santiparra/Trove/blob/main/docs/usage.md) · [Backup and recovery](https://github.com/Santiparra/Trove/blob/main/docs/backup-format.md)
