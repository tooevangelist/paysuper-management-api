package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

type Commission struct {
	Id                      bson.ObjectId `bson:"_id" json:"id"`
	PaymentMethodId         bson.ObjectId `bson:"pm_id" json:"pm_id"`
	ProjectId               bson.ObjectId `bson:"project_id" json:"project_id"`
	PaymentMethodCommission float64       `bson:"pm_commission" json:"pm_commission"`
	PspCommission           float64       `bson:"psp_commission" json:"psp_commission"`
	TotalCommissionToUser   float64       `bson:"total_commission_to_user" json:"total_commission_to_user"`
	StartDate               time.Time     `bson:"start_date" json:"start_date"`
	CreatedAt               time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt               *time.Time    `bson:"updated_at" json:"updated_at,omitempty"`
}

type CommissionOrder struct {
	PMCommission     float64
	PspCommission    float64
	ToUserCommission float64
}
