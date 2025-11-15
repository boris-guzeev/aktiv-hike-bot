-- name: CreateHike :one
INSERT INTO hikes (
    title_ru, 
    title_en, 
    description_ru, 
    description_en,
    starts_at, 
    ends_at, 
    photo_file_id, 
    is_published,
    created_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING *;

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
    updated_at     = $10
WHERE id = $1
RETURNING *;

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