CREATE TABLE hike_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    points INTEGER NOT NULL
);

INSERT INTO hike_types (name, points)
VALUES
    ('Однодневный хайк / выезд', 1),
    ('Поход 2-3 дня', 2),
    ('Треккинг 4-6 дней', 3),
    ('Большой треккинг / кэмп 7+ дней', 4),
    ('Восхождение 3000-4000 м', 3),
    ('Восхождение 4000-5000 м', 4),
    ('Восхождение 5000+ м', 5),
    ('Волонтёрство / помощь на мероприятии AktivHike', 1);

ALTER TABLE hikes
    ADD COLUMN hike_type_id INTEGER REFERENCES hike_types(id) ON DELETE RESTRICT;

CREATE INDEX idx_hikes_hike_type_id ON hikes (hike_type_id);
