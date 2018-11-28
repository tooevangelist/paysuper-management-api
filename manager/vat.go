package manager

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
)

const (
	vatErrorNotFound = "vat not found for user region"
)

type VatManager Manager

func InitVatManager(database dao.Database, logger *zap.SugaredLogger) *VatManager {
	vm := &VatManager{
		Database: database,
		Logger:   logger,
	}

	return vm
}

func (vm *VatManager) CalculateVat(geo *geoip2.City, amount float64) (float64, error) {
	var vat *model.Vat
	var err error

	vsFlag, ok := model.VatBySubdivisionCountries[geo.Country.IsoCode]

	if !ok || vsFlag == false || geo.Subdivisions[0].IsoCode == "" {
		vat, err = vm.Database.Repository(TableVat).FindVatByCountry(geo.Country.IsoCode)
	} else {
		vat, err = vm.Database.Repository(TableVat).FindVatByCountryAndSubdivision(geo.Country.IsoCode, geo.Subdivisions[0].IsoCode)
	}

	if err != nil {
		vm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableVat, err)
	}

	if vat == nil {
		return 0, errors.New(vatErrorNotFound)
	}

	amount = amount * (vat.Vat / 100)

	return amount, nil
}
