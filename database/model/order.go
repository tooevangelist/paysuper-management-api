package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

const (
	orderFieldProjectId     = "PP_PROJECT_ID"
	orderFieldSignature     = "PP_SIGNATURE"
	orderFieldAmount        = "PP_AMOUNT"
	orderFieldCurrency      = "PP_CURRENCY"
	orderFieldAccount       = "PP_ACCOUNT"
	orderFieldOrderId       = "PP_ORDER_ID"
	orderFieldPaymentMethod = "PP_PAYMENT_METHOD"
	orderFieldUrlVerify     = "PP_URL_VERIFY"
	orderFieldUrlNotify     = "PP_URL_NOTIFY"
	orderFieldUrlSuccess    = "PP_URL_SUCCESS"
	orderFieldUrlFail       = "PP_URL_FAIL"
	orderFieldPayerEmail    = "PP_PAYER_EMAIL"
	orderFieldPayerPhone    = "PP_PAYER_PHONE"
	orderFieldDescription   = "PP_DESCRIPTION"
	orderFieldRegion        = "PP_REGION"

	OrderStatusCreated  = 0
	OrderStatusComplete = 10
)

var OrderReservedWords = map[string]bool{
	orderFieldProjectId:     true,
	orderFieldSignature:     true,
	orderFieldAmount:        true,
	orderFieldCurrency:      true,
	orderFieldAccount:       true,
	orderFieldOrderId:       true,
	orderFieldDescription:   true,
	orderFieldPaymentMethod: true,
	orderFieldUrlVerify:     true,
	orderFieldUrlNotify:     true,
	orderFieldUrlSuccess:    true,
	orderFieldUrlFail:       true,
	orderFieldPayerEmail:    true,
	orderFieldPayerPhone:    true,
	orderFieldRegion:        true,
}

var OrderStatusesDescription = map[int]string{
	OrderStatusCreated:  "Order created",
	OrderStatusComplete: "Order successfully complete. Notification successfully send to project",
}

type PayerData struct {
	Ip            string
	CountryCodeA3 string
	City          string
	Timezone      string
	Phone         *string
	Email         *string
}

type OrderScalar struct {
	// ProjectId Project unique identifier in Protocol One payment solution
	ProjectId        string                 `query:"PP_PROJECT_ID" form:"PP_PROJECT_ID" json:"project" validate:"required,hexadecimal" swaggertype:"string"`
	// Signature Signature of request to verify that the data has not been changed. This field not required, BUT we're recommend send this field always
	Signature        *string                `query:"PP_SIGNATURE" form:"PP_SIGNATURE" json:"signature" validate:"omitempty,alphanum" swaggertype:"string"`
	// Amount Order amount
	Amount           float64                `query:"PP_AMOUNT" form:"PP_AMOUNT" json:"amount" validate:"required,numeric" swaggertype:"number"`
	// Currency Order currency by ISO 4217 (3 chars). If this field send, then we're process amount in this currency
	Currency         *string                `query:"PP_CURRENCY" form:"PP_CURRENCY" json:"currency" validate:"omitempty,alpha,len=3" swaggertype:"string"`
	// Account User unique account in project
	Account          string                 `query:"PP_ACCOUNT" form:"PP_ACCOUNT" json:"account" validate:"required"`
	// OrderId Unique order identifier in project. This field not required, BUT we're recommend send this field always
	OrderId          *string                `query:"PP_ORDER_ID" form:"PP_ORDER_ID" json:"order_id"`
	// Description Order description. If this field not send in request, then we're create standard order description
	Description      *string                `query:"PP_DESCRIPTION" form:"PP_DESCRIPTION" json:"description"`
	// PaymentMethod Payment method identifier in Protocol One payment solution
	PaymentMethod    *string                `query:"PP_PAYMENT_METHOD" form:"PP_PAYMENT_METHOD" json:"payment_method"`
	// UrlVerify URL for payment data verification request to project. This field can be send if it allowed in project admin panel
	UrlVerify        *string                `query:"PP_URL_VERIFY" form:"PP_URL_VERIFY" json:"url_verify" validate:"omitempty,url"`
	// UrlNotify URL for payment notification request to project. This field can be send if it allowed in project admin panel
	UrlNotify        *string                `query:"PP_URL_NOTIFY" form:"PP_URL_NOTIFY" json:"url_notify" validate:"omitempty,url"`
	// UrlSuccess URL for redirect user after successfully completed payment. This field can be send if it allowed in project admin panel
	UrlSuccess       *string                `query:"PP_URL_SUCCESS" form:"PP_URL_SUCCESS" json:"url_success" validate:"omitempty,url"`
	// UrlFail URL for redirect user after failed payment. This field can be send if it allowed in project admin panel
	UrlFail          *string                `query:"PP_URL_FAIL" form:"PP_URL_FAIL" json:"url_fail" validate:"omitempty,url"`
	// PayerEmail User (payer) email
	PayerEmail       *string                `query:"PP_PAYER_EMAIL" form:"PP_PAYER_EMAIL" json:"payer_email" validate:"omitempty,email"`
	// PayerPhone User (payer) phone
	PayerPhone       *string                `query:"PP_PAYER_PHONE" form:"PP_PAYER_PHONE" json:"payer_phone"`
	// Region User (payer) region code by ISO 3166-1 (2 chars) for check project packages. If this field not send, then user region will be get from user ip
	Region           *string                `query:"PP_REGION" form:"PP_REGION" json:"region" validate:"omitempty,alpha,len=2"`
	// CreateOrderIp User (payer) ip
	CreateOrderIp    string                 `json:"payer_ip"`
	// Other Object with any fields on the project side that do not match the names of the reserved fields
	Other            map[string]interface{} `json:"other"`
	RawRequestParams map[string]string      `json:"-"`
	RawRequestBody   string                 `json:"-"`
	IsJsonRequest    bool                   `json:"-"`
}

type Order struct {
	Id                           bson.ObjectId          `bson:"_id" json:"id"`
	ProjectId                    bson.ObjectId          `bson:"project_id" json:"project_id"`
	ProjectOrderId               *string                `bson:"project_order_id" json:"project_order_id"`
	ProjectAccount               string                 `bson:"project_account" json:"project_account"`
	ProjectIncomeAmount          float64                `bson:"project_income_amount" json:"project_income_amount"`
	ProjectIncomeCurrency        *Currency              `bson:"project_income_currency" json:"project_income_currency"`
	ProjectOutcomeAmount         float64                `bson:"project_outcome_amount" json:"project_outcome_amount"`
	ProjectOutcomeCurrency       *Currency              `bson:"project_outcome_currency" json:"project_outcome_currency"`
	ProjectFee                   float64                `bson:"project_fee" json:"project_fee"`
	ProjectLastRequestedAt       time.Time              `bson:"project_last_requested_at" json:"project_last_requested_at"`
	ProjectParams                map[string]interface{} `bson:"project_params" json:"project_params"`
	PayerData                    *PayerData             `bson:"payer_data" json:"payer_data"`
	PaymentMethodId              bson.ObjectId          `bson:"pm_id" json:"pm_id"`
	PaymentMethodTerminalId      string                 `bson:"pm_terminal_id" json:"pm_terminal_id"`
	PaymentMethodOrderId         string                 `bson:"pm_order_id" json:"pm_order_id"`
	PaymentMethodOutcomeAmount   float64                `bson:"pm_outcome_amount" json:"pm_outcome_amount"`
	PaymentMethodOutcomeCurrency *Currency              `bson:"pm_outcome_currency" json:"pm_outcome_currency"`
	PaymentMethodIncomeAmount    float64                `bson:"pm_income_amount" json:"pm_income_amount"`
	PaymentMethodIncomeCurrency  *Currency              `bson:"pm_income_currency" json:"pm_income_currency"`
	PaymentMethodFee             float64                `bson:"pm_fee" json:"pm_fee"`
	PaymentMethodOrderClosedAt   time.Time              `bson:"pm_order_close_date" json:"pm_order_close_date"`
	Status                       int                    `bson:"status" json:"status"`
	CreatedAt                    time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt                    time.Time              `bson:"updated_at" json:"created_at"`
	IsJsonRequest                bool                   `bson:"created_by_json" json:"created_by_json"`

	ProjectOutcomeAmountPrintable string   `bson:"-" json:"-"`
	OrderIdPrintable              string   `bson:"-" json:"-"`
	ProjectData                   *Project `bson:"-" json:"-"`
}

type OrderUrl struct {
	OrderUrl string `json:"order_url"`
}
