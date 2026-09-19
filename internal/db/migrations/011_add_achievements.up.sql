CREATE TABLE achievements (
    id SMALLSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL
);

CREATE TABLE telegram_users_to_achievements (
    telegram_user_id INT NOT NULL REFERENCES telegram_users(id) ON DELETE CASCADE,
    achievement_id SMALLINT NOT NULL REFERENCES achievements(id),
    PRIMARY KEY (telegram_user_id, achievement_id)
);

INSERT INTO achievements (name, description)
VALUES
    ('4000+', 'Вершина выше 4000 м с AktivHike.'),
    ('5000+', 'Вершина выше 5000 м с AktivHike.'),
    ('TREKKER', 'Не менее 3 многодневных треккингов.'),
    ('EXPLORER', 'Не менее 10 разных маршрутов AktivHike.'),
    ('FASTPACKER', 'Участие в fastpacking-проекте.'),
    ('ALL SEASON', 'Участие и в зимних, и в летних мероприятиях.'),
    ('EXPEDITIONER', 'Участие в зарубежной экспедиции AktivHike.'),
    ('100K / 250K / 500K / 1000K', 'Суммарная дистанция, пройденная с AktivHike.'),
    ('SEASONS', 'Количество сезонов участия: II, III, V и далее.');
