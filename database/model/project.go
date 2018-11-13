package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

type ProjectPaymentModes struct {
	Id      bson.ObjectId `bson:"id" json:"id"`
	AddedAt time.Time     `bson:"added_at" json:"added_at"`
}

type ProjectScalar struct {
	// Name project name
	Name                       string                      `json:"name" validate:"required,min=1,max=255"`
	// CallbackCurrency ISO 4217 numeric currency code to send payment notification
	CallbackCurrency           *int                        `json:"callback_currency,omitempty" validate:"omitempty,numeric"`
	// CallbackProtocol protocol identifier to send payment notification. Now available: default
	CallbackProtocol           string                      `json:"callback_protocol" validate:"required,oneof=default"`
	// CreateInvoiceAllowedUrls list of urls rom which you can send a request to create an order
	CreateInvoiceAllowedUrls   []string                    `json:"create_invoice_allowed_urls" validate:"unique"`
	// IsAllowDynamicNotifyUrls is allow send dynamic notification urls in request to create an order
	IsAllowDynamicNotifyUrls   bool                        `json:"is_allow_dynamic_notify_urls"`
	// IsAllowDynamicRedirectUrls is allow send dynamic user's redirect urls in request to create an order
	IsAllowDynamicRedirectUrls bool                        `json:"is_allow_dynamic_redirect_urls"`
	// LimitsCurrency ISO 4217 numeric currency code for limit amounts
	LimitsCurrency             *int                        `json:"limits_currency,omitempty" validate:"omitempty,numeric"`
	// MaxPaymentAmount maximal amount allowed for create order
	MaxPaymentAmount           *float64                    `json:"max_payment_amount,omitempty" validate:"omitempty,numeric,min=0,max=999999"`
	// MinPaymentAmount minimal amount allowed for create order
	MinPaymentAmount           *float64                    `json:"min_payment_amount,omitempty" validate:"omitempty,numeric,min=0,max=999999"`
	// NotifyEmails list of emails to send notifications about successfully completed payment operations
	NotifyEmails               []string                    `json:"notify_emails" validate:"unique"`
	// OnlyFixedAmounts is allow create order only with amounts from fixed packages list
	OnlyFixedAmounts           bool                        `json:"only_fixed_amounts"`
	// SecretKey secret key for create check hash for request about order statuses changes
	SecretKey                  string                      `json:"secret_key" validate:"max=255"`
	// SendNotifyEmail is allow send notifications about successfully completed payment operations to user's emails
	SendNotifyEmail            bool                        `json:"send_notify_email"`
	// URLCheckAccount default url to send request for verification payment data to project
	URLCheckAccount            *string                     `json:"url_check_account,omitempty" validate:"omitempty,url,max=255"`
	// URLProcessPayment default url to send request for notification about successfully completed payment to project
	URLProcessPayment          *string                     `json:"url_process_payment,omitempty" validate:"omitempty,url,max=255"`
	// URLRedirectFail default url to redirect user after failed payment
	URLRedirectFail            *string                     `json:"url_redirect_fail,omitempty" validate:"omitempty,url,max=255"`
	// URLRedirectSuccess default url to redirect user after successfully completed payment
	URLRedirectSuccess         *string                     `json:"url_redirect_success,omitempty" validate:"omitempty,url,max=255"`
	// IsActive is project active
	IsActive                   bool                        `json:"is_active,omitempty"`
	// FixedPackage list of project's fixed packages
	FixedPackage               map[string][]*FixedPackage `json:"fixed_package,omitempty"`

	Merchant *Merchant `json:"-"`
}

type Project struct {
	// Id unique project identifier
	Id                         bson.ObjectId                     `bson:"_id" json:"id"`
	// CallbackCurrency full object of currency which described currency to send payment notification
	CallbackCurrency           *Currency                         `bson:"callback_currency" json:"callback_currency"`
	// CallbackProtocol protocol identifier to send payment notification. Now available: default
	CallbackProtocol           string                            `bson:"callback_protocol" json:"callback_protocol"`
	// CreateInvoiceAllowedUrls list of urls rom which you can send a request to create an order
	CreateInvoiceAllowedUrls   []string                          `bson:"create_invoice_allowed_urls" json:"create_invoice_allowed_urls"`
	// Merchant full object of merchant which describes project's owner
	Merchant                   *Merchant                         `bson:"merchant" json:"-"`
	// IsAllowDynamicNotifyUrls is allow send dynamic notification urls in request to create an order
	IsAllowDynamicNotifyUrls   bool                              `bson:"is_allow_dynamic_notify_urls" json:"is_allow_dynamic_notify_urls"`
	// IsAllowDynamicRedirectUrls is allow send dynamic user's redirect urls in request to create an order
	IsAllowDynamicRedirectUrls bool                              `bson:"is_allow_dynamic_redirect_urls" json:"is_allow_dynamic_redirect_urls"`
	// LimitsCurrency full object of currency which describes currency for amount's limit
	LimitsCurrency             *Currency                         `bson:"limits_currency" json:"limits_currency"`
	// MaxPaymentAmount maximal amount allowed for create order
	MaxPaymentAmount           float64                           `bson:"max_payment_amount" json:"max_payment_amount"`
	// MinPaymentAmount minimal amount allowed for create order
	MinPaymentAmount           float64                           `bson:"min_payment_amount" json:"min_payment_amount"`
	// Name project name
	Name                       string                            `bson:"name" json:"name"`
	// NotifyEmails list of emails to send notifications about successfully completed payment operations
	NotifyEmails               []string                          `bson:"notify_emails" json:"notify_emails"`
	// OnlyFixedAmounts is allow create order only with amounts from fixed packages list
	OnlyFixedAmounts           bool                              `bson:"only_fixed_amounts" json:"only_fixed_amounts"`
	// SecretKey secret key for create check hash for request about order statuses changes
	SecretKey                  string                            `bson:"secret_key" json:"secret_key"`
	// SendNotifyEmail is allow send notifications about successfully completed payment operations to user's emails
	SendNotifyEmail            bool                              `bson:"send_notify_email" json:"send_notify_email"`
	// URLCheckAccount default url to send request for verification payment data to project
	URLCheckAccount            *string                           `bson:"url_check_account" json:"url_check_account"`
	// URLProcessPayment default url to send request for notification about successfully completed payment to project
	URLProcessPayment          *string                           `bson:"url_process_payment" json:"url_process_payment"`
	// URLRedirectFail default url to redirect user after failed payment
	URLRedirectFail            *string                           `bson:"url_redirect_fail" json:"url_redirect_fail"`
	// URLRedirectSuccess default url to redirect user after successfully completed payment
	URLRedirectSuccess         *string                           `bson:"url_redirect_success" json:"url_redirect_success"`
	// IsActive is project active
	IsActive                   bool                              `bson:"is_active" json:"is_active"`
	// CreatedAt date of create project
	CreatedAt                  time.Time                         `bson:"created_at" json:"created_at"`
	UpdatedAt                  time.Time                         `bson:"updated_at" json:"-"`
	// FixedPackage list of project's fixed packages
	FixedPackage               map[string][]*FixedPackage        `bson:"fixed_package" json:"fixed_package,omitempty"`
	// FixedPackage list of payment methods allowed to project
	PaymentMethods             map[string][]*ProjectPaymentModes `bson:"payment_methods" json:"payment_methods,omitempty"`
}
