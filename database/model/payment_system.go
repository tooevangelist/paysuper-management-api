package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

type PaymentSystem struct {
	Id                 bson.ObjectId `bson:"_id" json:"id"`
	Name               string        `bson:"name" json:"name"`
	Country            string        `bson:"country" json:"country"`
	AccountingCurrency *Currency     `bson:"accounting_currency" json:"accounting_currency"`
	AccountingPeriod   string        `bson:"accounting_period" json:"accounting_period"`
	IsActive           bool          `bson:"is_active" json:"is_active"`
	CreatedAt          time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time     `bson:"updated_at" json:"-"`
}
