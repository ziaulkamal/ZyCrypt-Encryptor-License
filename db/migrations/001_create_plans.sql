CREATE TABLE IF NOT EXISTS license_plans (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    slug        VARCHAR(50)  NOT NULL UNIQUE,
    site_limit  INT          NOT NULL DEFAULT 1,
    duration    VARCHAR(20)  NOT NULL DEFAULT '1y',
    price       DECIMAL(12,2) NOT NULL DEFAULT 0,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);
