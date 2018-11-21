package payment_system

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/model"
)

const (
	PaymentSystemHandlerCardPay = "cardpay"

	paymentSystemErrorHandlerNotFound            = "handler for specified payment system not found"
	paymentSystemErrorSettingsNotFound           = "payment system settings not found"
	paymentSystemErrorAuthenticateFailed         = "authentication failed"
	paymentSystemErrorUnknownPaymentMethod       = "unknown payment method"
	paymentSystemErrorCreateRequestFailed        = "order can't be create. try request later"
	paymentSystemErrorEWalletIdentifierIsInvalid = "wallet identifier is invalid"

	bankCardFieldPan       = "pan"
	bankCardFieldCvv       = "cvv"
	bankCardFieldMonth     = "month"
	bankCardFieldYear      = "year"
	bankCardFieldHolder    = "card_holder"
	eWalletFieldIdentifier = "ewallet"

	paymentSystemSettingsFieldNameCreatePaymentUrl = "create_payment_url"
)

var handlers = map[string]func(*model.Order, *Settings) PaymentSystem{
	PaymentSystemHandlerCardPay: NewCardPayHandler,
}

type PaymentSystem interface {
	CreatePayment() error
	ProcessPayment() error
}

type Settings struct {
	Url      string
	Settings interface{}
}

type Path struct {
	path   string
	method string
}

func GetPaymentHandler(order *model.Order, config map[string]interface{}) (PaymentSystem, error) {
	handler, ok := handlers[order.PaymentMethod.Params.Handler]

	if !ok {
		return nil, errors.New(paymentSystemErrorHandlerNotFound)
	}

	c, ok := config[order.PaymentMethod.Params.Handler]

	if !ok {
		return nil, errors.New(paymentSystemErrorSettingsNotFound)
	}

	cMap := c.(map[interface{}]interface{})

	s := &Settings{
		Url:      cMap[paymentSystemSettingsFieldNameCreatePaymentUrl].(string),
		Settings: cMap[order.PaymentMethod.Params.ExternalId],
	}

	return handler(order, s), nil
}
