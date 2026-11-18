ALTER TABLE editions ADD COLUMN identifiers_json TEXT NOT NULL DEFAULT '[]' CHECK(json_valid(identifiers_json));
ALTER TABLE editions ADD COLUMN catalog_json TEXT CHECK(catalog_json IS NULL OR json_valid(catalog_json));
