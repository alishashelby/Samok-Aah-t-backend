package handler

import (
	"context"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/authorized"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/models"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
	"github.com/go-playground/validator/v10"
)

type SlotService interface {
	CreateSlot(ctx context.Context, start, end time.Time) (*entity.Slot, error)
	UpdateSlot(ctx context.Context, slotID int64,
		start, end *time.Time) (*entity.Slot, error)
	DeactivateSlot(ctx context.Context, slotID int64) (*entity.Slot, error)
	GetSlotsWithModelIDByModel(ctx context.Context) ([]*entity.Slot, error)
	GetSlotsWithModelIDByClient(ctx context.Context, modelID int64) ([]*entity.Slot, error)
}

type SlotHandler struct {
	slotService SlotService
	logger      pkg.Logger
	validate    *validator.Validate
}

func NewSlotHandler(slotService SlotService, logger pkg.Logger) *SlotHandler {
	return &SlotHandler{
		slotService: slotService,
		logger:      logger,
		validate:    validator.New(),
	}
}

func (h *SlotHandler) CreateSlot(ctx context.Context,
	request authorized.PostModelSlotsRequestObject) (authorized.PostModelSlotsResponseObject, error) {

	h.logger.Info(ctx, "SlotHandler.CreateSlot")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.slotService.CreateSlot(ctx, request.Body.Start, request.Body.End)
	if err != nil {
		return nil, err
	}

	return &authorized.PostModelSlots201JSONResponse{
		Id:        res.ID,
		ModelId:   res.ModelID,
		StartTime: res.StartTime,
		EndTime:   res.EndTime,
		Status:    models.SlotStatus(res.Status),
		CreatedAt: res.CreatedAt,
	}, nil
}

func (h *SlotHandler) GetOwnModelSlots(ctx context.Context,
	request authorized.GetModelSlotsRequestObject) (authorized.GetModelSlotsResponseObject, error) {

	h.logger.Info(ctx, "SlotHandler.GetOwnModelSlots")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	slots, err := h.slotService.GetSlotsWithModelIDByModel(ctx)
	if err != nil {
		return nil, err
	}

	res := make(authorized.GetModelSlots200JSONResponse, len(slots))
	for i, s := range slots {
		res[i] = models.SlotResponse{
			Id:        s.ID,
			ModelId:   s.ModelID,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Status:    models.SlotStatus(s.Status),
			CreatedAt: s.CreatedAt,
		}
	}

	return res, nil
}

func (h *SlotHandler) UpdateSlot(ctx context.Context,
	request authorized.PatchModelSlotsSlotIdRequestObject) (authorized.PatchModelSlotsSlotIdResponseObject, error) {

	h.logger.Info(ctx, "SlotHandler.UpdateSlot")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.slotService.UpdateSlot(ctx, request.SlotId, request.Body.Start,
		request.Body.End)
	if err != nil {
		return nil, err
	}

	return authorized.PatchModelSlotsSlotId200JSONResponse{
		Id:        res.ID,
		ModelId:   res.ModelID,
		StartTime: res.StartTime,
		EndTime:   res.EndTime,
		Status:    models.SlotStatus(res.Status),
		CreatedAt: res.CreatedAt,
	}, nil
}

func (h *SlotHandler) DeactivateSlot(ctx context.Context,
	request authorized.PatchModelSlotsSlotIdDisableRequestObject,
) (authorized.PatchModelSlotsSlotIdDisableResponseObject, error) {

	h.logger.Info(ctx, "SlotHandler.DeactivateSlot")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	res, err := h.slotService.DeactivateSlot(ctx, request.SlotId)
	if err != nil {
		return nil, err
	}

	return authorized.PatchModelSlotsSlotIdDisable200JSONResponse{
		Id:        res.ID,
		ModelId:   res.ModelID,
		StartTime: res.StartTime,
		EndTime:   res.EndTime,
		Status:    models.SlotStatus(res.Status),
		CreatedAt: res.CreatedAt,
	}, nil
}

func (h *SlotHandler) GetModelSlotsForClient(ctx context.Context,
	request authorized.GetClientModelsModelIdSlotsRequestObject,
) (authorized.GetClientModelsModelIdSlotsResponseObject, error) {

	h.logger.Info(ctx, "SlotHandler.GetSlotsForClient")

	if err := h.validate.Struct(request); err != nil {
		h.logger.Error(ctx, "validation error",
			option.Error(err))

		return nil, err
	}

	slots, err := h.slotService.GetSlotsWithModelIDByClient(ctx, request.ModelId)
	if err != nil {
		return nil, err
	}

	res := make(authorized.GetClientModelsModelIdSlots200JSONResponse, 0, len(slots))
	for _, s := range slots {
		res = append(res, models.SlotResponse{
			Id:        s.ID,
			ModelId:   s.ModelID,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Status:    models.SlotStatus(s.Status),
			CreatedAt: s.CreatedAt,
		})
	}

	return res, nil
}
