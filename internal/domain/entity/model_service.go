package entity

import "time"

type ModelService struct {
	ID          int64
	ModelID     int64
	Title       string
	Description string
	Price       float32
	IsActive    bool
	CreatedAt   time.Time
}

func NewModelService(modelID int64, title, description string, price float32) *ModelService {
	return &ModelService{
		ModelID:     modelID,
		Title:       title,
		Description: description,
		Price:       price,
		IsActive:    true,
	}
}
