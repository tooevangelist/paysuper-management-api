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

func (rep *Repository) FindAllPaymentSystem() ([]*model.PaymentSystem, error) {
	var pss []*model.PaymentSystem
	err := rep.Collection.Find(bson.M{}).All(&pss)

	return pss, err
}
