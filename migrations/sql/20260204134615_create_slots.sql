-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS slots (
    slot_id BIGSERIAL PRIMARY KEY,
    model_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
        status VARCHAR(10) NOT NULL CHECK (
            status IN ('AVAILABLE', 'RESERVED', 'BOOKED', 'DISABLED')
    ),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS slots;
-- +goose StatementEnd
