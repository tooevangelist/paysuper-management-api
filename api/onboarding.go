package api

import (
	"context"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type onboardingRoute struct {
	*Api
}

func (api *Api) initOnboardingRoutes() *Api {
	route := &onboardingRoute{Api: api}

	api.authUserRouteGroup.GET("/merchants", route.listMerchants)
	api.authUserRouteGroup.GET("/merchants/:id", route.getMerchant)
	api.authUserRouteGroup.GET("/merchants/user", route.getMerchantByUser)
	api.authUserRouteGroup.POST("/merchants", route.changeMerchant)
	api.authUserRouteGroup.PUT("/merchants", route.changeMerchant)
	api.authUserRouteGroup.PUT("/merchants/:id/change-status", route.changeMerchantStatus)
	api.authUserRouteGroup.PATCH("/merchants/:id/agreement-type", route.changeAgreementType)
	api.authUserRouteGroup.PATCH("/merchants/:id/agreement-sign", route.singAgreement)

	api.authUserRouteGroup.POST("/merchants/:merchant_id/notifications", route.createNotification)
	api.authUserRouteGroup.GET("/merchants/:merchant_id/notifications/:notification_id", route.getNotification)
	api.authUserRouteGroup.GET("/merchants/:merchant_id/notifications", route.listNotifications)
	api.authUserRouteGroup.PUT("/merchants/:merchant_id/notifications/:notification_id/mark-as-read", route.markAsReadNotification)

	api.authUserRouteGroup.GET("/merchants/:merchant_id/methods/:method_id", route.getPaymentMethod)
	api.authUserRouteGroup.GET("/merchants/:merchant_id/methods", route.listPaymentMethods)
	api.authUserRouteGroup.PUT("/merchants/:merchant_id/methods/:method_id", route.changePaymentMethod)

	return api
}

func (r *onboardingRoute) getMerchant(ctx echo.Context) error {
	id := ctx.Param(requestParameterId)

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorIdIsEmpty)
	}

	req := &grpc.GetMerchantByRequest{MerchantId: id}
	rsp, err := r.billingService.GetMerchantBy(context.TODO(), req)

	if err != nil {
		r.logError("Call billing-server method GetMerchantBy failed", []interface{}{"error", err.Error(), "request", req})
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) getMerchantByUser(ctx echo.Context) error {
	if r.authUser.Id == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, errorMessageAccessDenied)
	}

	rsp, err := r.billingService.GetMerchantBy(context.TODO(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) listMerchants(ctx echo.Context) error {
	req := &grpc.MerchantListingRequest{}
	err := (&OnboardingMerchantListingBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	rsp, err := r.billingService.ListMerchants(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) changeMerchant(ctx echo.Context) error {
	req := &grpc.OnboardingRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	req.User = &billing.MerchantUser{
		Id:    r.authUser.Id,
		Email: r.authUser.Email,
	}
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ChangeMerchant(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) changeMerchantStatus(ctx echo.Context) error {
	req := &grpc.MerchantChangeStatusRequest{}
	err := (&OnboardingChangeMerchantStatusBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	req.UserId = r.authUser.Id
	rsp, err := r.billingService.ChangeMerchantStatus(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) changeAgreementType(ctx echo.Context) error {
	req := &grpc.ChangeMerchantAgreementTypeRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	req.MerchantId = ctx.Param(requestParameterId)
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ChangeMerchantAgreementType(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) createNotification(ctx echo.Context) error {
	req := &grpc.NotificationRequest{}
	err := (&OnboardingCreateNotificationBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	req.UserId = r.authUser.Id
	rsp, err := r.billingService.CreateNotification(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusCreated, rsp)
}

func (r *onboardingRoute) getNotification(ctx echo.Context) error {
	merchantId := ctx.Param(requestParameterMerchantId)
	notificationId := ctx.Param(requestParameterNotificationId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectMerchantId)
	}

	if notificationId == "" || bson.IsObjectIdHex(notificationId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectNotificationId)
	}

	req := &grpc.GetNotificationRequest{
		MerchantId:     merchantId,
		NotificationId: notificationId,
	}
	rsp, err := r.billingService.GetNotification(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) listNotifications(ctx echo.Context) error {
	req := &grpc.ListingNotificationRequest{}
	err := (&OnboardingNotificationsListBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	rsp, err := r.billingService.ListNotifications(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp.Notifications)
}

func (r *onboardingRoute) markAsReadNotification(ctx echo.Context) error {
	merchantId := ctx.Param(requestParameterMerchantId)
	notificationId := ctx.Param(requestParameterNotificationId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectMerchantId)
	}

	if notificationId == "" || bson.IsObjectIdHex(notificationId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectNotificationId)
	}

	req := &grpc.GetNotificationRequest{
		MerchantId:     merchantId,
		NotificationId: notificationId,
	}
	rsp, err := r.billingService.MarkNotificationAsRead(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) getPaymentMethod(ctx echo.Context) error {
	req := &grpc.GetMerchantPaymentMethodRequest{}
	err := (&OnboardingGetPaymentMethodBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	rsp, err := r.billingService.GetMerchantPaymentMethod(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) listPaymentMethods(ctx echo.Context) error {
	req := &grpc.ListMerchantPaymentMethodsRequest{}
	err := (&OnboardingListPaymentMethodsBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	rsp, err := r.billingService.ListMerchantPaymentMethods(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp.PaymentMethods)
}

func (r *onboardingRoute) changePaymentMethod(ctx echo.Context) error {
	req := &grpc.MerchantPaymentMethodRequest{}
	err := (&OnboardingChangePaymentMethodBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ChangeMerchantPaymentMethod(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) singAgreement(ctx echo.Context) error {
	req := &grpc.SignMerchantRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	req.MerchantId = ctx.Param(requestParameterId)
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ProcessMerchantAgreement(context.TODO(), req)

	if err != nil {
		r.logError(
			`Call billing server method "ProcessMerchantAgreement" failed`,
			[]interface{}{"error", err.Error(), "request", req},
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}
