-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders (
    order_id BIGSERIAL PRIMARY KEY,
    booking_id BIGINT NOT NULL UNIQUE REFERENCES bookings(booking_id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL CHECK (
        status IN ('CONFIRMED', 'IN_TRANSIT', 'COMPLETED', 'CANCELLED')
    ),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
