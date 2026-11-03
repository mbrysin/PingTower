-- +goose Up
CREATE TABLE stat (
    id               INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    site_id          INTEGER NOT NULL REFERENCES sites(id) ON DELETE RESTRICT,
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_up            BOOLEAN NOT NULL,
    status_code      INTEGER,
    response_time_ms INTEGER CHECK (response_time_ms IS NULL OR response_time_ms >= 0),
    error            TEXT
);

CREATE INDEX idx_stat_site_checked_at
    ON stat (site_id, checked_at DESC);

-- +goose Down
DROP TABLE IF EXISTS stat;
