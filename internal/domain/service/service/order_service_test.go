package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	metrics2 "github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/metrics"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/mocks"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_const"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/logger/config"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type orderServiceTest struct {
	ctrl             *gomock.Controller
	orderRepo        *mocks.MockOrderRepository
	bookingRepo      *mocks.MockBookingRepository
	slotRepo         *mocks.MockSlotRepository
	userRepo         *mocks.MockUserRepository
	modelServiceRepo *mocks.MockModelServiceRepository
	txManager        *mocks.MockTxManager
	metrics          *metrics2.Metrics
	service          *DefaultOrderService
}

func setUpOrderServiceTest(t *testing.T) *orderServiceTest {
	t.Helper()

	ctrl := gomock.NewController(t)
	orderRepo := mocks.NewMockOrderRepository(ctrl)
	bookingRepo := mocks.NewMockBookingRepository(ctrl)
	slotRepo := mocks.NewMockSlotRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	modelServiceRepo := mocks.NewMockModelServiceRepository(ctrl)
	mockTxManager := mocks.NewMockTxManager(ctrl)
	metrics := metrics2.NewMetrics()

	cfg := &config.LogConfig{}
	cfg.Logger.Level = "info"
	tmpDir := os.TempDir()
	cfg.Logger.LogsDir = tmpDir
	cfg.Logger.LogsFile = "test.log"
	log, err := pkg.NewDualLogger(cfg)
	if err != nil {
		t.Fatal(err)
	}

	orderService := NewDefaultOrderService(
		orderRepo, bookingRepo, slotRepo, userRepo, modelServiceRepo,
		mockTxManager, log, metrics,
	)

	return &orderServiceTest{
		ctrl:             ctrl,
		orderRepo:        orderRepo,
		bookingRepo:      bookingRepo,
		slotRepo:         slotRepo,
		userRepo:         userRepo,
		modelServiceRepo: modelServiceRepo,
		txManager:        mockTxManager,
		metrics:          metrics,
		service:          orderService,
	}
}

func TestOrderService_CompleteOrder(t *testing.T) {
	test := setUpOrderServiceTest(t)
	defer test.ctrl.Finish()

	ctxModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxModel = context.WithValue(ctxModel, service_const.RoleKey, "MODEL")

	verifiedModel := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	order := &entity.Order{
		ID:        1,
		BookingID: 2,
		Status:    entity.OrderInTransit,
	}

	booking := &entity.Booking{
		ID:             2,
		ModelServiceID: 3,
		SlotID:         4,
		ClientID:       5,
		Status:         entity.BookingApproved,
	}

	slot := &entity.Slot{
		ID:        4,
		ModelID:   1,
		StartTime: time.Now().Add(-2 * time.Hour),
		EndTime:   time.Now().Add(-1 * time.Hour),
		Status:    entity.SlotBooked,
	}

	modelService := &entity.ModelService{
		ID:      3,
		ModelID: 1,
	}

	completedOrder := &entity.Order{
		ID:        1,
		BookingID: 2,
		Status:    entity.OrderCompleted,
	}

	tests := []struct {
		name                string
		ctx                 context.Context
		orderID             int64
		mockModel           *entity.User
		mockModelErr        error
		mockOrder           *entity.Order
		mockOrderErr        error
		mockBooking         *entity.Booking
		mockBookingErr      error
		mockModelService    *entity.ModelService
		mockModelServiceErr error
		mockSlot            *entity.Slot
		mockSlotErr         error
		mockUpdateOrder     *entity.Order
		mockUpdateErr       error
		expectMetrics       bool
		expectedError       error
	}{
		{
			name:             "successful completion",
			ctx:              ctxModel,
			orderID:          1,
			mockModel:        verifiedModel,
			mockOrder:        order,
			mockBooking:      booking,
			mockModelService: modelService,
			mockSlot:         slot,
			mockUpdateOrder:  completedOrder,
			expectMetrics:    true,
		},
		{
			name:          "order not found",
			ctx:           ctxModel,
			orderID:       1,
			mockModel:     verifiedModel,
			mockOrderErr:  persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrOrderNotFound,
		},
		{
			name:             "cannot complete - slot not ended yet",
			ctx:              ctxModel,
			orderID:          1,
			mockModel:        verifiedModel,
			mockOrder:        order,
			mockBooking:      booking,
			mockModelService: modelService,
			mockSlot: &entity.Slot{
				ID:        4,
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now().Add(1 * time.Hour),
			},
			expectedError: service_errors.ErrCannotCompleteOrder,
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
				test.orderRepo.EXPECT().
					GetByID(gomock.Any(), tt.orderID).
					Return(tt.mockOrder, tt.mockOrderErr).
					Times(1)

				if tt.mockOrderErr == nil && tt.mockOrder != nil {
					test.bookingRepo.EXPECT().
						GetByID(gomock.Any(), tt.mockOrder.BookingID).
						Return(tt.mockBooking, tt.mockBookingErr).
						Times(1)

					if tt.mockBookingErr == nil && tt.mockBooking != nil {
						test.modelServiceRepo.EXPECT().
							GetByID(gomock.Any(), tt.mockBooking.ModelServiceID, false).
							Return(tt.mockModelService, tt.mockModelServiceErr).
							Times(1)

						if tt.mockModelServiceErr == nil && tt.mockModelService != nil {
							if tt.mockModelService.ModelID == tt.mockModel.ID {
								test.slotRepo.EXPECT().
									GetByID(gomock.Any(), tt.mockBooking.SlotID).
									Return(tt.mockSlot, tt.mockSlotErr).
									Times(1)

								if tt.mockSlotErr == nil && tt.mockSlot != nil {
									if tt.mockOrder.CanBeCompleted(time.Now(), tt.mockSlot.EndTime) {
										test.orderRepo.EXPECT().
											UpdateStatus(gomock.Any(), gomock.Any()).
											Return(tt.mockUpdateOrder, tt.mockUpdateErr).
											Times(1)
									}
								}
							}
						}
					}
				}
			}

			result, err := test.service.CompleteOrder(tt.ctx, tt.orderID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, entity.OrderCompleted, result.Status)
			}
		})
	}
}
