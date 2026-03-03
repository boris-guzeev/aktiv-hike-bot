-- name: ListActualHikes :many
SELECT id, title_ru, description_ru, starts_at, ends_at
FROM hikes
WHERE is_published = true AND ends_at >= now()
ORDER BY starts_at ASC
LIMIT $1 OFFSET $2;

-- name: UpsertTgUser :one
INSERT INTO tg_users (tg_user_id, tg_username, full_name, lang)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tg_user_id)
DO UPDATE SET
    tg_username = EXCLUDED.tg_username,
    full_name   = EXCLUDED.full_name,
    lang        = EXCLUDED.lang
RETURNING id;

-- name: GetHike :one
SELECT id, title_ru, starts_at, ends_at
FROM hikes
WHERE id = $1 AND is_published = true; 

-- name: CreateBookingPending :one
INSERT INTO bookings (hike_id, user_id, status)
VALUES ($1, $2, 'pending')
ON CONFLICT (hike_id, user_id) DO NOTHING
RETURNING id;