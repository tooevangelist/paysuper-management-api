package model

import (
	"github.com/globalsign/mgo/bson"
)

type PaymentMethodParams struct {
	Handler    string            `bson:"handler" json:"handler"`
	Terminal   string            `bson:"terminal" json:"terminal"`
	ExternalId string            `bson:"external_id" json:"external_id"`
	Other      map[string]string `bson:"other" json:"other"`
}

type OrderPaymentMethod struct {
	Id            bson.ObjectId        `bson:"id" json:"id"`
	Name          string               `bson:"name" json:"name"`
	Params        *PaymentMethodParams `bson:"params" json:"params"`
	PaymentSystem *PaymentSystem       `bson:"payment_system" json:"payment_system"`
	GroupAlias    string               `bson:"group_alias" json:"group_alias"`
}
