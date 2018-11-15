package model

import "github.com/globalsign/mgo/bson"

const (
	DefaultLimit  = 100
	DefaultOffset = 0
)

type Error struct {
	// text error description
	Message string `json:"message"`
}

type Paginate struct {
	// total count of selected items
	Count int `json:"count"`
	// array of selected items
	Items []interface{} `json:"items"`
}

type SimpleItem struct {
	// unique identifier of item
	Id bson.ObjectId `json:"id"`
	// item name
	Name string `json:"name"`
}

type Status struct {
	// numeric status code
	Status      int    `json:"status"`
	// text description
	Description string `json:"description"`
}

type SimpleCurrency struct {
	// numeric ISO 4217 currency code
	CodeInt int `json:"code_int"`
	// 3 chars ISO 4217 currency code
	CodeA3 string `json:"code_a3"`
	// list of currency names
	Name *Name `json:"name"`
}

type OrderSimpleAmountObject struct {
	// amount value
	Amount float64 `json:"amount"`
	// object which contains main information about currency
	Currency *SimpleCurrency `json:"currency"`
}
