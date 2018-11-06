package repository

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
)

func (rep *Repository) FindPaymentSystemById(id bson.ObjectId) (*model.PaymentSystem, error) {
	var ps *model.PaymentSystem
	err := rep.Collection.FindId(id).One(&ps)

	return ps, err
}
