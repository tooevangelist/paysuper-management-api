package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
)

func (rep *Repository) FindVatByCountry(cCodeA2 string) (*model.Vat, error) {
	var vat *model.Vat

	err := rep.Collection.Find(bson.M{"country.code_a2": cCodeA2}).
		Sort("-created_at").Limit(1).One(&vat)

	return vat, err
}

func (rep *Repository) FindVatByCountryAndSubdivision(cCodeA2 string, subdivision string) (*model.Vat, error) {
	var vat *model.Vat

	err := rep.Collection.Find(bson.M{"country.code_a2": cCodeA2, "subdivision_code": subdivision}).
		Sort("-created_at").Limit(1).One(&vat)

	return vat, err
}
