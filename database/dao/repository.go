package dao

import "github.com/ProtocolONE/p1pay.api/database/model"

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
}
