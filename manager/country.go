package manager

import (
	"github.com/paysuper/paysuper-management-api/database/dao"
	"github.com/paysuper/paysuper-management-api/database/model"
	"go.uber.org/zap"
)

type CountryManager Manager

func InitCountryManager(database dao.Database, logger *zap.SugaredLogger) *CountryManager {
	return &CountryManager{Database: database, Logger: logger}
}

func (cm *CountryManager) FindByIsoCodeA2(isoCodeA2 string) *model.Country {
	c, err := cm.Database.Repository(TableCountry).FindCountryByIsoCodeA2(isoCodeA2)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableCountry, err)
	}

	return c
}

func (cm *CountryManager) FindAll(limit int32, offset int32) *model.CountryItems {
	c, err := cm.Database.Repository(TableCountry).FindAllCountries(limit, offset)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableCountry, err)
	}

	return c
}
