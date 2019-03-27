package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/manager"
	"net/http"
)

type PaymentMethodApiV1 struct {
	*Api
	projectManager *manager.ProjectManager
}

func (api *Api) InitPaymentMethodRoutes() *Api {
	pmApiV1 := PaymentMethodApiV1{
		Api:            api,
		projectManager: manager.InitProjectManager(api.database, api.logger),
	}

	api.accessRouteGroup.GET("/payment_method/merchant", pmApiV1.getMerchantPaymentMethodsForFilters)

	return api
}

func (pmApiV1 *PaymentMethodApiV1) getMerchantPaymentMethodsForFilters(ctx echo.Context) error {
	p := pmApiV1.projectManager.GetProjectsPaymentMethodsByMerchantMainData(pmApiV1.Merchant.Identifier)

	return ctx.JSON(http.StatusOK, p)
}
