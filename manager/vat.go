package manager

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
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

func (vm *VatManager) CalculateVat(countryCodeA2 string, subdivision string, amount float64) (float64, error) {
	var vat *model.Vat
	var err error

	vsFlag, ok := model.VatBySubdivisionCountries[countryCodeA2]

	if !ok || vsFlag == false || subdivision == "" {
		vat, err = vm.Database.Repository(TableVat).FindVatByCountry(countryCodeA2)
	} else {
		vat, err = vm.Database.Repository(TableVat).FindVatByCountryAndSubdivision(countryCodeA2, subdivision)
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
