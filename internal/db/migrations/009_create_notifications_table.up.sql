CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,

    recipient_tg_user_id BIGINT NOT NULL,
    type VARCHAR(100) NOT NULL,

    entity_type VARCHAR(50),
    entity_id BIGINT,

    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK(status IN('pending', 'processing', 'sent', 'failed')),

    attempts INT NOT NULL DEFAULT 0,
    last_error TEXT,

    last_attempt_at TIMESTAMPTZ,
    next_retry_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ
);

CREATE INDEX notifications_pending_idx
    ON notifications (created_at, id)
    WHERE status = 'pending';
