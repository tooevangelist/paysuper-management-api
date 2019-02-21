package api

import (
	"context"
	"github.com/labstack/echo"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/paysuper/paysuper-management-api/payment_system/entity"
	"net/http"
)

const (
	cardPayWebHookPaymentNotifyPath = "/cardpay/notify"
)

type CardPayWebHook struct {
	*Api
}

func (api *Api) InitCardPayWebHookRoutes() *Api {
	cpWebHook := &CardPayWebHook{Api: api}
	api.webhookRouteGroup.POST(cardPayWebHookPaymentNotifyPath, cpWebHook.paymentCallback)

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

	rsp, err := h.billingService.PaymentCallbackProcess(context.TODO(), req)

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
