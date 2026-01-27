package service_const

type ctxKey string

const (
	AuthIDKey   ctxKey = "auth_id"
	RoleKey     ctxKey = "role"
	IssuedAtKey ctxKey = "iat"
	ExpiryKey   ctxKey = "exp"

	DotEnvJWTSecret     = "JWT_SECRET"
	DotEnvJWTExpiration = "JWT_TTL"
)

const (
	DotEnvBookingExpiration = "BOOKING_TTL"
)
