ALTER TABLE telegram_users
    ADD COLUMN initial_hikes_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN initial_distance_km NUMERIC(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN initial_elevation_gain_m INTEGER NOT NULL DEFAULT 0;
