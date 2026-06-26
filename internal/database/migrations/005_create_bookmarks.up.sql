CREATE TABLE bookmarks (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    circular_id UUID NOT NULL REFERENCES circulars(id) ON DELETE CASCADE,
    note        TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, circular_id)
);

CREATE INDEX idx_bookmarks_user     ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_circular ON bookmarks(circular_id);
