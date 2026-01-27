package service

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/common"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/interfaces"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_const"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

type DefaultBookingService struct {
	bookingRepo      interfaces.BookingRepository
	slotRepo         interfaces.SlotRepository
	orderRepo        interfaces.OrderRepository
	userRepo         interfaces.UserRepository
	modelServiceRepo interfaces.ModelServiceRepository
	txManager        database.TxManager
	logger           pkg.Logger
	bookingTtl       time.Duration
}

func NewDefaultBookingService(bookingRepo interfaces.BookingRepository, slotRepo interfaces.SlotRepository,
	userRepo interfaces.UserRepository, modelServiceRepo interfaces.ModelServiceRepository,
	orderRepo interfaces.OrderRepository, txManager database.TxManager, logger pkg.Logger,
) (*DefaultBookingService, error) {

	ttl := os.Getenv(service_const.DotEnvBookingExpiration)
	if ttl == "" {
		return nil, service_errors.ErrLoadingTTL
	}

	ttlInSeconds, err := strconv.Atoi(ttl)
	if err != nil {
		return nil, service_errors.ErrParsingTTL
	}
	if ttlInSeconds < 0 {
		return nil, service_errors.ErrNotPositiveTTL
	}

	return &DefaultBookingService{
		bookingRepo:      bookingRepo,
		slotRepo:         slotRepo,
		orderRepo:        orderRepo,
		userRepo:         userRepo,
		modelServiceRepo: modelServiceRepo,
		txManager:        txManager,
		logger:           logger,
		bookingTtl:       time.Duration(ttlInSeconds) * time.Second,
	}, nil
}

func (d *DefaultBookingService) CreateBooking(ctx context.Context, modelServiceID,
	slotID int64, street string, house int, apartment, entrance, floor *int, comment *string) (*entity.Booking, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	client, err := d.checkClientRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	slot, err := d.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot not found by id",
				option.Any("slot_id", slotID),
				option.Any("model_service_id", modelServiceID),
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "failed to find slot  by id",
			option.Any("slot_id", slotID),
			option.Any("model_service_id", modelServiceID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if !slot.IsAvailable() {
		d.logger.Error(ctx, "slot not available",
			option.Any("slot_id", slotID),
			option.Any("model_service_id", modelServiceID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrSlotNotAvailable))

		return nil, service_errors.ErrSlotNotAvailable
	}

	if !slot.IsCorrectTransition(entity.SlotReserved) {
		d.logger.Error(ctx, "slot not correct",
			option.Any("slot_id", slotID),
			option.Any("model_service_id", modelServiceID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrInvalidSlotStatusTransition))

		return nil, service_errors.ErrInvalidSlotStatusTransition
	}

	var resApartment, resEntrance, resFloor int
	var resComment string
	if apartment == nil {
		resApartment = 0
	}
	if entrance == nil {
		resEntrance = 0
	}
	if floor == nil {
		resFloor = 0
	}
	if comment == nil {
		resComment = ""
	}

	address := entity.NewAddress(street, house, resApartment, resEntrance, resFloor, resComment)
	booking := entity.NewBooking(client.ID, modelServiceID, slotID, address, d.bookingTtl)

	var res *entity.Booking
	err = d.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		slot.Status = entity.SlotReserved
		if slot, err = d.slotRepo.Update(ctx, slot); err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "slot not found by id",
					option.Any("slot_id", slotID),
					option.Any("model_service_id", modelServiceID),
					option.Any("auth_id", authID),
					option.Error(service_errors.ErrSlotIsNotFound))

				return service_errors.ErrSlotIsNotFound
			}

			d.logger.Error(ctx, "failed to update slot",
				option.Any("slot_id", slotID),
				option.Any("model_service_id", modelServiceID),
				option.Any("auth_id", authID),
				option.Error(err))

			return err
		}

		if err = d.bookingRepo.Save(ctx, booking); err != nil {
			d.logger.Error(ctx, "failed to save booking",
				option.Any("slot_id", slotID),
				option.Any("model_service_id", modelServiceID),
				option.Any("auth_id", authID),
				option.Error(err))

			return err
		}

		res = booking

		return nil
	})

	if err != nil {
		return nil, err
	}

	return res, nil
}

func (d *DefaultBookingService) ApproveBooking(ctx context.Context, bookingID int64) (*entity.Booking, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	booking, err := d.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "booking not found by id",
				option.Any("booking_id", bookingID),
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrBookingNotFound))

			return nil, service_errors.ErrBookingNotFound
		}

		d.logger.Error(ctx, "failed to find booking by id",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	err = d.checkIfModelIsAnOwner(ctx, model.ID, booking.ModelServiceID)
	if err != nil {
		return nil, err
	}

	if booking.IsExpired(time.Now()) {
		d.logger.Error(ctx, "booking is expired",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrBookingExpired))

		return nil, service_errors.ErrBookingExpired
	}

	if !booking.CanBeApproved() {
		d.logger.Error(ctx, "booking cannot be approved",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrBookingAlreadyProcessed))

		return nil, service_errors.ErrBookingAlreadyProcessed
	}

	slot, err := d.slotRepo.GetByID(ctx, booking.SlotID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot not found by id",
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrSlotIsNotFound))
		}

		d.logger.Error(ctx, "failed to find slot by id",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if !slot.IsCorrectTransition(entity.SlotBooked) {
		d.logger.Error(ctx, "not correct slot transition",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrInvalidSlotStatusTransition))

		return nil, service_errors.ErrInvalidSlotStatusTransition
	}

	var res *entity.Booking
	err = d.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		slot.Status = entity.SlotBooked
		if _, err = d.slotRepo.Update(ctx, slot); err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "slot not found by id",
					option.Any("slot_id", slot.ID),
					option.Any("auth_id", authID),
					option.Error(service_errors.ErrSlotIsNotFound))

				return service_errors.ErrSlotIsNotFound
			}

			d.logger.Error(ctx, "failed to update slot",
				option.Any("slot_id", slot.ID),
				option.Any("auth_id", authID),
				option.Error(err))

			return err
		}

		booking.Status = entity.BookingApproved
		res, err = d.bookingRepo.Update(ctx, booking)
		if err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "booking not found by id",
					option.Any("booking_id", booking.ID),
					option.Any("auth_id", authID),
					option.Error(service_errors.ErrBookingNotFound))

				return service_errors.ErrBookingNotFound
			}

			d.logger.Error(ctx, "failed to update booking",
				option.Any("booking_id", booking.ID),
				option.Any("auth_id", authID),
				option.Error(err))

			return err
		}

		order := entity.NewOrder(bookingID)

		if err = d.orderRepo.Save(ctx, order); err != nil {
			d.logger.Error(ctx, "failed to save order",
				option.Any("order_id", order.ID),
				option.Any("booking_id", booking.ID),
				option.Any("auth_id", authID),
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

func (d *DefaultBookingService) RejectBooking(ctx context.Context, bookingID int64) (*entity.Booking, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	booking, err := d.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "booking not found by id",
				option.Any("booking_id", bookingID),
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrBookingNotFound))

			return nil, service_errors.ErrBookingNotFound
		}

		d.logger.Error(ctx, "failed to find booking by id",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	err = d.checkIfModelIsAnOwner(ctx, model.ID, booking.ModelServiceID)
	if err != nil {
		return nil, err
	}

	if !booking.CanBeRejected() {
		d.logger.Error(ctx, "booking cannot be rejected",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrBookingAlreadyProcessed))

		return nil, service_errors.ErrBookingAlreadyProcessed
	}

	if booking.IsExpired(time.Now()) {
		d.logger.Error(ctx, "booking is expired",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrBookingExpired))

		return nil, service_errors.ErrBookingExpired
	}

	slot, err := d.slotRepo.GetByID(ctx, booking.SlotID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot not found by id",
				option.Any("slot_id", slot.ID),
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "failed to find slot by id",
			option.Any("slot_id", slot.ID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if !slot.IsCorrectTransition(entity.SlotAvailable) {
		d.logger.Error(ctx, "not correct slot transition",
			option.Any("auth_id", authID),
			option.Any("slot_id", slot.ID),
			option.Error(service_errors.ErrInvalidSlotStatusTransition))

		return nil, service_errors.ErrInvalidSlotStatusTransition
	}

	var res *entity.Booking
	err = d.txManager.WithTransaction(ctx, func(ctx context.Context) error {

		slot.Status = entity.SlotAvailable
		if _, err = d.slotRepo.Update(ctx, slot); err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "slot not found by id",
					option.Any("slot_id", slot.ID),
					option.Any("auth_id", authID),
					option.Error(service_errors.ErrSlotIsNotFound))

				return service_errors.ErrSlotIsNotFound
			}

			d.logger.Error(ctx, "failed to save slot by id",
				option.Any("slot_id", slot.ID),
				option.Any("auth_id", authID),
				option.Error(err))

			return err
		}

		booking.Status = entity.BookingRejected
		res, err = d.bookingRepo.Update(ctx, booking)
		if err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "booking not found by id",
					option.Any("booking_id", booking.ID),
					option.Any("auth_id", authID),
					option.Error(service_errors.ErrBookingNotFound))

				return service_errors.ErrBookingNotFound
			}

			d.logger.Error(ctx, "failed to update booking by id",
				option.Any("booking_id", booking.ID),
				option.Any("auth_id", authID),
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

func (d *DefaultBookingService) CancelBookingByClient(ctx context.Context, bookingID int64) (*entity.Booking, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	client, err := d.checkClientRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	booking, err := d.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "booking not found by id",
				option.Any("booking_id", bookingID),
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrBookingNotFound))

			return nil, service_errors.ErrBookingNotFound
		}

		d.logger.Error(ctx, "failed to find booking by id",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if booking.ClientID != client.ID {
		d.logger.Error(ctx, "client is not an owner of this booking",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrClientIsNotOwnerOfBooking))

		return nil, service_errors.ErrClientIsNotOwnerOfBooking
	}

	if booking.IsExpired(time.Now()) {
		d.logger.Error(ctx, "booking is expired",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrBookingExpired))

		return nil, service_errors.ErrBookingExpired
	}

	if !booking.CanBeCancelledByClient() {
		d.logger.Error(ctx, "booking cannot be cancelled by client",
			option.Any("booking_id", bookingID),
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrInvalidBookingState))

		return nil, service_errors.ErrInvalidBookingState
	}

	slot, err := d.slotRepo.GetByID(ctx, booking.SlotID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot not found by id",
				option.Any("slot_id", slot.ID),
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "failed to find slot by id",
			option.Any("slot_id", slot.ID),
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	if !slot.IsCorrectTransition(entity.SlotAvailable) {
		d.logger.Error(ctx, "not correct slot transition",
			option.Any("auth_id", authID),
			option.Any("slot_id", slot.ID),
			option.Error(service_errors.ErrInvalidSlotStatusTransition))

		return nil, service_errors.ErrInvalidSlotStatusTransition
	}

	var res *entity.Booking
	err = d.txManager.WithTransaction(ctx, func(ctx context.Context) error {
		slot.Status = entity.SlotAvailable
		if _, err = d.slotRepo.Update(ctx, slot); err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "slot not found by id",
					option.Any("slot_id", slot.ID),
					option.Any("auth_id", authID),
					option.Error(service_errors.ErrSlotIsNotFound))

				return service_errors.ErrSlotIsNotFound
			}

			d.logger.Error(ctx, "failed to update booking by id",
				option.Any("slot_id", slot.ID),
				option.Any("auth_id", authID),
				option.Error(err))

			return err
		}

		booking.Status = entity.BookingCancelled
		res, err = d.bookingRepo.Update(ctx, booking)
		if err != nil {
			if errors.Is(err, persistence.ErrNoRowsFound) {
				d.logger.Error(ctx, "booking not found by id",
					option.Any("booking_id", booking.ID),
					option.Any("auth_id", authID),
					option.Error(service_errors.ErrBookingNotFound))

				return service_errors.ErrBookingNotFound
			}

			d.logger.Error(ctx, "failed to find booking by id",
				option.Any("booking_id", booking.ID),
				option.Any("auth_id", authID),
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

func (d *DefaultBookingService) checkModelRestrictions(ctx context.Context, authID *int64) (*entity.User, error) {
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

func (d *DefaultBookingService) checkClientRestrictions(ctx context.Context, authID *int64) (*entity.User, error) {
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

func (d *DefaultBookingService) checkIfModelIsAnOwner(ctx context.Context, modelID, modelServiceID int64) error {
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
