-- +goose Up
-- +goose StatementBegin
CREATE TABLE sites (
    id          INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    url         TEXT NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE sites;
-- +goose StatementEnd
