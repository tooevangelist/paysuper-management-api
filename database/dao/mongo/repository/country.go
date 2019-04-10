package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
)

func (rep *Repository) FindCountryById(codeInt int) (*model.Country, error) {
	var c *model.Country
	err := rep.Collection.Find(bson.M{"code_int": codeInt, "is_active": true}).One(&c)

	return c, err
}

func (rep *Repository) FindCountryByName(name string) ([]*model.Country, error) {
	var c []*model.Country

	r := bson.RegEx{Pattern: ".*" + name + ".*", Options: "i"}
	err := rep.Collection.Find(
		bson.M{
			"$or": []bson.M{
				{"code_a2": r},
				{"code_a3": r},
				{"name.en": r},
				{"name.ru": r},
			},
			"is_active": true,
		}).All(&c)

	return c, err
}

func (rep *Repository) FindAllCountries(limit int32, offset int32) (*model.CountryItems, error) {
	var c []*model.Country
	query := bson.M{"is_active": true}

	count, err := rep.Collection.Find(query).Count()

	if err != nil {
		return nil, err
	}

	if count <= 0 {
		return &model.CountryItems{Items: []*model.Country{}}, nil
	}

	err = rep.Collection.Find(query).Limit(int(limit)).Skip(int(offset)).All(&c)

	if err != nil {
		return nil, err
	}

	return &model.CountryItems{Count: count, Items: c}, nil
}
