package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
)

func (rep *Repository) FindCountryByIsoCodeA2(isoCodeA2 string) (*model.Country, error) {
	var c *model.Country
	err := rep.Collection.Find(bson.M{"iso_code_a2": isoCodeA2}).One(&c)

	return c, err
}

func (rep *Repository) FindAllCountries(limit int32, offset int32) (*model.CountryItems, error) {
	var c []*model.Country
	count, err := rep.Collection.Find(nil).Count()

	if err != nil {
		return nil, err
	}

	if count <= 0 {
		return &model.CountryItems{Items: []*model.Country{}}, nil
	}

	err = rep.Collection.Find(nil).Limit(int(limit)).Skip(int(offset)).All(&c)

	if err != nil {
		return nil, err
	}

	return &model.CountryItems{Count: count, Items: c}, nil
}
