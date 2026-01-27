package service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/mocks"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_const"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/logger/config"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type slotServiceTest struct {
	ctrl        *gomock.Controller
	slotRepo    *mocks.MockSlotRepository
	bookingRepo *mocks.MockBookingRepository
	userRepo    *mocks.MockUserRepository
	txManager   *mocks.MockTxManager
	service     *DefaultSlotService
}

func setUpSlotServiceTest(t *testing.T) *slotServiceTest {
	t.Helper()

	ctrl := gomock.NewController(t)
	slotRepo := mocks.NewMockSlotRepository(ctrl)
	bookingRepo := mocks.NewMockBookingRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	mockTxManager := mocks.NewMockTxManager(ctrl)

	cfg := &config.LogConfig{}
	cfg.Logger.Level = "info"
	tmpDir := os.TempDir()
	cfg.Logger.LogsDir = tmpDir
	cfg.Logger.LogsFile = "test.log"
	log, err := pkg.NewDualLogger(cfg)
	if err != nil {
		t.Fatal(err)
	}

	slotService := NewDefaultSlotService(
		slotRepo, bookingRepo, userRepo, mockTxManager, log,
	)

	return &slotServiceTest{
		ctrl:        ctrl,
		slotRepo:    slotRepo,
		bookingRepo: bookingRepo,
		userRepo:    userRepo,
		txManager:   mockTxManager,
		service:     slotService,
	}
}

func TestSlotService_UpdateSlot(t *testing.T) {
	test := setUpSlotServiceTest(t)
	defer test.ctrl.Finish()

	ctxModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxModel = context.WithValue(ctxModel, service_const.RoleKey, "MODEL")

	verifiedModel := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	existingSlot := &entity.Slot{
		ID:        1,
		ModelID:   1,
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(26 * time.Hour),
		Status:    entity.SlotAvailable,
	}

	newStart := existingSlot.StartTime.Add(1 * time.Hour)
	newEnd := existingSlot.EndTime.Add(1 * time.Hour)

	updatedSlot := &entity.Slot{
		ID:        1,
		ModelID:   1,
		StartTime: newStart,
		EndTime:   newEnd,
		Status:    entity.SlotAvailable,
	}

	tests := []struct {
		name            string
		ctx             context.Context
		slotID          int64
		start           *time.Time
		end             *time.Time
		mockModel       *entity.User
		mockModelErr    error
		mockSlot        *entity.Slot
		mockSlotErr     error
		mockOverlaps    []*entity.Slot
		mockOverlapsErr error
		mockUpdateSlot  *entity.Slot
		mockUpdateErr   error
		expectedError   error
	}{
		{
			name:           "successful update",
			ctx:            ctxModel,
			slotID:         1,
			start:          &newStart,
			end:            &newEnd,
			mockModel:      verifiedModel,
			mockSlot:       existingSlot,
			mockOverlaps:   []*entity.Slot{},
			mockUpdateSlot: updatedSlot,
		},
		{
			name:          "slot not found",
			ctx:           ctxModel,
			slotID:        1,
			start:         &newStart,
			mockModel:     verifiedModel,
			mockSlotErr:   persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrSlotIsNotFound,
		},
		{
			name:          "model is not owner",
			ctx:           ctxModel,
			slotID:        1,
			start:         &newStart,
			mockModel:     verifiedModel,
			mockSlot:      &entity.Slot{ID: 1, ModelID: 999, Status: entity.SlotAvailable},
			expectedError: service_errors.ErrModelIsNotAnOwnerOfSlot,
		},
		{
			name:          "slot not available",
			ctx:           ctxModel,
			slotID:        1,
			start:         &newStart,
			mockModel:     verifiedModel,
			mockSlot:      &entity.Slot{ID: 1, ModelID: 1, Status: entity.SlotBooked},
			expectedError: service_errors.ErrSlotNotAvailable,
		},
		{
			name:          "update repo error",
			ctx:           ctxModel,
			slotID:        1,
			start:         &newStart,
			end:           &newEnd,
			mockModel:     verifiedModel,
			mockSlot:      existingSlot,
			mockOverlaps:  []*entity.Slot{},
			mockUpdateErr: errors.New("update failed"),
			expectedError: errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "MODEL" {
				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(tt.mockModel, tt.mockModelErr).
					Times(1)
			}

			if tt.mockModelErr == nil && tt.mockModel != nil && tt.mockModel.IsVerified {
				test.slotRepo.EXPECT().
					GetByID(gomock.Any(), tt.slotID).
					Return(tt.mockSlot, tt.mockSlotErr).
					Times(1)

				if tt.mockSlotErr == nil && tt.mockSlot != nil {
					if tt.mockSlot.ModelID == tt.mockModel.ID && tt.mockSlot.IsAvailable() {
						timeChanged := (tt.start != nil && *tt.start != tt.mockSlot.StartTime) ||
							(tt.end != nil && *tt.end != tt.mockSlot.EndTime)

						if timeChanged {
							newStartTime := tt.mockSlot.StartTime
							newEndTime := tt.mockSlot.EndTime
							if tt.start != nil {
								newStartTime = *tt.start
							}
							if tt.end != nil {
								newEndTime = *tt.end
							}

							test.slotRepo.EXPECT().
								GetOverlappingSlots(gomock.Any(), tt.mockSlot.ModelID, newStartTime, newEndTime).
								Return(tt.mockOverlaps, tt.mockOverlapsErr).
								Times(1)

							if tt.mockOverlapsErr == nil && len(tt.mockOverlaps) == 0 {
								test.slotRepo.EXPECT().
									Update(gomock.Any(), gomock.Any()).
									Return(tt.mockUpdateSlot, tt.mockUpdateErr).
									Times(1)
							}
						} else {
							test.slotRepo.EXPECT().
								Update(gomock.Any(), gomock.Any()).
								Return(tt.mockUpdateSlot, tt.mockUpdateErr).
								Times(1)
						}
					}
				}
			}

			slot, err := test.service.UpdateSlot(tt.ctx, tt.slotID, tt.start, tt.end)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, slot)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, slot)
			}
		})
	}
}

func TestSlotService_DeactivateSlot(t *testing.T) {
	test := setUpSlotServiceTest(t)
	defer test.ctrl.Finish()

	ctxModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxModel = context.WithValue(ctxModel, service_const.RoleKey, "MODEL")

	verifiedModel := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	availableSlot := &entity.Slot{
		ID:        1,
		ModelID:   1,
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(26 * time.Hour),
		Status:    entity.SlotAvailable,
	}

	disabledSlot := &entity.Slot{
		ID:        1,
		ModelID:   1,
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(26 * time.Hour),
		Status:    entity.SlotDisabled,
	}

	tests := []struct {
		name           string
		ctx            context.Context
		slotID         int64
		mockModel      *entity.User
		mockModelErr   error
		mockSlot       *entity.Slot
		mockSlotErr    error
		mockUpdateSlot *entity.Slot
		mockUpdateErr  error
		expectedError  error
	}{
		{
			name:           "successful deactivation",
			ctx:            ctxModel,
			slotID:         1,
			mockModel:      verifiedModel,
			mockSlot:       availableSlot,
			mockUpdateSlot: disabledSlot,
		},
		{
			name:          "slot not found",
			ctx:           ctxModel,
			slotID:        1,
			mockModel:     verifiedModel,
			mockSlotErr:   persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrSlotIsNotFound,
		},
		{
			name:          "model is not owner",
			ctx:           ctxModel,
			slotID:        1,
			mockModel:     verifiedModel,
			mockSlot:      &entity.Slot{ID: 1, ModelID: 999, Status: entity.SlotAvailable},
			expectedError: service_errors.ErrModelIsNotAnOwnerOfSlot,
		},
		{
			name:          "slot not available",
			ctx:           ctxModel,
			slotID:        1,
			mockModel:     verifiedModel,
			mockSlot:      &entity.Slot{ID: 1, ModelID: 1, Status: entity.SlotBooked},
			expectedError: service_errors.ErrSlotNotAvailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "MODEL" {
				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(tt.mockModel, tt.mockModelErr).
					Times(1)
			}

			if tt.mockModelErr == nil && tt.mockModel != nil && tt.mockModel.IsVerified {
				test.slotRepo.EXPECT().
					GetByID(gomock.Any(), tt.slotID).
					Return(tt.mockSlot, tt.mockSlotErr).
					Times(1)

				if tt.mockSlotErr == nil && tt.mockSlot != nil {
					if tt.mockSlot.ModelID == tt.mockModel.ID && tt.mockSlot.IsAvailable() {
						test.slotRepo.EXPECT().
							Update(gomock.Any(), gomock.Any()).
							Return(tt.mockUpdateSlot, tt.mockUpdateErr).
							Times(1)
					}
				}
			}

			slot, err := test.service.DeactivateSlot(tt.ctx, tt.slotID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, slot)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, slot)
				assert.Equal(t, entity.SlotDisabled, slot.Status)
			}
		})
	}
}

func TestSlotService_GetSlotsWithModelIDByClient(t *testing.T) {
	test := setUpSlotServiceTest(t)
	defer test.ctrl.Finish()

	ctxClient := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxClient = context.WithValue(ctxClient, service_const.RoleKey, "CLIENT")

	verifiedClient := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	modelID := int64(2)
	verifiedModel := &entity.User{
		ID:         modelID,
		IsVerified: true,
	}

	slots := []*entity.Slot{
		{ID: 1, ModelID: modelID, Status: entity.SlotAvailable},
		{ID: 2, ModelID: modelID, Status: entity.SlotDisabled},
		{ID: 3, ModelID: modelID, Status: entity.SlotBooked},
	}

	tests := []struct {
		name          string
		ctx           context.Context
		modelID       int64
		mockClient    *entity.User
		mockClientErr error
		mockModel     *entity.User
		mockModelErr  error
		mockSlots     []*entity.Slot
		mockSlotsErr  error
		expectedError error
		expectedCount int
	}{
		{
			name:          "successful get slots",
			ctx:           ctxClient,
			modelID:       modelID,
			mockClient:    verifiedClient,
			mockModel:     verifiedModel,
			mockSlots:     slots,
			expectedCount: 2,
		},
		{
			name:          "client not verified",
			ctx:           ctxClient,
			modelID:       modelID,
			mockClient:    &entity.User{ID: 1, IsVerified: false},
			expectedError: service_errors.ErrNotVerifiedClient,
		},
		{
			name:          "model not found",
			ctx:           ctxClient,
			modelID:       modelID,
			mockClient:    verifiedClient,
			mockModelErr:  persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrNotAModel,
		},
		{
			name:          "model not verified",
			ctx:           ctxClient,
			modelID:       modelID,
			mockClient:    verifiedClient,
			mockModel:     &entity.User{ID: modelID, IsVerified: false},
			expectedError: service_errors.ErrNotAModel,
		},
		{
			name:          "slots repo error",
			ctx:           ctxClient,
			modelID:       modelID,
			mockClient:    verifiedClient,
			mockModel:     verifiedModel,
			mockSlotsErr:  errors.New("db error"),
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "CLIENT" {
				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(tt.mockClient, tt.mockClientErr).
					Times(1)
			}

			if tt.mockClientErr == nil && tt.mockClient != nil && tt.mockClient.IsVerified {
				test.userRepo.EXPECT().
					GetByID(gomock.Any(), tt.modelID).
					Return(tt.mockModel, tt.mockModelErr).
					Times(1)

				if tt.mockModelErr == nil && tt.mockModel != nil && tt.mockModel.IsVerified {
					test.slotRepo.EXPECT().
						GetByModelID(gomock.Any(), tt.modelID).
						Return(tt.mockSlots, tt.mockSlotsErr).
						Times(1)
				}
			}

			result, err := test.service.GetSlotsWithModelIDByClient(tt.ctx, tt.modelID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.expectedCount)

				for _, slot := range result {
					assert.NotEqual(t, entity.SlotDisabled, slot.Status)
				}
			}
		})
	}
}

func TestSlotService_CheckModelRestrictions_EdgeCases(t *testing.T) {
	test := setUpSlotServiceTest(t)
	defer test.ctrl.Finish()

	tests := []struct {
		name          string
		ctx           context.Context
		mockUser      *entity.User
		mockUserErr   error
		expectedError error
	}{
		{
			name:          "successful model verification",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "MODEL"),
			mockUser:      &entity.User{ID: 1, AuthID: int64(1), IsVerified: true},
			expectedError: nil,
		},
		{
			name:          "not a model role - client",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "CLIENT"),
			expectedError: service_errors.ErrNotAModel,
		},
		{
			name:          "not a model role - admin",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "ADMIN"),
			expectedError: service_errors.ErrNotAModel,
		},
		{
			name:          "model not found in repository",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "MODEL"),
			mockUserErr:   persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrNotAModel,
		},
		{
			name:          "repository error when getting model",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "MODEL"),
			mockUserErr:   errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:          "model not verified",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "MODEL"),
			mockUser:      &entity.User{ID: 1, AuthID: int64(1), IsVerified: false},
			expectedError: service_errors.ErrNotVerifiedModel,
		},
		{
			name:          "model with zero ID",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(0)), service_const.RoleKey, "MODEL"),
			mockUser:      &entity.User{ID: 0, AuthID: int64(0), IsVerified: true},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "MODEL" && tt.ctx.Value(service_const.AuthIDKey) != nil {
				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(tt.mockUser, tt.mockUserErr).
					Times(1)
			}

			model, err := test.service.checkModelRestrictions(tt.ctx, &[]int64{tt.ctx.Value(service_const.AuthIDKey).(int64)}[0])

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, model)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, model)
				assert.True(t, model.IsVerified)
			}
		})
	}
}

func TestSlotService_CheckClientRestrictions_EdgeCases(t *testing.T) {
	test := setUpSlotServiceTest(t)
	defer test.ctrl.Finish()

	tests := []struct {
		name          string
		ctx           context.Context
		mockUser      *entity.User
		mockUserErr   error
		expectedError error
	}{
		{
			name:          "successful client verification",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "CLIENT"),
			mockUser:      &entity.User{ID: 1, AuthID: int64(1), IsVerified: true},
			expectedError: nil,
		},
		{
			name:          "not a client role - model",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "MODEL"),
			expectedError: service_errors.ErrNotClient,
		},
		{
			name:          "not a client role - admin",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "ADMIN"),
			expectedError: service_errors.ErrNotClient,
		},
		{
			name:          "client not found in repository",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "CLIENT"),
			mockUserErr:   persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrNotClient,
		},
		{
			name:          "repository error when getting client",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "CLIENT"),
			mockUserErr:   errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:          "client not verified",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "CLIENT"),
			mockUser:      &entity.User{ID: 1, AuthID: int64(1), IsVerified: false},
			expectedError: service_errors.ErrNotVerifiedClient,
		},
		{
			name:          "client with zero ID",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(0)), service_const.RoleKey, "CLIENT"),
			mockUser:      &entity.User{ID: 0, AuthID: int64(0), IsVerified: true},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "CLIENT" && tt.ctx.Value(service_const.AuthIDKey) != nil {
				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(tt.mockUser, tt.mockUserErr).
					Times(1)
			}

			var authIDPtr *int64
			if tt.ctx.Value(service_const.AuthIDKey) != nil {
				id := tt.ctx.Value(service_const.AuthIDKey).(int64)
				authIDPtr = &id
			}

			err := test.service.checkClientRestrictions(tt.ctx, authIDPtr)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
