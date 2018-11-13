package model

import "time"

type Name struct {
	// EN english name
	EN string `bson:"en" json:"en"`
	// EN russian name
	RU string `bson:"ru" json:"ru"`
}

type Currency struct {
	// CodeInt numeric ISO 4217 currency code
	CodeInt   int       `bson:"code_int" json:"code_int"`
	// CodeA3 3 chars ISO 4217 currency code
	CodeA3    string    `bson:"code_a3" json:"code_a3"`
	// Name list of currency names
	Name      *Name     `bson:"name" json:"name"`
	// IsActive is currency active
	IsActive  bool      `bson:"is_active" json:"is_active"`
	// CreatedAt date of create currency in system8
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"-"`
}
