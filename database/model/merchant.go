package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

const (
	MerchantStatusCreated   = 0
	MerchantStatusCompleted = 1
	MerchantStatusActive    = 2
	MerchantStatusDeleted   = 3
)

type MerchantScalar struct {
	// unique merchant identifier in auth system
	Id string `json:"id" validate:"required"`
	// merchant email
	Email *string `json:"email" validate:"omitempty,email"`
	// merchant company legal name
	Name *string `json:"name" validate:"omitempty,min=3,max=255"`
	// numeric ISO 3166-1 country code
	Country *int `json:"country" validate:"omitempty,numeric"`
	// period payout of money to merchant bank account. Now available next values: day - every day, week - every week, 2week - every two week, month - every month, quarter - every quarter, half-year - every half-year, year - every year
	AccountingPeriod *string `json:"accounting_period" validate:"omitempty,oneof=day week 2week month quarter half-year year"`
	// numeric ISO 4217 currency code describes merchant's accounting currency
	Currency *int `json:"currency" validate:"omitempty,numeric"`
	// merchant status in system. Now available next statuses: 0 - created, 1 - verified, 2 - active, 3 - deleted
	Status *int `bson:"status" json:"status" validate:"omitempty,numeric"`
}

type Merchant struct {
	// unique merchant identifier
	Id bson.ObjectId `bson:"_id" json:"id"`
	// unique merchant identifier in auth system
	ExternalId string `bson:"external_id" json:"external_id"`
	// merchant email
	Email string `bson:"email" json:"email"`
	// merchant company legal name
	Name *string `bson:"name" json:"name"`
	// full object of country where merchant company registered
	Country *Country `bson:"country" json:"country"`
	// period payout of money to merchant bank account. Now available next values: day - every day, week - every week, 2week - every two week, month - every month, quarter - every quarter, half-year - every half-year, year - every year
	AccountingPeriod *string `bson:"accounting_period" json:"accounting_period"`
	// full object describes merchant's accounting currency
	Currency *Currency `bson:"currency" json:"currency"`
	// vat calculation enabled
	IsVatEnabled bool `bson:"is_vat_enabled" json:"is_vat_enabled"`
	// enable to add commission payment method and commission PSP (P1) to payment amount
	IsCommissionToUserEnabled bool `bson:"is_commission_to_user_enabled" json:"is_commission_to_user_enabled"`
	// merchant status in system. Now available next statuses: 0 - created, 1 - verified, 2 - active, 3 - deleted
	Status int `bson:"status" json:"status"`
	// date of create merchant in system
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"-"`
}
