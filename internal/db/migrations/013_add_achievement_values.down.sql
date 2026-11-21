UPDATE achievements
SET description = 'Количество сезонов участия: II, III, V и далее.'
WHERE name = 'SEASONS';

ALTER TABLE telegram_users_to_achievements
    DROP COLUMN value;

ALTER TABLE achievements
    DROP COLUMN has_value;
