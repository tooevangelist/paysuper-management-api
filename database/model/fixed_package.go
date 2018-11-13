package model

import "time"

type FixedPackage struct {
	// package name
	Name        string    `bson:"name" json:"name" validate:"required,url,max=255"`
	// numeric ISO 4217 currency code to package price
	CurrencyInt int       `bson:"currency_int" json:"currency_int" validate:"required,numeric"`
	// package price in chosen currency
	Price       float64   `bson:"price" json:"price" validate:"required,numeric,min=0,max=100000"`
	// is package active
	IsActive    bool      `bson:"is_active" json:"is_active" validate:"required"`
	// date of create package
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"-"`
	// full object of currency to package price
	Currency    *Currency `json:"currency"`
}