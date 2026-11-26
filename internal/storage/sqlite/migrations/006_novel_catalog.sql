-- Stable provider identities belong to one local novel/volume, including concurrent writers.
CREATE UNIQUE INDEX novels_ranobe_identity ON novels(json_extract(data_json, '$.catalog.id'))
WHERE json_extract(data_json, '$.catalog.id') IS NOT NULL;
CREATE UNIQUE INDEX novel_volumes_ranobe_identity ON novel_volumes(json_extract(data_json, '$.catalog.bookID'))
WHERE json_extract(data_json, '$.catalog.bookID') IS NOT NULL;
