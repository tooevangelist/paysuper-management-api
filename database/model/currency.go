package model

import "time"

type Name struct {
	// english name
	EN string `bson:"en" json:"en"`
	// russian name
	RU string `bson:"ru" json:"ru"`
}

type Currency struct {
	// numeric ISO 4217 currency code
	CodeInt int `bson:"code_int" json:"code_int"`
	// 3 chars ISO 4217 currency code
	CodeA3 string `bson:"code_a3" json:"code_a3"`
	// list of currency names
	Name *Name `bson:"name" json:"name"`
	// is currency active
	IsActive bool `bson:"is_active" json:"is_active"`
	// date of create currency in system8
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"-"`
}
