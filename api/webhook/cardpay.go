package webhook

import (
	"github.com/ProtocolONE/p1pay.api/api"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/ProtocolONE/p1pay.api/payment_system/entity"
	"github.com/labstack/echo"
	"net/http"
)

const (
	cardPayWebHookPaymentNotifyPath = "/cardpay/notify"
)

type CardPayWebHook struct {
	api          *api.Api
	orderManager *manager.OrderManager
}

func (wh *WebHook) InitCardPayWebHookRoutes() *WebHook {
	cpWebHook := &CardPayWebHook{
		api:          wh.Api,
		orderManager: manager.InitOrderManager(wh.Api.Database, wh.Api.Logger, wh.Api.GeoDbReader, wh.Api.PSPAccountingCurrencyA3),
	}

	wh.Api.WebHookGroup.POST(cardPayWebHookPaymentNotifyPath, cpWebHook.paymentNotify)

	return wh
}

func (cpWebHook *CardPayWebHook) paymentNotify(ctx echo.Context) error {
	req := &entity.CardPayPaymentNotificationWebHookRequest{
		Signature: ctx.Request().Header.Get(entity.CardPayPaymentResponseHeaderSignature),
	}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.ResponseMessageInvalidRequestData)
	}

	if err := cpWebHook.api.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cpWebHook.api.GetFirstValidationError(err))
	}

	oPaymentNotification := &model.OrderPaymentNotification{
		Id:         req.MerchantOrder.Id,
		Request:    req,
		RawRequest: cpWebHook.api.WebHookRawBody,
	}

	order, err := cpWebHook.orderManager.ProcessNotifyPayment(oPaymentNotification, cpWebHook.api.PaymentSystemConfig)

	if err != nil && order == nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err != nil && order != nil {
		return ctx.JSON(http.StatusOK, err.Error())
	}

	return ctx.JSON(http.StatusOK, "Payment successfully complete")
}
