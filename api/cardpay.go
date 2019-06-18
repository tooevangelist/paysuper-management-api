package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/paysuper/paysuper-management-api/payment_system/entity"
	"net/http"
)

const (
	cardPayWebHookPaymentNotifyPath = "/cardpay/payment"
	cardPayWebHookRefundNotifyPath  = "/cardpay/refund"
)

type CardPayWebHook struct {
	*Api
}

func (api *Api) InitCardPayWebHookRoutes() *Api {
	cpWebHook := &CardPayWebHook{Api: api}
	api.webhookRouteGroup.POST(cardPayWebHookPaymentNotifyPath, cpWebHook.paymentCallback)
	api.webhookRouteGroup.POST(cardPayWebHookRefundNotifyPath, cpWebHook.refundCallback)

	return api
}

func (h *CardPayWebHook) paymentCallback(ctx echo.Context) error {
	st := &billing.CardPayPaymentCallback{}

	if err := ctx.Bind(st); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageInvalidRequestData)
	}

	if err := h.validate.Struct(st); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
	}

	req := &grpc.PaymentNotifyRequest{
		OrderId:   st.MerchantOrder.Id,
		Request:   []byte(h.rawBody),
		Signature: ctx.Request().Header.Get(entity.CardPayPaymentResponseHeaderSignature),
	}

	rsp, err := h.billingService.PaymentCallbackProcess(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageUnknownError)
	}

	var httpStatus int
	var message = map[string]string{"message": rsp.Error}

	switch rsp.Status {
	case pkg.StatusErrorValidation:
		httpStatus = http.StatusBadRequest
		break
	case pkg.StatusErrorSystem:
		httpStatus = http.StatusInternalServerError
		break
	case pkg.StatusTemporary:
		httpStatus = http.StatusGone
		break
	default:
		httpStatus = http.StatusOK
		message["message"] = "Payment successfully complete"
	}

	return ctx.JSON(httpStatus, message)
}

func (h *CardPayWebHook) refundCallback(ctx echo.Context) error {
	st := &billing.CardPayRefundCallback{}
	err := ctx.Bind(st)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	err = h.validate.Struct(st)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
	}

	req := &grpc.CallbackRequest{
		Handler:   pkg.PaymentSystemHandlerCardPay,
		Body:      []byte(h.rawBody),
		Signature: ctx.Request().Header.Get(entity.CardPayPaymentResponseHeaderSignature),
	}

	rsp, err := h.billingService.ProcessRefundCallback(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Error)
	}

	if rsp.Error != "" {
		return ctx.JSON(http.StatusOK, map[string]string{"message": rsp.Error})
	}

	return ctx.NoContent(http.StatusOK)
}
