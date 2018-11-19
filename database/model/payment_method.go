package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

type PaymentMethod struct {
	Id               bson.ObjectId          `bson:"_id" json:"id"`
	Name             string                 `bson:"name" json:"name"`
	GroupAlias       string                 `bson:"group_alias" json:"group_alias"`
	Currency         *Currency              `bson:"currency" json:"currency"`
	MinPaymentAmount float64                `bson:"min_payment_amount" json:"min_payment_amount"`
	MaxPaymentAmount float64                `bson:"max_payment_amount" json:"min_payment_amount"`
	Params           map[string]interface{} `bson:"params" json:"params"`
	IsActive         bool                   `bson:"is_active" json:"is_active"`
	CreatedAt        time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time              `bson:"updated_at" json:"-"`
	PaymentSystem    *PaymentSystem         `bson:"payment_system" json:"payment_system"`
	TerminalId       string                 `bson:"terminal_id" json:"terminal_id"`
}

type OrderPaymentMethod struct {
	Id            bson.ObjectId          `bson:"id"`
	Name          string                 `bson:"name"`
	TerminalId    string                 `bson:"terminal_id"`
	Params        map[string]interface{} `bson:"params"`
	PaymentSystem *PaymentSystem         `bson:"payment_system"`
}
