package model

import "github.com/globalsign/mgo/bson"

const (
	DefaultLimit  = 100
	DefaultOffset = 0

	EmptyString = ""

	QueryParameterNameLimit  = "limit"
	QueryParameterNameOffset = "offset"
	QueryParameterNameSort   = "sort[]"

	ResponseMessageInvalidRequestData = "Invalid request data"
	ResponseMessageAccessDenied       = "Access denied"
	ResponseMessageNotFound           = "Not found"
	ResponseMessageProjectIdIsInvalid = "one or more project identifier is invalid"
	ResponseMessageUnknownDbError     = "err: 1, unknown error. try request later"
	ResponseMessageUnknownError       = "unknown error. try request later"

	ApiRequestParameterProjectId = "project_id"
	ApiRequestParameterRegion    = "region"

	DBFieldId          = "id"
	DBFieldName        = "name"
	DBFieldPrice       = "price"
	DBFieldCurrencyInt = "currency_int"

	BankCardFieldBrand         = "card_brand"
	BankCardFieldType          = "card_type"
	BankCardFieldCategory      = "card_category"
	BankCardFieldIssuerName    = "bank_issuer_name"
	BankCardFieldIssuerCountry = "bank_issuer_country"
)

var DefaultSort = []string{"_id"}

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
	CodeA2 string `bson:"code_a2" json:"code_a2"`
}

type OrderSimpleAmountObject struct {
	// amount value
	Amount float64 `json:"amount"`
	// object which contains main information about currency
	Currency *SimpleCurrency `json:"currency"`
}
