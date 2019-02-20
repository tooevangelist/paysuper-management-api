package manager

import (
	"github.com/paysuper/paysuper-management-api/database/dao"
	"github.com/paysuper/paysuper-management-api/database/model"
	"go.uber.org/zap"
)

type CurrencyManager Manager

func InitCurrencyManager(database dao.Database, logger *zap.SugaredLogger) *CurrencyManager {
	return &CurrencyManager{Database: database, Logger: logger}
}

func (cm *CurrencyManager) FindByCodeInt(codeInt int) *model.Currency {
	c, err := cm.Database.Repository(TableCurrency).FindCurrencyById(codeInt)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableCurrency, err)
	}

	return c
}

func (cm *CurrencyManager) FindByCodeA3(codeA3 string) *model.Currency {
	c, err := cm.Database.Repository(TableCurrency).FindCurrencyByCodeA3(codeA3)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableCurrency, err)
	}

	return c
}

func (cm *CurrencyManager) FindByName(name string) []*model.Currency {
	c, err := cm.Database.Repository(TableCurrency).FindCurrenciesByName(name)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableCurrency, err)
	}

	if c == nil {
		return []*model.Currency{}
	}

	return c
}

func (cm *CurrencyManager) FindAll(limit int, offset int) []*model.Currency {
	c, err := cm.Database.Repository(TableCurrency).FindAllCurrencies(limit, offset)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableCurrency, err)
	}

	if c == nil {
		return []*model.Currency{}
	}

	return c
}
