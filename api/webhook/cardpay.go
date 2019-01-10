package webhook

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/ProtocolONE/p1pay.api/payment_system"
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
		orderManager: manager.InitOrderManager(wh.database, wh.logger, wh.geoDbReader, wh.pspAccountingCurrencyA3, wh.paymentSystemSettings, wh.publisher, wh.centrifugoSecret),
	}

	wh.webHookGroup.POST(cardPayWebHookPaymentNotifyPath, cpWebHook.paymentNotify)

	return wh
}

func (cpWebHook *CardPayWebHook) paymentNotify(ctx echo.Context) error {
	req := &entity.CardPayPaymentNotificationWebHookRequest{
		Signature: ctx.Request().Header.Get(entity.CardPayPaymentResponseHeaderSignature),
	}

	// temporary hack to change content-type header to correct
	ctx.Request().Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageInvalidRequestData)
	}

	if err := cpWebHook.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
	}

	oPaymentNotification := &model.OrderPaymentNotification{
		Id:         req.MerchantOrder.Id,
		Request:    req,
		RawRequest: cpWebHook.rawBody,
	}

	res := cpWebHook.orderManager.ProcessNotifyPayment(oPaymentNotification, cpWebHook.paymentSystemConfig)

	var httpStatus int
	var message = map[string]string{"message": res.Error}

	switch res.Status {
	case payment_system.PaymentStatusErrorValidation:
		httpStatus = http.StatusBadRequest
		break
	case payment_system.PaymentStatusErrorSystem:
		httpStatus = http.StatusInternalServerError
		break
	case payment_system.PaymentStatusTemporary:
		httpStatus = http.StatusGone
		break
	default:
		httpStatus = http.StatusOK
		message["message"] = "Payment successfully complete"
	}

	return ctx.JSON(httpStatus, message)
}
