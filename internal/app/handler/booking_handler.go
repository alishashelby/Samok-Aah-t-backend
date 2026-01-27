package handler

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/authorized"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/models"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/mapping"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
	"github.com/go-playground/validator/v10"
)

type BookingService interface {
	CreateBooking(ctx context.Context, modelServiceID,
		slotID int64, street string, house int, apartment, entrance, floor *int, comment *string) (*entity.Booking, error)
	ApproveBooking(ctx context.Context, bookingID int64) (*entity.Booking, error)
	RejectBooking(ctx context.Context, bookingID int64) (*entity.Booking, error)
	CancelBookingByClient(ctx context.Context, bookingID int64) (*entity.Booking, error)
}

type BookingHandler struct {
	bookingService BookingService
	logger         pkg.Logger
	validate       *validator.Validate
}

func NewBookingHandler(bookingService BookingService, logger pkg.Logger) *BookingHandler {
	return &BookingHandler{
		bookingService: bookingService,
		logger:         logger,
		validate:       validator.New(),
	}
}

func (h *BookingHandler) CreateBooking(ctx context.Context,
	request authorized.PostClientBookingsRequestObject) (authorized.PostClientBookingsResponseObject, error) {

	h.logger.Info(ctx, "BookingHandler.CreateBooking")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.bookingService.CreateBooking(ctx, request.Body.ModelServiceID,
		request.Body.SlotID, request.Body.Address.Street, request.Body.Address.House,
		request.Body.Address.Apartment, request.Body.Address.Entrance, request.Body.Address.Floor,
		request.Body.Address.Comment)
	if err != nil {
		return nil, err
	}

	return &authorized.PostClientBookings201JSONResponse{
		Id:             res.ID,
		ModelServiceID: res.ModelServiceID,
		ClientID:       res.ClientID,
		SlotID:         res.SlotID,
		Address:        mapping.ToGeneratedAddress(res.Address),
		Status:         models.BookingStatus(res.Status),
		ExpiresAt:      res.ExpiresAt,
		CreatedAt:      res.CreatedAt,
	}, nil
}

func (h *BookingHandler) CancelBookingByClient(ctx context.Context,
	request authorized.PatchClientBookingsIdCancelRequestObject,
) (authorized.PatchClientBookingsIdCancelResponseObject, error) {

	h.logger.Info(ctx, "BookingHandler.CancelBookingByClient")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.bookingService.CancelBookingByClient(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return &authorized.PatchClientBookingsIdCancel200JSONResponse{
		Id:             res.ID,
		ModelServiceID: res.ModelServiceID,
		ClientID:       res.ClientID,
		SlotID:         res.SlotID,
		Address:        mapping.ToGeneratedAddress(res.Address),
		Status:         models.BookingStatus(res.Status),
		ExpiresAt:      res.ExpiresAt,
		CreatedAt:      res.CreatedAt,
	}, nil
}

func (h *BookingHandler) ApproveBooking(ctx context.Context,
	request authorized.PatchModelBookingsIdApproveRequestObject,
) (authorized.PatchModelBookingsIdApproveResponseObject, error) {

	h.logger.Info(ctx, "BookingHandler.ApproveBooking")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.bookingService.ApproveBooking(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return &authorized.PatchModelBookingsIdApprove200JSONResponse{
		Id:             res.ID,
		ModelServiceID: res.ModelServiceID,
		ClientID:       res.ClientID,
		SlotID:         res.SlotID,
		Address:        mapping.ToGeneratedAddress(res.Address),
		Status:         models.BookingStatus(res.Status),
		ExpiresAt:      res.ExpiresAt,
		CreatedAt:      res.CreatedAt,
	}, nil
}

func (h *BookingHandler) RejectBooking(ctx context.Context,
	request authorized.PatchModelBookingsIdRejectRequestObject,
) (authorized.PatchModelBookingsIdRejectResponseObject, error) {

	h.logger.Info(ctx, "BookingHandler.RejectBooking")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.bookingService.RejectBooking(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return &authorized.PatchModelBookingsIdReject200JSONResponse{
		Id:             res.ID,
		ModelServiceID: res.ModelServiceID,
		ClientID:       res.ClientID,
		SlotID:         res.SlotID,
		Address:        mapping.ToGeneratedAddress(res.Address),
		Status:         models.BookingStatus(res.Status),
		ExpiresAt:      res.ExpiresAt,
		CreatedAt:      res.CreatedAt,
	}, nil
}
