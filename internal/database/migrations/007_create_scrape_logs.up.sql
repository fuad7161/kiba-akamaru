CREATE TABLE scrape_logs (
    id            SERIAL PRIMARY KEY,
    source        VARCHAR(60) NOT NULL,
    started_at    TIMESTAMPTZ DEFAULT NOW(),
    finished_at   TIMESTAMPTZ,
    status        VARCHAR(20) DEFAULT 'running',
    total_fetched INTEGER DEFAULT 0,
    new_inserted  INTEGER DEFAULT 0,
    updated       INTEGER DEFAULT 0,
    skipped       INTEGER DEFAULT 0,
    error_message TEXT,
    meta          JSONB
);
