package entity

import "time"

type BookingStatus string

const (
	BookingPending   BookingStatus = "PENDING"
	BookingApproved  BookingStatus = "APPROVED"
	BookingRejected  BookingStatus = "REJECTED"
	BookingCancelled BookingStatus = "CANCELLED"
	BookingExpired   BookingStatus = "EXPIRED"
)

type Booking struct {
	ID             int64
	ClientID       int64
	ModelServiceID int64
	SlotID         int64
	Address        Address
	Status         BookingStatus
	ExpiresAt      time.Time
	CreatedAt      time.Time
}

func NewBooking(clientID, modelServiceID, slotID int64, address Address, ttl time.Duration) *Booking {
	return &Booking{
		ClientID:       clientID,
		ModelServiceID: modelServiceID,
		SlotID:         slotID,
		Address:        address,
		Status:         BookingPending,
		ExpiresAt:      time.Now().Add(ttl),
		CreatedAt:      time.Now(),
	}
}

type Address struct {
	Street    string `json:"street"`
	House     int    `json:"house"`
	Apartment int    `json:"apartment"`
	Entrance  int    `json:"entrance"`
	Floor     int    `json:"floor"`
	Comment   string `json:"comment"`
}

func NewAddress(street string, house int, apartment, entrance, floor int, comment string) Address {
	return Address{
		Street:    street,
		House:     house,
		Apartment: apartment,
		Entrance:  entrance,
		Floor:     floor,
		Comment:   comment,
	}
}

func (b Booking) IsExpired(now time.Time) bool {
	return b.Status == BookingPending && now.After(b.ExpiresAt)
}

func (b Booking) CanBeApproved() bool {
	return b.Status == BookingPending
}

func (b Booking) CanBeRejected() bool {
	return b.Status == BookingPending
}

func (b Booking) CanBeCancelledByClient() bool {
	return b.Status == BookingPending
}
