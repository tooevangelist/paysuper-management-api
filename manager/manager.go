package manager

import (
	"fmt"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
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
	TableLog           = "log"
	TableVat           = "vat"
	TableCommission    = "commission"

	errorMessageMask = "Field validation for '%s' failed on the '%s' tag"
)

type Manager struct {
	Database dao.Database
	Logger   *zap.SugaredLogger
}

func FormatAmount(amount float64) float64 {
	return math.Floor(amount*100) / 100
}

func GetFirstValidationError(err error) string {
	vErr := err.(validator.ValidationErrors)[0]

	return fmt.Sprintf(errorMessageMask, vErr.Field(), vErr.Tag())
}
