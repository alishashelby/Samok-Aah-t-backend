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

type userServiceTest struct {
	ctrl      *gomock.Controller
	userRepo  *mocks.MockUserRepository
	txManager *mocks.MockTxManager
	service   *DefaultUserService
}

func setUpUserServiceTest(t *testing.T) *userServiceTest {
	t.Helper()

	ctrl := gomock.NewController(t)
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

	userService := NewDefaultUserService(userRepo, mockTxManager, log)

	return &userServiceTest{
		ctrl:      ctrl,
		userRepo:  userRepo,
		txManager: mockTxManager,
		service:   userService,
	}
}

func TestUserService_Create(t *testing.T) {
	test := setUpUserServiceTest(t)
	defer test.ctrl.Finish()

	ctx := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))

	name := "John Doe"
	birthDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		ctx           context.Context
		nameParam     string
		birthDate     time.Time
		mockSaveErr   error
		expectedError error
	}{
		{
			name:      "successful user creation",
			ctx:       ctx,
			nameParam: name,
			birthDate: birthDate,
		},
		{
			name:          "repo save error",
			ctx:           ctx,
			nameParam:     name,
			birthDate:     birthDate,
			mockSaveErr:   errors.New("save failed"),
			expectedError: errors.New("save failed"),
		},
		{
			name:          "missing auth id in context",
			ctx:           context.Background(),
			nameParam:     name,
			birthDate:     birthDate,
			expectedError: service_errors.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.AuthIDKey) != nil {
				test.userRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(tt.mockSaveErr).
					Times(1)
			}

			result, err := test.service.Create(tt.ctx, tt.nameParam, tt.birthDate)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.nameParam, result.Name)
				assert.Equal(t, tt.birthDate, result.BirthDate)
				assert.Equal(t, int64(1), result.AuthID)
			}
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	test := setUpUserServiceTest(t)
	defer test.ctrl.Finish()

	ctx := context.Background()

	user := &entity.User{
		ID:         1,
		AuthID:     2,
		Name:       "John Doe",
		IsVerified: true,
	}

	tests := []struct {
		name          string
		ctx           context.Context
		userID        int64
		mockUser      *entity.User
		mockErr       error
		expectedError error
	}{
		{
			name:     "successful get user",
			ctx:      ctx,
			userID:   1,
			mockUser: user,
		},
		{
			name:          "user not found",
			ctx:           ctx,
			userID:        2,
			mockErr:       persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrUserNotFound,
		},
		{
			name:          "repo error",
			ctx:           ctx,
			userID:        3,
			mockErr:       errors.New("db error"),
			expectedError: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.userRepo.EXPECT().
				GetByID(gomock.Any(), tt.userID).
				Return(tt.mockUser, tt.mockErr).
				Times(1)

			result, err := test.service.GetByID(tt.ctx, tt.userID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockUser, result)
			}
		})
	}
}

func TestUserService_GetByAuthID(t *testing.T) {
	test := setUpUserServiceTest(t)
	defer test.ctrl.Finish()

	ctx := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))

	user := &entity.User{
		ID:         1,
		AuthID:     1,
		Name:       "John Doe",
		IsVerified: true,
	}

	tests := []struct {
		name          string
		ctx           context.Context
		mockUser      *entity.User
		mockErr       error
		expectedError error
	}{
		{
			name:     "successful get user by auth id",
			ctx:      ctx,
			mockUser: user,
		},
		{
			name:          "user not found",
			ctx:           ctx,
			mockErr:       persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrUserNotFound,
		},
		{
			name:          "repo error",
			ctx:           ctx,
			mockErr:       errors.New("db error"),
			expectedError: errors.New("db error"),
		},
		{
			name:          "missing auth id in context",
			ctx:           context.Background(),
			expectedError: service_errors.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.AuthIDKey) != nil {
				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(tt.mockUser, tt.mockErr).
					Times(1)
			}

			result, err := test.service.GetByAuthID(tt.ctx)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockUser, result)
			}
		})
	}
}

func TestUserService_Update(t *testing.T) {
	test := setUpUserServiceTest(t)
	defer test.ctrl.Finish()

	ctx := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))

	existingUser := &entity.User{
		ID:         1,
		AuthID:     1,
		Name:       "Old Name",
		IsVerified: true,
	}

	updatedUser := &entity.User{
		ID:         1,
		AuthID:     1,
		Name:       "New Name",
		IsVerified: true,
	}

	tests := []struct {
		name           string
		ctx            context.Context
		nameParam      string
		mockGetUser    *entity.User
		mockGetErr     error
		mockUpdateUser *entity.User
		mockUpdateErr  error
		expectedError  error
	}{
		{
			name:           "successful update",
			ctx:            ctx,
			nameParam:      "New Name",
			mockGetUser:    existingUser,
			mockUpdateUser: updatedUser,
		},
		{
			name:          "user not found by auth id",
			ctx:           ctx,
			nameParam:     "New Name",
			mockGetErr:    persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrUserNotFound,
		},
		{
			name:          "get repo error",
			ctx:           ctx,
			nameParam:     "New Name",
			mockGetErr:    errors.New("db error"),
			expectedError: errors.New("db error"),
		},
		{
			name:          "update repo error",
			ctx:           ctx,
			nameParam:     "New Name",
			mockGetUser:   existingUser,
			mockUpdateErr: errors.New("update failed"),
			expectedError: errors.New("update failed"),
		},
		{
			name:          "missing auth id in context",
			ctx:           context.Background(),
			nameParam:     "New Name",
			expectedError: service_errors.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ctx.Value(service_const.AuthIDKey) != nil {
				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(tt.mockGetUser, tt.mockGetErr).
					Times(1)

				if tt.mockGetErr == nil && tt.mockGetUser != nil {
					test.userRepo.EXPECT().
						Update(gomock.Any(), gomock.Any()).
						Return(tt.mockUpdateUser, tt.mockUpdateErr).
						Times(1)
				}
			}

			result, err := test.service.Update(tt.ctx, tt.nameParam)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.nameParam, result.Name)
			}
		})
	}
}
