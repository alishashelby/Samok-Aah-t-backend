package handler

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/authorized"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/models"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
	"github.com/go-playground/validator/v10"
)

type OrderService interface {
	GetModelOrders(ctx context.Context, page, limit *int64) ([]*entity.Order, error)
	CancelOrderByModel(ctx context.Context, orderID int64) (*entity.Order, error)
	CompleteOrder(ctx context.Context, orderID int64) (*entity.Order, error)
	CancelOrderByClient(ctx context.Context, orderID int64) (*entity.Order, error)
}

type OrderHandler struct {
	orderService OrderService
	logger       pkg.Logger
	validate     *validator.Validate
}

func NewOrderHandler(orderService OrderService, logger pkg.Logger) OrderHandler {
	return OrderHandler{
		orderService: orderService,
		logger:       logger,
		validate:     validator.New(),
	}
}

func (h *OrderHandler) GetModelOrders(ctx context.Context,
	request authorized.GetModelOrdersRequestObject) (authorized.GetModelOrdersResponseObject, error) {

	h.logger.Info(ctx, "OrderHandler.GetModelOrders")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	orders, err := h.orderService.GetModelOrders(ctx, request.Params.Page, request.Params.Limit)
	if err != nil {
		return nil, err
	}

	res := make(authorized.GetModelOrders200JSONResponse, len(orders))
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

func (h *OrderHandler) CancelOrderByModel(ctx context.Context,
	request authorized.PatchModelOrdersIdCancelRequestObject,
) (authorized.PatchModelOrdersIdCancelResponseObject, error) {

	h.logger.Info(ctx, "OrderHandler.CancelOrderByModel")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.orderService.CancelOrderByModel(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return authorized.PatchModelOrdersIdCancel200JSONResponse{
		Id:        res.ID,
		BookingID: res.BookingID,
		Status:    models.OrderStatus(res.Status),
		CreatedAt: res.CreatedAt,
	}, nil
}

func (h *OrderHandler) CompleteOrder(ctx context.Context,
	request authorized.PatchModelOrdersIdCompleteRequestObject,
) (authorized.PatchModelOrdersIdCompleteResponseObject, error) {

	h.logger.Info(ctx, "OrderHandler.CompleteOrder")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.orderService.CompleteOrder(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return authorized.PatchModelOrdersIdComplete200JSONResponse{
		Id:        res.ID,
		BookingID: res.BookingID,
		Status:    models.OrderStatus(res.Status),
		CreatedAt: res.CreatedAt,
	}, nil
}

func (h *OrderHandler) CancelOrderByClient(ctx context.Context,
	request authorized.PatchClientOrdersIdCancelRequestObject,
) (authorized.PatchClientOrdersIdCancelResponseObject, error) {

	h.logger.Info(ctx, "OrderHandler.CancelOrderByClient")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.orderService.CancelOrderByClient(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	return authorized.PatchClientOrdersIdCancel200JSONResponse{
		Id:        res.ID,
		BookingID: res.BookingID,
		Status:    models.OrderStatus(res.Status),
		CreatedAt: res.CreatedAt,
	}, nil
}
