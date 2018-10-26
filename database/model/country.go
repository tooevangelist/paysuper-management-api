package model

import "time"

type Country struct {
	CodeInt   int       `bson:"code_int" json:"code_int"`
	CodeA2    string    `bson:"code_a2" json:"code_a2"`
	CodeA3    string    `bson:"code_a3" json:"code_a3"`
	Name      *Name     `bson:"name" json:"name"`
	IsActive  bool      `bson:"is_active" json:"is_active"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"-"`
}
