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
	openapitypes "github.com/oapi-codegen/runtime/types"
)

type AdminService interface {
	Create(ctx context.Context,
		newAdminAuthID int64, permissions map[string]bool) (*entity.Admin, error)
	GetAdminByID(ctx context.Context, id int64) (*entity.Admin, error)
	UpdateAdmin(ctx context.Context,
		updatingAdminAuthID int64, permissions map[string]bool) (*entity.Admin, error)
	VerifyUser(ctx context.Context, userID int64) (*entity.User, error)
	GetBookingByID(ctx context.Context, bookingID int64) (*entity.Booking, error)
	UpdateBookingStatus(ctx context.Context, bookingID int64, status entity.BookingStatus) (*entity.Booking, error)
	GetOrderByID(ctx context.Context, orderID int64) (*entity.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status entity.OrderStatus) (*entity.Order, error)
	GetAllUsers(ctx context.Context, page, limit *int64) ([]*entity.User, error)
	GetAllBookings(ctx context.Context, page, limit *int64) ([]*entity.Booking, error)
	GetAllOrders(ctx context.Context, page, limit *int64) ([]*entity.Order, error)
}

type AdminHandler struct {
	service  AdminService
	logger   pkg.Logger
	validate *validator.Validate
}

func NewAdminHandler(service AdminService, logger pkg.Logger) *AdminHandler {
	return &AdminHandler{
		service:  service,
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *AdminHandler) CreateAdmin(ctx context.Context,
	request authorized.PostAdminRequestObject) (authorized.PostAdminResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.CreateAdmin")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	admin, err := h.service.Create(ctx, request.Body.AuthId, request.Body.Permissions)
	if err != nil {
		return nil, err
	}

	return authorized.PostAdmin201JSONResponse{
		Id:          admin.ID,
		AuthID:      admin.AuthID,
		Permissions: admin.Permissions,
	}, nil
}

func (h *AdminHandler) GetAdminById(ctx context.Context,
	request authorized.GetAdminIdRequestObject) (authorized.GetAdminIdResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.GetAdminById")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	admin, err := h.service.GetAdminByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return authorized.GetAdminId200JSONResponse{
		Id:          admin.ID,
		AuthID:      admin.AuthID,
		Permissions: admin.Permissions,
	}, nil
}

func (h *AdminHandler) UpdateAdmin(ctx context.Context,
	request authorized.PatchAdminIdRequestObject) (authorized.PatchAdminIdResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.UpdateAdmin")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	admin, err := h.service.UpdateAdmin(
		ctx, request.Id, request.Body.Permissions)
	if err != nil {
		return nil, err
	}

	return authorized.PatchAdminId200JSONResponse{
		Id:          admin.ID,
		AuthID:      admin.AuthID,
		Permissions: admin.Permissions,
	}, nil
}

func (h *AdminHandler) GetAllUsers(ctx context.Context,
	request authorized.GetAdminUsersRequestObject) (authorized.GetAdminUsersResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.GetAllUsers")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	users, err := h.service.GetAllUsers(ctx, request.Params.Page, request.Params.Limit)
	if err != nil {
		return nil, err
	}

	res := make(authorized.GetAdminUsers200JSONResponse, len(users))
	for i, u := range users {
		res[i] = models.UserResponse{
			Id:   u.ID,
			Name: u.Name,
			BirthDate: openapitypes.Date{
				Time: u.BirthDate,
			},
			IsVerified: u.IsVerified,
		}
	}

	return res, nil
}

func (h *AdminHandler) VerifyUser(
	ctx context.Context,
	request authorized.PatchAdminUsersIdVerifyRequestObject,
) (authorized.PatchAdminUsersIdVerifyResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.VerifyUser")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	user, err := h.service.VerifyUser(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return authorized.PatchAdminUsersIdVerify200JSONResponse{
		Id:   user.ID,
		Name: user.Name,
		BirthDate: openapitypes.Date{
			Time: user.BirthDate,
		},
		IsVerified: user.IsVerified,
	}, nil
}

func (h *AdminHandler) GetAllBookings(
	ctx context.Context,
	request authorized.GetAdminBookingsRequestObject,
) (authorized.GetAdminBookingsResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.GetAllBookings")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	bookings, err := h.service.GetAllBookings(ctx, request.Params.Page, request.Params.Limit)
	if err != nil {
		return nil, err
	}

	res := make(authorized.GetAdminBookings200JSONResponse, len(bookings))
	for i, b := range bookings {
		res[i] = models.BookingResponse{
			Id:             b.ID,
			ModelServiceID: b.ModelServiceID,
			ClientID:       b.ClientID,
			SlotID:         b.SlotID,
			Address:        mapping.ToGeneratedAddress(b.Address),
			Status:         models.BookingStatus(b.Status),
			ExpiresAt:      b.ExpiresAt,
			CreatedAt:      b.CreatedAt,
		}
	}

	return res, nil
}

func (h *AdminHandler) UpdateBookingStatus(
	ctx context.Context,
	request authorized.PatchAdminBookingsIdStatusRequestObject,
) (authorized.PatchAdminBookingsIdStatusResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.UpdateBookingStatus")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.service.UpdateBookingStatus(
		ctx, request.Id, entity.BookingStatus(request.Body.Status))
	if err != nil {
		return nil, err
	}

	return authorized.PatchAdminBookingsIdStatus200JSONResponse{
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

func (h *AdminHandler) GetAllOrders(
	ctx context.Context,
	request authorized.GetAdminOrdersRequestObject,
) (authorized.GetAdminOrdersResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.GetAllOrders")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	orders, err := h.service.GetAllOrders(ctx, request.Params.Page, request.Params.Limit)
	if err != nil {
		return nil, err
	}

	res := make(authorized.GetAdminOrders200JSONResponse, len(orders))
	for i, o := range orders {
		res[i] = models.OrderResponse{
			Id:        o.ID,
			BookingID: o.BookingID,
			Status:    models.OrderStatus(o.Status),
			CreatedAt: o.CreatedAt,
		}
	}

	return res, nil
}

func (h *AdminHandler) UpdateOrderStatus(
	ctx context.Context,
	request authorized.PatchAdminOrdersIdStatusRequestObject,
) (authorized.PatchAdminOrdersIdStatusResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.UpdateOrderStatus")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.service.UpdateOrderStatus(ctx,
		request.Id, entity.OrderStatus(request.Body.Status))
	if err != nil {
		return nil, err
	}

	return authorized.PatchAdminOrdersIdStatus200JSONResponse{
		Id:        res.ID,
		BookingID: res.BookingID,
		Status:    models.OrderStatus(res.Status),
		CreatedAt: res.CreatedAt,
	}, nil
}

func (h *AdminHandler) GetBookingByID(ctx context.Context,
	request authorized.GetAdminBookingsIdRequestObject) (authorized.GetAdminBookingsIdResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.GetBookingByID")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.service.GetBookingByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return authorized.GetAdminBookingsId200JSONResponse{
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

func (h *AdminHandler) GetOrderByID(ctx context.Context,
	request authorized.GetAdminOrdersIdRequestObject) (authorized.GetAdminOrdersIdResponseObject, error) {

	h.logger.Info(ctx, "AdminHandler.GetOrderByID")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.service.GetOrderByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return authorized.GetAdminOrdersId200JSONResponse{
		Id:        res.ID,
		BookingID: res.BookingID,
		Status:    models.OrderStatus(res.Status),
		CreatedAt: res.CreatedAt,
	}, nil
}
