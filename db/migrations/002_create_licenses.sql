CREATE TABLE IF NOT EXISTS licenses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_key     VARCHAR(40)  NOT NULL UNIQUE,
    plan_id         UUID         NOT NULL REFERENCES license_plans(id),
    customer_name   VARCHAR(150) NOT NULL,
    customer_email  VARCHAR(150) NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'active',
    purchase_date   DATE         NOT NULL DEFAULT CURRENT_DATE,
    expires_at      TIMESTAMP,
    is_lifetime     BOOLEAN      NOT NULL DEFAULT FALSE,
    banned_reason   TEXT,
    notes           TEXT,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_licenses_status ON licenses(status);
CREATE INDEX IF NOT EXISTS idx_licenses_plan_id ON licenses(plan_id);
