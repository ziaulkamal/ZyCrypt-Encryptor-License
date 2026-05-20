CREATE TABLE IF NOT EXISTS plan_themes (
    plan_id  UUID NOT NULL REFERENCES license_plans(id) ON DELETE CASCADE,
    theme_id UUID NOT NULL REFERENCES themes(id)        ON DELETE CASCADE,
    PRIMARY KEY (plan_id, theme_id)
);
