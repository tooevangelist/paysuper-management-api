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
	// Id unique merchant identifier in auth system
	Id               string  `json:"id" validate:"required"`
	// Email merchant email
	Email            *string `json:"email" validate:"omitempty,email"`
	// Name merchant company legal name
	Name             *string `json:"name" validate:"omitempty,min=3,max=255"`
	// Country numeric ISO 3166-1 country code
	Country          *int    `json:"country" validate:"omitempty,numeric"`
	// AccountingPeriod period payout of money to merchant bank account. Now available next values: day - every day, week - every week, 2week - every two week, month - every month, quarter - every quarter, half-year - every half-year, year - every year
	AccountingPeriod *string `json:"accounting_period" validate:"omitempty,oneof=day week 2week month quarter half-year year"`
	// Currency numeric ISO 4217 currency code describes merchant's accounting currency
	Currency         *int    `json:"currency" validate:"omitempty,numeric"`
	// Status merchant status in system. Now available next statuses: 0 - created, 1 - verified, 2 - active, 3 - deleted
	Status           *int    `bson:"status" json:"status" validate:"omitempty,numeric"`
}

type Merchant struct {
	// Id unique merchant identifier
	Id               bson.ObjectId `bson:"_id" json:"id"`
	// ExternalId unique merchant identifier in auth system
	ExternalId       string        `bson:"external_id" json:"external_id"`
	// Email merchant email
	Email            string        `bson:"email" json:"email"`
	// Name merchant company legal name
	Name             *string       `bson:"name" json:"name"`
	// Country full object of country where merchant company registered
	Country          *Country      `bson:"country" json:"country"`
	// AccountingPeriod period payout of money to merchant bank account. Now available next values: day - every day, week - every week, 2week - every two week, month - every month, quarter - every quarter, half-year - every half-year, year - every year
	AccountingPeriod *string       `bson:"accounting_period" json:"accounting_period"`
	// Country full object describes merchant's accounting currency
	Currency         *Currency     `bson:"currency" json:"currency"`
	// Status merchant status in system. Now available next statuses: 0 - created, 1 - verified, 2 - active, 3 - deleted
	Status           int           `bson:"status" json:"status"`
	// CreatedAt date of create merchant in system
	CreatedAt        time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time     `bson:"updated_at" json:"-"`
}
