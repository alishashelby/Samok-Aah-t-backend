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

type bookingServiceTest struct {
	ctrl             *gomock.Controller
	bookingRepo      *mocks.MockBookingRepository
	slotRepo         *mocks.MockSlotRepository
	orderRepo        *mocks.MockOrderRepository
	userRepo         *mocks.MockUserRepository
	modelServiceRepo *mocks.MockModelServiceRepository
	txManager        *mocks.MockTxManager
	service          *DefaultBookingService
}

func setUpBookingServiceTest(t *testing.T) *bookingServiceTest {
	t.Helper()

	ctrl := gomock.NewController(t)
	bookingRepo := mocks.NewMockBookingRepository(ctrl)
	slotRepo := mocks.NewMockSlotRepository(ctrl)
	orderRepo := mocks.NewMockOrderRepository(ctrl)
	userRepo := mocks.NewMockUserRepository(ctrl)
	modelServiceRepo := mocks.NewMockModelServiceRepository(ctrl)
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

	if err = os.Setenv(service_const.DotEnvBookingExpiration, "300"); err != nil {
		t.Fatal(err)
	}

	bookingService, err := NewDefaultBookingService(
		bookingRepo, slotRepo, userRepo, modelServiceRepo, orderRepo, mockTxManager, log,
	)
	if err != nil {
		t.Fatal(err)
	}

	return &bookingServiceTest{
		ctrl:             ctrl,
		bookingRepo:      bookingRepo,
		slotRepo:         slotRepo,
		orderRepo:        orderRepo,
		userRepo:         userRepo,
		modelServiceRepo: modelServiceRepo,
		txManager:        mockTxManager,
		service:          bookingService,
	}
}
func TestBookingService_CreateBooking(t *testing.T) {
	test := setUpBookingServiceTest(t)
	defer test.ctrl.Finish()

	ctxClient := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxClient = context.WithValue(ctxClient, service_const.RoleKey, "CLIENT")

	verifiedClient := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	availableSlot := &entity.Slot{
		ID:     1,
		Status: entity.SlotAvailable,
	}

	reservedSlot := &entity.Slot{
		ID:     1,
		Status: entity.SlotReserved,
	}

	tests := []struct {
		name           string
		ctx            context.Context
		modelServiceID int64
		slotID         int64
		mockUser       *entity.User
		mockUserErr    error
		mockSlot       *entity.Slot
		mockSlotErr    error
		mockUpdateSlot *entity.Slot
		mockUpdateErr  error
		mockSaveErr    error
		expectedError  error
	}{
		{
			name:           "successful booking creation",
			ctx:            ctxClient,
			modelServiceID: 1,
			slotID:         1,
			mockUser:       verifiedClient,
			mockSlot:       availableSlot,
			mockUpdateSlot: reservedSlot,
		},
		{
			name:           "client not verified",
			ctx:            ctxClient,
			modelServiceID: 1,
			slotID:         1,
			mockUser:       &entity.User{ID: 1, IsVerified: false},
			expectedError:  service_errors.ErrNotVerifiedClient,
		},
		{
			name:           "slot not found",
			ctx:            ctxClient,
			modelServiceID: 1,
			slotID:         1,
			mockUser:       verifiedClient,
			mockSlotErr:    persistence.ErrNoRowsFound,
			expectedError:  service_errors.ErrSlotIsNotFound,
		},
		{
			name:           "slot not available",
			ctx:            ctxClient,
			modelServiceID: 1,
			slotID:         1,
			mockUser:       verifiedClient,
			mockSlot:       &entity.Slot{ID: 1, Status: entity.SlotBooked},
			expectedError:  service_errors.ErrSlotNotAvailable,
		},
		{
			name:           "slot already reserved",
			ctx:            ctxClient,
			modelServiceID: 1,
			slotID:         1,
			mockUser:       verifiedClient,
			mockSlot:       &entity.Slot{ID: 1, Status: entity.SlotReserved},
			expectedError:  service_errors.ErrSlotNotAvailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "CLIENT" {
				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(tt.mockUser, tt.mockUserErr).
					Times(1)
			}

			if tt.mockUserErr == nil && tt.mockUser != nil && tt.mockUser.IsVerified {
				test.slotRepo.EXPECT().
					GetByID(gomock.Any(), tt.slotID).
					Return(tt.mockSlot, tt.mockSlotErr).
					Times(1)

				if tt.mockSlotErr == nil && tt.mockSlot != nil && tt.mockSlot.IsAvailable() {
					test.txManager.EXPECT().
						WithTransaction(gomock.Any(), gomock.Any()).
						DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
							return fn(ctx)
						}).
						Times(1)

					test.slotRepo.EXPECT().
						Update(gomock.Any(), gomock.Any()).
						Return(tt.mockUpdateSlot, tt.mockUpdateErr).
						Times(1)

					if tt.mockUpdateErr == nil {
						test.bookingRepo.EXPECT().
							Save(gomock.Any(), gomock.Any()).
							Return(tt.mockSaveErr).
							Times(1)
					}
				}
			}

			apartment := 5
			entrance := 1
			floor := 2
			comment := "Test comment"

			booking, err := test.service.CreateBooking(
				tt.ctx,
				tt.modelServiceID,
				tt.slotID,
				"Test Street",
				10,
				&apartment,
				&entrance,
				&floor,
				&comment,
			)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, booking)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, booking)
				assert.Equal(t, tt.modelServiceID, booking.ModelServiceID)
				assert.Equal(t, tt.slotID, booking.SlotID)
			}
		})
	}
}

func TestBookingService_ApproveBooking(t *testing.T) {
	test := setUpBookingServiceTest(t)
	defer test.ctrl.Finish()

	ctxModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxModel = context.WithValue(ctxModel, service_const.RoleKey, "MODEL")

	verifiedModel := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	pendingBooking := entity.NewBooking(2, 3, 4, entity.Address{}, 5*time.Minute)
	pendingBooking.ID = 1

	reservedSlot := &entity.Slot{
		ID:     4,
		Status: entity.SlotReserved,
	}

	bookedSlot := &entity.Slot{
		ID:     4,
		Status: entity.SlotBooked,
	}

	modelService := &entity.ModelService{
		ID:      3,
		ModelID: 1,
	}

	tests := []struct {
		name                 string
		ctx                  context.Context
		bookingID            int64
		mockModel            *entity.User
		mockModelErr         error
		mockBooking          *entity.Booking
		mockBookingErr       error
		mockModelService     *entity.ModelService
		mockModelServiceErr  error
		mockSlot             *entity.Slot
		mockSlotErr          error
		mockUpdateSlot       *entity.Slot
		mockUpdateSlotErr    error
		mockUpdateBooking    *entity.Booking
		mockUpdateBookingErr error
		mockSaveOrderErr     error
		expectedError        error
		expectTransaction    bool
	}{
		{
			name:              "successful booking approval",
			ctx:               ctxModel,
			bookingID:         1,
			mockModel:         verifiedModel,
			mockBooking:       pendingBooking,
			mockModelService:  modelService,
			mockSlot:          reservedSlot,
			mockUpdateSlot:    bookedSlot,
			mockUpdateBooking: &entity.Booking{ID: 1, Status: entity.BookingApproved},
			expectTransaction: true,
		},
		{
			name:              "model not verified",
			ctx:               ctxModel,
			bookingID:         1,
			mockModel:         &entity.User{ID: 1, IsVerified: false},
			expectedError:     service_errors.ErrNotVerifiedModel,
			expectTransaction: false,
		},
		{
			name:              "booking not found",
			ctx:               ctxModel,
			bookingID:         1,
			mockModel:         verifiedModel,
			mockBookingErr:    persistence.ErrNoRowsFound,
			expectedError:     service_errors.ErrBookingNotFound,
			expectTransaction: false,
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
				test.bookingRepo.EXPECT().
					GetByID(gomock.Any(), tt.bookingID).
					Return(tt.mockBooking, tt.mockBookingErr).
					Times(1)

				if tt.mockBookingErr == nil && tt.mockBooking != nil && tt.mockBooking.CanBeApproved() {
					test.modelServiceRepo.EXPECT().
						GetByID(gomock.Any(), tt.mockBooking.ModelServiceID, false).
						Return(tt.mockModelService, tt.mockModelServiceErr).
						Times(1)

					if tt.mockModelServiceErr == nil && tt.mockModelService != nil &&
						tt.mockModelService.ModelID == tt.mockModel.ID {

						if !tt.mockBooking.IsExpired(time.Now()) {
							test.slotRepo.EXPECT().
								GetByID(gomock.Any(), tt.mockBooking.SlotID).
								Return(tt.mockSlot, tt.mockSlotErr).
								Times(1)

							if tt.mockSlotErr == nil && tt.mockSlot != nil &&
								tt.mockSlot.IsCorrectTransition(entity.SlotBooked) {

								if tt.expectTransaction {
									test.txManager.EXPECT().
										WithTransaction(gomock.Any(), gomock.Any()).
										DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
											return fn(ctx)
										}).
										Times(1)

									test.slotRepo.EXPECT().
										Update(gomock.Any(), gomock.Any()).
										Return(tt.mockUpdateSlot, tt.mockUpdateSlotErr).
										Times(1)

									if tt.mockUpdateSlotErr == nil {
										test.bookingRepo.EXPECT().
											Update(gomock.Any(), gomock.Any()).
											Return(tt.mockUpdateBooking, tt.mockUpdateBookingErr).
											Times(1)

										if tt.mockUpdateBookingErr == nil {
											test.orderRepo.EXPECT().
												Save(gomock.Any(), gomock.Any()).
												Return(tt.mockSaveOrderErr).
												Times(1)
										}
									}
								}
							}
						}
					}
				}
			}

			booking, err := test.service.ApproveBooking(tt.ctx, tt.bookingID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, booking)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, booking)
				assert.Equal(t, entity.BookingApproved, booking.Status)
			}
		})
	}
}

func TestBookingService_NewDefaultBookingService_Errors(t *testing.T) {
	tests := []struct {
		name          string
		envValue      string
		expectedError error
	}{
		{
			name:          "missing env variable",
			envValue:      "",
			expectedError: service_errors.ErrLoadingTTL,
		},
		{
			name:          "invalid ttl format",
			envValue:      "not-a-number",
			expectedError: service_errors.ErrParsingTTL,
		},
		{
			name:          "negative ttl",
			envValue:      "-1",
			expectedError: service_errors.ErrNotPositiveTTL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalValue := os.Getenv(service_const.DotEnvBookingExpiration)
			defer func() {
				if err := os.Setenv(service_const.DotEnvBookingExpiration, originalValue); err != nil {
					t.Fatal(err)
				}
			}()

			if err := os.Setenv(service_const.DotEnvBookingExpiration, tt.envValue); err != nil {
				t.Fatal(err)
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			cfg := &config.LogConfig{}
			cfg.Logger.Level = "info"
			tmpDir := os.TempDir()
			cfg.Logger.LogsDir = tmpDir
			cfg.Logger.LogsFile = "test.log"
			log, _ := pkg.NewDualLogger(cfg)

			_, err := NewDefaultBookingService(
				mocks.NewMockBookingRepository(ctrl),
				mocks.NewMockSlotRepository(ctrl),
				mocks.NewMockUserRepository(ctrl),
				mocks.NewMockModelServiceRepository(ctrl),
				mocks.NewMockOrderRepository(ctrl),
				mocks.NewMockTxManager(ctrl),
				log,
			)

			assert.EqualError(t, err, tt.expectedError.Error())
		})
	}
}

func TestBookingService_RejectBooking(t *testing.T) {
	test := setUpBookingServiceTest(t)
	defer test.ctrl.Finish()

	ctxModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxModel = context.WithValue(ctxModel, service_const.RoleKey, "MODEL")

	ctxNotModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxNotModel = context.WithValue(ctxNotModel, service_const.RoleKey, "CLIENT")

	verifiedModel := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	pendingBooking := entity.NewBooking(2, 3, 4, entity.Address{}, 5*time.Minute)
	pendingBooking.ID = 1
	pendingBooking.Status = entity.BookingPending
	pendingBooking.CreatedAt = time.Now().Add(-1 * time.Minute)

	expiredBooking := entity.NewBooking(2, 3, 4, entity.Address{}, 1*time.Minute)
	expiredBooking.ID = 2
	expiredBooking.Status = entity.BookingPending
	expiredBooking.CreatedAt = time.Now().Add(-2 * time.Minute)

	reservedSlot := &entity.Slot{
		ID:     4,
		Status: entity.SlotReserved,
	}

	availableSlot := &entity.Slot{
		ID:     4,
		Status: entity.SlotAvailable,
	}

	modelService := &entity.ModelService{
		ID:      3,
		ModelID: 1,
	}

	tests := []struct {
		name                 string
		ctx                  context.Context
		bookingID            int64
		mockModel            *entity.User
		mockModelErr         error
		mockBooking          *entity.Booking
		mockBookingErr       error
		mockModelService     *entity.ModelService
		mockModelServiceErr  error
		mockSlot             *entity.Slot
		mockSlotErr          error
		mockUpdateSlot       *entity.Slot
		mockUpdateSlotErr    error
		mockUpdateBooking    *entity.Booking
		mockUpdateBookingErr error
		expectedError        error
		expectTransaction    bool
		expectSlotTransition bool
	}{
		{
			name:                 "successful booking rejection",
			ctx:                  ctxModel,
			bookingID:            1,
			mockModel:            verifiedModel,
			mockBooking:          pendingBooking,
			mockModelService:     modelService,
			mockSlot:             reservedSlot,
			mockUpdateSlot:       availableSlot,
			mockUpdateBooking:    &entity.Booking{ID: 1, Status: entity.BookingRejected},
			expectTransaction:    true,
			expectSlotTransition: true,
		},
		{
			name:          "not a model error",
			ctx:           ctxNotModel,
			bookingID:     1,
			expectedError: service_errors.ErrNotAModel,
		},
		{
			name:          "model not verified",
			ctx:           ctxModel,
			bookingID:     1,
			mockModel:     &entity.User{ID: 1, IsVerified: false},
			expectedError: service_errors.ErrNotVerifiedModel,
		},
		{
			name:           "booking not found",
			ctx:            ctxModel,
			bookingID:      1,
			mockModel:      verifiedModel,
			mockBookingErr: persistence.ErrNoRowsFound,
			expectedError:  service_errors.ErrBookingNotFound,
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
				test.bookingRepo.EXPECT().
					GetByID(gomock.Any(), tt.bookingID).
					Return(tt.mockBooking, tt.mockBookingErr).
					Times(1)

				if tt.mockBookingErr == nil && tt.mockBooking != nil && tt.mockBooking.CanBeRejected() {
					test.modelServiceRepo.EXPECT().
						GetByID(gomock.Any(), tt.mockBooking.ModelServiceID, false).
						Return(tt.mockModelService, tt.mockModelServiceErr).
						Times(1)

					if tt.mockModelServiceErr == nil && tt.mockModelService != nil &&
						tt.mockModelService.ModelID == tt.mockModel.ID && !tt.mockBooking.IsExpired(time.Now()) {

						test.slotRepo.EXPECT().
							GetByID(gomock.Any(), tt.mockBooking.SlotID).
							Return(tt.mockSlot, tt.mockSlotErr).
							Times(1)

						if tt.mockSlotErr == nil && tt.mockSlot != nil &&
							tt.mockSlot.IsCorrectTransition(entity.SlotAvailable) {

							if tt.expectTransaction {
								test.txManager.EXPECT().
									WithTransaction(gomock.Any(), gomock.Any()).
									DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
										return fn(ctx)
									}).
									Times(1)

								test.slotRepo.EXPECT().
									Update(gomock.Any(), gomock.Any()).
									Return(tt.mockUpdateSlot, tt.mockUpdateSlotErr).
									Times(1)

								if tt.mockUpdateSlotErr == nil {
									test.bookingRepo.EXPECT().
										Update(gomock.Any(), gomock.Any()).
										Return(tt.mockUpdateBooking, tt.mockUpdateBookingErr).
										Times(1)
								}
							}
						}
					}
				}
			}

			booking, err := test.service.RejectBooking(tt.ctx, tt.bookingID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, booking)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, booking)
				assert.Equal(t, entity.BookingRejected, booking.Status)
			}
		})
	}
}

func TestBookingService_CancelBookingByClient(t *testing.T) {
	test := setUpBookingServiceTest(t)
	defer test.ctrl.Finish()

	ctxClient := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxClient = context.WithValue(ctxClient, service_const.RoleKey, "CLIENT")

	ctxNotClient := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxNotClient = context.WithValue(ctxNotClient, service_const.RoleKey, "MODEL")

	verifiedClient := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	pendingBooking := entity.NewBooking(1, 3, 4, entity.Address{}, 5*time.Minute)
	pendingBooking.ID = 1
	pendingBooking.Status = entity.BookingPending
	pendingBooking.CreatedAt = time.Now().Add(-1 * time.Minute)

	approvedBooking := entity.NewBooking(1, 3, 4, entity.Address{}, 5*time.Minute)
	approvedBooking.ID = 2
	approvedBooking.Status = entity.BookingApproved
	approvedBooking.CreatedAt = time.Now().Add(-1 * time.Minute)

	expiredBooking := entity.NewBooking(1, 3, 4, entity.Address{}, 1*time.Minute)
	expiredBooking.ID = 3
	expiredBooking.Status = entity.BookingPending
	expiredBooking.CreatedAt = time.Now().Add(-2 * time.Minute)

	reservedSlot := &entity.Slot{
		ID:     4,
		Status: entity.SlotReserved,
	}

	availableSlot := &entity.Slot{
		ID:     4,
		Status: entity.SlotAvailable,
	}

	tests := []struct {
		name                 string
		ctx                  context.Context
		bookingID            int64
		mockClient           *entity.User
		mockClientErr        error
		mockBooking          *entity.Booking
		mockBookingErr       error
		mockSlot             *entity.Slot
		mockSlotErr          error
		mockUpdateSlot       *entity.Slot
		mockUpdateSlotErr    error
		mockUpdateBooking    *entity.Booking
		mockUpdateBookingErr error
		expectedError        error
		expectTransaction    bool
	}{
		{
			name:              "successful booking cancellation by client",
			ctx:               ctxClient,
			bookingID:         1,
			mockClient:        verifiedClient,
			mockBooking:       pendingBooking,
			mockSlot:          reservedSlot,
			mockUpdateSlot:    availableSlot,
			mockUpdateBooking: &entity.Booking{ID: 1, Status: entity.BookingCancelled},
			expectTransaction: true,
		},
		{
			name:          "not a client error",
			ctx:           ctxNotClient,
			bookingID:     1,
			expectedError: service_errors.ErrNotClient,
		},
		{
			name:          "client not verified",
			ctx:           ctxClient,
			bookingID:     1,
			mockClient:    &entity.User{ID: 1, IsVerified: false},
			expectedError: service_errors.ErrNotVerifiedClient,
		},
		{
			name:           "booking not found",
			ctx:            ctxClient,
			bookingID:      1,
			mockClient:     verifiedClient,
			mockBookingErr: persistence.ErrNoRowsFound,
			expectedError:  service_errors.ErrBookingNotFound,
		},
		{
			name:          "client not owner of booking",
			ctx:           ctxClient,
			bookingID:     1,
			mockClient:    verifiedClient,
			mockBooking:   &entity.Booking{ID: 1, ClientID: 999, Status: entity.BookingPending},
			expectedError: service_errors.ErrClientIsNotOwnerOfBooking,
		},
		{
			name:          "booking already approved",
			ctx:           ctxClient,
			bookingID:     2,
			mockClient:    verifiedClient,
			mockBooking:   approvedBooking,
			expectedError: service_errors.ErrInvalidBookingState,
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
				test.bookingRepo.EXPECT().
					GetByID(gomock.Any(), tt.bookingID).
					Return(tt.mockBooking, tt.mockBookingErr).
					Times(1)

				if tt.mockBookingErr == nil && tt.mockBooking != nil &&
					tt.mockBooking.ClientID == tt.mockClient.ID &&
					tt.mockBooking.CanBeCancelledByClient() && !tt.mockBooking.IsExpired(time.Now()) {

					test.slotRepo.EXPECT().
						GetByID(gomock.Any(), tt.mockBooking.SlotID).
						Return(tt.mockSlot, tt.mockSlotErr).
						Times(1)

					if tt.mockSlotErr == nil && tt.mockSlot != nil &&
						tt.mockSlot.IsCorrectTransition(entity.SlotAvailable) {

						if tt.expectTransaction {
							test.txManager.EXPECT().
								WithTransaction(gomock.Any(), gomock.Any()).
								DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
									return fn(ctx)
								}).
								Times(1)

							test.slotRepo.EXPECT().
								Update(gomock.Any(), gomock.Any()).
								Return(tt.mockUpdateSlot, tt.mockUpdateSlotErr).
								Times(1)

							if tt.mockUpdateSlotErr == nil {
								test.bookingRepo.EXPECT().
									Update(gomock.Any(), gomock.Any()).
									Return(tt.mockUpdateBooking, tt.mockUpdateBookingErr).
									Times(1)
							}
						}
					}
				}
			}

			booking, err := test.service.CancelBookingByClient(tt.ctx, tt.bookingID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, booking)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, booking)
				assert.Equal(t, entity.BookingCancelled, booking.Status)
			}
		})
	}
}

func TestBookingService_CheckModelRestrictions_EdgeCases(t *testing.T) {
	test := setUpBookingServiceTest(t)
	defer test.ctrl.Finish()

	tests := []struct {
		name          string
		ctx           context.Context
		mockUser      *entity.User
		mockUserErr   error
		expectedError error
	}{
		{
			name:     "successful model verification",
			ctx:      context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "MODEL"),
			mockUser: &entity.User{ID: 1, AuthID: int64(1), IsVerified: true},
		},
		{
			name:          "not a model role",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "CLIENT"),
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
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, model)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, model)
				assert.True(t, model.IsVerified)
			}
		})
	}
}

func TestBookingService_CheckClientRestrictions_EdgeCases(t *testing.T) {
	test := setUpBookingServiceTest(t)
	defer test.ctrl.Finish()

	tests := []struct {
		name          string
		ctx           context.Context
		mockUser      *entity.User
		mockUserErr   error
		expectedError error
	}{
		{
			name:     "successful client verification",
			ctx:      context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "CLIENT"),
			mockUser: &entity.User{ID: 1, AuthID: int64(1), IsVerified: true},
		},
		{
			name:          "not a client role",
			ctx:           context.WithValue(context.WithValue(context.Background(), service_const.AuthIDKey, int64(1)), service_const.RoleKey, "MODEL"),
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

			client, err := test.service.checkClientRestrictions(tt.ctx, &[]int64{tt.ctx.Value(service_const.AuthIDKey).(int64)}[0])

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, client)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.True(t, client.IsVerified)
			}
		})
	}
}

func TestBookingService_CheckIfModelIsAnOwner_EdgeCases(t *testing.T) {
	test := setUpBookingServiceTest(t)
	defer test.ctrl.Finish()

	tests := []struct {
		name                string
		modelID             int64
		modelServiceID      int64
		mockModelService    *entity.ModelService
		mockModelServiceErr error
		expectedError       error
	}{
		{
			name:             "model is owner of service",
			modelID:          1,
			modelServiceID:   100,
			mockModelService: &entity.ModelService{ID: 100, ModelID: 1},
		},
		{
			name:             "model is not owner of service",
			modelID:          1,
			modelServiceID:   100,
			mockModelService: &entity.ModelService{ID: 100, ModelID: 999},
			expectedError:    service_errors.ErrModelIsNotAnOwnerOfService,
		},
		{
			name:                "model service not found",
			modelID:             1,
			modelServiceID:      100,
			mockModelServiceErr: persistence.ErrNoRowsFound,
			expectedError:       service_errors.ErrServiceIsNotFound,
		},
		{
			name:                "repository error when getting model service",
			modelID:             1,
			modelServiceID:      100,
			mockModelServiceErr: errors.New("database error"),
			expectedError:       errors.New("database error"),
		},
		{
			name:             "model service with zero ID",
			modelID:          1,
			modelServiceID:   0,
			mockModelService: &entity.ModelService{ID: 0, ModelID: 1},
		},
		{
			name:             "model with zero ID checking ownership",
			modelID:          0,
			modelServiceID:   100,
			mockModelService: &entity.ModelService{ID: 100, ModelID: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			test.modelServiceRepo.EXPECT().
				GetByID(ctx, tt.modelServiceID, false).
				Return(tt.mockModelService, tt.mockModelServiceErr).
				Times(1)

			err := test.service.checkIfModelIsAnOwner(ctx, tt.modelID, tt.modelServiceID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
