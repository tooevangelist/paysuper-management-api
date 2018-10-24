package dao

import "github.com/ProtocolONE/p1pay.api/database/model"

type Repository interface {
	FindCurrencyById(int) (*model.Currency, error)
	FindCurrenciesByName(string) ([]*model.Currency, error)
	FindAllCurrencies(int, int) ([]*model.Currency, error)
}
