package service

import (
	"context"
	"errors"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/common"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/interfaces"
	metrics2 "github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/metrics"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

type DefaultOrderService struct {
	orderRepo        interfaces.OrderRepository
	bookingRepo      interfaces.BookingRepository
	slotRepo         interfaces.SlotRepository
	userRepo         interfaces.UserRepository
	modelServiceRepo interfaces.ModelServiceRepository
	txManager        database.TxManager
	logger           pkg.Logger
	metrics          *metrics2.Metrics
}

func NewDefaultOrderService(orderRepo interfaces.OrderRepository, bookingRepo interfaces.BookingRepository,
	slotRepo interfaces.SlotRepository, userRepo interfaces.UserRepository, modelServiceRepo interfaces.ModelServiceRepository,
	txManager database.TxManager, logger pkg.Logger, metrics *metrics2.Metrics) *DefaultOrderService {
	return &DefaultOrderService{
		orderRepo:        orderRepo,
		bookingRepo:      bookingRepo,
		slotRepo:         slotRepo,
		userRepo:         userRepo,
		modelServiceRepo: modelServiceRepo,
		txManager:        txManager,
		logger:           logger,
		metrics:          metrics,
	}
}

func (d *DefaultOrderService) GetModelOrders(ctx context.Context, page, limit *int64) ([]*entity.Order, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	res, err := d.orderRepo.GetAllByModelID(ctx, model.ID, entity.NewOptions(common.CheckPagination(page, limit)))
	if err != nil {
		d.logger.Error(ctx, "orders are not found by model id",
			option.Any("model_id", model.ID),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultOrderService) CancelOrderByModel(ctx context.Context, orderID int64) (*entity.Order, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	order, err := d.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "order is not found by id",
				option.Any("order_id", orderID),
				option.Error(service_errors.ErrOrderNotFound))

			return nil, service_errors.ErrOrderNotFound
		}

		d.logger.Error(ctx, "failed to get order by id",
			option.Any("order_id", orderID),
			option.Error(err))

		return nil, err
	}

	booking, err := d.bookingRepo.GetByID(ctx, order.BookingID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "booking is not found by id",
				option.Any("booking_id", order.BookingID),
				option.Error(service_errors.ErrBookingNotFound))

			return nil, service_errors.ErrBookingNotFound
		}

		d.logger.Error(ctx, "failed to get booking by id",
			option.Any("booking_id", order.BookingID),
			option.Error(err))

		return nil, err
	}

	err = d.checkIfModelIsAnOwner(ctx, model.ID, booking.ModelServiceID)
	if err != nil {
		return nil, err
	}

	slot, err := d.slotRepo.GetByID(ctx, booking.SlotID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot is not found by id",
				option.Any("slot_id", booking.SlotID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "failed to get slot by id",
			option.Any("slot_id", booking.SlotID),
			option.Error(err))

		return nil, err
	}

	if !order.CanBeCancelled(time.Now(), slot.StartTime) {
		d.logger.Error(ctx, "order cannot be canceled ",
			option.Any("order_id", order.ID),
			option.Error(service_errors.ErrCannotCancelOrderNow))

		return nil, service_errors.ErrCannotCancelOrderNow
	}

	return d.cancelOrder(ctx, booking, slot, order)
}

func (d *DefaultOrderService) CompleteOrder(ctx context.Context, orderID int64) (*entity.Order, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	order, err := d.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "order is not found by id",
				option.Any("order_id", orderID),
				option.Error(service_errors.ErrOrderNotFound))

			return nil, service_errors.ErrOrderNotFound
		}

		d.logger.Error(ctx, "failed to get order by id",
			option.Any("order_id", orderID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	booking, err := d.bookingRepo.GetByID(ctx, order.BookingID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "booking is not found by id",
				option.Any("booking_id", order.BookingID),
				option.Error(service_errors.ErrBookingNotFound))

			return nil, service_errors.ErrBookingNotFound
		}

		d.logger.Error(ctx, "failed to get booking by id",
			option.Any("booking_id", order.BookingID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	err = d.checkIfModelIsAnOwner(ctx, model.ID, booking.ModelServiceID)
	if err != nil {
		return nil, err
	}

	slot, err := d.slotRepo.GetByID(ctx, booking.SlotID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot is not found by id",
				option.Any("slot_id", booking.SlotID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "failed to get slot by id",
			option.Any("slot_id", booking.SlotID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if !order.CanBeCompleted(time.Now(), slot.EndTime) {
		d.logger.Error(ctx, "order cannot be completed ",
			option.Any("order_id", order.ID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrCannotCompleteOrder))

		return nil, service_errors.ErrCannotCompleteOrder
	}

	order.Status = entity.OrderCompleted
	res, err := d.orderRepo.UpdateStatus(ctx, order)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "order is not found by id",
				option.Any("order_id", order.ID),
				option.Error(service_errors.ErrOrderNotFound))

			return nil, service_errors.ErrOrderNotFound
		}

		d.logger.Error(ctx, "failed to update order status",
			option.Any("order_id", order.ID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	d.metrics.IncCompletedOrders()

	return res, nil
}

func (d *DefaultOrderService) CancelOrderByClient(ctx context.Context, orderID int64) (*entity.Order, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	client, err := d.checkClientRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	order, err := d.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "order is not found by id",
				option.Any("order_id", orderID),
				option.Error(service_errors.ErrOrderNotFound))

			return nil, service_errors.ErrOrderNotFound
		}

		d.logger.Error(ctx, "failed to get order by id",
			option.Any("order_id", orderID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	booking, err := d.bookingRepo.GetByID(ctx, order.BookingID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "booking is not found by id",
				option.Any("booking_id", order.BookingID),
				option.Error(service_errors.ErrBookingNotFound))

			return nil, service_errors.ErrBookingNotFound
		}

		d.logger.Error(ctx, "failed to get booking by id",
			option.Any("booking_id", order.BookingID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if booking.ClientID != client.ID {
		d.logger.Error(ctx, "order is not owned by this client",
			option.Any("booking_id", order.BookingID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrClientIsNotOwnerOfOrder))

		return nil, service_errors.ErrClientIsNotOwnerOfOrder
	}

	slot, err := d.slotRepo.GetByID(ctx, booking.SlotID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot is not found by id",
				option.Any("slot_id", booking.SlotID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "failed to get slot by id",
			option.Any("slot_id", booking.SlotID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if !order.CanBeCancelled(time.Now(), slot.StartTime) {
		d.logger.Error(ctx, "order cannot be cancelled",
			option.Any("order_id", order.ID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrCannotCancelOrderNow))

		return nil, service_errors.ErrCannotCancelOrderNow
	}

	return d.cancelOrder(ctx, booking, slot, order)
}

func (d *DefaultOrderService) checkModelRestrictions(ctx context.Context, authID *int64) (*entity.User, error) {
	role, err := common.GetRoleFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if *role != entity.RoleModel.String() {
		d.logger.Error(ctx, "access denied",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotAdmin))

		return nil, service_errors.ErrNotAModel
	}

	model, err := d.userRepo.GetByAuthID(ctx, *authID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "model is not found by authID",
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrNotAModel))

			return nil, service_errors.ErrNotAModel
		}

		d.logger.Error(ctx, "check model restrictions failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if !model.IsUserVerified() {
		d.logger.Error(ctx, "model is not verified",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotVerifiedModel))

		return nil, service_errors.ErrNotVerifiedModel
	}

	return model, nil
}

func (d *DefaultOrderService) checkClientRestrictions(ctx context.Context, authID *int64) (*entity.User, error) {
	role, err := common.GetRoleFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if *role != entity.RoleClient.String() {
		d.logger.Error(ctx, "access denied",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotClient))

		return nil, service_errors.ErrNotClient
	}

	client, err := d.userRepo.GetByAuthID(ctx, *authID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "client is not found by authID",
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrNotClient))

			return nil, service_errors.ErrNotClient
		}

		d.logger.Error(ctx, "check client restrictions failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if !client.IsUserVerified() {
		d.logger.Error(ctx, "client is not verified",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotVerifiedClient))

		return nil, service_errors.ErrNotVerifiedClient
	}

	return client, nil
}

func (d *DefaultOrderService) checkIfModelIsAnOwner(ctx context.Context, modelID, modelServiceID int64) error {
	service, err := d.modelServiceRepo.GetByID(ctx, modelServiceID, false)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "model service is not found by id",
				option.Any("model_service_id", modelServiceID),
				option.Any("model_id", modelID),
				option.Error(service_errors.ErrServiceIsNotFound))

			return service_errors.ErrServiceIsNotFound
		}

		d.logger.Error(ctx, "check model is an owner failed",
			option.Any("model_service_id", modelServiceID),
			option.Any("model_id", modelID),
			option.Error(err))

		return err
	}

	if service.ModelID != modelID {
		d.logger.Error(ctx, "model service is not owned by this model",
			option.Any("model_id", modelID),
			option.Any("model_service_id", modelServiceID),
			option.Error(service_errors.ErrModelIsNotAnOwnerOfService))

		return service_errors.ErrModelIsNotAnOwnerOfService
	}

	return nil
}

func (d *DefaultOrderService) cancelOrder(ctx context.Context, booking *entity.Booking,
	slot *entity.Slot, order *entity.Order) (*entity.Order, error) {

	var res *entity.Order
	err := d.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		var err error
		order.Status = entity.OrderCancelled
		if res, err = d.orderRepo.UpdateStatus(ctx, order); err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "order is not found by id",
					option.Any("order_id", order.ID),
					option.Error(service_errors.ErrOrderNotFound))

				return service_errors.ErrOrderNotFound
			}

			d.logger.Error(ctx, "failed to find order by id",
				option.Any("order_id", order.ID),
				option.Error(err))

			return err
		}

		booking.Status = entity.BookingCancelled
		if _, err = d.bookingRepo.Update(ctx, booking); err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "booking is not found by id",
					option.Any("booking_id", booking.ID),
					option.Error(service_errors.ErrBookingNotFound))

				return service_errors.ErrBookingNotFound
			}

			d.logger.Error(ctx, "failed to find booking by id",
				option.Any("booking_id", booking.ID),
				option.Error(err))

			return err
		}

		slot.Status = entity.SlotAvailable
		if _, err = d.slotRepo.Update(ctx, slot); err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "slot is not found by id",
					option.Any("slot_id", slot.ID),
					option.Error(service_errors.ErrSlotIsNotFound))

				return service_errors.ErrSlotIsNotFound
			}

			d.logger.Error(ctx, "failed to find slot by id",
				option.Any("slot_id", slot.ID),
				option.Error(err))

			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}
