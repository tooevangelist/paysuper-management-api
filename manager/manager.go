package manager

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"go.uber.org/zap"
)

const (
	TableCountry  = "country"
	TableCurrency = "currency"
	TableMerchant = "merchant"
	TableProject  = "project"
)

type Manager struct {
	Database dao.Database
	Logger   *zap.SugaredLogger
}
