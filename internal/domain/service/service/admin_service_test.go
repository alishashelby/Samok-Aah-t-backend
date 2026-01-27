package service

import (
	"context"
	"errors"
	"os"
	"testing"

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

type adminServiceTest struct {
	ctrl        *gomock.Controller
	adminRepo   *mocks.MockAdminRepository
	userRepo    *mocks.MockUserRepository
	bookingRepo *mocks.MockBookingRepository
	orderRepo   *mocks.MockOrderRepository
	service     *DefaultAdminService
	txManager   *mocks.MockTxManager
}

func setUpAdminServiceTest(t *testing.T) *adminServiceTest {
	t.Helper()

	ctrl := gomock.NewController(t)
	admin := mocks.NewMockAdminRepository(ctrl)
	user := mocks.NewMockUserRepository(ctrl)
	booking := mocks.NewMockBookingRepository(ctrl)
	order := mocks.NewMockOrderRepository(ctrl)
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

	adminService := NewDefaultAdminService(admin, user, booking, order, mockTxManager, log)

	return &adminServiceTest{
		ctrl:        ctrl,
		adminRepo:   admin,
		userRepo:    user,
		bookingRepo: booking,
		orderRepo:   order,
		service:     adminService,
		txManager:   mockTxManager,
	}
}

func TestAdminService_Create(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	ctxNotAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxNotAdmin = context.WithValue(ctxNotAdmin, service_const.RoleKey, "CLIENT")

	tests := []struct {
		name          string
		ctx           context.Context
		newAdminID    int64
		permissions   map[string]bool
		mockSaveErr   error
		expectedError error
	}{
		{
			name:        "successful create",
			ctx:         ctxAdmin,
			newAdminID:  2,
			permissions: map[string]bool{"can_edit": true},
		},
		{
			name:          "repo save error",
			ctx:           ctxAdmin,
			newAdminID:    3,
			permissions:   map[string]bool{"can_edit": false},
			mockSaveErr:   errors.New("save failed"),
			expectedError: errors.New("save failed"),
		},
		{
			name:          "not admin error",
			ctx:           ctxNotAdmin,
			newAdminID:    4,
			permissions:   map[string]bool{},
			expectedError: service_errors.ErrNotAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" {
				test.adminRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(tt.mockSaveErr).
					Times(1)
			}

			admin, err := test.service.Create(tt.ctx, tt.newAdminID, tt.permissions)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, admin)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, admin)
				assert.Equal(t, tt.newAdminID, admin.AuthID)
				assert.Equal(t, tt.permissions, admin.Permissions)
			}
		})
	}
}

func TestAdminService_GetAdminByID(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	tests := []struct {
		name          string
		ctx           context.Context
		adminID       int64
		mockAdmin     *entity.Admin
		mockErr       error
		expectedError error
	}{
		{
			name:      "found admin",
			ctx:       ctxAdmin,
			adminID:   2,
			mockAdmin: &entity.Admin{AuthID: 2, Permissions: map[string]bool{"can_edit": true}},
		},
		{
			name:          "admin not found",
			ctx:           ctxAdmin,
			adminID:       3,
			mockErr:       persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrAdminNotFound,
		},
		{
			name:          "repo error",
			ctx:           ctxAdmin,
			adminID:       4,
			mockErr:       errors.New("db error"),
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" {
				test.adminRepo.EXPECT().
					GetByID(tt.ctx, tt.adminID).
					Return(tt.mockAdmin, tt.mockErr)
			}

			admin, err := test.service.GetAdminByID(tt.ctx, tt.adminID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, admin)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, admin)
				assert.Equal(t, tt.mockAdmin, admin)
			}
		})
	}
}

func TestAdminService_UpdateAdmin(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	existingAdmin := &entity.Admin{
		AuthID:      2,
		Permissions: map[string]bool{"can_edit": false},
	}

	tests := []struct {
		name           string
		ctx            context.Context
		updatingAuthID int64
		permissions    map[string]bool
		mockGetAdmin   *entity.Admin
		mockGetErr     error
		mockUpdateRes  *entity.Admin
		mockUpdateErr  error
		expectedError  error
	}{
		{
			name:           "successful update",
			ctx:            ctxAdmin,
			updatingAuthID: 2,
			permissions:    map[string]bool{"can_edit": true, "can_delete": true},
			mockGetAdmin:   existingAdmin,
			mockUpdateRes:  &entity.Admin{AuthID: 2, Permissions: map[string]bool{"can_edit": true, "can_delete": true}},
		},
		{
			name:           "admin not found for update",
			ctx:            ctxAdmin,
			updatingAuthID: 3,
			permissions:    map[string]bool{"can_edit": true},
			mockGetErr:     persistence.ErrNoRowsFound,
			expectedError:  service_errors.ErrAdminNotFound,
		},
		{
			name:           "repo get error",
			ctx:            ctxAdmin,
			updatingAuthID: 4,
			permissions:    map[string]bool{"can_edit": true},
			mockGetErr:     errors.New("db error"),
			expectedError:  errors.New("db error"),
		},
		{
			name:           "repo update error",
			ctx:            ctxAdmin,
			updatingAuthID: 2,
			permissions:    map[string]bool{"can_edit": true},
			mockGetAdmin:   existingAdmin,
			mockUpdateRes:  nil,
			mockUpdateErr:  errors.New("update failed"),
			expectedError:  errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" {
				test.adminRepo.EXPECT().
					GetByAuthID(gomock.Any(), tt.updatingAuthID).
					Return(tt.mockGetAdmin, tt.mockGetErr).
					Times(1)

				if tt.mockGetErr == nil && tt.expectedError == nil || (tt.mockGetErr == nil && tt.mockUpdateErr != nil) {
					test.adminRepo.EXPECT().
						Update(gomock.Any(), gomock.Any()).
						Return(tt.mockUpdateRes, tt.mockUpdateErr).
						Times(1)
				}
			}

			admin, err := test.service.UpdateAdmin(tt.ctx, tt.updatingAuthID, tt.permissions)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, admin)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, admin)
				assert.Equal(t, tt.permissions, admin.Permissions)
			}
		})
	}
}

func TestAdminService_VerifyUser(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	unverifiedUser := &entity.User{
		ID:         2,
		IsVerified: false,
	}

	verifiedUser := &entity.User{
		ID:         2,
		IsVerified: true,
	}

	tests := []struct {
		name          string
		ctx           context.Context
		userID        int64
		mockGetUser   *entity.User
		mockGetErr    error
		mockUpdateRes *entity.User
		mockUpdateErr error
		expectedError error
	}{
		{
			name:          "successful verification",
			ctx:           ctxAdmin,
			userID:        2,
			mockGetUser:   unverifiedUser,
			mockUpdateRes: verifiedUser,
		},
		{
			name:          "user not found",
			ctx:           ctxAdmin,
			userID:        3,
			mockGetErr:    persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrUserNotFound,
		},
		{
			name:          "repo get error",
			ctx:           ctxAdmin,
			userID:        4,
			mockGetErr:    errors.New("db error"),
			expectedError: errors.New("db error"),
		},
		{
			name:          "repo update error",
			ctx:           ctxAdmin,
			userID:        2,
			mockGetUser:   unverifiedUser,
			mockUpdateRes: nil,
			mockUpdateErr: errors.New("update failed"),
			expectedError: errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" {
				test.userRepo.EXPECT().
					GetByID(gomock.Any(), tt.userID).
					Return(tt.mockGetUser, tt.mockGetErr).
					Times(1)

				if tt.mockGetErr == nil && tt.expectedError == nil || (tt.mockGetErr == nil && tt.mockUpdateErr != nil) {
					test.userRepo.EXPECT().
						Update(gomock.Any(), gomock.Any()).
						Return(tt.mockUpdateRes, tt.mockUpdateErr).
						Times(1)
				}
			}

			user, err := test.service.VerifyUser(tt.ctx, tt.userID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.True(t, user.IsVerified)
			}
		})
	}
}

func TestAdminService_GetBookingByID(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	booking := &entity.Booking{
		ID:             1,
		ClientID:       2,
		ModelServiceID: 3,
		Status:         entity.BookingPending,
	}

	tests := []struct {
		name          string
		ctx           context.Context
		bookingID     int64
		mockBooking   *entity.Booking
		mockErr       error
		expectedError error
	}{
		{
			name:        "found booking",
			ctx:         ctxAdmin,
			bookingID:   1,
			mockBooking: booking,
		},
		{
			name:          "booking not found",
			ctx:           ctxAdmin,
			bookingID:     2,
			mockErr:       persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrBookingNotFound,
		},
		{
			name:          "repo error",
			ctx:           ctxAdmin,
			bookingID:     3,
			mockErr:       errors.New("db error"),
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" {
				test.bookingRepo.EXPECT().
					GetByID(gomock.Any(), tt.bookingID).
					Return(tt.mockBooking, tt.mockErr).
					Times(1)
			}

			result, err := test.service.GetBookingByID(tt.ctx, tt.bookingID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockBooking, result)
			}
		})
	}
}

func TestAdminService_UpdateBookingStatus(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	existingBooking := &entity.Booking{
		ID:     1,
		Status: entity.BookingPending,
	}

	updatedBooking := &entity.Booking{
		ID:     1,
		Status: entity.BookingApproved,
	}

	tests := []struct {
		name           string
		ctx            context.Context
		bookingID      int64
		status         entity.BookingStatus
		mockGetBooking *entity.Booking
		mockGetErr     error
		mockUpdateRes  *entity.Booking
		mockUpdateErr  error
		expectedError  error
	}{
		{
			name:           "successful status update",
			ctx:            ctxAdmin,
			bookingID:      1,
			status:         entity.BookingApproved,
			mockGetBooking: existingBooking,
			mockUpdateRes:  updatedBooking,
		},
		{
			name:          "booking not found",
			ctx:           ctxAdmin,
			bookingID:     2,
			status:        entity.BookingApproved,
			mockGetErr:    persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrBookingNotFound,
		},
		{
			name:          "repo get error",
			ctx:           ctxAdmin,
			bookingID:     3,
			status:        entity.BookingApproved,
			mockGetErr:    errors.New("db error"),
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" {
				test.bookingRepo.EXPECT().
					GetByID(gomock.Any(), tt.bookingID).
					Return(tt.mockGetBooking, tt.mockGetErr).
					Times(1)

				if tt.mockGetErr == nil && tt.expectedError == nil {
					test.bookingRepo.EXPECT().
						Update(gomock.Any(), gomock.Any()).
						Return(tt.mockUpdateRes, tt.mockUpdateErr).
						Times(1)
				}
			}

			booking, err := test.service.UpdateBookingStatus(tt.ctx, tt.bookingID, tt.status)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, booking)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, booking)
				assert.Equal(t, tt.status, booking.Status)
			}
		})
	}
}

func TestAdminService_GetOrderByID(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	ctxNotAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxNotAdmin = context.WithValue(ctxNotAdmin, service_const.RoleKey, "USER")

	order := &entity.Order{
		ID:     1,
		Status: entity.OrderInTransit,
	}

	tests := []struct {
		name          string
		ctx           context.Context
		orderID       int64
		mockOrder     *entity.Order
		mockErr       error
		expectedError error
	}{
		{
			name:      "successful get order",
			ctx:       ctxAdmin,
			orderID:   1,
			mockOrder: order,
		},
		{
			name:          "order not found",
			ctx:           ctxAdmin,
			orderID:       2,
			mockErr:       persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrOrderNotFound,
		},
		{
			name:          "repo error",
			ctx:           ctxAdmin,
			orderID:       3,
			mockErr:       errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:          "not admin error",
			ctx:           ctxNotAdmin,
			orderID:       1,
			expectedError: service_errors.ErrNotAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" {
				test.orderRepo.EXPECT().
					GetByID(gomock.Any(), tt.orderID).
					Return(tt.mockOrder, tt.mockErr).
					Times(1)
			}

			result, err := test.service.GetOrderByID(tt.ctx, tt.orderID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockOrder, result)
			}
		})
	}
}

func TestAdminService_UpdateOrderStatus(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	ctxNotAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxNotAdmin = context.WithValue(ctxNotAdmin, service_const.RoleKey, "USER")

	existingOrder := &entity.Order{
		ID:     1,
		Status: entity.OrderInTransit,
	}

	updatedOrder := &entity.Order{
		ID:     1,
		Status: entity.OrderCompleted,
	}

	tests := []struct {
		name          string
		ctx           context.Context
		orderID       int64
		status        entity.OrderStatus
		mockGetOrder  *entity.Order
		mockGetErr    error
		mockUpdateRes *entity.Order
		mockUpdateErr error
		expectedError error
	}{
		{
			name:          "successful order status update",
			ctx:           ctxAdmin,
			orderID:       1,
			status:        entity.OrderCompleted,
			mockGetOrder:  existingOrder,
			mockUpdateRes: updatedOrder,
		},
		{
			name:          "order not found",
			ctx:           ctxAdmin,
			orderID:       2,
			status:        entity.OrderCompleted,
			mockGetErr:    persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrOrderNotFound,
		},
		{
			name:          "repo get error",
			ctx:           ctxAdmin,
			orderID:       3,
			status:        entity.OrderCompleted,
			mockGetErr:    errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:          "repo update error",
			ctx:           ctxAdmin,
			orderID:       1,
			status:        entity.OrderCompleted,
			mockGetOrder:  existingOrder,
			mockUpdateErr: errors.New("update failed"),
			expectedError: errors.New("update failed"),
		},
		{
			name:          "not admin error",
			ctx:           ctxNotAdmin,
			orderID:       1,
			status:        entity.OrderCompleted,
			expectedError: service_errors.ErrNotAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" {
				test.orderRepo.EXPECT().
					GetByID(gomock.Any(), tt.orderID).
					Return(tt.mockGetOrder, tt.mockGetErr).
					Times(1)

				if tt.mockGetErr == nil && tt.expectedError == nil || (tt.mockGetErr == nil && tt.mockUpdateErr != nil) {
					test.orderRepo.EXPECT().
						UpdateStatus(gomock.Any(), gomock.Any()).
						Return(tt.mockUpdateRes, tt.mockUpdateErr).
						Times(1)
				}
			}

			order, err := test.service.UpdateOrderStatus(tt.ctx, tt.orderID, tt.status)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, order)
				assert.Equal(t, tt.status, order.Status)
			}
		})
	}
}

func TestAdminService_GetAllUsers(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	ctxNotAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxNotAdmin = context.WithValue(ctxNotAdmin, service_const.RoleKey, "USER")

	users := []*entity.User{
		{ID: 1, IsVerified: true},
		{ID: 2, IsVerified: false},
		{ID: 3, IsVerified: true},
	}

	tests := []struct {
		name          string
		ctx           context.Context
		page          *int64
		limit         *int64
		mockUsers     []*entity.User
		mockErr       error
		expectedError error
		expectCall    bool
	}{
		{
			name:       "successful get all users without pagination",
			ctx:        ctxAdmin,
			page:       nil,
			limit:      nil,
			mockUsers:  users,
			expectCall: true,
		},
		{
			name:       "successful get all users with pagination",
			ctx:        ctxAdmin,
			page:       int64Ptr(1),
			limit:      int64Ptr(10),
			mockUsers:  users,
			expectCall: true,
		},
		{
			name:          "repo error",
			ctx:           ctxAdmin,
			page:          nil,
			limit:         nil,
			mockErr:       errors.New("database error"),
			expectedError: errors.New("database error"),
			expectCall:    true,
		},
		{
			name:          "not admin error",
			ctx:           ctxNotAdmin,
			page:          nil,
			limit:         nil,
			expectedError: service_errors.ErrNotAdmin,
			expectCall:    false,
		},
		{
			name:       "successful get all users with empty result",
			ctx:        ctxAdmin,
			page:       nil,
			limit:      nil,
			mockUsers:  []*entity.User{},
			expectCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" && tt.expectCall {
				test.userRepo.EXPECT().
					GetAll(gomock.Any(), gomock.Any()).
					Return(tt.mockUsers, tt.mockErr).
					Times(1)
			}

			result, err := test.service.GetAllUsers(tt.ctx, tt.page, tt.limit)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockUsers, result)
			}
		})
	}
}

func TestAdminService_GetAllBookings(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	ctxNotAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxNotAdmin = context.WithValue(ctxNotAdmin, service_const.RoleKey, "USER")

	bookings := []*entity.Booking{
		{ID: 1, ClientID: 1, Status: entity.BookingPending},
		{ID: 2, ClientID: 2, Status: entity.BookingApproved},
		{ID: 3, ClientID: 3, Status: entity.BookingApproved},
	}

	tests := []struct {
		name          string
		ctx           context.Context
		page          *int64
		limit         *int64
		mockBookings  []*entity.Booking
		mockErr       error
		expectedError error
		expectCall    bool
	}{
		{
			name:         "successful get all bookings without pagination",
			ctx:          ctxAdmin,
			page:         nil,
			limit:        nil,
			mockBookings: bookings,
			expectCall:   true,
		},
		{
			name:         "successful get all bookings with pagination",
			ctx:          ctxAdmin,
			page:         int64Ptr(1),
			limit:        int64Ptr(5),
			mockBookings: bookings,
			expectCall:   true,
		},
		{
			name:          "repo error",
			ctx:           ctxAdmin,
			page:          nil,
			limit:         nil,
			mockErr:       errors.New("database error"),
			expectedError: errors.New("database error"),
			expectCall:    true,
		},
		{
			name:          "not admin error",
			ctx:           ctxNotAdmin,
			page:          nil,
			limit:         nil,
			expectedError: service_errors.ErrNotAdmin,
			expectCall:    false,
		},
		{
			name:         "successful get all bookings with empty result",
			ctx:          ctxAdmin,
			page:         nil,
			limit:        nil,
			mockBookings: []*entity.Booking{},
			expectCall:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" && tt.expectCall {
				test.bookingRepo.EXPECT().
					GetAll(gomock.Any(), gomock.Any()).
					Return(tt.mockBookings, tt.mockErr).
					Times(1)
			}

			result, err := test.service.GetAllBookings(tt.ctx, tt.page, tt.limit)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockBookings, result)
			}
		})
	}
}

func TestAdminService_GetAllOrders(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	ctxNotAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxNotAdmin = context.WithValue(ctxNotAdmin, service_const.RoleKey, "USER")

	orders := []*entity.Order{
		{ID: 1, Status: entity.OrderInTransit},
		{ID: 2, Status: entity.OrderConfirmed},
		{ID: 3, Status: entity.OrderCompleted},
	}

	tests := []struct {
		name          string
		ctx           context.Context
		page          *int64
		limit         *int64
		mockOrders    []*entity.Order
		mockErr       error
		expectedError error
		expectCall    bool
	}{
		{
			name:       "successful get all orders without pagination",
			ctx:        ctxAdmin,
			page:       nil,
			limit:      nil,
			mockOrders: orders,
			expectCall: true,
		},
		{
			name:       "successful get all orders with pagination",
			ctx:        ctxAdmin,
			page:       int64Ptr(2),
			limit:      int64Ptr(20),
			mockOrders: orders,
			expectCall: true,
		},
		{
			name:          "repo error",
			ctx:           ctxAdmin,
			page:          nil,
			limit:         nil,
			mockErr:       errors.New("database error"),
			expectedError: errors.New("database error"),
			expectCall:    true,
		},
		{
			name:          "not admin error",
			ctx:           ctxNotAdmin,
			page:          nil,
			limit:         nil,
			expectedError: service_errors.ErrNotAdmin,
			expectCall:    false,
		},
		{
			name:       "successful get all orders with empty result",
			ctx:        ctxAdmin,
			page:       nil,
			limit:      nil,
			mockOrders: []*entity.Order{},
			expectCall: true,
		},
		{
			name:       "successful get all orders with nil page and limit",
			ctx:        ctxAdmin,
			page:       nil,
			limit:      nil,
			mockOrders: orders,
			expectCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.RoleKey) == "ADMIN" && tt.expectCall {
				test.orderRepo.EXPECT().
					GetAll(gomock.Any(), gomock.Any()).
					Return(tt.mockOrders, tt.mockErr).
					Times(1)
			}

			result, err := test.service.GetAllOrders(tt.ctx, tt.page, tt.limit)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockOrders, result)
			}
		})
	}
}

func TestAdminService_GetAllMethods_EdgeCases(t *testing.T) {
	test := setUpAdminServiceTest(t)
	defer test.ctrl.Finish()

	ctxAdmin := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxAdmin = context.WithValue(ctxAdmin, service_const.RoleKey, "ADMIN")

	t.Run("get all users with negative pagination", func(t *testing.T) {
		page := int64Ptr(-1)
		limit := int64Ptr(-10)

		users := []*entity.User{
			{ID: 1},
		}

		test.userRepo.EXPECT().
			GetAll(gomock.Any(), gomock.Any()).
			Return(users, nil).
			Times(1)

		result, err := test.service.GetAllUsers(ctxAdmin, page, limit)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, users, result)
	})

	t.Run("get all bookings with zero pagination", func(t *testing.T) {
		page := int64Ptr(0)
		limit := int64Ptr(0)

		bookings := []*entity.Booking{
			{ID: 1, ClientID: 1, Status: entity.BookingPending},
		}

		test.bookingRepo.EXPECT().
			GetAll(gomock.Any(), gomock.Any()).
			Return(bookings, nil).
			Times(1)

		result, err := test.service.GetAllBookings(ctxAdmin, page, limit)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, bookings, result)
	})

	t.Run("get all orders with large pagination", func(t *testing.T) {
		page := int64Ptr(1000)
		limit := int64Ptr(1000)

		orders := []*entity.Order{}

		test.orderRepo.EXPECT().
			GetAll(gomock.Any(), gomock.Any()).
			Return(orders, nil).
			Times(1)

		result, err := test.service.GetAllOrders(ctxAdmin, page, limit)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})
}

func int64Ptr(i int64) *int64 {
	return &i
}
