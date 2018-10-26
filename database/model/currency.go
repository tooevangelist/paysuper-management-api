package model

import "time"

type Name struct {
	EN string `bson:"en" json:"en"`
	RU string `bson:"ru" json:"ru"`
}

type Currency struct {
	CodeInt   int       `bson:"code_int" json:"code_int"`
	CodeA3    string    `bson:"code_a3" json:"code_a3"`
	Name      *Name     `bson:"name" json:"name"`
	IsActive  bool      `bson:"is_active" json:"is_active"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"-"`
}
