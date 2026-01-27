package service

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_const"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTService() (*JWTService, error) {
	secret := []byte(os.Getenv(service_const.DotEnvJWTSecret))
	if secret == nil {
		return nil, service_errors.ErrLoadingSecret
	}

	ttl := os.Getenv(service_const.DotEnvJWTExpiration)
	if ttl == "" {
		return nil, service_errors.ErrLoadingTTL
	}

	ttlInSeconds, err := strconv.Atoi(ttl)
	if err != nil {
		return nil, service_errors.ErrParsingTTL
	}
	if ttlInSeconds < 0 {
		return nil, service_errors.ErrNotPositiveTTL
	}

	return &JWTService{
		secret: secret,
		ttl:    time.Duration(ttlInSeconds) * time.Second,
	}, nil
}

func (s *JWTService) GenerateToken(auth *entity.Auth) (*string, error) {
	claims := jwt.MapClaims{
		string(service_const.AuthIDKey):   auth.ID,
		string(service_const.RoleKey):     auth.Role,
		string(service_const.IssuedAtKey): time.Now().Unix(),
		string(service_const.ExpiryKey):   time.Now().Add(s.ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(s.secret)
	if err != nil {
		return nil, err
	}

	return &signedToken, nil
}

func (s *JWTService) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	payload, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return payload, nil
}
