package manager

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"go.uber.org/zap"
	"math"
)

const (
	TableCountry       = "country"
	TableCurrency      = "currency"
	TableMerchant      = "merchant"
	TableProject       = "project"
	TablePaymentSystem = "payment_system"
	TablePaymentMethod = "payment_method"
	TableOrder         = "order"
	TableCurrencyRate  = "currency_rate"
)

type Manager struct {
	Database dao.Database
	Logger   *zap.SugaredLogger
}

func FormatAmount(amount float64) float64 {
	return math.Floor(amount*100) / 100
}
