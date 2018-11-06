package repository

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
)

func (rep *Repository) FindCurrenciesPair(from int, to int) (*model.CurrencyRate, error) {
	var cr *model.CurrencyRate
	err := rep.Collection.Find(bson.M{"currency_from": from, "currency_to": to}).One(&cr)

	return cr, err
}
