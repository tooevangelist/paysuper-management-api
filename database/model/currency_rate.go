package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

type CurrencyRate struct {
	Id           bson.ObjectId `bson:"_id"`
	CurrencyFrom int           `bson:"currency_from"`
	CurrencyTo   int           `bson:"currency_to"`
	Rate         float64       `bson:"rate"`
	Date         time.Time     `bson:"date"`
	CreatedAt    time.Time     `bson:"created_at"`
}
