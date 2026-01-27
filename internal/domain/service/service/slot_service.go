package service

import (
	"context"
	"errors"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/common"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/interfaces"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

type DefaultSlotService struct {
	slotRepo    interfaces.SlotRepository
	bookingRepo interfaces.BookingRepository
	userRepo    interfaces.UserRepository
	txManager   database.TxManager
	logger      pkg.Logger
}

func NewDefaultSlotService(slotRepo interfaces.SlotRepository, bookingRepo interfaces.BookingRepository,
	userRepo interfaces.UserRepository, txManager database.TxManager, logger pkg.Logger) *DefaultSlotService {
	return &DefaultSlotService{
		slotRepo:    slotRepo,
		bookingRepo: bookingRepo,
		userRepo:    userRepo,
		txManager:   txManager,
		logger:      logger,
	}
}

func (d *DefaultSlotService) CreateSlot(ctx context.Context, start, end time.Time) (*entity.Slot, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if !start.Before(end) {
		d.logger.Error(ctx, "start time must be before end time",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrIncorrectSlotTime))

		return nil, service_errors.ErrIncorrectSlotTime
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	overlaps, err := d.slotRepo.GetOverlappingSlots(ctx, model.ID, start, end)
	if err != nil {
		d.logger.Error(ctx, "cannot get overlaps slot for model",
			option.Any("model_id", model.ID),
			option.Error(err))

		return nil, err
	}
	if len(overlaps) > 0 {
		d.logger.Error(ctx, "there are overlapping slots for model",
			option.Any("model_id", model.ID),
			option.Error(service_errors.ErrSlotOverlap))

		return nil, service_errors.ErrSlotOverlap
	}

	slot := entity.NewSlot(model.ID, start, end)
	if err = d.slotRepo.Save(ctx, slot); err != nil {
		d.logger.Error(ctx, "cannot save slot for model",
			option.Any("auth_id", authID),
			option.Error(err))

		return nil, err
	}

	return slot, nil
}

func (d *DefaultSlotService) UpdateSlot(ctx context.Context,
	slotID int64, start, end *time.Time) (*entity.Slot, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	slot, err := d.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot is not found by id",
				option.Any("auth_id", authID),
				option.Any("slot_id", slotID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "failed to find slot",
			option.Any("auth_id", authID),
			option.Any("slot_id", slotID),
			option.Error(err))

		return nil, err
	}

	if slot.ModelID != model.ID {
		d.logger.Error(ctx, "slot is not found by this model",
			option.Any("auth_id", authID),
			option.Any("model_id", model.ID),
			option.Any("slot_id", slotID),
			option.Error(service_errors.ErrModelIsNotAnOwnerOfSlot))

		return nil, service_errors.ErrModelIsNotAnOwnerOfSlot
	}

	if !slot.IsAvailable() {
		d.logger.Error(ctx, "slot is not available",
			option.Any("auth_id", authID),
			option.Any("slot_id", slotID),
			option.Error(service_errors.ErrSlotNotAvailable))

		return nil, service_errors.ErrSlotNotAvailable
	}

	newStart := slot.StartTime
	newEnd := slot.EndTime
	if start != nil {
		newStart = *start
	}
	if end != nil {
		newEnd = *end
	}

	if slot.StartTime != newStart || slot.EndTime != newEnd {
		overlaps, err := d.slotRepo.GetOverlappingSlots(ctx, slot.ModelID, newStart, newEnd)
		if err != nil {
			d.logger.Error(ctx, "cannot get overlaps slot for model",
				option.Any("model_id", model.ID),
				option.Error(err))

			return nil, err
		}
		filtered := make([]*entity.Slot, 0, len(overlaps))
		for _, s := range overlaps {
			if s.ID != slot.ID {
				filtered = append(filtered, s)
			}
		}
		if len(filtered) > 0 {
			return nil, service_errors.ErrSlotOverlap
		}

		slot.StartTime = newStart
		slot.EndTime = newEnd
	}

	res, err := d.slotRepo.Update(ctx, slot)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot is not found by id",
				option.Any("auth_id", authID),
				option.Any("slot_id", slotID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "cannot update slot for model",
			option.Any("auth_id", authID),
			option.Any("slot_id", slotID),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultSlotService) DeactivateSlot(ctx context.Context, slotID int64) (*entity.Slot, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	slot, err := d.slotRepo.GetByID(ctx, slotID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot is not found by id",
				option.Any("auth_id", authID),
				option.Any("slot_id", slotID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "failed to find slot",
			option.Any("auth_id", authID),
			option.Any("slot_id", slotID),
			option.Error(err))

		return nil, err
	}

	if slot.ModelID != model.ID {
		d.logger.Error(ctx, "slot is not owned by this model",
			option.Any("auth_id", authID),
			option.Any("slot_id", slotID),
			option.Any("model_id", model.ID),
			option.Error(service_errors.ErrModelIsNotAnOwnerOfSlot))

		return nil, service_errors.ErrModelIsNotAnOwnerOfSlot
	}

	if !slot.IsAvailable() {
		d.logger.Error(ctx, "slot is not available",
			option.Any("auth_id", authID),
			option.Any("slot_id", slotID),
			option.Error(service_errors.ErrSlotNotAvailable))

		return nil, service_errors.ErrSlotNotAvailable
	}

	slot.Status = entity.SlotDisabled

	res, err := d.slotRepo.Update(ctx, slot)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "slot is not found by id",
				option.Any("auth_id", authID),
				option.Any("slot_id", slotID),
				option.Error(service_errors.ErrSlotIsNotFound))

			return nil, service_errors.ErrSlotIsNotFound
		}

		d.logger.Error(ctx, "cannot update slot for model",
			option.Any("auth_id", authID),
			option.Any("slot_id", slotID),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultSlotService) GetSlotsWithModelIDByModel(ctx context.Context) ([]*entity.Slot, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	model, err := d.checkModelRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	slots, err := d.slotRepo.GetByModelID(ctx, model.ID)
	if err != nil {
		d.logger.Error(ctx, "failed to find slots for model",
			option.Any("auth_id", authID),
			option.Any("model_id", model.ID),
			option.Error(err))

		return nil, err
	}

	return slots, nil
}

func (d *DefaultSlotService) GetSlotsWithModelIDByClient(ctx context.Context, modelID int64) ([]*entity.Slot, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkClientRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	model, err := d.userRepo.GetByID(ctx, modelID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "model is not found by modelID",
				option.Any("model_id", modelID),
				option.Error(service_errors.ErrNotAModel))

			return nil, service_errors.ErrNotAModel
		}

		d.logger.Error(ctx, "check model restrictions failed",
			option.Any("model_id", modelID),
			option.Error(err))

		return nil, err
	}

	if !model.IsUserVerified() {
		d.logger.Error(ctx, "model is not a model",
			option.Any("model_id", modelID),
			option.Error(service_errors.ErrNotAModel))

		return nil, service_errors.ErrNotAModel
	}

	slots, err := d.slotRepo.GetByModelID(ctx, modelID)
	if err != nil {
		d.logger.Error(ctx, "failed to find slots for model",
			option.Any("auth_id", authID),
			option.Any("model_id", modelID),
			option.Error(err))

		return nil, err
	}

	activeSlots := make([]*entity.Slot, 0, len(slots))
	for _, slot := range slots {
		if slot.Status != entity.SlotDisabled {
			activeSlots = append(activeSlots, slot)
		}
	}

	return activeSlots, nil
}

func (d *DefaultSlotService) checkModelRestrictions(ctx context.Context, authID *int64) (*entity.User, error) {
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

func (d *DefaultSlotService) checkClientRestrictions(ctx context.Context, authID *int64) error {
	role, err := common.GetRoleFromContext(ctx)
	if err != nil {
		return err
	}

	if *role != entity.RoleClient.String() {
		d.logger.Error(ctx, "access denied",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotClient))

		return service_errors.ErrNotClient
	}

	client, err := d.userRepo.GetByAuthID(ctx, *authID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "client is not found by authID",
				option.Any("auth_id", authID),
				option.Error(service_errors.ErrNotClient))

			return service_errors.ErrNotClient
		}

		d.logger.Error(ctx, "check client restrictions failed",
			option.Any("auth_id", authID),
			option.Error(err))

		return err
	}

	if !client.IsUserVerified() {
		d.logger.Error(ctx, "client is not verified",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotVerifiedClient))

		return service_errors.ErrNotVerifiedClient
	}

	return nil
}
