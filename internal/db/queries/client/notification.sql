-- name: ClaimPendingNotifications :many
WITH pending AS (
    SELECT id
    FROM notifications
    WHERE status = 'pending'
    AND (
        next_retry_at IS NULL
        OR next_retry_at <= NOW()
    )
    ORDER BY created_at, id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg('limit')
)
UPDATE notifications AS n
SET
    status = 'processing',
    attempts = n.attempts + 1,
    last_attempt_at = NOW()
FROM pending
WHERE n.id = pending.id
RETURNING
    n.id,
    n.recipient_tg_user_id,
    n.type,
    n.payload,
    n.attempts,
    n.created_at;


-- name: MarkNotificationSent :execrows
UPDATE notifications
SET
    status = 'sent',
    sent_at = NOW(),
    last_error = NULL,
    next_retry_at = NULL

WHERE id = sqlc.arg('id')
    AND status = 'processing';

-- name: MarkNotificationFailed :execrows
UPDATE notifications
SET
    status = 'failed',
    last_error = sqlc.arg('last_error'),
    failed_at = NOW(),
    next_retry_at = NULL
WHERE id = sqlc.arg('id')
    AND status = 'processing';

-- name: ReleaseNotificationForRetry :execrows
UPDATE notifications
SET
    status = 'pending',
    last_error = sqlc.arg('last_error'),
    next_retry_at = sqlc.arg('next_retry_at')
WHERE id = sqlc.arg('id')
    AND status = 'processing';