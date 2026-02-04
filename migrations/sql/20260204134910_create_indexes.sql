-- +goose Up
-- +goose StatementBegin
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_admins_auth_id;
DROP INDEX IF EXISTS idx_users_auth_id;
DROP INDEX IF EXISTS idx_model_services_model_id;
DROP INDEX IF EXISTS idx_model_services_model_id_active;
DROP INDEX IF EXISTS idx_model_services_active;
DROP INDEX IF EXISTS idx_slots_model_id;
DROP INDEX IF EXISTS idx_slots_model_id_start_time;
DROP INDEX IF EXISTS idx_slots_model_id_time_range;
DROP INDEX IF EXISTS idx_bookings_model_service_id;
DROP INDEX IF EXISTS idx_bookings_client_id;
DROP INDEX IF EXISTS idx_bookings_slot_id;
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_bookings_client_id_booking_id;
-- +goose StatementEnd
