package dao

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
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
	FindProjectsByMerchantId(string, int, int) ([]*model.Project, error)
	FindProjectByMerchantIdAndName(bson.ObjectId, string) (*model.Project, error)
	FindProjectById(bson.ObjectId) (*model.Project, error)
}
