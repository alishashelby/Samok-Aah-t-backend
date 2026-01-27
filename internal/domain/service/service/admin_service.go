package service

import (
	"context"
	"errors"
	"time"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/entity"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/common"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/interfaces"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/domain/service/service_errors"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/database"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/infrastructure/persistence"
	pkg "github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/pkg/logger/option"
)

type DefaultAdminService struct {
	adminRepo   interfaces.AdminRepository
	userRepo    interfaces.UserRepository
	bookingRepo interfaces.BookingRepository
	orderRepo   interfaces.OrderRepository
	txManager   database.TxManager
	logger      pkg.Logger
}

func NewDefaultAdminService(adminRepo interfaces.AdminRepository, userRepo interfaces.UserRepository,
	bookingRepo interfaces.BookingRepository, orderRepo interfaces.OrderRepository,
	txManager database.TxManager, logger pkg.Logger) *DefaultAdminService {
	return &DefaultAdminService{
		adminRepo:   adminRepo,
		userRepo:    userRepo,
		bookingRepo: bookingRepo,
		orderRepo:   orderRepo,
		txManager:   txManager,
		logger:      logger,
	}
}

func (d *DefaultAdminService) Create(ctx context.Context,
	newAdminAuthID int64, permissions map[string]bool) (*entity.Admin, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	admin := entity.NewAdmin(newAdminAuthID, permissions)
	err = d.adminRepo.Save(ctx, admin)
	if err != nil {
		d.logger.Error(ctx, "failed to save admin",
			option.Any("new_admin_id", newAdminAuthID),
			option.Error(err))

		return nil, err
	}

	return admin, nil
}

func (d *DefaultAdminService) GetAdminByID(ctx context.Context, id int64) (*entity.Admin, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	admin, err := d.adminRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "admin not found by id",
				option.Any("admin_id", id),
				option.Error(service_errors.ErrAdminNotFound))

			return nil, service_errors.ErrAdminNotFound
		}

		d.logger.Error(ctx, "failed to get admin by id",
			option.Any("admin_id", id),
			option.Error(err))

		return nil, err
	}

	return admin, nil
}

func (d *DefaultAdminService) UpdateAdmin(ctx context.Context,
	updatingAdminAuthID int64, permissions map[string]bool) (*entity.Admin, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	admin, err := d.adminRepo.GetByAuthID(ctx, updatingAdminAuthID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "admin not found by id",
				option.Any("auth_id", updatingAdminAuthID),
				option.Error(service_errors.ErrAdminNotFound))
			return nil, service_errors.ErrAdminNotFound
		}

		d.logger.Error(ctx, "failed to get admin by id",
			option.Any("auth_id", updatingAdminAuthID),
			option.Error(err))

		return nil, err
	}

	admin.Permissions = permissions

	res, err := d.adminRepo.Update(ctx, admin)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "admin not found by id",
				option.Any("auth_id", updatingAdminAuthID),
				option.Error(service_errors.ErrAdminNotFound))

			return nil, service_errors.ErrAdminNotFound
		}

		d.logger.Error(ctx, "failed to update admin",
			option.Any("auth_id", updatingAdminAuthID),
			option.Error(err))

		return nil, err
	}
	return res, nil
}

func (d *DefaultAdminService) VerifyUser(ctx context.Context, userID int64) (*entity.User, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	user, err := d.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "user not found by id",
				option.Any("user_id", userID),
				option.Error(service_errors.ErrUserNotFound))

			return nil, service_errors.ErrUserNotFound
		}

		d.logger.Error(ctx, "failed to get user by id",
			option.Any("user_id", userID),
			option.Error(err))

		return nil, err
	}

	if !user.IsAnAdult(time.Now()) {
		return nil, service_errors.ErrIsNotAnAdult
	}

	user.IsVerified = true

	res, err := d.userRepo.Update(ctx, user)
	if err != nil {
		d.logger.Error(ctx, "failed to update user",
			option.Any("user_id", userID),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultAdminService) GetBookingByID(ctx context.Context, bookingID int64) (*entity.Booking, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	booking, err := d.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "booking not found by id",
				option.Any("booking_id", bookingID),
				option.Error(service_errors.ErrBookingNotFound))

			return nil, service_errors.ErrBookingNotFound
		}

		d.logger.Error(ctx, "failed to get booking by id",
			option.Any("booking_id", bookingID),
			option.Error(err))

		return nil, err
	}

	return booking, nil
}

func (d *DefaultAdminService) UpdateBookingStatus(ctx context.Context,
	bookingID int64, status entity.BookingStatus) (*entity.Booking, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	booking, err := d.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "booking not found by id",
				option.Any("booking_id", bookingID),
				option.Error(service_errors.ErrBookingNotFound))

			return nil, service_errors.ErrBookingNotFound
		}

		d.logger.Error(ctx, "failed to get booking by id",
			option.Any("booking_id", bookingID),
			option.Error(err))

		return nil, err
	}

	booking.Status = status

	res, err := d.bookingRepo.Update(ctx, booking)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "booking not found by id",
				option.Any("booking_id", bookingID),
				option.Error(service_errors.ErrBookingNotFound))

			return nil, service_errors.ErrBookingNotFound
		}

		d.logger.Error(ctx, "failed to update booking",
			option.Any("booking_id", bookingID),
			option.Error(err))

		return nil, err
	}
	return res, nil
}

func (d *DefaultAdminService) GetOrderByID(ctx context.Context, orderID int64) (*entity.Order, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	order, err := d.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "order not found by id",
				option.Any("order_id", orderID),
				option.Error(service_errors.ErrOrderNotFound))

			return nil, service_errors.ErrOrderNotFound
		}

		d.logger.Error(ctx, "failed to get order by id",
			option.Any("order_id", orderID),
			option.Error(err))

		return nil, err
	}

	return order, nil
}

func (d *DefaultAdminService) UpdateOrderStatus(ctx context.Context,
	orderID int64, status entity.OrderStatus) (*entity.Order, error) {

	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	order, err := d.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "order not found by id",
				option.Any("order_id", orderID),
				option.Error(service_errors.ErrOrderNotFound))

			return nil, service_errors.ErrOrderNotFound
		}

		d.logger.Error(ctx, "failed to get order by id",
			option.Any("order_id", orderID),
			option.Error(err))

		return nil, err
	}

	order.Status = status

	res, err := d.orderRepo.UpdateStatus(ctx, order)
	if err != nil {
		if errors.Is(err, persistence.ErrNoRowsFound) {
			d.logger.Error(ctx, "order not found by id",
				option.Any("order_id", orderID),
				option.Error(service_errors.ErrOrderNotFound))

			return nil, service_errors.ErrOrderNotFound
		}

		d.logger.Error(ctx, "failed to update order status",
			option.Any("order_id", orderID),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultAdminService) GetAllUsers(ctx context.Context, page, limit *int64) ([]*entity.User, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	res, err := d.userRepo.GetAll(ctx, entity.NewOptions(common.CheckPagination(page, limit)))
	if err != nil {
		d.logger.Error(ctx, "failed to get all users",
			option.Any("page", page),
			option.Any("limit", limit),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultAdminService) GetAllBookings(ctx context.Context, page, limit *int64) ([]*entity.Booking, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	res, err := d.bookingRepo.GetAll(ctx, entity.NewOptions(common.CheckPagination(page, limit)))
	if err != nil {
		d.logger.Error(ctx, "failed to get all bookings",
			option.Any("page", page),
			option.Any("limit", limit),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultAdminService) GetAllOrders(ctx context.Context, page, limit *int64) ([]*entity.Order, error) {
	authID, err := common.GetAuthIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = d.checkAdminRestrictions(ctx, authID)
	if err != nil {
		return nil, err
	}

	res, err := d.orderRepo.GetAll(ctx, entity.NewOptions(common.CheckPagination(page, limit)))
	if err != nil {
		d.logger.Error(ctx, "failed to get all orders",
			option.Any("page", page),
			option.Any("limit", limit),
			option.Error(err))

		return nil, err
	}

	return res, nil
}

func (d *DefaultAdminService) checkAdminRestrictions(ctx context.Context, authID *int64) error {
	role, err := common.GetRoleFromContext(ctx)
	if err != nil {
		return err
	}

	if *role != entity.RoleAdmin.String() {
		d.logger.Error(ctx, "access denied",
			option.Any("auth_id", authID),
			option.Error(service_errors.ErrNotAdmin))

		return service_errors.ErrNotAdmin
	}

	return nil
}
