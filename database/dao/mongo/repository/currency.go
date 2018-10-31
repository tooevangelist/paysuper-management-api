package repository

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
)

func (rep *Repository) FindCurrencyById(codeInt int) (*model.Currency, error) {
	var c *model.Currency
	err := rep.Collection.Find(bson.M{"code_int": codeInt, "is_active": true}).One(&c)

	return c, err
}

func (rep *Repository) FindCurrenciesByName(name string) ([]*model.Currency, error) {
	var c []*model.Currency

	r := bson.RegEx{Pattern: ".*" + name + ".*", Options: "i"}
	err := rep.Collection.Find(
		bson.M{
			"$or": []bson.M{ { "code_a3": r }, { "name.en": r }, { "name.ru": r }},
			"is_active": true,
		}).All(&c)

	return c, err
}

func (rep *Repository) FindAllCurrencies(limit int, offset int) ([]*model.Currency, error) {
	var c []*model.Currency
	err := rep.Collection.Find(bson.M{"is_active": true}).Limit(limit).Skip(offset).All(&c)

	return c, err
}
