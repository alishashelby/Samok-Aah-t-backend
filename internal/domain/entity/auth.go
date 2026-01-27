package entity

import (
	"time"
)

type Role string

const (
	RoleClient Role = "CLIENT"
	RoleModel  Role = "MODEL"
	RoleAdmin  Role = "ADMIN"
)

func (r Role) String() string {
	return string(r)
}

type Auth struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
}

func NewAuth(email, passwordHash string, role Role) *Auth {
	return &Auth{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
	}
}
