ALTER TABLE achievements
    ADD COLUMN has_value BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE telegram_users_to_achievements
    ADD COLUMN value INTEGER,
    ADD CONSTRAINT telegram_users_to_achievements_value_positive
        CHECK (value IS NULL OR value > 0);

UPDATE achievements
SET has_value = TRUE,
    description = 'Количество сезонов участия.'
WHERE name = 'SEASONS';
