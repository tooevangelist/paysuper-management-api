package manager

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"go.uber.org/zap"
)

type CountryManager Manager

func InitCountryManager(database dao.Database, logger *zap.SugaredLogger) *CountryManager {
	return &CountryManager{Database: database, Logger: logger}
}

func (cm *CountryManager) FindByCodeInt(codeInt int) *model.Country {
	c, err := cm.Database.Repository(tableCountry).FindCountryById(codeInt)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", err)
	}

	return c
}

func (cm *CountryManager) FindByName(name string) []*model.Country {
	c, err := cm.Database.Repository(tableCountry).FindCountryByName(name)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", err)
	}

	if c == nil {
		return []*model.Country{}
	}

	return c
}

func (cm *CountryManager) FindAll(limit int, offset int) []*model.Country {
	c, err := cm.Database.Repository(tableCountry).FindAllCountries(limit, offset)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", err)
	}

	if c == nil {
		return []*model.Country{}
	}

	return c
}

