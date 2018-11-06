package repository

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
)

func (rep *Repository) FindPaymentMethodById(id bson.ObjectId) (*model.PaymentMethod, error) {
	var pm *model.PaymentMethod
	err := rep.Collection.FindId(id).One(&pm)

	return pm, err
}
