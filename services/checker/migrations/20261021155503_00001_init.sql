-- +goose Up
-- +goose StatementBegin
CREATE TABLE check_results (
    id               INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    site_id          INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    checked_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_up            BOOLEAN NOT NULL,
    status_code      INTEGER,
    response_time_ms INTEGER,
    error            TEXT
);

CREATE INDEX idx_check_results_site_checked_at
    ON check_results (site_id, checked_at DESC);

CREATE TABLE site_daily_stats (
    id                INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    site_id           INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    date              DATE NOT NULL,
    status            TEXT NOT NULL DEFAULT 'unknown',
    checks_total      INTEGER NOT NULL DEFAULT 0,
    checks_failed     INTEGER NOT NULL DEFAULT 0,
    uptime_percent    NUMERIC(5, 2),
    last_checked_at   TIMESTAMPTZ,

    UNIQUE (site_id, date),
    CHECK (status IN ('unknown', 'up', 'degraded', 'down')),
    CHECK (checks_total >= 0),
    CHECK (checks_failed >= 0),
    CHECK (checks_failed <= checks_total),
    CHECK (uptime_percent IS NULL OR uptime_percent BETWEEN 0 AND 100)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE check_results;
DROP TABLE site_daily_stats;
-- +goose StatementEnd
