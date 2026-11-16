CREATE TABLE series (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    name_key TEXT NOT NULL UNIQUE
);
CREATE TABLE books (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    authors_json TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    added_on TEXT NOT NULL,
    series_id TEXT REFERENCES series(id),
    series_position REAL
);
CREATE TABLE editions (
    id TEXT PRIMARY KEY,
    book_id TEXT NOT NULL REFERENCES books(id),
    language TEXT NOT NULL DEFAULT '',
    format TEXT NOT NULL DEFAULT '',
    identifier TEXT NOT NULL DEFAULT ''
);
CREATE INDEX editions_book ON editions(book_id);
CREATE TABLE reading_cycles (
    id TEXT PRIMARY KEY,
    edition_id TEXT NOT NULL REFERENCES editions(id),
    status TEXT NOT NULL,
    started_on TEXT NOT NULL DEFAULT '',
    finished_on TEXT NOT NULL DEFAULT ''
);
CREATE INDEX cycles_edition ON reading_cycles(edition_id);
