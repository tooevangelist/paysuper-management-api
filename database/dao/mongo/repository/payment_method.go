package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
)

func (rep *Repository) FindPaymentMethodById(id bson.ObjectId) (*model.PaymentMethod, error) {
	var pm *model.PaymentMethod
	err := rep.Collection.FindId(id).One(&pm)

	return pm, err
}

func (rep *Repository) FindAllPaymentMethods() ([]*model.PaymentMethod, error) {
	var pms []*model.PaymentMethod
	err := rep.Collection.Find(bson.M{}).All(&pms)

	return pms, err
}

func (rep *Repository) FindPaymentMethodsByIds(ids []bson.ObjectId) ([]*model.PaymentMethod, error) {
	var pms []*model.PaymentMethod
	err := rep.Collection.Find(bson.M{"_id": bson.M{"$in": ids}}).All(&pms)

	return pms, err
}
