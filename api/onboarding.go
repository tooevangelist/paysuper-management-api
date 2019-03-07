package api

import (
	"context"
	"github.com/labstack/echo"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type onboardingRoute struct {
	*Api
}

func (api *Api) initOnboardingRoutes() *Api {
	route := &onboardingRoute{Api: api}

	api.authUserRouteGroup.GET("/merchant", route.listMerchants)
	api.authUserRouteGroup.GET("/merchant/:id", route.getMerchant)
	api.authUserRouteGroup.POST("/merchant", route.changeMerchant)
	api.authUserRouteGroup.PUT("/merchant", route.changeMerchant)
	api.authUserRouteGroup.PUT("/merchant/:id/change-status", route.changeMerchantStatus)

	api.authUserRouteGroup.POST("/notification", route.createNotification)
	api.authUserRouteGroup.GET("/notification/:id", route.getNotification)
	api.authUserRouteGroup.GET("/notification", route.listNotifications)
	api.authUserRouteGroup.PUT("/notification/:id/mark-as-read", route.markAsReadNotification)

	return api
}

func (r *onboardingRoute) getMerchant(ctx echo.Context) error {
	id := ctx.Param(requestParameterId)

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorIdIsEmpty)
	}

	rsp, err := r.billingService.GetMerchantById(context.TODO(), &grpc.FindByIdRequest{Id: id})

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

	return ctx.JSON(http.StatusOK, rsp.Merchants)
}

func (r *onboardingRoute) changeMerchant(ctx echo.Context) error {
	req := &grpc.OnboardingRequest{}
	httpErr := r.onboardingBeforeHandler(req, ctx)

	if httpErr != nil {
		return httpErr
	}

	rsp, err := r.billingService.ChangeMerchant(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) changeMerchantStatus(ctx echo.Context) error {
	req := &grpc.MerchantChangeStatusRequest{}
	httpErr := r.onboardingBeforeHandler(req, ctx)

	if httpErr != nil {
		return httpErr
	}

	req.UserId = r.authUser.Id
	rsp, err := r.billingService.ChangeMerchantStatus(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) createNotification(ctx echo.Context) error {
	req := &grpc.NotificationRequest{}
	httpErr := r.onboardingBeforeHandler(req, ctx)

	if httpErr != nil {
		return httpErr
	}

	req.UserId = r.authUser.Id
	rsp, err := r.billingService.CreateNotification(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusCreated, rsp)
}

func (r *onboardingRoute) getNotification(ctx echo.Context) error {
	id := ctx.Param(requestParameterId)

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorIdIsEmpty)
	}

	rsp, err := r.billingService.GetNotification(context.TODO(), &grpc.FindByIdRequest{Id: id})

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
	id := ctx.Param(requestParameterId)

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorIdIsEmpty)
	}

	rsp, err := r.billingService.MarkNotificationAsRead(context.TODO(), &grpc.FindByIdRequest{Id: id})

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) getPaymentMethod(ctx echo.Context) error {
	req := &grpc.GetMerchantPaymentMethodRequest{}
	httpErr := r.onboardingBeforeHandler(req, ctx)

	if httpErr != nil {
		return httpErr
	}

	rsp, err := r.billingService.GetMerchantPaymentMethod(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) listPaymentMethods(ctx echo.Context) error {
	req := &grpc.ListMerchantPaymentMethodsRequest{}
	httpErr := r.onboardingBeforeHandler(req, ctx)

	if httpErr != nil {
		return httpErr
	}

	rsp, err := r.billingService.ListMerchantPaymentMethods(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp.PaymentMethods)
}

func (r *onboardingRoute) changePaymentMethod(ctx echo.Context) error {
	req := &grpc.MerchantPaymentMethodRequest{}
	httpErr := r.onboardingBeforeHandler(req, ctx)

	if httpErr != nil {
		return httpErr
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
