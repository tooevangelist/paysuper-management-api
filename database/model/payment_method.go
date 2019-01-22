package model

import (
	"github.com/ProtocolONE/payone-repository/pkg/proto/billing"
	"github.com/globalsign/mgo/bson"
	"time"
)

const (
	PaymentMethodTypeBankCard = "bank_card"
	PaymentMethodTypeEWallet  = "ewallet"
	PaymentMethodTypeCrypto   = "crypto"
)

type PaymentMethodParams struct {
	Handler    string            `bson:"handler" json:"handler"`
	Terminal   string            `bson:"terminal" json:"terminal"`
	ExternalId string            `bson:"external_id" json:"external_id"`
	Other      map[string]string `bson:"other" json:"other"`
}

type PaymentMethod struct {
	Id               bson.ObjectId        `bson:"_id" json:"id"`
	Name             string               `bson:"name" json:"name"`
	GroupAlias       string               `bson:"group_alias" json:"group_alias"`
	Currency         *Currency            `bson:"currency" json:"currency"`
	MinPaymentAmount float64              `bson:"min_payment_amount" json:"min_payment_amount"`
	MaxPaymentAmount float64              `bson:"max_payment_amount" json:"min_payment_amount"`
	Params           *PaymentMethodParams `bson:"params" json:"params"`
	Icon             string               `bson:"icon" json:"icon"`
	IsActive         bool                 `bson:"is_active" json:"is_active"`
	CreatedAt        time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time            `bson:"updated_at" json:"-"`
	PaymentSystem    *PaymentSystem       `bson:"payment_system" json:"payment_system"`
	Currencies       []int32              `bson:"currencies" json:"currencies"`
	// type of payment method. allowed at current time: bank_card, ewallet, crypto
	Type string `bson:"type" json:"type"`
	// regexp mask for check main requisite of payment method
	AccountRegexp string `bson:"account_regexp" json:"account_regexp"`
}

type OrderPaymentMethod struct {
	Id            bson.ObjectId        `bson:"id" json:"id"`
	Name          string               `bson:"name" json:"name"`
	Params        *PaymentMethodParams `bson:"params" json:"params"`
	PaymentSystem *PaymentSystem       `bson:"payment_system" json:"payment_system"`
	GroupAlias    string               `bson:"group_alias" json:"group_alias"`
}

// Contain data about payment methods to render payment form from client library
type PaymentMethodJsonOrderResponse struct {
	// payment method unique identifier
	Id string `json:"id"`
	// payment method name
	Name string `json:"name"`
	// url to payment method icon
	Icon string `json:"icon"`
	// payment method type. allowed: bank_card, ewallet, crypto
	Type string `json:"type"`
	// payment method group alias
	GroupAlias string `json:"group_alias"`
	// regexp mask for check main requisite of payment method
	AccountRegexp string `json:"account_regexp"`
	// total amount to payment in payment method currency include all possible commissions
	AmountWithCommissions float64 `json:"amount_with_commissions"`
	// original price of item in currency of payment method  which buy user
	AmountWithoutCommissions float64 `json:"amount_without_commissions"`
	// 3 symbols currency code of payment method by  ISO 4217
	Currency string `json:"currency"`
	// amount of commission in payment method currency, which the user pays for the project
	UserCommissionAmount float64 `json:"user_commission_amount"`
	// amount of VAT in payment method currency, which the user pays
	VatAmount     float64              `json:"vat_amount"`
	HasSavedCards *bool                `json:"has_saved_cards,omitempty"`
	SavedCards    []*SavedCardResponse `json:"saved_cards,omitempty"`
}

type SavedCardResponse struct {
	Id     string              `json:"id"`
	Pan    string              `json:"pan"`
	Expire *billing.CardExpire `json:"expire"`
}

// Temporary struct for save backward compatibility for self hosted payment form
type PaymentMethodJsonOrderResponseOrderFormRendering struct {
	*PaymentMethodJsonOrderResponse
	GroupAlias string
}

// Temporary struct for save backward compatibility for self hosted payment form
type OrderFormRendering struct {
	SlicePaymentMethodJsonOrderResponse []*PaymentMethodJsonOrderResponse
	MapPaymentMethodJsonOrderResponse   []*PaymentMethodJsonOrderResponseOrderFormRendering
}
