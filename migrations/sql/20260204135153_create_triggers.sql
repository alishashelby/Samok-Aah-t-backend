-- +goose Up
-- +goose StatementBegin
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
DROP TRIGGER IF EXISTS trg_booking_expired ON bookings;
DROP TRIGGER IF EXISTS trg_order_in_transit ON orders;

DROP FUNCTION IF EXISTS expire_booking_update_slot();
DROP FUNCTION IF EXISTS update_order_in_transit();
-- +goose StatementEnd
