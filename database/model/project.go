package model

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type ProjectScalar struct {
	Name                       string   `json:"name" validate:"required,min=1,max=255"`
	CallbackCurrency           *int     `json:"callback_currency,omitempty" validate:"omitempty,numeric"`
	CallbackProtocol           string   `json:"callback_protocol" validate:"oneof=default"`
	CreateInvoiceAllowedUrls   []string `json:"create_invoice_allowed_urls" validate:"unique"`
	IsAllowDynamicNotifyUrls   bool     `json:"is_allow_dynamic_notify_urls"`
	IsAllowDynamicRedirectUrls bool     `json:"is_allow_dynamic_redirect_urls"`
	LimitsCurrency             *int     `json:"limits_currency,omitempty" validate:"omitempty,numeric"`
	MaxPaymentAmount           *float64 `json:"max_payment_amount,omitempty" validate:"omitempty,numeric,min=0,max=15000"`
	MinPaymentAmount           *float64 `json:"min_payment_amount,omitempty" validate:"omitempty,numeric,min=0,max=15000"`
	NotifyEmails               []string `json:"notify_emails" validate:"unique"`
	OnlyFixedAmounts           bool     `json:"only_fixed_amounts"`
	SecretKey                  string   `json:"secret_key" validate:"required,min=1,max=255"`
	SendNotifyEmail            bool     `json:"send_notify_email"`
	URLCheckAccount            *string  `json:"url_check_account,omitempty" validate:"omitempty,url,max=255"`
	URLProcessPayment          *string  `json:"url_process_payment,omitempty" validate:"omitempty,url,max=255"`
	URLRedirectFail            *string  `json:"url_redirect_fail,omitempty" validate:"omitempty,url,max=255"`
	URLRedirectSuccess         *string  `json:"url_redirect_success,omitempty" validate:"omitempty,url,max=255"`
	IsActive                   bool     `json:"is_active,omitempty"`

	Merchant *Merchant `json:"-"`
}

type Project struct {
	Id                         bson.ObjectId `bson:"_id" json:"id"`
	CallbackCurrency           *Currency     `bson:"callback_currency" json:"callback_currency"`
	CallbackProtocol           string        `bson:"callback_protocol" json:"callback_protocol"`
	CreateInvoiceAllowedUrls   []string      `bson:"create_invoice_allowed_urls" json:"create_invoice_allowed_urls"`
	Merchant                   *Merchant     `bson:"merchant" json:"-"`
	IsAllowDynamicNotifyUrls   bool          `bson:"is_allow_dynamic_notify_urls" json:"is_allow_dynamic_notify_urls"`
	IsAllowDynamicRedirectUrls bool          `bson:"is_allow_dynamic_redirect_urls" json:"is_allow_dynamic_redirect_urls"`
	LimitsCurrency             *Currency     `bson:"limits_currency" json:"limits_currency"`
	MaxPaymentAmount           float64       `bson:"max_payment_amount" json:"max_payment_amount"`
	MinPaymentAmount           float64       `bson:"min_payment_amount" json:"min_payment_amount"`
	Name                       string        `bson:"name" json:"name"`
	NotifyEmails               []string      `bson:"notify_emails" json:"notify_emails"`
	OnlyFixedAmounts           bool          `bson:"only_fixed_amounts" json:"only_fixed_amounts"`
	SecretKey                  string        `bson:"secret_key" json:"secret_key"`
	SendNotifyEmail            bool          `bson:"send_notify_email" json:"send_notify_email"`
	URLCheckAccount            *string       `bson:"url_check_account" json:"url_check_account"`
	URLProcessPayment          *string       `bson:"url_process_payment" json:"url_process_payment"`
	URLRedirectFail            *string       `bson:"url_redirect_fail" json:"url_redirect_fail"`
	URLRedirectSuccess         *string       `bson:"url_redirect_success" json:"url_redirect_success"`
	IsActive                   bool          `bson:"is_active" json:"is_active"`
	CreatedAt                  time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt                  time.Time     `bson:"updated_at" json:"-"`
}
