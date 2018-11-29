package entity

import "time"

const (
	CardPayPaymentResponseStatusNew = "NEW"
	CardPayPaymentResponseStatusInProgress = "IN_PROGRESS"
	CardPayPaymentResponseStatusDeclined = "DECLINED"
	CardPayPaymentResponseStatusAuthorized = "AUTHORIZED"
	CardPayPaymentResponseStatusCompleted = "COMPLETED"
	CardPayPaymentResponseStatusCancelled = "CANCELLED"
	CardPayPaymentResponseStatusRefunded = "REFUNDED"
	CardPayPaymentResponseStatusPartiallyRefunded = "PARTIALLY_REFUNDED"
	CardPayPaymentResponseStatusVoided = "VOIDED"
	CardPayPaymentResponseStatusChargedBack = "CHARGED_BACK"
	CardPayPaymentResponseStatusChargebackResolved = "CHARGEBACK_RESOLVED"

	CardPayPaymentResponseHeaderSignature = "Signature"
)

type CardPayBankCardAccount struct {
	Pan        string `json:"pan"`
	HolderName string `json:"holder"`
	Cvv        string `json:"security_code"`
	Expire     string `json:"expiration"`
}

type CardPayEWalletAccount struct {
	Id string `json:"id"`
}

type CardPayPaymentData struct {
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
	Descriptor string  `json:"dynamic_descriptor"`
	Note       string  `json:"note"`
}

type CardPayCustomer struct {
	Email   string `json:"email"`
	Ip      string `json:"ip"`
	Account string `json:"id"`
}

type CardPayItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Count       int     `json:"count"`
	Price       float64 `json:"price"`
}

type CardPayRequest struct {
	Id   string `json:"id"`
	Time string `json:"time"`
}

type CardPayAddress struct {
	Country string `json:"country"`
	City    string `json:"city,omitempty"`
	Phone   string `json:"phone,omitempty"`
	State   string `json:"state,omitempty"`
	Street  string `json:",omitempty"`
	Zip     string `json:"zip,omitempty"`
}

type CardPayMerchantOrder struct {
	Id              string          `json:"id" validate:"required,hexadecimal"`
	Description     string          `json:"description" validate:"required"`
	Items           []*CardPayItem  `json:"items,omitempty"`
	ShippingAddress *CardPayAddress `json:"shipping_address,omitempty"`
}

type CardPayCardAccount struct {
	BillingAddress *CardPayAddress         `json:"billing_address,omitempty"`
	Card           *CardPayBankCardAccount `json:"card"`
	Token          string                  `json:"token,omitempty"`
}

type CardPayCryptoCurrencyAccount struct {
	RollbackAddress string `json:"rollback_address"`
}

type CardPayOrder struct {
	Request               *CardPayRequest               `json:"request"`
	MerchantOrder         *CardPayMerchantOrder         `json:"merchant_order"`
	Description           string                        `json:"description"`
	PaymentMethod         string                        `json:"payment_method"`
	PaymentData           *CardPayPaymentData           `json:"payment_data"`
	CardAccount           *CardPayCardAccount           `json:"card_account,omitempty"`
	Customer              *CardPayCustomer              `json:"customer"`
	EWalletAccount        *CardPayEWalletAccount        `json:"ewallet_account,omitempty"`
	CryptoCurrencyAccount *CardPayCryptoCurrencyAccount `json:"cryptocurrency_account,omitempty"`
}

type CardPayOrderResponse struct {
	RedirectUrl string `json:"redirect_url"`
}

type CardPayPaymentNotificationWebHookRequest struct {
	MerchantOrder         *CardPayMerchantOrder                 `json:"merchant_order" validate:"required"`
	PaymentMethod         string                                `json:"payment_method" validate:"required"`
	CallbackTime          string                                `json:"callback_time" validate:"required"`
	CallbackTimeTime      time.Time                             `json:"-"`
	CardAccount           *CardPayBankCardAccountResponse       `json:"card_account,omitempty" validate:"omitempty"`
	CryptoCurrencyAccount *CardPayCryptoCurrencyAccountResponse `json:"cryptocurrency_account,omitempty" validate:"omitempty"`
	Customer              *CardPayCustomer                      `json:"customer" validate:"required"`
	EWalletAccount        *CardPayEWalletAccount                `json:"ewallet_account,omitempty" validate:"omitempty"`
	PaymentData           *CardPayPaymentDataResponse           `json:"payment_data" validate:"required"`
	Signature             string                                `json:"-"`
}

type CardPayCryptoCurrencyAccountResponse struct {
	CryptoAddress       string  `json:"crypto_address" validate:"required,btc_addr"`
	CryptoTransactionId string  `json:"crypto_transaction_id" validate:"required"`
	PrcAmount           float64 `json:"prc_amount"`
	PrcCurrency         string  `json:"prc_currency"`
}

type CardPayBankCardAccountResponse struct {
	Holder             string `json:"holder" validate:"required"`
	IssuingCountryCode string `json:"issuing_country_code" validate:"required"`
	MaskedPan          string `json:"masked_pan" validate:"required"`
	Token              string `json:"token" validate:"required"`
}

type CardPayPaymentDataResponse struct {
	Id            string  `json:"id" validate:"required"`
	Amount        float64 `json:"amount" validate:"required,numeric"`
	AuthCode      string  `json:"auth_code,omitempty"`
	Created       string  `json:"created" validate:"required"`
	Currency      string  `json:"currency" validate:"required,alpha"`
	DeclineCode   string  `json:"decline_code,omitempty"`
	DeclineReason string  `json:"decline_reason,omitempty"`
	Description   string  `json:"description" validate:"required"`
	Is3d          bool    `json:"is_3d,omitempty"`
	Note          string  `json:"note"`
	Rrn           string  `json:"rrn,omitempty"`
	Status        string  `json:"status" validate:"required,alpha"`
}

func (cpReq *CardPayPaymentNotificationWebHookRequest) IsPaymentAllowedStatus() bool {
	return cpReq.PaymentData.Status == CardPayPaymentResponseStatusCompleted ||
		cpReq.PaymentData.Status == CardPayPaymentResponseStatusDeclined || cpReq.PaymentData.Status == CardPayPaymentResponseStatusCancelled
}

func (cpReq *CardPayPaymentNotificationWebHookRequest) GetBankCardTxnParams() map[string]interface{} {
	params := make(map[string]interface{})

	params[BankCardFieldPan] = cpReq.CardAccount.MaskedPan
	params[BankCardFieldHolder] = cpReq.CardAccount.Holder
	params[TxnParamsFieldBankCardEmissionCountry] = cpReq.CardAccount.IssuingCountryCode
	params[TxnParamsFieldBankCardToken] = cpReq.CardAccount.Token
	params[TxnParamsFieldBankCardIs3DS] = cpReq.PaymentData.Is3d
	params[TxnParamsFieldBankCardRrn] = cpReq.PaymentData.Rrn

	if cpReq.PaymentData.Status == CardPayPaymentResponseStatusDeclined {
		params[TxnParamsFieldDeclineCode] = cpReq.PaymentData.DeclineCode
		params[TxnParamsFieldDeclineReason] = cpReq.PaymentData.DeclineReason
	}

	return params
}

func (cpReq *CardPayPaymentNotificationWebHookRequest) GetEWalletTxnParams() map[string]interface{} {
	params := make(map[string]interface{})

	params[EWalletFieldIdentifier] = cpReq.EWalletAccount.Id

	if cpReq.PaymentData.Status == CardPayPaymentResponseStatusDeclined {
		params[TxnParamsFieldDeclineCode] = cpReq.PaymentData.DeclineCode
		params[TxnParamsFieldDeclineReason] = cpReq.PaymentData.DeclineReason
	}

	return params
}

func (cpReq *CardPayPaymentNotificationWebHookRequest) GetCryptoCurrencyTxnParams() map[string]interface{} {
	params := make(map[string]interface{})

	params[CryptoFieldIdentifier] = cpReq.CryptoCurrencyAccount.CryptoAddress
	params[TxnParamsFieldCryptoTransactionId] = cpReq.CryptoCurrencyAccount.CryptoTransactionId
	params[TxnParamsFieldCryptoAmount] = cpReq.CryptoCurrencyAccount.PrcAmount
	params[TxnParamsFieldCryptoCurrency] = cpReq.CryptoCurrencyAccount.PrcCurrency

	if cpReq.PaymentData.Status == CardPayPaymentResponseStatusDeclined {
		params[TxnParamsFieldDeclineCode] = cpReq.PaymentData.DeclineCode
		params[TxnParamsFieldDeclineReason] = cpReq.PaymentData.DeclineReason
	}

	return params
}
