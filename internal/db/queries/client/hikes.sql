-- name: ListActualHikes :many
SELECT 
    id, 
    title_ru, 
    starts_at, 
    ends_at, 
    price_gel,
    distance_km,
    elevation_gain_m
FROM hikes
WHERE is_published = true AND ends_at >= now()
ORDER BY starts_at ASC
LIMIT $1 OFFSET $2;

-- name: GetHikeCard :one
SELECT 
    id, 
    title_ru, 
    starts_at, 
    ends_at, 
    price_gel,
    distance_km,
    elevation_gain_m,
    preview_ru,
    image_path
FROM hikes
WHERE id = $1 AND is_published = true; 

-- name: GetHikeDetails :one
SELECT 
    id, 
    title_ru,
    description_ru
FROM hikes
WHERE id = $1 AND is_published = true; 

-- name: CreateBooking :one
INSERT INTO bookings (hike_id, user_id, status)
VALUES ($1, $2, $3)
ON CONFLICT (hike_id, user_id) DO NOTHING
RETURNING id;

-- name: GetBookingByID :one
SELECT id, hike_id, user_id, status, taken_by_admin_id, taken_at
FROM bookings WHERE id = $1;

-- name: TakeBookingInProgress :one
UPDATE bookings
SET
    status = $2,
    taken_by_admin_id = $3,
    taken_at = now()
WHERE id = $1 AND status = sqlc.arg(expected_status)
RETURNING id;

-- name: UpdateBookingStatus :one
UPDATE bookings
SET
    status = $2
WHERE id = $1 AND status = $3
RETURNING id;

-- name: GetTelegramUserByID :one
SELECT id, tg_user_id, tg_username, full_name,
       initial_hikes_count, initial_distance_km, initial_elevation_gain_m
FROM telegram_users
WHERE id = $1;

-- name: UpsertTelegramUser :one
WITH claimed AS (
    UPDATE telegram_users
    SET tg_user_id = $1, tg_username = $2, full_name = $3, lang = $4
    WHERE tg_user_id < 0
      AND $2::text IS NOT NULL
      AND lower(trim(leading '@' from tg_username)) = lower(trim(leading '@' from $2::text))
      AND NOT EXISTS (SELECT 1 FROM telegram_users WHERE tg_user_id = $1)
    RETURNING id
), upserted AS (
    INSERT INTO telegram_users (tg_user_id, tg_username, full_name, lang)
    SELECT $1, $2, $3, $4 WHERE NOT EXISTS (SELECT 1 FROM claimed)
    ON CONFLICT (tg_user_id) DO UPDATE SET
        tg_username = EXCLUDED.tg_username,
        full_name = EXCLUDED.full_name,
        lang = EXCLUDED.lang
    RETURNING id
)
SELECT id FROM claimed UNION ALL SELECT id FROM upserted LIMIT 1;

-- name: CreateAdminIfNotExists :exec
INSERT INTO admins (id)
VALUES ($1)
ON CONFLICT DO NOTHING;
