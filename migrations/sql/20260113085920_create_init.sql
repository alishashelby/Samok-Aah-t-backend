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

CREATE TABLE IF NOT EXISTS admins (
    admin_id BIGSERIAL PRIMARY KEY,
    auth_id BIGINT NOT NULL UNIQUE REFERENCES auth(auth_id) ON DELETE CASCADE,
    permissions json NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
    user_id BIGSERIAL PRIMARY KEY,
    name VARCHAR(30) NOT NULL,
    auth_id BIGINT NOT NULL UNIQUE REFERENCES auth(auth_id) ON DELETE CASCADE,
    birth_date DATE,
    is_verified BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS model_services (
    model_service_id BIGSERIAL PRIMARY KEY,
    model_id BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    title VARCHAR(50) NOT NULL,
    description VARCHAR(1000),
    price DECIMAL(9,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

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

CREATE TABLE IF NOT EXISTS orders (
    order_id BIGSERIAL PRIMARY KEY,
    booking_id BIGINT NOT NULL UNIQUE REFERENCES bookings(booking_id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL CHECK (
        status IN ('CONFIRMED', 'IN_TRANSIT', 'COMPLETED', 'CANCELLED')
    ),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE INDEX idx_admins_auth_id ON admins(auth_id);

CREATE INDEX idx_users_auth_id ON users(auth_id);

CREATE INDEX idx_model_services_model_id ON model_services(model_id);
CREATE INDEX idx_model_services_model_id_active ON model_services(model_id, is_active);
CREATE INDEX idx_model_services_active ON model_services(is_active);

CREATE INDEX idx_slots_model_id ON slots(model_id);
CREATE INDEX idx_slots_model_id_start_time ON slots(model_id, start_time);
CREATE INDEX idx_slots_model_id_time_range ON slots(model_id, start_time, end_time);

CREATE INDEX idx_bookings_model_service_id ON bookings(model_service_id);
CREATE INDEX idx_bookings_client_id ON bookings(client_id);
CREATE INDEX idx_bookings_slot_id ON bookings(slot_id);
CREATE INDEX idx_bookings_status ON bookings(status);

CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_bookings_client_id_booking_id ON bookings(client_id, booking_id);

CREATE OR REPLACE FUNCTION expire_booking_update_slot() RETURNS trigger AS $$
BEGIN
    IF NEW.status = 'EXPIRED' THEN
UPDATE slots
SET status = 'AVAILABLE'
WHERE slot_id = NEW.slot_id
  AND status = 'RESERVED';
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_booking_expired
    AFTER UPDATE OF status ON bookings
    FOR EACH ROW
    WHEN (NEW.status = 'EXPIRED')
    EXECUTE FUNCTION expire_booking_update_slot();

CREATE OR REPLACE FUNCTION update_order_in_transit() RETURNS trigger AS $$
DECLARE
slot_start TIMESTAMP WITH TIME ZONE;
BEGIN
SELECT s.start_time INTO slot_start
FROM bookings b
    JOIN slots s ON b.slot_id = s.slot_id
WHERE b.booking_id = NEW.booking_id;

IF NEW.status = 'CONFIRMED' AND slot_start <= now() THEN
   UPDATE orders
   SET status = 'IN_TRANSIT'
   WHERE order_id = NEW.order_id;
END IF;

RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_order_in_transit
    AFTER INSERT OR UPDATE ON orders
        FOR EACH ROW
        EXECUTE FUNCTION update_order_in_transit();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS model_services;
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
