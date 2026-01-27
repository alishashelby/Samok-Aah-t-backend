package entity

import "time"

type SlotStatus string

const (
	SlotAvailable SlotStatus = "AVAILABLE"
	SlotReserved  SlotStatus = "RESERVED"
	SlotBooked    SlotStatus = "BOOKED"
	SlotDisabled  SlotStatus = "DISABLED"
)

type Slot struct {
	ID        int64
	ModelID   int64
	StartTime time.Time
	EndTime   time.Time
	Status    SlotStatus
	CreatedAt time.Time
}

func NewSlot(modelID int64, start time.Time, end time.Time) *Slot {
	return &Slot{
		ModelID:   modelID,
		StartTime: start,
		EndTime:   end,
		Status:    SlotAvailable,
	}
}

func (s *Slot) IsCorrectTransition(next SlotStatus) bool {
	switch s.Status {
	case SlotAvailable:
		return next == SlotDisabled || next == SlotReserved
	case SlotReserved:
		return next == SlotBooked || next == SlotAvailable
	case SlotBooked:
		return false
	case SlotDisabled:
		return next == SlotAvailable
	default:
		return false
	}
}

func (s *Slot) IsAvailable() bool {
	return s.Status == SlotAvailable
}
