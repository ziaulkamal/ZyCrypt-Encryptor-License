CREATE TABLE IF NOT EXISTS license_domains (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_id    UUID         NOT NULL REFERENCES licenses(id) ON DELETE CASCADE,
    domain        VARCHAR(255) NOT NULL,
    is_primary    BOOLEAN      NOT NULL DEFAULT FALSE,
    registered_at TIMESTAMP    NOT NULL DEFAULT NOW(),
    UNIQUE (license_id, domain)
);

CREATE INDEX IF NOT EXISTS idx_license_domains_domain ON license_domains(domain);
CREATE INDEX IF NOT EXISTS idx_license_domains_license_id ON license_domains(license_id);
