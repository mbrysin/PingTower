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

-- CREATE TABLE daily_stat (
--     id                  INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
--     site_id             INTEGER NOT NULL REFERENCES sites(id) ON DELETE RESTRICT,
--     stat_date           DATE NOT NULL,
--     checks              INTEGER NOT NULL DEFAULT 0 CHECK (checks >= 0),
--     failed_checks       INTEGER NOT NULL DEFAULT 0 CHECK (failed_checks >= 0),
--     success_percent     NUMERIC(5,2)
--         CHECK (success_percent IS NULL OR success_percent BETWEEN 0 AND 100),
--     last_check          TIMESTAMPTZ,
--     avg_response_time_ms NUMERIC(10,2),
--     min_response_time_ms INTEGER,
--     max_response_time_ms INTEGER,
--
--     CONSTRAINT daily_stat_site_date_uq UNIQUE (site_id, stat_date),
--     CONSTRAINT daily_stat_failed_le_checks CHECK (failed_checks <= checks),
--     CONSTRAINT daily_stat_percent_when_checks CHECK (checks > 0 OR success_percent IS NULL)
-- );

-- +goose Down
DROP TABLE IF EXISTS daily_stat;
-- DROP TABLE IF EXISTS stat;
