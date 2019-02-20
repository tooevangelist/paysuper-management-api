package payment_system

import (
	"errors"
	"github.com/paysuper/paysuper-management-api/database/model"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	PaymentSystemHandlerCardPay = "cardpay"

	paymentSystemErrorHandlerNotFound                = "handler for specified payment system not found"
	paymentSystemErrorSettingsNotFound               = "payment system settings not found"
	paymentSystemErrorAuthenticateFailed             = "authentication failed"
	paymentSystemErrorUnknownPaymentMethod           = "unknown payment method"
	paymentSystemErrorCreateRequestFailed            = "order can't be create. try request later"
	paymentSystemErrorEWalletIdentifierIsInvalid     = "wallet identifier is invalid"
	paymentSystemErrorCryptoCurrencyAddressIsInvalid = "crypto currency address is invalid"
	paymentSystemErrorRequestSignatureIsInvalid      = "request signature is invalid"
	paymentSystemErrorRequestTimeFieldIsInvalid      = "time field in request is invalid"
	paymentSystemErrorRequestStatusIsInvalid         = "status is invalid"
	paymentSystemErrorRequestPaymentMethodIsInvalid  = "payment method from request not equal value in order"
	paymentSystemErrorRequestTemporarySkipped        = "notification skipped with temporary status"

	paymentSystemSettingsFieldNameCreatePaymentUrl = "create_payment_url"

	PaymentStatusOK                       = 0
	PaymentStatusErrorValidation          = 1
	PaymentStatusErrorSystem              = 2
	CreatePaymentStatusErrorPaymentSystem = 3
	PaymentStatusTemporary                = 4

	settingsFieldTerminalId         = "terminal_id"
	settingsFieldSecretWord         = "secret_word"
	settingsFieldCallbackSecretWord = "callback_secret_word"
)

var handlers = map[string]func(*model.Order, *Settings) PaymentSystem{
	PaymentSystemHandlerCardPay: NewCardPayHandler,
}

type PaymentSystem interface {
	CreatePayment() *PaymentResponse
	ProcessPayment(*model.Order, *model.OrderPaymentNotification) *PaymentResponse
}

type PaymentSystemSetting struct {
	Logger *zap.SugaredLogger
}

type Settings struct {
	Url      string
	Settings interface{}
	*PaymentSystemSetting
}

type Path struct {
	path   string
	method string
}

type PaymentResponse struct {
	Status      int    `json:"-"`
	RedirectUrl string `json:"redirect_url,omitempty"`
	Error       string `json:"error,omitempty"`
	*model.Order
}

func (pss *PaymentSystemSetting) GetPaymentHandler(order *model.Order, config map[string]interface{}) (PaymentSystem, error) {
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
		Url:                  cMap[paymentSystemSettingsFieldNameCreatePaymentUrl].(string),
		Settings:             cMap[order.PaymentMethod.Params.ExternalId],
		PaymentSystemSetting: pss,
	}

	return handler(order, s), nil
}

func (pss *PaymentSystemSetting) GetLoggableHttpClient() *http.Client {
	return &http.Client{
		Transport: &Transport{Logger: pss.Logger},
		Timeout:   time.Duration(defaultHttpClientTimeout * time.Second),
	}
}

func NewPaymentResponse(status int, error string) *PaymentResponse {
	cpResp := &PaymentResponse{Status: status}

	if error != "" {
		cpResp.Error = error
	}

	return cpResp
}

func (pr *PaymentResponse) SetOrder(o *model.Order) *PaymentResponse  {
	pr.Order = o

	return pr
}

func (pr *PaymentResponse) SetError(error string) *PaymentResponse  {
	pr.Error = error

	return pr
}
