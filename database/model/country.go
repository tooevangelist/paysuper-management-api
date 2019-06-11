package model

import (
	"time"
)

type Country struct {
	Id              string    `json:"id" bson:"_id"`
	IsoCodeA2       string    `json:"iso_code_a2" bson:"iso_code_a2"`
	Region          string    `json:"region" bson:"region"`
	Currency        string    `json:"currency" bson:"currency"`
	PaymentsAllowed bool      `json:"payments_allowed" bson:"payments_allowed"`
	ChangeAllowed   bool      `json:"change_allowed" bson:"change_allowed"`
	VatEnabled      bool      `json:"vat_enabled" bson:"vat_enabled"`
	VatCurrency     string    `json:"vat_currency" bson:"vat_currency"`
	PriceGroupId    string    `json:"price_group_id" bson:"price_group_id"`
	CreatedAt       time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" bson:"updated_at"`
}

type CountryItems struct {
	Count int        `json:"count"`
	Items []*Country `json:"items"`
}
