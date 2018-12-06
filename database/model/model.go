package model

import "github.com/globalsign/mgo/bson"

const (
	DefaultLimit  = 100
	DefaultOffset = 0

	ResponseMessageInvalidRequestData = "Invalid request data"
	ResponseMessageAccessDenied       = "Access denied"
	ResponseMessageNotFound           = "Not found"
	ResponseMessageProjectIdIsInvalid = "one or more project identifier is invalid"

	ApiRequestParameterProjectId = "project_id"
	ApiRequestParameterRegion    = "region"

	DBFieldId          = "id"
	DBFieldName        = "name"
	DBFieldPrice       = "price"
	DBFieldCurrencyInt = "currency_int"
	DBFieldCurrency    = "currency"
)

type Error struct {
	// text error description
	Message string `json:"message"`
}

type SimpleItem struct {
	// unique identifier of item
	Id bson.ObjectId `json:"id"`
	// item name
	Name string `json:"name"`
}

type Status struct {
	// numeric status code
	Status int `json:"status"`
	// status name
	Name string `json:"name"`
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

type SimpleCountry struct {
	CodeInt int    `bson:"code_int" json:"code_int"`
	CodeA2  string `bson:"code_a2" json:"code_a2"`
	CodeA3  string `bson:"code_a3" json:"code_a3"`
	Name    *Name  `bson:"code_name" json:"name"`
}

type OrderSimpleAmountObject struct {
	// amount value
	Amount float64 `json:"amount"`
	// object which contains main information about currency
	Currency *SimpleCurrency `json:"currency"`
}
