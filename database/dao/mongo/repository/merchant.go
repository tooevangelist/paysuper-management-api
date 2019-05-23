package repository

import (
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-management-api/database/model"
	"gopkg.in/mgo.v2/bson"
)

func (rep *Repository) FindMerchantById(id string) (*billing.Merchant, error) {
	var m *billing.Merchant
	err := rep.Collection.Find(bson.M{"user.id": id}).One(&m)

	return m, err
}

func (rep *Repository) InsertMerchant(m *model.Merchant) error {
	return rep.Collection.Insert(m)
}

func (rep *Repository) UpdateMerchant(m *billing.Merchant) error {
	return rep.Collection.UpdateId(m.Id, m)
}
