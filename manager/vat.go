package manager

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
)

type VatManager Manager

func InitVatManager(database dao.Database, logger *zap.SugaredLogger) *VatManager {
	vm := &VatManager{
		Database: database,
		Logger:   logger,
	}

	return vm
}

func (vm *VatManager) CalculateVat(geoCity *geoip2.City, amount float64) {

}

