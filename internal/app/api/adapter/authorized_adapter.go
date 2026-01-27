package adapter

import (
	"context"

	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/api/generated/authorized"
	"github.com/alishashelby/Samok-Aah-t/backend/internal/app/handler"
)

type AuthorizedAdapter struct {
	User         *handler.UserHandler
	ModelService *handler.ModelServiceHandler
	Slot         *handler.SlotHandler
	Booking      *handler.BookingHandler
	Order        *handler.OrderHandler
	Admin        *handler.AdminHandler
}

func NewAuthorizedAdapter(user *handler.UserHandler, modelService *handler.ModelServiceHandler,
	slot *handler.SlotHandler, booking *handler.BookingHandler,
	order *handler.OrderHandler, admin *handler.AdminHandler) *AuthorizedAdapter {

	return &AuthorizedAdapter{
		User:         user,
		ModelService: modelService,
		Slot:         slot,
		Booking:      booking,
		Order:        order,
		Admin:        admin,
	}

}

func (a *AuthorizedAdapter) PostAdmin(ctx context.Context,
	request authorized.PostAdminRequestObject) (authorized.PostAdminResponseObject, error) {
	return a.Admin.CreateAdmin(ctx, request)
}

func (a *AuthorizedAdapter) GetAdminBookings(ctx context.Context,
	request authorized.GetAdminBookingsRequestObject) (authorized.GetAdminBookingsResponseObject, error) {
	return a.Admin.GetAllBookings(ctx, request)
}

func (a *AuthorizedAdapter) GetAdminBookingsId(ctx context.Context,
	request authorized.GetAdminBookingsIdRequestObject) (authorized.GetAdminBookingsIdResponseObject, error) {
	return a.Admin.GetBookingByID(ctx, request)
}

func (a *AuthorizedAdapter) PatchAdminBookingsIdStatus(ctx context.Context,
	request authorized.PatchAdminBookingsIdStatusRequestObject,
) (authorized.PatchAdminBookingsIdStatusResponseObject, error) {
	return a.Admin.UpdateBookingStatus(ctx, request)
}

func (a *AuthorizedAdapter) GetAdminOrders(ctx context.Context,
	request authorized.GetAdminOrdersRequestObject) (authorized.GetAdminOrdersResponseObject, error) {
	return a.Admin.GetAllOrders(ctx, request)
}

func (a *AuthorizedAdapter) GetAdminOrdersId(ctx context.Context,
	request authorized.GetAdminOrdersIdRequestObject) (authorized.GetAdminOrdersIdResponseObject, error) {
	return a.Admin.GetOrderByID(ctx, request)
}

func (a *AuthorizedAdapter) PatchAdminOrdersIdStatus(ctx context.Context,
	request authorized.PatchAdminOrdersIdStatusRequestObject,
) (authorized.PatchAdminOrdersIdStatusResponseObject, error) {
	return a.Admin.UpdateOrderStatus(ctx, request)
}

func (a *AuthorizedAdapter) GetAdminUsers(ctx context.Context,
	request authorized.GetAdminUsersRequestObject) (authorized.GetAdminUsersResponseObject, error) {
	return a.Admin.GetAllUsers(ctx, request)
}

func (a *AuthorizedAdapter) PatchAdminUsersIdVerify(ctx context.Context,
	request authorized.PatchAdminUsersIdVerifyRequestObject,
) (authorized.PatchAdminUsersIdVerifyResponseObject, error) {
	return a.Admin.VerifyUser(ctx, request)
}

func (a *AuthorizedAdapter) GetAdminId(ctx context.Context,
	request authorized.GetAdminIdRequestObject) (authorized.GetAdminIdResponseObject, error) {
	return a.Admin.GetAdminById(ctx, request)
}

func (a *AuthorizedAdapter) PatchAdminId(ctx context.Context,
	request authorized.PatchAdminIdRequestObject) (authorized.PatchAdminIdResponseObject, error) {
	return a.Admin.UpdateAdmin(ctx, request)
}

func (a *AuthorizedAdapter) PostClientBookings(ctx context.Context,
	request authorized.PostClientBookingsRequestObject) (authorized.PostClientBookingsResponseObject, error) {
	return a.Booking.CreateBooking(ctx, request)
}

func (a *AuthorizedAdapter) PatchClientBookingsIdCancel(ctx context.Context,
	request authorized.PatchClientBookingsIdCancelRequestObject,
) (authorized.PatchClientBookingsIdCancelResponseObject, error) {
	return a.Booking.CancelBookingByClient(ctx, request)
}

func (a *AuthorizedAdapter) PatchClientOrdersIdCancel(ctx context.Context,
	request authorized.PatchClientOrdersIdCancelRequestObject,
) (authorized.PatchClientOrdersIdCancelResponseObject, error) {
	return a.Order.CancelOrderByClient(ctx, request)
}

func (a *AuthorizedAdapter) GetClientServices(ctx context.Context,
	request authorized.GetClientServicesRequestObject,
) (authorized.GetClientServicesResponseObject, error) {
	return a.ModelService.GetAllServices(ctx, request)
}

func (a *AuthorizedAdapter) GetClientServicesId(ctx context.Context,
	request authorized.GetClientServicesIdRequestObject) (authorized.GetClientServicesIdResponseObject, error) {
	return a.ModelService.GetServiceByID(ctx, request)
}

func (a *AuthorizedAdapter) GetClientModelsModelIdSlots(ctx context.Context,
	request authorized.GetClientModelsModelIdSlotsRequestObject,
) (authorized.GetClientModelsModelIdSlotsResponseObject, error) {
	return a.Slot.GetModelSlotsForClient(ctx, request)
}

func (a *AuthorizedAdapter) PatchModelBookingsIdApprove(ctx context.Context,
	request authorized.PatchModelBookingsIdApproveRequestObject,
) (authorized.PatchModelBookingsIdApproveResponseObject, error) {
	return a.Booking.ApproveBooking(ctx, request)
}

func (a *AuthorizedAdapter) PatchModelBookingsIdReject(ctx context.Context,
	request authorized.PatchModelBookingsIdRejectRequestObject,
) (authorized.PatchModelBookingsIdRejectResponseObject, error) {
	return a.Booking.RejectBooking(ctx, request)
}

func (a *AuthorizedAdapter) GetModelOrders(ctx context.Context,
	request authorized.GetModelOrdersRequestObject) (authorized.GetModelOrdersResponseObject, error) {
	return a.Order.GetModelOrders(ctx, request)
}

func (a *AuthorizedAdapter) PatchModelOrdersIdCancel(ctx context.Context,
	request authorized.PatchModelOrdersIdCancelRequestObject,
) (authorized.PatchModelOrdersIdCancelResponseObject, error) {
	return a.Order.CancelOrderByModel(ctx, request)
}

func (a *AuthorizedAdapter) PatchModelOrdersIdComplete(ctx context.Context,
	request authorized.PatchModelOrdersIdCompleteRequestObject,
) (authorized.PatchModelOrdersIdCompleteResponseObject, error) {
	return a.Order.CompleteOrder(ctx, request)
}

func (a *AuthorizedAdapter) GetModelServices(ctx context.Context,
	request authorized.GetModelServicesRequestObject) (authorized.GetModelServicesResponseObject, error) {
	return a.ModelService.GetModelServices(ctx, request)
}

func (a *AuthorizedAdapter) PostModelServices(ctx context.Context,
	request authorized.PostModelServicesRequestObject) (authorized.PostModelServicesResponseObject, error) {
	return a.ModelService.CreateService(ctx, request)
}

func (a *AuthorizedAdapter) PatchModelServicesId(ctx context.Context,
	request authorized.PatchModelServicesIdRequestObject) (authorized.PatchModelServicesIdResponseObject, error) {
	return a.ModelService.UpdateService(ctx, request)
}

func (a *AuthorizedAdapter) PatchModelServicesIdDeactivate(ctx context.Context,
	request authorized.PatchModelServicesIdDeactivateRequestObject,
) (authorized.PatchModelServicesIdDeactivateResponseObject, error) {
	return a.ModelService.DeactivateService(ctx, request)
}

func (a *AuthorizedAdapter) GetModelSlots(ctx context.Context,
	request authorized.GetModelSlotsRequestObject) (authorized.GetModelSlotsResponseObject, error) {
	return a.Slot.GetOwnModelSlots(ctx, request)
}

func (a *AuthorizedAdapter) PostModelSlots(ctx context.Context,
	request authorized.PostModelSlotsRequestObject) (authorized.PostModelSlotsResponseObject, error) {
	return a.Slot.CreateSlot(ctx, request)
}

func (a *AuthorizedAdapter) PatchModelSlotsSlotId(ctx context.Context,
	request authorized.PatchModelSlotsSlotIdRequestObject,
) (authorized.PatchModelSlotsSlotIdResponseObject, error) {
	return a.Slot.UpdateSlot(ctx, request)
}

func (a *AuthorizedAdapter) PatchModelSlotsSlotIdDisable(ctx context.Context,
	request authorized.PatchModelSlotsSlotIdDisableRequestObject,
) (authorized.PatchModelSlotsSlotIdDisableResponseObject, error) {
	return a.Slot.DeactivateSlot(ctx, request)
}

func (a *AuthorizedAdapter) PostUsers(ctx context.Context,
	request authorized.PostUsersRequestObject) (authorized.PostUsersResponseObject, error) {
	return a.User.CreateProfile(ctx, request)
}

func (a *AuthorizedAdapter) GetUsersMe(ctx context.Context,
	request authorized.GetUsersMeRequestObject) (authorized.GetUsersMeResponseObject, error) {
	return a.User.GetOwnProfile(ctx, request)
}

func (a *AuthorizedAdapter) PatchUsersMe(ctx context.Context,
	request authorized.PatchUsersMeRequestObject) (authorized.PatchUsersMeResponseObject, error) {
	return a.User.UpdateProfile(ctx, request)
}

func (a *AuthorizedAdapter) GetUsersId(ctx context.Context,
	request authorized.GetUsersIdRequestObject) (authorized.GetUsersIdResponseObject, error) {
	return a.User.GetSomeoneProfile(ctx, request)
}
