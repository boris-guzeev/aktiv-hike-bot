-- name: ListActualHikes :many
SELECT id, title_ru, description_ru, starts_at, ends_at
FROM hikes
WHERE is_published = true AND ends_at >= now()
ORDER BY starts_at ASC
LIMIT $1 OFFSET $2;