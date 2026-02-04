-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    user_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL,
    auth_id BIGINT NOT NULL UNIQUE REFERENCES auth(auth_id) ON DELETE CASCADE,
    birth_date DATE,
    is_verified BOOLEAN NOT NULL DEFAULT false
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
