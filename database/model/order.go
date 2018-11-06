package model

import "github.com/globalsign/mgo/bson"

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

	orderStatusCreated  = 0
	orderStatusComplete = 10
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
}

var OrderStatusesDescription = map[int]string{
	orderStatusCreated:  "Order created",
	orderStatusComplete: "Order successfully complete. Notification successfully send to project",
}

type CreateOrderData struct {
	Ip            string
	CountryCodeA3 string
	City          string
	Timezone      string
}

type OrderScalar struct {
	ProjectId     string  `query:"PP_PROJECT_ID" form:"PP_PROJECT_ID" validate:"required,hexadecimal"`
	Signature     *string `query:"PP_SIGNATURE" form:"PP_SIGNATURE" validate:"omitempty,alphanum"`
	Amount        float64 `query:"PP_AMOUNT" form:"PP_AMOUNT" validate:"required,numeric"`
	Currency      *string `query:"PP_CURRENCY" form:"PP_CURRENCY" validate:"omitempty,alpha,len=3"`
	Account       string  `query:"PP_ACCOUNT" form:"PP_ACCOUNT" validate:"required"`
	OrderId       *string `query:"PP_ORDER_ID" form:"PP_ORDER_ID"`
	Description   *string `query:"PP_DESCRIPTION" form:"PP_DESCRIPTION"`
	PaymentMethod *string `query:"PP_PAYMENT_METHOD" form:"PP_PAYMENT_METHOD" validate:"omitempty,alphanum"`
	UrlVerify     *string `query:"PP_URL_VERIFY" form:"PP_URL_VERIFY" validate:"omitempty,url"`
	UrlNotify     *string `query:"PP_URL_NOTIFY" form:"PP_URL_NOTIFY" validate:"omitempty,url"`
	UrlSuccess    *string `query:"PP_URL_SUCCESS" form:"PP_URL_SUCCESS" validate:"omitempty,url"`
	UrlFail       *string `query:"PP_URL_FAIL" form:"PP_URL_FAIL" validate:"omitempty,url"`
	PayerEmail    *string `query:"PP_PAYER_EMAIL" form:"PP_PAYER_EMAIL" validate:"omitempty,email"`
	PayerPhone    *string `query:"PP_PAYER_PHONE" form:"PP_PAYER_PHONE"`
	Region        *string `query:"PP_REGION" form:"PP_REGION" validate:"omitempty,alpha,len=3"`
	CreateOrderIp string
	Other         map[string]string
}

type Order struct {
	Id              bson.ObjectId `bson:"_id" json:"id"`
	ProjectOrderId  *string       `bson:"project_order_id" json:"project_order_id"`
	PaymentMethodId bson.ObjectId `bson:"payment_method_id" json:"payment_method_id"`
}
