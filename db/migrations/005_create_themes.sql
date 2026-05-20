CREATE TABLE IF NOT EXISTS themes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug       VARCHAR(100) UNIQUE NOT NULL,
    name       VARCHAR(255) NOT NULL,
    version    VARCHAR(20)  NOT NULL DEFAULT '1.0.0',
    bundle     BYTEA        NOT NULL,
    checksum   VARCHAR(64)  NOT NULL,
    file_size  BIGINT       NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
