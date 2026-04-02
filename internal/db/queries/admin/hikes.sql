-- =========================================
-- HIKES
-- =========================================

-- name: CreateHike :one
INSERT INTO hikes (
    title_ru, 
    title_en, 
    description_ru, 
    description_en,
    starts_at, 
    ends_at, 
    photo_file_id,
    price_gel,
    distance_km,
    elevation_gain_m,
    is_published
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
RETURNING id;

-- name: UpdateHike :one
UPDATE hikes SET
    title_ru       = $2,
    title_en       = $3,
    description_ru = $4,
    description_en = $5,
    starts_at      = $6,
    ends_at        = $7,
    photo_file_id  = $8,
    is_published   = $9,
    distance_km    = $10,
    elevation_gain_m = $11,
    updated_at       = $12
WHERE id = $1
RETURNING *;

-- name: UpdateImagePath :exec
UPDATE hikes SET image_path = $2 WHERE id = $1;

-- name: DeleteHike :exec
DELETE FROM hikes WHERE id = $1;

-- name: GetHikeByID :one
SELECT * FROM hikes WHERE id = $1;

-- name: ListHikes :many
SELECT 
    id, 
    title_ru, 
    description_ru,
    title_en,
    description_en,
    starts_at, 
    ends_at, 
    is_published, 
    created_at 
FROM 
    hikes 
ORDER BY is_published DESC, created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListActualHikes :many
SELECT id, title_ru, starts_at, ends_at, is_published
FROM hikes
WHERE is_published = true AND ends_at >= now()
ORDER BY starts_at ASC
LIMIT $1 OFFSET $2;

-- name: SetPublished :exec
UPDATE hikes
SET is_published = $1, updated_at = now()
WHERE id = $2;

-- =========================================
-- TELEGRAM USERS
-- =========================================

-- name: UpsertTelegramUser :one
INSERT INTO telegram_users (tg_user_id, tg_username, full_name, lang)
VALUES ($1, $2, $3, $4)
ON CONFLICT (tg_user_id)
DO UPDATE SET
    tg_username = EXCLUDED.tg_username,
    full_name   = EXCLUDED.full_name,
    lang        = EXCLUDED.lang
RETURNING id;

-- =========================================
-- BOOKINGS
-- =========================================

-- name: GetBookingByID :one
SELECT id, hike_id, user_id, status, taken_by_admin_id
FROM bookings WHERE id = $1;

-- name: UpdateBookingStatus :one
UPDATE bookings
SET status = sqlc.arg(new_status)
WHERE id = $1
RETURNING *;

-- name: ListAdminBookings :many
SELECT
    b.id,
    b.hike_id,
    h.title_ru AS hike_title,
    b.user_id,
    COALESCE(u.full_name, '') AS user_name,
    u.tg_user_id AS user_tg_id,
    b.status,
    b.taken_at,
    b.created_at
FROM bookings b
JOIN hikes h ON h.id = b.hike_id
JOIN telegram_users u ON u.id = b.user_id
WHERE
    b.taken_by_admin_id = $1
    AND b.status IN ('in_progress', 'confirmed')
ORDER BY
    b.created_at DESC;