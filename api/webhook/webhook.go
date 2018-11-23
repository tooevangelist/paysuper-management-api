package webhook

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/labstack/echo"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type WebHook struct {
	database                dao.Database
	logger                  *zap.SugaredLogger
	validate                *validator.Validate
	geoDbReader             *geoip2.Reader
	pspAccountingCurrencyA3 string
	webHookGroup            *echo.Group
	webHookRawBody          string
	paymentSystemConfig     map[string]interface{}
}

func InitWebHook(
	database dao.Database,
	logger *zap.SugaredLogger,
	validate *validator.Validate,
	geoDbReader *geoip2.Reader,
	pspAccountingCurrencyA3 string,
	webHookGroup *echo.Group,
	webHookRawBody string,
	paymentSystemConfig map[string]interface{},
) *WebHook {
	return &WebHook{
		database:                database,
		logger:                  logger,
		validate:                validate,
		geoDbReader:             geoDbReader,
		pspAccountingCurrencyA3: pspAccountingCurrencyA3,
		webHookGroup:            webHookGroup,
		webHookRawBody:          webHookRawBody,
		paymentSystemConfig:     paymentSystemConfig,
	}
}
