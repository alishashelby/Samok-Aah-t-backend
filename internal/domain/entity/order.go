package entity

import "time"

type OrderStatus string

const (
	OrderConfirmed OrderStatus = "CONFIRMED"
	OrderInTransit OrderStatus = "INTRANSIT"
	OrderCompleted OrderStatus = "COMPLETED"
	OrderCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID        int64
	BookingID int64
	Status    OrderStatus
	CreatedAt time.Time
}

func NewOrder(bookingID int64) *Order {
	return &Order{
		BookingID: bookingID,
		Status:    OrderConfirmed,
	}
}

func (o Order) CanBeCancelled(now time.Time, slotStart time.Time) bool {
	return o.Status == OrderConfirmed && now.Before(slotStart.Add(-24*time.Hour))
}

func (o Order) CanBeCompleted(now time.Time, slotEnd time.Time) bool {
	return o.Status == OrderInTransit && now.After(slotEnd)
}
