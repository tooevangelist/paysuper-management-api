package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	paymentMethodPath           = "/payment_method"
	paymentMethodIdPath         = "/payment_method/:id"
	paymentMethodProductionPath = "/payment_method/:id/production"
	paymentMethodTestPath       = "/payment_method/:id/test"
)

type PaymentMethodApiV1 struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewPaymentMethodApiV1(set common.HandlerSet, cfg *common.Config) *PaymentMethodApiV1 {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PaymentMethodApiV1"})
	return &PaymentMethodApiV1{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *PaymentMethodApiV1) Route(groups *common.Groups) {
	groups.SystemUser.POST(paymentMethodPath, h.create)
	groups.SystemUser.PUT(paymentMethodIdPath, h.update)
	groups.SystemUser.POST(paymentMethodProductionPath, h.createProductionSettings)
	groups.SystemUser.PUT(paymentMethodProductionPath, h.updateProductionSettings)
	groups.SystemUser.GET(paymentMethodProductionPath, h.getProductionSettings)
	groups.SystemUser.DELETE(paymentMethodProductionPath, h.deleteProductionSettings)
	groups.SystemUser.POST(paymentMethodTestPath, h.createTestSettings)
	groups.SystemUser.PUT(paymentMethodTestPath, h.updateTestSettings)
	groups.SystemUser.GET(paymentMethodTestPath, h.getTestSettings)
	groups.SystemUser.DELETE(paymentMethodTestPath, h.deleteTestSettings)
}

func (h *PaymentMethodApiV1) create(ctx echo.Context) error {
	return h.createOrUpdatePaymentMethod(ctx)
}

func (h *PaymentMethodApiV1) update(ctx echo.Context) error {
	return h.createOrUpdatePaymentMethod(ctx)
}

func (h *PaymentMethodApiV1) createOrUpdatePaymentMethod(ctx echo.Context) error {
	req := &billing.PaymentMethod{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreateOrUpdatePaymentMethod(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *PaymentMethodApiV1) getProductionSettings(ctx echo.Context) error {
	req := &grpc.GetPaymentMethodSettingsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaymentMethodProductionSettings(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *PaymentMethodApiV1) createProductionSettings(ctx echo.Context) error {
	return h.createOrUpdateProductionSettings(ctx)
}

func (h *PaymentMethodApiV1) updateProductionSettings(ctx echo.Context) error {
	return h.createOrUpdateProductionSettings(ctx)
}

func (h *PaymentMethodApiV1) createOrUpdateProductionSettings(ctx echo.Context) error {

	req := &grpc.ChangePaymentMethodParamsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreateOrUpdatePaymentMethodProductionSettings(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *PaymentMethodApiV1) deleteProductionSettings(ctx echo.Context) error {
	req := &grpc.GetPaymentMethodSettingsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeletePaymentMethodProductionSettings(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *PaymentMethodApiV1) getTestSettings(ctx echo.Context) error {
	req := &grpc.GetPaymentMethodSettingsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaymentMethodTestSettings(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *PaymentMethodApiV1) createTestSettings(ctx echo.Context) error {
	return h.createOrUpdateTestSettings(ctx)
}

func (h *PaymentMethodApiV1) updateTestSettings(ctx echo.Context) error {
	return h.createOrUpdateTestSettings(ctx)
}

func (h *PaymentMethodApiV1) createOrUpdateTestSettings(ctx echo.Context) error {
	req := &grpc.ChangePaymentMethodParamsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreateOrUpdatePaymentMethodTestSettings(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *PaymentMethodApiV1) deleteTestSettings(ctx echo.Context) error {
	req := &grpc.GetPaymentMethodSettingsRequest{
		PaymentMethodId: ctx.Param("id"),
	}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeletePaymentMethodTestSettings(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}
