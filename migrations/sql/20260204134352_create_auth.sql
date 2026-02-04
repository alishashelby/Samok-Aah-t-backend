-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS auth (
    auth_id BIGSERIAL PRIMARY KEY,
    email varchar(320) NOT NULL UNIQUE,
    password_hash varchar(255) NOT NULL,
    role VARCHAR(6) NOT NULL CHECK (
        role IN ('CLIENT', 'MODEL', 'ADMIN')
    ),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth;
-- +goose StatementEnd
