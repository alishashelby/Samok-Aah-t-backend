package mapping

import (
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/models"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
)

func ToGeneratedAddress(a entity.Address) models.Address {
	return models.Address{
		Street:    a.Street,
		House:     a.House,
		Apartment: &a.Apartment,
		Entrance:  &a.Entrance,
		Floor:     &a.Floor,
		Comment:   &a.Comment,
	}
}
