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
	ProjectId        string                 `query:"PP_PROJECT_ID" form:"PP_PROJECT_ID" json:"project" validate:"required,hexadecimal"`
	Signature        *string                `query:"PP_SIGNATURE" form:"PP_SIGNATURE" json:"signature" validate:"omitempty,alphanum"`
	Amount           float64                `query:"PP_AMOUNT" form:"PP_AMOUNT" json:"amount" validate:"required,numeric"`
	Currency         *string                `query:"PP_CURRENCY" form:"PP_CURRENCY" json:"currency" validate:"omitempty,alpha,len=3"`
	Account          string                 `query:"PP_ACCOUNT" form:"PP_ACCOUNT" json:"account" validate:"required"`
	OrderId          *string                `query:"PP_ORDER_ID" form:"PP_ORDER_ID" json:"order_id"`
	Description      *string                `query:"PP_DESCRIPTION" form:"PP_DESCRIPTION" json:"description"`
	PaymentMethod    *string                `query:"PP_PAYMENT_METHOD" form:"PP_PAYMENT_METHOD" json:"payment_method"`
	UrlVerify        *string                `query:"PP_URL_VERIFY" form:"PP_URL_VERIFY" json:"url_verify" validate:"omitempty,url"`
	UrlNotify        *string                `query:"PP_URL_NOTIFY" form:"PP_URL_NOTIFY" json:"url_notify" validate:"omitempty,url"`
	UrlSuccess       *string                `query:"PP_URL_SUCCESS" form:"PP_URL_SUCCESS" json:"url_success" validate:"omitempty,url"`
	UrlFail          *string                `query:"PP_URL_FAIL" form:"PP_URL_FAIL" json:"url_fail" validate:"omitempty,url"`
	PayerEmail       *string                `query:"PP_PAYER_EMAIL" form:"PP_PAYER_EMAIL" json:"payer_email" validate:"omitempty,email"`
	PayerPhone       *string                `query:"PP_PAYER_PHONE" form:"PP_PAYER_PHONE" json:"payer_phone"`
	Region           *string                `query:"PP_REGION" form:"PP_REGION" json:"region" validate:"omitempty,alpha,len=2"`
	CreateOrderIp    string                 `json:"payer_ip"`
	Other            map[string]interface{} `json:"other"`
	RawRequestParams map[string]string
	RawRequestBody   string
	IsJsonRequest    bool
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
	PaymentMethodOrderClosedAt   time.Time              `bson:"pm_order_close_dat" json:"pm_order_close_dat"`
	Status                       int                    `bson:"status" json:"status"`
	CreatedAt                    time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt                    time.Time              `bson:"updated_at" json:"created_at"`

	ProjectOutcomeAmountPrintable string   `bson:"-" json:"-"`
	OrderIdPrintable              string   `bson:"-" json:"-"`
	ProjectData                   *Project `bson:"-" json:"-"`
}

type OrderUrl struct {
	OrderUrl string `json:"order_url"`
}
