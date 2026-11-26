ALTER TABLE reading_cycles ADD COLUMN is_current INTEGER NOT NULL DEFAULT 1;
ALTER TABLE reading_cycles ADD COLUMN position INTEGER NOT NULL DEFAULT 0;
ALTER TABLE reading_cycles ADD COLUMN details_json TEXT NOT NULL DEFAULT '{}';
CREATE UNIQUE INDEX current_cycle_edition ON reading_cycles(edition_id) WHERE is_current=1;
