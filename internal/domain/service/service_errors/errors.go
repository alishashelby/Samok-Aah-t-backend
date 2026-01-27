package service_errors

import "errors"

var (
	ErrUnauthorized = errors.New("invalid or missing auth ID")
	ErrNoRole       = errors.New("role is not fount in ctx")
)

var (
	ErrLoadingSecret  = errors.New("error loading JWT_SECRET environment variable")
	ErrLoadingTTL     = errors.New("error loading JWT_TTL environment variable")
	ErrParsingTTL     = errors.New("error parsing JWT_TTL environment variable")
	ErrNotPositiveTTL = errors.New("JWT_TTL environment variable should be positive")
)

var (
	ErrEmailExists                = errors.New("this email already was used")
	ErrAuthWithEmailDoesNotExists = errors.New("an account with this email does not exist")
	ErrIDDoesNotExists            = errors.New("this id does not exist")
	ErrInvalidPasswordOrEmail     = errors.New("invalid password or email")
)

var (
	ErrUserNotFound = errors.New("user is not found")
	ErrIsNotAnAdult = errors.New("age is under 18")
)

var (
	ErrAdminNotFound   = errors.New("admin does not exist")
	ErrBookingNotFound = errors.New("booking does not exist")
	ErrOrderNotFound   = errors.New("order does not exist")
)

var (
	ErrNotVerifiedModel   = errors.New("only verified model can do this action")
	ErrNotVerifiedClient  = errors.New("only verified client can do this action")
	ErrInvalidPrice       = errors.New("price should be greater than zero")
	ErrDescriptionTooLong = errors.New("description too long, must be not greater than 1000 characters")
	ErrServiceIsNotActive = errors.New("service is not active")
)

var (
	ErrNotAModel                  = errors.New("user is not a model")
	ErrModelIsNotAnOwnerOfService = errors.New("model is not an owner of service")
	ErrServiceIsNotFound          = errors.New("service is not found")
)

var (
	ErrIncorrectSlotTime           = errors.New("start time must be before end time")
	ErrSlotOverlap                 = errors.New("slot overlaps with existing slot")
	ErrModelIsNotAnOwnerOfSlot     = errors.New("model is not an owner of slot")
	ErrInvalidSlotStatusTransition = errors.New("invalid slot status transition")
)

var (
	ErrSlotNotAvailable          = errors.New("slot is not available")
	ErrBookingAlreadyProcessed   = errors.New("booking already processed")
	ErrClientIsNotOwnerOfBooking = errors.New("client is not owner of booking")
	ErrInvalidBookingState       = errors.New("invalid booking state")
	ErrBookingExpired            = errors.New("booking ttl expired")
	ErrSlotIsNotFound            = errors.New("slot is not found")
)

var (
	ErrCannotCancelOrderNow    = errors.New("cannot cancel order less than 24h before slot start")
	ErrCannotCompleteOrder     = errors.New("cannot complete order: either wrong status or slot not finished")
	ErrClientIsNotOwnerOfOrder = errors.New("client is not owner of this order")
)

var (
	ErrNotAdmin  = errors.New("this is not an admin")
	ErrNotClient = errors.New("this is not a client")
)
