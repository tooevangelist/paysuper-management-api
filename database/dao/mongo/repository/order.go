package repository

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
)

func (rep *Repository) FindOrderByProjectOrderId(prjOrderId string) (*model.Order, error) {
	var o *model.Order
	err := rep.Collection.Find(bson.M{"project_order_id": prjOrderId}).One(&o)

	return o, err
}
