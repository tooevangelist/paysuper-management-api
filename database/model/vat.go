package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

var VatBySubdivisionCountries = map[string]bool{"US": true, "CA": true}

type Vat struct {
	Id              bson.ObjectId  `bson:"_id" json:"id"`
	Country         *SimpleCountry `bson:"country" json:"country"`
	SubdivisionCode string         `bson:"subdivision_code" json:"subdivision_code,omitempty"`
	Vat             float64        `bson:"vat" json:"vat"`
	CreatedAt       time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt       *time.Time     `bson:"updated_at" json:"updated_at,omitempty"`
}
