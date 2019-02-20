package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
)

func (rep *Repository) FindCurrenciesPair(from int, to int) (*model.CurrencyRate, error) {
	var cr *model.CurrencyRate
	err := rep.Collection.Find(bson.M{"currency_from": from, "currency_to": to}).One(&cr)

	return cr, err
}
