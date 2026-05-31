CREATE TABLE outbox_events (
    id             BIGSERIAL PRIMARY KEY,
    event_id       UUID        NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id   BIGINT      NOT NULL,
    event_type     VARCHAR(100) NOT NULL,
    payload        JSONB       NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at   TIMESTAMPTZ DEFAULT NULL,
    retry_count    INT         NOT NULL DEFAULT 0,
    last_error     TEXT        DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_outbox_unpublished
    ON outbox_events (created_at ASC)
    WHERE published_at IS NULL;
