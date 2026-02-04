-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS model_services (
    model_service_id BIGSERIAL PRIMARY KEY,
    model_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    title VARCHAR(50) NOT NULL,
    description VARCHAR(1000),
    price DECIMAL(9,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS model_services;
-- +goose StatementEnd
