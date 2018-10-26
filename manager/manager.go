package manager

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"go.uber.org/zap"
)

const (
	tableCountry  = "country"
	tableCurrency = "currency"
	tableMerchant = "merchant"
	tableProject  = "project"
)

type Manager struct {
	Database dao.Database
	Logger   *zap.SugaredLogger
}
