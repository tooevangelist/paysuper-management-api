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

func (rep *Repository) FindOrderById(id bson.ObjectId) (*model.Order, error) {
	var o *model.Order
	err := rep.Collection.FindId(id).One(&o)

	return o, err
}

func (rep *Repository) InsertOrder(order *model.Order) error {
	return rep.Collection.Insert(order)
}
