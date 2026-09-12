DROP INDEX IF EXISTS idx_hikes_hike_type_id;

ALTER TABLE hikes
    DROP COLUMN IF EXISTS hike_type_id;

DROP TABLE IF EXISTS hike_types;
