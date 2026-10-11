-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id            INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username      TEXT NOT NULL,
    password_hash TEXT NOT NULL CHECK (password_hash <> ''),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX users_username_lower_key
    ON users (lower(username));
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
