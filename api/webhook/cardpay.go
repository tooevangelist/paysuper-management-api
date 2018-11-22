package webhook

import (
	"github.com/ProtocolONE/p1pay.api/api"
	"github.com/ProtocolONE/p1pay.api/payment_system/entity"
	"github.com/labstack/echo"
	"net/http"
)

const (
	cardPayWebHookPaymentNotifyPath = "/cardpay/notify"
)

type CardPayWebHook struct {
}

func (wh *WebHook) InitCardPayWebHookRoutes() *WebHook {
	cpWebHook := &CardPayWebHook{}

	wh.Api.WebHookGroup.POST(cardPayWebHookPaymentNotifyPath, cpWebHook.paymentNotify)

	return wh
}

func (cpWebHook *CardPayWebHook) paymentNotify(ctx echo.Context) error {
	req := &entity.CardPayPaymentNotificationWebHookRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.ResponseMessageInvalidRequestData)
	}
}
