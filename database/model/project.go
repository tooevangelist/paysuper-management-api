package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
)

type ProjectScalar struct {
	// project name
	Name string `json:"name" validate:"required,min=1,max=255"`
	// ISO 4217 numeric currency code to send payment notification
	CallbackCurrency *int `json:"callback_currency,omitempty" validate:"omitempty,numeric"`
	// protocol identifier to send payment notification. Now available: default
	CallbackProtocol string `json:"callback_protocol" validate:"required,oneof=default empty"`
	// list of urls rom which you can send a request to create an order
	CreateInvoiceAllowedUrls []string `json:"create_invoice_allowed_urls" validate:"unique"`
	// is allow send dynamic notification urls in request to create an order
	IsAllowDynamicNotifyUrls bool `json:"is_allow_dynamic_notify_urls"`
	// is allow send dynamic user's redirect urls in request to create an order
	IsAllowDynamicRedirectUrls bool `json:"is_allow_dynamic_redirect_urls"`
	// ISO 4217 numeric currency code for limit amounts
	LimitsCurrency *int `json:"limits_currency,omitempty" validate:"omitempty,numeric"`
	// maximal amount allowed for create order
	MaxPaymentAmount *float64 `json:"max_payment_amount,omitempty" validate:"omitempty,numeric,min=0,max=999999"`
	// minimal amount allowed for create order
	MinPaymentAmount *float64 `json:"min_payment_amount,omitempty" validate:"omitempty,numeric,min=0,max=999999"`
	// list of emails to send notifications about successfully completed payment operations
	NotifyEmails []string `json:"notify_emails" validate:"unique"`
	// is allow create order only with amounts from fixed packages list
	IsProductsCheckout bool `json:"is_products_checkout"`
	// secret key for create check hash for request about order statuses changes
	SecretKey string `json:"secret_key" validate:"max=255"`
	// is allow send notifications about successfully completed payment operations to user's emails
	SendNotifyEmail bool `json:"send_notify_email"`
	// default url to send request for verification payment data to project
	URLCheckAccount string `json:"url_check_account,omitempty" validate:"omitempty,url,max=255"`
	// default url to send request for notification about successfully completed payment to project
	URLProcessPayment string `json:"url_process_payment,omitempty" validate:"omitempty,url,max=255"`
	// default url to redirect user after failed payment
	URLRedirectFail string `json:"url_redirect_fail,omitempty" validate:"omitempty,url,max=255"`
	// default url to redirect user after successfully completed payment
	URLRedirectSuccess string `json:"url_redirect_success,omitempty" validate:"omitempty,url,max=255"`
	// is project active
	IsActive bool `json:"is_active,omitempty"`
	// list of project's fixed packages
	FixedPackage map[string][]*FixedPackage `json:"fixed_package,omitempty"`

	Merchant *billing.Merchant `json:"-"`
}

type ProjectOrder struct {
	Id                bson.ObjectId     `bson:"id" json:"id"`
	Name              string            `bson:"name" json:"name"`
	UrlSuccess        string            `bson:"url_success" json:"url_success"`
	UrlFail           string            `bson:"url_fail" json:"url_fail"`
	NotifyEmails      []string          `bson:"notify_emails" json:"notify_emails"`
	SecretKey         string            `bson:"secret_key" json:"secret_key"`
	SendNotifyEmail   bool              `bson:"send_notify_email" json:"send_notify_email"`
	URLCheckAccount   string            `bson:"url_check_account" json:"url_check_account"`
	URLProcessPayment string            `bson:"url_process_payment" json:"url_process_payment"`
	CallbackProtocol  string            `bson:"callback_protocol" json:"callback_protocol"`
	Merchant          *billing.Merchant `bson:"merchant" json:"merchant"`
}
