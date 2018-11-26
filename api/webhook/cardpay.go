package webhook

import (
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
	*WebHook
	orderManager *manager.OrderManager
}

func (wh *WebHook) InitCardPayWebHookRoutes() *WebHook {
	cpWebHook := &CardPayWebHook{
		WebHook:      wh,
		orderManager: manager.InitOrderManager(wh.database, wh.logger, wh.geoDbReader, wh.pspAccountingCurrencyA3, wh.paymentSystemSettings),
	}

	wh.webHookGroup.POST(cardPayWebHookPaymentNotifyPath, cpWebHook.paymentNotify)

	return wh
}

func (cpWebHook *CardPayWebHook) paymentNotify(ctx echo.Context) error {
	req := &entity.CardPayPaymentNotificationWebHookRequest{
		Signature: ctx.Request().Header.Get(entity.CardPayPaymentResponseHeaderSignature),
	}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageInvalidRequestData)
	}

	if err := cpWebHook.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
	}

	oPaymentNotification := &model.OrderPaymentNotification{
		Id:         req.MerchantOrder.Id,
		Request:    req,
		RawRequest: cpWebHook.webHookRawBody,
	}

	order, err := cpWebHook.orderManager.ProcessNotifyPayment(oPaymentNotification, cpWebHook.paymentSystemConfig)

	if err != nil && order == nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err != nil && order != nil {
		return ctx.JSON(http.StatusOK, err.Error())
	}

	return ctx.JSON(http.StatusOK, "Payment successfully complete")
}
