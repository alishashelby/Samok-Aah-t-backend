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

type modelServiceServiceTest struct {
	ctrl             *gomock.Controller
	modelServiceRepo *mocks.MockModelServiceRepository
	userRepo         *mocks.MockUserRepository
	txManager        *mocks.MockTxManager
	service          *DefaultModelServiceService
}

func setUpModelServiceServiceTest(t *testing.T) *modelServiceServiceTest {
	t.Helper()

	ctrl := gomock.NewController(t)
	modelServiceRepo := mocks.NewMockModelServiceRepository(ctrl)
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

	modelServiceService := NewDefaultModelServiceService(
		modelServiceRepo, userRepo, mockTxManager, log,
	)

	return &modelServiceServiceTest{
		ctrl:             ctrl,
		modelServiceRepo: modelServiceRepo,
		userRepo:         userRepo,
		txManager:        mockTxManager,
		service:          modelServiceService,
	}
}

func TestModelServiceService_GetServiceByID(t *testing.T) {
	test := setUpModelServiceServiceTest(t)
	defer test.ctrl.Finish()

	ctxClient := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxClient = context.WithValue(ctxClient, service_const.RoleKey, "CLIENT")

	verifiedClient := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	modelService := &entity.ModelService{
		ID:          1,
		ModelID:     2,
		Title:       "Test Service",
		Description: "Test Description",
		Price:       100.0,
		IsActive:    true,
	}

	tests := []struct {
		name           string
		ctx            context.Context
		serviceID      int64
		mockClient     *entity.User
		mockClientErr  error
		mockService    *entity.ModelService
		mockServiceErr error
		expectedError  error
	}{
		{
			name:        "successful get service",
			ctx:         ctxClient,
			serviceID:   1,
			mockClient:  verifiedClient,
			mockService: modelService,
		},
		{
			name:          "client not verified",
			ctx:           ctxClient,
			serviceID:     1,
			mockClient:    &entity.User{ID: 1, IsVerified: false},
			expectedError: service_errors.ErrNotVerifiedClient,
		},
		{
			name:           "service not found",
			ctx:            ctxClient,
			serviceID:      1,
			mockClient:     verifiedClient,
			mockServiceErr: persistence.ErrNoRowsFound,
			expectedError:  service_errors.ErrServiceIsNotFound,
		},
		{
			name:           "repo get error",
			ctx:            ctxClient,
			serviceID:      1,
			mockClient:     verifiedClient,
			mockServiceErr: errors.New("db error"),
			expectedError:  errors.New("db error"),
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
				test.modelServiceRepo.EXPECT().
					GetByID(gomock.Any(), tt.serviceID, false).
					Return(tt.mockService, tt.mockServiceErr).
					Times(1)
			}

			service, err := test.service.GetServiceByID(tt.ctx, tt.serviceID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.Equal(t, tt.mockService, service)
			}
		})
	}
}

func TestModelServiceService_DeactivateService(t *testing.T) {
	test := setUpModelServiceServiceTest(t)
	defer test.ctrl.Finish()

	ctxModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxModel = context.WithValue(ctxModel, service_const.RoleKey, "MODEL")

	verifiedModel := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	activeService := &entity.ModelService{
		ID:       1,
		ModelID:  1,
		IsActive: true,
	}

	inactiveService := &entity.ModelService{
		ID:       1,
		ModelID:  1,
		IsActive: false,
	}

	tests := []struct {
		name              string
		ctx               context.Context
		serviceID         int64
		mockModel         *entity.User
		mockModelErr      error
		mockService       *entity.ModelService
		mockServiceErr    error
		mockDeactivateErr error
		expectedError     error
	}{
		{
			name:        "successful deactivation",
			ctx:         ctxModel,
			serviceID:   1,
			mockModel:   verifiedModel,
			mockService: activeService,
		},
		{
			name:           "service not found",
			ctx:            ctxModel,
			serviceID:      1,
			mockModel:      verifiedModel,
			mockServiceErr: persistence.ErrNoRowsFound,
			expectedError:  service_errors.ErrServiceIsNotFound,
		},
		{
			name:          "model is not owner",
			ctx:           ctxModel,
			serviceID:     1,
			mockModel:     verifiedModel,
			mockService:   &entity.ModelService{ID: 1, ModelID: 999, IsActive: true},
			expectedError: service_errors.ErrModelIsNotAnOwnerOfService,
		},
		{
			name:          "service already inactive",
			ctx:           ctxModel,
			serviceID:     1,
			mockModel:     verifiedModel,
			mockService:   inactiveService,
			expectedError: service_errors.ErrServiceIsNotActive,
		},
		{
			name:              "deactivate repo error",
			ctx:               ctxModel,
			serviceID:         1,
			mockModel:         verifiedModel,
			mockService:       activeService,
			mockDeactivateErr: errors.New("deactivate failed"),
			expectedError:     errors.New("deactivate failed"),
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
				test.modelServiceRepo.EXPECT().
					GetByID(gomock.Any(), tt.serviceID, true).
					Return(tt.mockService, tt.mockServiceErr).
					Times(1)

				if tt.mockServiceErr == nil && tt.mockService != nil {
					if tt.mockService.ModelID == tt.mockModel.ID && tt.mockService.IsActive {
						test.modelServiceRepo.EXPECT().
							Deactivate(gomock.Any(), tt.serviceID).
							Return(tt.mockDeactivateErr).
							Times(1)
					}
				}
			}

			err := test.service.DeactivateService(tt.ctx, tt.serviceID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModelServiceService_GetAllServices(t *testing.T) {
	test := setUpModelServiceServiceTest(t)
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

	services := []*entity.ModelService{
		{
			ID:          1,
			ModelID:     2,
			Title:       "Service 1",
			Description: "Description 1",
			Price:       100.0,
			IsActive:    true,
		},
		{
			ID:          2,
			ModelID:     3,
			Title:       "Service 2",
			Description: "Description 2",
			Price:       200.0,
			IsActive:    true,
		},
		{
			ID:          3,
			ModelID:     2,
			Title:       "Service 3",
			Description: "Description 3",
			Price:       150.0,
			IsActive:    false,
		},
	}

	tests := []struct {
		name          string
		ctx           context.Context
		page          *int64
		limit         *int64
		mockClient    *entity.User
		mockClientErr error
		mockServices  []*entity.ModelService
		mockErr       error
		expectedError error
		expectCall    bool
	}{
		{
			name:          "client not verified",
			ctx:           ctxClient,
			page:          nil,
			limit:         nil,
			mockClient:    &entity.User{ID: 1, IsVerified: false},
			expectedError: service_errors.ErrNotVerifiedClient,
			expectCall:    false,
		},
		{
			name:          "not a client role",
			ctx:           ctxNotClient,
			page:          nil,
			limit:         nil,
			expectedError: service_errors.ErrNotClient,
			expectCall:    false,
		},
		{
			name:          "repository error",
			ctx:           ctxClient,
			page:          nil,
			limit:         nil,
			mockClient:    verifiedClient,
			mockErr:       errors.New("database error"),
			expectedError: errors.New("database error"),
			expectCall:    true,
		},
		{
			name:         "successful get all services with empty result",
			ctx:          ctxClient,
			page:         nil,
			limit:        nil,
			mockClient:   verifiedClient,
			mockServices: []*entity.ModelService{},
			expectCall:   true,
		},
		{
			name:         "successful get all services with only active services",
			ctx:          ctxClient,
			page:         int64Ptr(2),
			limit:        int64Ptr(5),
			mockClient:   verifiedClient,
			mockServices: []*entity.ModelService{services[0], services[1]},
			expectCall:   true,
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

			if tt.mockClientErr == nil && tt.mockClient != nil && tt.mockClient.IsVerified && tt.expectCall {
				test.modelServiceRepo.EXPECT().
					GetAll(gomock.Any(), gomock.Any(), false).
					Return(tt.mockServices, tt.mockErr).
					Times(1)
			}

			result, err := test.service.GetAllServices(tt.ctx, tt.page, tt.limit)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockServices, result)

				if tt.mockServices != nil {
					for _, service := range result {
						assert.True(t, service.IsActive)
					}
				}
			}
		})
	}
}

func TestModelServiceService_GetAllServicesByModelID(t *testing.T) {
	test := setUpModelServiceServiceTest(t)
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

	services := []*entity.ModelService{
		{
			ID:          1,
			ModelID:     1,
			Title:       "Service 1",
			Description: "Description 1",
			Price:       100.0,
			IsActive:    true,
		},
		{
			ID:          2,
			ModelID:     1,
			Title:       "Service 2",
			Description: "Description 2",
			Price:       200.0,
			IsActive:    false,
		},
		{
			ID:          3,
			ModelID:     1,
			Title:       "Service 3",
			Description: "Description 3",
			Price:       150.0,
			IsActive:    true,
		},
	}

	tests := []struct {
		name          string
		ctx           context.Context
		page          *int64
		limit         *int64
		mockModel     *entity.User
		mockModelErr  error
		mockServices  []*entity.ModelService
		mockErr       error
		expectedError error
		expectCall    bool
	}{
		{
			name:         "successful get all services by model id without pagination",
			ctx:          ctxModel,
			page:         nil,
			limit:        nil,
			mockModel:    verifiedModel,
			mockServices: services,
			expectCall:   true,
		},
		{
			name:         "successful get all services by model id with pagination",
			ctx:          ctxModel,
			page:         int64Ptr(1),
			limit:        int64Ptr(10),
			mockModel:    verifiedModel,
			mockServices: services,
			expectCall:   true,
		},
		{
			name:          "model not verified",
			ctx:           ctxModel,
			page:          nil,
			limit:         nil,
			mockModel:     &entity.User{ID: 1, IsVerified: false},
			expectedError: service_errors.ErrNotVerifiedModel,
			expectCall:    false,
		},
		{
			name:          "not a model role",
			ctx:           ctxNotModel,
			page:          nil,
			limit:         nil,
			expectedError: service_errors.ErrNotAModel,
			expectCall:    false,
		},
		{
			name:          "repository error",
			ctx:           ctxModel,
			page:          nil,
			limit:         nil,
			mockModel:     verifiedModel,
			mockErr:       errors.New("database error"),
			expectedError: errors.New("database error"),
			expectCall:    true,
		},
		{
			name:         "successful get all services by model id with empty result",
			ctx:          ctxModel,
			page:         nil,
			limit:        nil,
			mockModel:    verifiedModel,
			mockServices: []*entity.ModelService{},
			expectCall:   true,
		},
		{
			name:         "successful get all services by model id with mixed active/inactive",
			ctx:          ctxModel,
			page:         int64Ptr(2),
			limit:        int64Ptr(5),
			mockModel:    verifiedModel,
			mockServices: services,
			expectCall:   true,
		},
		{
			name:          "repository get user error",
			ctx:           ctxModel,
			page:          nil,
			limit:         nil,
			mockModelErr:  errors.New("user not found"),
			expectedError: errors.New("user not found"),
			expectCall:    false,
		},
		{
			name:          "repository get user not found",
			ctx:           ctxModel,
			page:          nil,
			limit:         nil,
			mockModelErr:  persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrNotAModel,
			expectCall:    false,
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

			if tt.mockModelErr == nil && tt.mockModel != nil && tt.mockModel.IsVerified && tt.expectCall {
				test.modelServiceRepo.EXPECT().
					GetByModelID(gomock.Any(), tt.mockModel.ID, gomock.Any(), true).
					Return(tt.mockServices, tt.mockErr).
					Times(1)
			}

			result, err := test.service.GetAllServicesByModelID(tt.ctx, tt.page, tt.limit)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockServices, result)

				for _, service := range result {
					assert.Equal(t, tt.mockModel.ID, service.ModelID)
				}
			}
		})
	}
}

func TestModelServiceService_GetAllMethods_EdgeCases(t *testing.T) {
	test := setUpModelServiceServiceTest(t)
	defer test.ctrl.Finish()

	ctxClient := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxClient = context.WithValue(ctxClient, service_const.RoleKey, "CLIENT")

	ctxModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxModel = context.WithValue(ctxModel, service_const.RoleKey, "MODEL")

	verifiedClient := &entity.User{ID: 1, AuthID: int64(1), IsVerified: true}
	verifiedModel := &entity.User{ID: 1, AuthID: int64(1), IsVerified: true}

	tests := []struct {
		name          string
		ctx           context.Context
		method        string
		page          *int64
		limit         *int64
		expectedCount int
	}{
		{
			name:          "GetAllServices with negative pagination",
			ctx:           ctxClient,
			method:        "GetAllServices",
			page:          int64Ptr(-1),
			limit:         int64Ptr(-10),
			expectedCount: 3,
		},
		{
			name:          "GetAllServices with zero pagination",
			ctx:           ctxClient,
			method:        "GetAllServices",
			page:          int64Ptr(0),
			limit:         int64Ptr(0),
			expectedCount: 3,
		},
		{
			name:          "GetAllServices with large pagination",
			ctx:           ctxClient,
			method:        "GetAllServices",
			page:          int64Ptr(1000),
			limit:         int64Ptr(1000),
			expectedCount: 3,
		},
		{
			name:          "GetAllServicesByModelID with nil pagination",
			ctx:           ctxModel,
			method:        "GetAllServicesByModelID",
			page:          nil,
			limit:         nil,
			expectedCount: 2,
		},
		{
			name:          "GetAllServicesByModelID with page only",
			ctx:           ctxModel,
			method:        "GetAllServicesByModelID",
			page:          int64Ptr(1),
			limit:         nil,
			expectedCount: 2,
		},
		{
			name:          "GetAllServicesByModelID with limit only",
			ctx:           ctxModel,
			method:        "GetAllServicesByModelID",
			page:          nil,
			limit:         int64Ptr(5),
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := []*entity.ModelService{
				{ID: 1, ModelID: 1, Title: "Service 1", IsActive: true},
				{ID: 2, ModelID: 1, Title: "Service 2", IsActive: true},
			}

			if tt.method == "GetAllServices" && tt.ctx.Value(service_const.RoleKey) == "CLIENT" {
				services = []*entity.ModelService{
					{ID: 1, ModelID: 1, Title: "Service 1", IsActive: true},
					{ID: 2, ModelID: 2, Title: "Service 2", IsActive: true},
					{ID: 3, ModelID: 3, Title: "Service 3", IsActive: true},
				}

				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(verifiedClient, nil).
					Times(1)

				test.modelServiceRepo.EXPECT().
					GetAll(gomock.Any(), gomock.Any(), false).
					Return(services, nil).
					Times(1)
			} else if tt.method == "GetAllServicesByModelID" && tt.ctx.Value(service_const.RoleKey) == "MODEL" {
				authID := tt.ctx.Value(service_const.AuthIDKey).(int64)
				test.userRepo.EXPECT().
					GetByAuthID(gomock.Any(), authID).
					Return(verifiedModel, nil).
					Times(1)

				test.modelServiceRepo.EXPECT().
					GetByModelID(gomock.Any(), verifiedModel.ID, gomock.Any(), true).
					Return(services, nil).
					Times(1)
			}

			var result []*entity.ModelService
			var err error

			switch tt.method {
			case "GetAllServices":
				result, err = test.service.GetAllServices(tt.ctx, tt.page, tt.limit)
			case "GetAllServicesByModelID":
				result, err = test.service.GetAllServicesByModelID(tt.ctx, tt.page, tt.limit)
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestModelServiceService_GetAllServices_Filtering(t *testing.T) {
	test := setUpModelServiceServiceTest(t)
	defer test.ctrl.Finish()

	ctxClient := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxClient = context.WithValue(ctxClient, service_const.RoleKey, "CLIENT")

	verifiedClient := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	activeServices := []*entity.ModelService{
		{ID: 1, ModelID: 1, Title: "Active Service 1", IsActive: true},
		{ID: 2, ModelID: 2, Title: "Active Service 2", IsActive: true},
		{ID: 4, ModelID: 3, Title: "Active Service 3", IsActive: true},
	}

	t.Run("GetAllServices returns only active services", func(t *testing.T) {
		authID := ctxClient.Value(service_const.AuthIDKey).(int64)
		test.userRepo.EXPECT().
			GetByAuthID(gomock.Any(), authID).
			Return(verifiedClient, nil).
			Times(1)

		test.modelServiceRepo.EXPECT().
			GetAll(gomock.Any(), gomock.Any(), false).
			Return(activeServices, nil).
			Times(1)

		result, err := test.service.GetAllServices(ctxClient, nil, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 3)

		for _, service := range result {
			assert.True(t, service.IsActive)
		}
	})

	t.Run("GetAllServicesByModelID returns all services including inactive", func(t *testing.T) {
		ctxModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
		ctxModel = context.WithValue(ctxModel, service_const.RoleKey, "MODEL")

		verifiedModel := &entity.User{
			ID:         1,
			AuthID:     int64(1),
			IsVerified: true,
		}

		modelServices := []*entity.ModelService{
			{ID: 1, ModelID: 1, Title: "Active Service 1", IsActive: true},
			{ID: 3, ModelID: 1, Title: "Inactive Service 1", IsActive: false},
		}

		authID := ctxModel.Value(service_const.AuthIDKey).(int64)
		test.userRepo.EXPECT().
			GetByAuthID(gomock.Any(), authID).
			Return(verifiedModel, nil).
			Times(1)

		test.modelServiceRepo.EXPECT().
			GetByModelID(gomock.Any(), verifiedModel.ID, gomock.Any(), true).
			Return(modelServices, nil).
			Times(1)

		result, err := test.service.GetAllServicesByModelID(ctxModel, nil, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)

		hasActive := false
		hasInactive := false
		for _, service := range result {
			assert.Equal(t, verifiedModel.ID, service.ModelID)
			if service.IsActive {
				hasActive = true
			} else {
				hasInactive = true
			}
		}
		assert.True(t, hasActive)
		assert.True(t, hasInactive)
	})
}

func TestModelServiceService_CheckPayloadRestrictions(t *testing.T) {
	test := setUpModelServiceServiceTest(t)
	defer test.ctrl.Finish()

	tests := []struct {
		name          string
		price         float32
		description   string
		expectedError error
	}{
		{
			name:          "valid payload",
			price:         100.0,
			description:   "Valid description",
			expectedError: nil,
		},
		{
			name:          "zero price",
			price:         0,
			description:   "Valid description",
			expectedError: service_errors.ErrInvalidPrice,
		},
		{
			name:          "negative price",
			price:         -50.0,
			description:   "Valid description",
			expectedError: service_errors.ErrInvalidPrice,
		},
		{
			name:          "price equals zero",
			price:         0.0,
			description:   "Valid description",
			expectedError: service_errors.ErrInvalidPrice,
		},
		{
			name:          "description too long",
			price:         100.0,
			description:   string(make([]byte, 1001)),
			expectedError: service_errors.ErrDescriptionTooLong,
		},
		{
			name:          "description exactly 1000 characters",
			price:         100.0,
			description:   string(make([]byte, 1000)),
			expectedError: nil,
		},
		{
			name:          "empty description",
			price:         100.0,
			description:   "",
			expectedError: nil,
		},
		{
			name:          "very small positive price",
			price:         0.01,
			description:   "Valid description",
			expectedError: nil,
		},
		{
			name:          "both price and description invalid",
			price:         0,
			description:   string(make([]byte, 1001)),
			expectedError: service_errors.ErrInvalidPrice,
		},
		{
			name:          "description with special characters",
			price:         150.0,
			description:   "Description with спецсимволы: 测试 テスト тест 🚀",
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := test.service.checkPayloadRestrictions(tt.price, tt.description)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestModelServiceService_CreateService_PayloadValidationIntegration(t *testing.T) {
	test := setUpModelServiceServiceTest(t)
	defer test.ctrl.Finish()

	ctxModel := context.WithValue(context.Background(), service_const.AuthIDKey, int64(1))
	ctxModel = context.WithValue(ctxModel, service_const.RoleKey, "MODEL")

	verifiedModel := &entity.User{
		ID:         1,
		AuthID:     int64(1),
		IsVerified: true,
	}

	tests := []struct {
		name          string
		title         string
		description   string
		price         float32
		mockSaveErr   error
		expectedError error
		expectSave    bool
	}{
		{
			name:          "successful creation with valid payload",
			title:         "Valid Service",
			description:   "Valid description",
			price:         100.0,
			expectSave:    true,
			expectedError: nil,
		},
		{
			name:          "creation with zero price should fail before saving",
			title:         "Service with zero price",
			description:   "Valid description",
			price:         0,
			expectSave:    false,
			expectedError: service_errors.ErrInvalidPrice,
		},
		{
			name:          "creation with too long description should fail before saving",
			title:         "Service with long description",
			description:   string(make([]byte, 1001)),
			price:         100.0,
			expectSave:    false,
			expectedError: service_errors.ErrDescriptionTooLong,
		},
		{
			name:          "creation with negative price should fail before saving",
			title:         "Service with negative price",
			description:   "Valid description",
			price:         -10.0,
			expectSave:    false,
			expectedError: service_errors.ErrInvalidPrice,
		},
		{
			name:          "creation with valid payload but repo save fails",
			title:         "Valid Service",
			description:   "Valid description",
			price:         100.0,
			mockSaveErr:   errors.New("database error"),
			expectSave:    true,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authID := ctxModel.Value(service_const.AuthIDKey).(int64)
			test.userRepo.EXPECT().
				GetByAuthID(gomock.Any(), authID).
				Return(verifiedModel, nil).
				Times(1)

			if tt.expectSave {
				test.modelServiceRepo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Return(tt.mockSaveErr).
					Times(1)
			}

			service, err := test.service.CreateService(ctxModel, tt.title, tt.description, tt.price)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, service)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
				assert.Equal(t, tt.title, service.Title)
				assert.Equal(t, tt.description, service.Description)
				assert.Equal(t, tt.price, service.Price)
			}
		})
	}
}
