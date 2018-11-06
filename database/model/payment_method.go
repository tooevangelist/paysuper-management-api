package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

type PaymentMethod struct {
	Id               bson.ObjectId          `bson:"_id" json:"id"`
	Name             string                 `bson:"name" json:"name"`
	PaymentSystemId  bson.ObjectId          `bson:"payment_system_id" json:"payment_system_id"`
	GroupAlias       string                 `bson:"group_alias" json:"group_alias"`
	Currency         *Currency              `bson:"currency" json:"currency"`
	MinPaymentAmount float64                `bson:"min_payment_amount" json:"min_payment_amount"`
	MaxPaymentAmount float64                `bson:"max_payment_amount" json:"min_payment_amount"`
	Params           map[string]interface{} `bson:"params" json:"params"`
	IsActive         bool                   `bson:"is_active" json:"is_active"`
	CreatedAt        time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time              `bson:"updated_at" json:"-"`
}
