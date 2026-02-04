-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bookings (
    booking_id BIGSERIAL PRIMARY KEY,
    client_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    model_service_id BIGINT NOT NULL REFERENCES model_services(model_service_id) ON DELETE CASCADE,
    slot_id BIGINT NOT NULL REFERENCES slots(slot_id) ON DELETE CASCADE,
    address JSONB NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (
        status IN ('PENDING', 'REJECTED', 'APPROVED', 'CANCELLED', 'EXPIRED')
    ),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bookings;
-- +goose StatementEnd
