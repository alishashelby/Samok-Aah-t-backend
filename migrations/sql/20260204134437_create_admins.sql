-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS admins (
    admin_id BIGSERIAL PRIMARY KEY,
    auth_id BIGINT NOT NULL UNIQUE REFERENCES auth(auth_id) ON DELETE CASCADE,
    permissions json NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admins;
-- +goose StatementEnd
