package model

import "time"

type FixedPackage struct {
	// Name package name
	Name        string    `bson:"name" json:"name" validate:"required,url,max=255"`
	// CurrencyInt numeric ISO 4217 currency code to package price
	CurrencyInt int       `bson:"currency_int" json:"currency_int" validate:"required,numeric"`
	// Price package price in chosen currency
	Price       float64   `bson:"price" json:"price" validate:"required,numeric,min=0,max=100000"`
	// IsActive is package active
	IsActive    bool      `bson:"is_active" json:"is_active" validate:"required"`
	// CreatedAt date of create package
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"-"`
	// Currency full object of currency to package price
	Currency    *Currency `json:"currency"`
}