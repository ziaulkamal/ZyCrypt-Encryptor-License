CREATE TABLE IF NOT EXISTS license_audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_id  UUID         NOT NULL REFERENCES licenses(id) ON DELETE CASCADE,
    event       VARCHAR(50)  NOT NULL,
    domain      VARCHAR(255),
    ip_address  INET,
    pkg_version VARCHAR(20),
    payload     JSONB,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_license_id ON license_audit_logs(license_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event ON license_audit_logs(event);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON license_audit_logs(created_at);
