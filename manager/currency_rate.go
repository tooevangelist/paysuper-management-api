package manager

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"go.uber.org/zap"
)

type CurrencyRateManager Manager

func InitCurrencyRateManager(database dao.Database, logger *zap.SugaredLogger) *CurrencyRateManager {
	crm := &CurrencyRateManager{
		Database: database,
		Logger:   logger,
	}

	return crm
}

func (crm *CurrencyRateManager) convert(from int, to int, value float64) (float64, error) {
	cr, err := crm.Database.Repository(TableCurrencyRate).FindCurrenciesPair(from, to)

	if err != nil {
		crm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableCurrencyRate, err)
	}

	if cr == nil {
		return 0, errors.New("currencies pair not found")
	}

	value = value / cr.Rate

	return value, nil
}
