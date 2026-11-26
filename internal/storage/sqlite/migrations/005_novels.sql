CREATE TABLE novels (
    id TEXT PRIMARY KEY,
    data_json TEXT NOT NULL
);
CREATE TABLE novel_volumes (
    id TEXT PRIMARY KEY,
    novel_id TEXT NOT NULL REFERENCES novels(id) ON DELETE CASCADE,
    number INTEGER NOT NULL CHECK(number > 0),
    data_json TEXT NOT NULL,
    UNIQUE(novel_id, number)
);
