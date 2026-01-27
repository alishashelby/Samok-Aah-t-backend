package handler

import (
	"errors"
	"net/http"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/models"
	errors2 "github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/go-playground/validator/v10"
)

type Error struct {
	Status int
	Code   models.ErrorResponseCode
}

type ErrorMapper struct {
	registry map[error]Error
}

func NewErrorMapper() *ErrorMapper {
	return &ErrorMapper{
		registry: map[error]Error{
			errors2.ErrEmailExists:                 {http.StatusConflict, models.EMAILALREADYEXISTS},
			errors2.ErrAuthWithEmailDoesNotExists:  {http.StatusUnauthorized, models.INVALIDCREDENTIALS},
			errors2.ErrInvalidPasswordOrEmail:      {http.StatusUnauthorized, models.INVALIDCREDENTIALS},
			errors2.ErrUserNotFound:                {http.StatusNotFound, models.NOTFOUND},
			errors2.ErrUnauthorized:                {http.StatusUnauthorized, models.UNAUTHORIZED},
			errors2.ErrNotVerifiedModel:            {http.StatusForbidden, models.FORBIDDEN},
			errors2.ErrNotVerifiedClient:           {http.StatusForbidden, models.FORBIDDEN},
			errors2.ErrNotAModel:                   {http.StatusForbidden, models.NOTAMODEL},
			errors2.ErrNotAdmin:                    {http.StatusForbidden, models.NOTADMIN},
			errors2.ErrNotClient:                   {http.StatusForbidden, models.NOTCLIENT},
			errors2.ErrBookingNotFound:             {http.StatusNotFound, models.BOOKINGNOTFOUND},
			errors2.ErrOrderNotFound:               {http.StatusNotFound, models.ORDERNOTFOUND},
			errors2.ErrAdminNotFound:               {http.StatusNotFound, models.NOTFOUND},
			errors2.ErrClientIsNotOwnerOfBooking:   {http.StatusForbidden, models.FORBIDDEN},
			errors2.ErrClientIsNotOwnerOfOrder:     {http.StatusForbidden, models.FORBIDDEN},
			errors2.ErrBookingExpired:              {http.StatusConflict, models.BOOKINGEXPIRED},
			errors2.ErrBookingAlreadyProcessed:     {http.StatusConflict, models.BOOKINGALREADYPROCESSED},
			errors2.ErrInvalidBookingState:         {http.StatusConflict, models.INVALIDBOOKINGSTATE},
			errors2.ErrCannotCancelOrderNow:        {http.StatusConflict, models.CANNOTCANCELORDER},
			errors2.ErrCannotCompleteOrder:         {http.StatusConflict, models.CANNOTCOMPLETEORDER},
			errors2.ErrServiceIsNotFound:           {http.StatusNotFound, models.SERVICENOTFOUND},
			errors2.ErrModelIsNotAnOwnerOfService:  {http.StatusForbidden, models.NOTSERVICEOWNER},
			errors2.ErrServiceIsNotActive:          {http.StatusConflict, models.SERVICENOTACTIVE},
			errors2.ErrSlotOverlap:                 {http.StatusConflict, models.SLOTOVERLAP},
			errors2.ErrInvalidSlotStatusTransition: {http.StatusConflict, models.INVALIDSLOTSTATUSTRANSITION},
			errors2.ErrSlotNotAvailable:            {http.StatusConflict, models.SLOTNOTAVAILABLE},
			errors2.ErrModelIsNotAnOwnerOfSlot:     {http.StatusForbidden, models.NOTSLOTOWNER},
			errors2.ErrIncorrectSlotTime:           {http.StatusBadRequest, models.INCORRECTSLOTTIME},
			errors2.ErrInvalidPrice:                {http.StatusBadRequest, models.INVALIDPRICE},
			errors2.ErrDescriptionTooLong:          {http.StatusBadRequest, models.DESCRIPTIONTOOLONG},
			errors2.ErrSlotIsNotFound:              {http.StatusNotFound, models.SLOTNOTFOUND},
			errors2.ErrIsNotAnAdult:                {http.StatusBadRequest, models.USERISNOTANADULT},
		},
	}
}

func (e *ErrorMapper) MapError(err error) (status int, code models.ErrorResponseCode, msg string) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		return http.StatusBadRequest, models.VALIDATIONERROR, err.Error()
	}

	for i, registeredError := range e.registry {
		if errors.Is(err, i) {
			return registeredError.Status, registeredError.Code, i.Error()
		}
	}

	return http.StatusInternalServerError, models.INTERNALERROR, err.Error()
}
