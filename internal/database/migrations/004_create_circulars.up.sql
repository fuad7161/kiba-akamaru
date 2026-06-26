CREATE TABLE circulars (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id          VARCHAR(100),
    source               VARCHAR(60) NOT NULL,
    source_url           TEXT,

    title                VARCHAR(500) NOT NULL,
    title_bn             VARCHAR(500),
    organization_id      INTEGER REFERENCES organizations(id) ON DELETE SET NULL,
    organization_name    VARCHAR(255) NOT NULL,
    category_id          INTEGER REFERENCES categories(id) ON DELETE SET NULL,

    vacancy              INTEGER,
    job_type             VARCHAR(50) DEFAULT 'permanent',
    gender               VARCHAR(20),
    age_min              INTEGER,
    age_max              INTEGER,
    age_note             TEXT,
    education_level      VARCHAR(100),
    education_detail     TEXT,
    experience_years     INTEGER,
    experience_note      TEXT,

    salary_min           DECIMAL(12,2),
    salary_max           DECIMAL(12,2),
    salary_grade         VARCHAR(20),
    salary_display       VARCHAR(200),

    location             VARCHAR(255) DEFAULT 'Bangladesh',
    district             VARCHAR(60),
    division             VARCHAR(60),

    published_date       DATE NOT NULL,
    application_deadline DATE,
    exam_date            DATE,

    apply_url            TEXT,
    apply_via            VARCHAR(60),
    teletalk_code        VARCHAR(50),

    description          TEXT,
    requirements         TEXT,
    circular_image_url   TEXT,
    circular_pdf_url     TEXT,

    status               VARCHAR(30) DEFAULT 'active',
    is_featured          BOOLEAN DEFAULT FALSE,
    is_verified          BOOLEAN DEFAULT FALSE,
    view_count           INTEGER DEFAULT 0,
    content_hash         VARCHAR(64),

    created_at           TIMESTAMPTZ DEFAULT NOW(),
    updated_at           TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_circulars_status    ON circulars(status);
CREATE INDEX idx_circulars_deadline  ON circulars(application_deadline);
CREATE INDEX idx_circulars_published ON circulars(published_date DESC);
CREATE INDEX idx_circulars_category  ON circulars(category_id);
CREATE INDEX idx_circulars_org       ON circulars(organization_id);
CREATE INDEX idx_circulars_hash      ON circulars(content_hash);
CREATE UNIQUE INDEX uq_circular_hash ON circulars(content_hash);

CREATE INDEX idx_circulars_fts ON circulars
  USING gin(to_tsvector('english',
    coalesce(title,'') || ' ' || coalesce(organization_name,'')));

CREATE INDEX idx_circulars_title_trgm ON circulars USING gin(title gin_trgm_ops);

CREATE TRIGGER set_timestamp BEFORE UPDATE ON circulars
  FOR EACH ROW EXECUTE FUNCTION trigger_set_timestamp();
