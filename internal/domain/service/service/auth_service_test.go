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
	"golang.org/x/crypto/bcrypt"
)

type authServiceTest struct {
	ctrl       *gomock.Controller
	authRepo   *mocks.MockAuthRepository
	jwtService *JWTService
	txManager  *mocks.MockTxManager
	service    *DefaultAuthService
}

func setUpAuthServiceTest(t *testing.T) *authServiceTest {
	t.Helper()

	t.Setenv(service_const.DotEnvJWTExpiration, "21600")
	t.Setenv(service_const.DotEnvJWTSecret, "jwt-secret")

	ctrl := gomock.NewController(t)
	authRepo := mocks.NewMockAuthRepository(ctrl)
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

	JWTService, err := NewJWTService()
	if err != nil {
		t.Log("Failed to create JWT service")
	}
	authService := NewDefaultAuthService(authRepo, JWTService, mockTxManager, log)

	return &authServiceTest{
		ctrl:       ctrl,
		authRepo:   authRepo,
		jwtService: JWTService,
		txManager:  mockTxManager,
		service:    authService,
	}
}

func TestAuthService_Register(t *testing.T) {
	test := setUpAuthServiceTest(t)
	defer test.ctrl.Finish()

	tests := []struct {
		name          string
		email         string
		password      string
		role          string
		mockSaveErr   error
		expectedError error
		expectToken   bool
	}{
		{
			name:        "successful registration",
			email:       "test@example.com",
			password:    "password123",
			role:        "CLIENT",
			expectToken: true,
		},
		{
			name:          "duplicate email",
			email:         "duplicate@example.com",
			password:      "password123",
			role:          "CLIENT",
			mockSaveErr:   persistence.ErrDuplicateKey,
			expectedError: service_errors.ErrEmailExists,
			expectToken:   false,
		},
		{
			name:          "repo save error",
			email:         "error@example.com",
			password:      "password123",
			role:          "CLIENT",
			mockSaveErr:   errors.New("db error"),
			expectedError: errors.New("db error"),
			expectToken:   false,
		},
		{
			name:        "invalid role - should still save",
			email:       "test@example.com",
			password:    "password123",
			role:        "INVALID_ROLE",
			expectToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.authRepo.EXPECT().
				Save(gomock.Any(), gomock.Any()).
				Return(tt.mockSaveErr).
				Times(1)

			token, err := test.service.Register(context.Background(), tt.email, tt.password, tt.role)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				if tt.expectToken {
					assert.NotNil(t, token)
					assert.NotEmpty(t, *token)
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	test := setUpAuthServiceTest(t)
	defer test.ctrl.Finish()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

	auth := &entity.Auth{
		ID:           1,
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		Role:         entity.RoleClient,
	}

	tests := []struct {
		name          string
		email         string
		password      string
		mockAuth      *entity.Auth
		mockErr       error
		expectedError error
	}{
		{
			name:     "successful login",
			email:    "test@example.com",
			password: "correctpassword",
			mockAuth: auth,
		},
		{
			name:          "auth not found",
			email:         "notfound@example.com",
			password:      "password123",
			mockErr:       persistence.ErrNoRowsFound,
			expectedError: service_errors.ErrAuthWithEmailDoesNotExists,
		},
		{
			name:          "repo error",
			email:         "error@example.com",
			password:      "password123",
			mockErr:       errors.New("db error"),
			expectedError: errors.New("db error"),
		},
		{
			name:          "wrong password",
			email:         "test@example.com",
			password:      "wrongpassword",
			mockAuth:      auth,
			expectedError: service_errors.ErrInvalidPasswordOrEmail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test.authRepo.EXPECT().
				GetByEmail(gomock.Any(), tt.email).
				Return(tt.mockAuth, tt.mockErr).
				Times(1)

			token, err := test.service.Login(context.Background(), tt.email, tt.password)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, token)
				assert.NotEmpty(t, *token)
			}
		})
	}
}
