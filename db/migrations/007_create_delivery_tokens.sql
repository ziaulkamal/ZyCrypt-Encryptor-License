CREATE TABLE IF NOT EXISTS delivery_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    token      VARCHAR(64) UNIQUE NOT NULL,
    license_id UUID        NOT NULL REFERENCES licenses(id) ON DELETE CASCADE,
    theme_id   UUID        NOT NULL REFERENCES themes(id)   ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    used       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_delivery_tokens_token ON delivery_tokens (token);
