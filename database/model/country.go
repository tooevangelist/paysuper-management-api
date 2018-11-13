package model

import "time"

type Country struct {
	// numeric ISO 3166-1 country code
	CodeInt   int       `bson:"code_int" json:"code_int"`
	// 2 chars ISO 3166-1 country code
	CodeA2    string    `bson:"code_a2" json:"code_a2"`
	// 3 chars ISO 3166-1 country code
	CodeA3    string    `bson:"code_a3" json:"code_a3"`
	// list of country names
	Name      *Name     `bson:"name" json:"name"`
	// is country active
	IsActive  bool      `bson:"is_active" json:"is_active"`
	// date of create country in system
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"-"`
}
