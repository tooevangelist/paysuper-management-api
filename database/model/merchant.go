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
	Id               string  `json:"id" validate:"required"`
	Email            *string `json:"email" validate:"omitempty,email"`
	Name             *string `json:"name" validate:"omitempty,min=3,max=255"`
	Country          *int    `json:"country" validate:"omitempty,numeric"`
	AccountingPeriod *string `json:"accounting_period" validate:"omitempty,oneof=day week 2week month quarter half-year year"`
	Currency         *int    `json:"currency" validate:"omitempty,numeric"`
	Status           *int    `bson:"status" json:"status" validate:"omitempty,numeric"`
}

type Merchant struct {
	Id               bson.ObjectId `bson:"_id" json:"id"`
	ExternalId       string        `bson:"external_id" json:"external_id"`
	Email            string        `bson:"email" json:"email"`
	Name             *string       `bson:"name" json:"name"`
	Country          *Country      `bson:"country" json:"country"`
	AccountingPeriod *string       `bson:"accounting_period" json:"accounting_period"`
	Currency         *Currency     `bson:"currency" json:"currency"`
	Status           int           `bson:"status" json:"status"`
	CreatedAt        time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time     `bson:"updated_at" json:"-"`
}
