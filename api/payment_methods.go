package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type PaymentMethodApiV1 struct {
	*Api
}

func (api *Api) initPaymentMethodRoutes() *Api {
	pmApiV1 := PaymentMethodApiV1{
		Api: api,
	}

	api.accessRouteGroup.POST("/payment_method", pmApiV1.create)
	api.accessRouteGroup.PUT("/payment_method/:id", pmApiV1.update)
	api.accessRouteGroup.POST("/payment_method/:id/production", pmApiV1.createProductionSettings)
	api.accessRouteGroup.PUT("/payment_method/:id/production", pmApiV1.updateProductionSettings)
	api.accessRouteGroup.GET("/payment_method/:id/production", pmApiV1.getProductionSettings)
	api.accessRouteGroup.DELETE("/payment_method/:id/production", pmApiV1.deleteProductionSettings)
	api.accessRouteGroup.POST("/payment_method/:id/test", pmApiV1.createTestSettings)
	api.accessRouteGroup.PUT("/payment_method/:id/test", pmApiV1.updateTestSettings)
	api.accessRouteGroup.GET("/payment_method/:id/test", pmApiV1.getTestSettings)
	api.accessRouteGroup.DELETE("/payment_method/:id/test", pmApiV1.deleteTestSettings)

	return api
}

// Create new payment method
// POST /api/v1/payment_method/:id
func (pmApiV1 *PaymentMethodApiV1) create(ctx echo.Context) error {
	return pmApiV1.createOrUpdatePaymentMethod(ctx)
}

// Update exists payment method
// PUT /api/v1/payment_method/:id
func (pmApiV1 *PaymentMethodApiV1) update(ctx echo.Context) error {
	return pmApiV1.createOrUpdatePaymentMethod(ctx)
}

func (pmApiV1 *PaymentMethodApiV1) createOrUpdatePaymentMethod(ctx echo.Context) error {
	req := &billing.PaymentMethod{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = pmApiV1.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pmApiV1.getValidationError(err))
	}

	res, err := pmApiV1.billingService.CreateOrUpdatePaymentMethod(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get production settings for payment method
// GET /api/v1/payment_method/:id/production
func (pmApiV1 *PaymentMethodApiV1) getProductionSettings(ctx echo.Context) error {
	req := &grpc.GetPaymentMethodSettingsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = pmApiV1.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pmApiV1.getValidationError(err))
	}

	res, err := pmApiV1.billingService.GetPaymentMethodProductionSettings(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Create new production settings for payment method
// POST /payment_method/:id/production
func (pmApiV1 *PaymentMethodApiV1) createProductionSettings(ctx echo.Context) error {
	return pmApiV1.createOrUpdateProductionSettings(ctx)
}

// Update exists production settings for payment method
// PUT /api/v1/payment_method/:id/production
func (pmApiV1 *PaymentMethodApiV1) updateProductionSettings(ctx echo.Context) error {
	return pmApiV1.createOrUpdateProductionSettings(ctx)
}

func (pmApiV1 *PaymentMethodApiV1) createOrUpdateProductionSettings(ctx echo.Context) error {
	req := &grpc.ChangePaymentMethodParamsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = pmApiV1.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pmApiV1.getValidationError(err))
	}

	res, err := pmApiV1.billingService.CreateOrUpdatePaymentMethodProductionSettings(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Delete production settings for payment method
// DELETE /api/v1/payment_method/:id/production
func (pmApiV1 *PaymentMethodApiV1) deleteProductionSettings(ctx echo.Context) error {
	req := &grpc.GetPaymentMethodSettingsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = pmApiV1.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pmApiV1.getValidationError(err))
	}

	res, err := pmApiV1.billingService.DeletePaymentMethodProductionSettings(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get test settings for payment method
// GET /api/v1/payment_method/:id/test
func (pmApiV1 *PaymentMethodApiV1) getTestSettings(ctx echo.Context) error {
	req := &grpc.GetPaymentMethodSettingsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = pmApiV1.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pmApiV1.getValidationError(err))
	}

	res, err := pmApiV1.billingService.GetPaymentMethodTestSettings(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Create new test settings for payment method
// POST /payment_method/:id/test
func (pmApiV1 *PaymentMethodApiV1) createTestSettings(ctx echo.Context) error {
	return pmApiV1.createOrUpdateProductionSettings(ctx)
}

// Update exists test settings for payment method
// PUT /api/v1/payment_method/:id/test
func (pmApiV1 *PaymentMethodApiV1) updateTestSettings(ctx echo.Context) error {
	return pmApiV1.createOrUpdateProductionSettings(ctx)
}

func (pmApiV1 *PaymentMethodApiV1) createOrUpdateTestSettings(ctx echo.Context) error {
	req := &grpc.ChangePaymentMethodParamsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = pmApiV1.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pmApiV1.getValidationError(err))
	}

	res, err := pmApiV1.billingService.CreateOrUpdatePaymentMethodTestSettings(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Delete test settings for payment method
// DELETE /api/v1/payment_method/:id/test
func (pmApiV1 *PaymentMethodApiV1) deleteTestSettings(ctx echo.Context) error {
	req := &grpc.GetPaymentMethodSettingsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = pmApiV1.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pmApiV1.getValidationError(err))
	}

	res, err := pmApiV1.billingService.DeletePaymentMethodTestSettings(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}
