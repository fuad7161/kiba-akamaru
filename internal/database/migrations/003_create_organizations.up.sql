CREATE TABLE organizations (
    id             SERIAL PRIMARY KEY,
    name           VARCHAR(255) NOT NULL UNIQUE,
    name_bn        VARCHAR(255),
    type           VARCHAR(60),
    website        TEXT,
    logo_url       TEXT,
    apply_base_url TEXT,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_org_name ON organizations USING gin(name gin_trgm_ops);
