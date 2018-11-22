package entity

const (
	CardPayPaymentResponseStatusApproved           = "APPROVED"
	CardPayPaymentResponseStatusDeclined           = "DECLINED"
	CardPayPaymentResponseStatusPending            = "PENDING"
	CardPayPaymentResponseStatusVoided             = "VOIDED"
	CardPayPaymentResponseStatusRefunded           = "REFUNDED"
	CardPayPaymentResponseStatusChargeBack         = "CHARGEBACK"
	CardPayPaymentResponseStatusChargeBackResolved = "CHARGEBACK RESOLVED"
)

var CardPayPaymentResponseStatusDescription = map[string]string{
	CardPayPaymentResponseStatusApproved:           "Transaction successfully completed",
	CardPayPaymentResponseStatusDeclined:           "Transaction denied",
	CardPayPaymentResponseStatusPending:            "Transaction successfully authorized, but needs some time to be verified",
	CardPayPaymentResponseStatusVoided:             "Transaction was voided",
	CardPayPaymentResponseStatusRefunded:           "Transaction was refunded",
	CardPayPaymentResponseStatusChargeBack:         "Customer's chargeback claim was received",
	CardPayPaymentResponseStatusChargeBackResolved: "Customer's claim was rejected, same as APPROVED",
}

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
	CardAccount           *CardPayBankCardAccountResponse       `json:"card_account,omitempty" validate:"omitempty"`
	CryptoCurrencyAccount *CardPayCryptoCurrencyAccountResponse `json:"cryptocurrency_account,omitempty" validate:"omitempty"`
	Customer              *CardPayCustomer                      `json:"customer" validate:"required"`
	EWalletAccount        *CardPayEWalletAccount                `json:"ewallet_account,omitempty" validate:"omitempty"`
	PaymentData           *CardPayPaymentDataResponse           `json:"payment_data" validate:"required"`
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
