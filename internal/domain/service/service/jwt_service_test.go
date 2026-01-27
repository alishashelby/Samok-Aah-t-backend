package service

import (
	"os"
	"testing"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_const"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_NewJWTService(t *testing.T) {
	tests := []struct {
		name          string
		secretEnv     string
		ttlEnv        string
		setupEnv      func()
		cleanupEnv    func()
		expectedError error
	}{
		{
			name:      "successful creation",
			secretEnv: "test-secret-key-very-long-for-testing-purposes-123",
			ttlEnv:    "3600",
			setupEnv: func() {
				os.Setenv(service_const.DotEnvJWTSecret, "test-secret-key-very-long-for-testing-purposes-123")
				os.Setenv(service_const.DotEnvJWTExpiration, "3600")
			},
			cleanupEnv: func() {
				os.Unsetenv(service_const.DotEnvJWTSecret)
				os.Unsetenv(service_const.DotEnvJWTExpiration)
			},
		},
		{
			name:      "missing ttl",
			secretEnv: "test-secret",
			ttlEnv:    "",
			setupEnv: func() {
				os.Setenv(service_const.DotEnvJWTSecret, "test-secret")
				os.Unsetenv(service_const.DotEnvJWTExpiration)
			},
			cleanupEnv: func() {
				os.Unsetenv(service_const.DotEnvJWTSecret)
			},
			expectedError: service_errors.ErrLoadingTTL,
		},
		{
			name:      "invalid ttl format",
			secretEnv: "test-secret",
			ttlEnv:    "not-a-number",
			setupEnv: func() {
				os.Setenv(service_const.DotEnvJWTSecret, "test-secret")
				os.Setenv(service_const.DotEnvJWTExpiration, "not-a-number")
			},
			cleanupEnv: func() {
				os.Unsetenv(service_const.DotEnvJWTSecret)
				os.Unsetenv(service_const.DotEnvJWTExpiration)
			},
			expectedError: service_errors.ErrParsingTTL,
		},
		{
			name:      "negative ttl",
			secretEnv: "test-secret",
			ttlEnv:    "-1",
			setupEnv: func() {
				os.Setenv(service_const.DotEnvJWTSecret, "test-secret")
				os.Setenv(service_const.DotEnvJWTExpiration, "-1")
			},
			cleanupEnv: func() {
				os.Unsetenv(service_const.DotEnvJWTSecret)
				os.Unsetenv(service_const.DotEnvJWTExpiration)
			},
			expectedError: service_errors.ErrNotPositiveTTL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			jwtService, err := NewJWTService()

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, jwtService)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, jwtService)
				assert.NotNil(t, jwtService.secret)
				assert.Greater(t, jwtService.ttl, time.Duration(0))
			}
		})
	}
}

func TestJWTService_GenerateAndParseToken(t *testing.T) {
	os.Setenv(service_const.DotEnvJWTSecret, "aeaebf84c8216bb5e45e2b18c78d8c26f9988ce31b26ecc7ee760f9e45c377cb")
	os.Setenv(service_const.DotEnvJWTExpiration, "3600")
	defer func() {
		os.Unsetenv(service_const.DotEnvJWTSecret)
		os.Unsetenv(service_const.DotEnvJWTExpiration)
	}()

	jwtService, err := NewJWTService()
	require.NoError(t, err)
	require.NotNil(t, jwtService)

	auth := &entity.Auth{
		ID:   1,
		Role: entity.RoleClient,
	}

	t.Run("successful token generation and parsing", func(t *testing.T) {
		token, err := jwtService.GenerateToken(auth)
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.NotEmpty(t, *token)

		claims, err := jwtService.ParseToken(*token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)

		authID, ok := claims[string(service_const.AuthIDKey)].(float64)
		assert.True(t, ok)
		assert.Equal(t, float64(auth.ID), authID)

		role, ok := claims[string(service_const.RoleKey)].(string)
		assert.True(t, ok)
		assert.Equal(t, auth.Role.String(), role)
	})

	t.Run("invalid token parsing", func(t *testing.T) {
		invalidToken := "invalid.token.string"

		claims, err := jwtService.ParseToken(invalidToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("tampered token", func(t *testing.T) {
		token, err := jwtService.GenerateToken(auth)
		assert.NoError(t, err)

		tamperedToken := *token + "tampered"

		claims, err := jwtService.ParseToken(tamperedToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("expired token", func(t *testing.T) {
		os.Setenv(service_const.DotEnvJWTExpiration, "1")
		shortJWTService, err := NewJWTService()
		assert.NoError(t, err)

		token, err := shortJWTService.GenerateToken(auth)
		assert.NoError(t, err)

		time.Sleep(2 * time.Second)

		claims, err := shortJWTService.ParseToken(*token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("different signing method", func(t *testing.T) {
		wrongMethodToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

		claims, err := jwtService.ParseToken(wrongMethodToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}
