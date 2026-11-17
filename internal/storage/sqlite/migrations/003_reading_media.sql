ALTER TABLE books ADD COLUMN medium TEXT NOT NULL DEFAULT 'books' CHECK(medium IN ('books','light-novels'));
ALTER TABLE series ADD COLUMN medium TEXT NOT NULL DEFAULT 'books' CHECK(medium IN ('books','light-novels'));
-- name_key is a private uniqueness key, now namespaced without changing IDs/names.
UPDATE series SET name_key = 'books:' || name_key;
CREATE INDEX books_medium ON books(medium);
CREATE INDEX series_medium ON series(medium);
