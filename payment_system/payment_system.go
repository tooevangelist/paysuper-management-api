package payment_system

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/model"
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

	paymentSystemSettingsFieldNameCreatePaymentUrl = "create_payment_url"

	CreatePaymentStatusOK                 = 0
	CreatePaymentStatusErrorValidation    = 1
	CreatePaymentStatusErrorSystem        = 2
	CreatePaymentStatusErrorPaymentSystem = 3

	settingsFieldTerminalId         = "terminal_id"
	settingsFieldSecretWord         = "secret_word"
	settingsFieldCallbackSecretWord = "callback_secret_word"
)

var handlers = map[string]func(*model.Order, *Settings) PaymentSystem{
	PaymentSystemHandlerCardPay: NewCardPayHandler,
}

type PaymentSystem interface {
	CreatePayment() *CreatePaymentResponse
	ProcessPayment(*model.Order, *model.OrderPaymentNotification) (*model.Order, error)
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

type CreatePaymentResponse struct {
	Status      int    `json:"-"`
	RedirectUrl string `json:"redirect_url,omitempty"`
	Error       string `json:"error,omitempty"`
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
		Url:      cMap[paymentSystemSettingsFieldNameCreatePaymentUrl].(string),
		Settings: cMap[order.PaymentMethod.Params.ExternalId],
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

func GetCreatePaymentResponse(status int, error string, url string) *CreatePaymentResponse {
	cpResp := &CreatePaymentResponse{Status: status}

	if error != "" {
		cpResp.Error = error
	}

	if url != "" {
		cpResp.RedirectUrl = url
	}

	return cpResp
}
