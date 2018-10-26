package dao

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"gopkg.in/mgo.v2/bson"
)

type Repository interface {
	FindCurrencyById(int) (*model.Currency, error)
	FindCurrenciesByName(string) ([]*model.Currency, error)
	FindAllCurrencies(int, int) ([]*model.Currency, error)

	FindCountryById(int) (*model.Country, error)
	FindCountryByName(string) ([]*model.Country, error)
	FindAllCountries(int, int) ([]*model.Country, error)

	FindMerchantById(id string) (*model.Merchant, error)
	InsertMerchant(m *model.Merchant) error
	UpdateMerchant(m *model.Merchant) error

	InsertProject(p *model.Project) error
	UpdateProject(p *model.Project) error
	FindProjectsByMerchantId(id bson.ObjectId) ([]*model.Project, error)
	FindProjectsByMerchantIdAndName(bson.ObjectId, string) *model.Project
	FindProjectById(bson.ObjectId) (*model.Project, error)
}
