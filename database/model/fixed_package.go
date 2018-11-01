package model

import "time"

type FixedPackage struct {
	Name        string    `bson:"name" json:"name" validate:"required,url,max=255"`
	CurrencyInt int       `bson:"currency_int" json:"currency_int" validate:"required,numeric"`
	Price       float64   `bson:"price" json:"price" validate:"required,numeric,min=0,max=100000"`
	IsActive    bool      `bson:"is_active" json:"is_active" validate:"required"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"-"`
	Currency    *Currency `json:"currency"`
}