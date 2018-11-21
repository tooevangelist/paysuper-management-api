package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
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
}

type OrderPaymentMethod struct {
	Id            bson.ObjectId        `bson:"id" json:"id"`
	Name          string               `bson:"name" json:"name"`
	Params        *PaymentMethodParams `bson:"params" json:"params"`
	PaymentSystem *PaymentSystem       `bson:"payment_system" json:"payment_system"`
	GroupAlias    string               `bson:"group_alias" json:"group_alias"`
}
