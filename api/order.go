package api

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/labstack/echo"
	"net/http"
)

type OrderApiV1 struct {
	*Api
	orderManager *manager.OrderManager
}

func (api *Api) InitOrderV1Routes() *Api {
	oApiV1 := OrderApiV1{
		Api:          api,
		orderManager: manager.InitOrderManager(api.database, api.logger, api.geoDbReader),
	}

	api.Http.GET("/order/create", oApiV1.createFromFormData)
	api.Http.POST("/order/create", oApiV1.createFromFormData)

	return api
}

func (oApiV1 *OrderApiV1) createFromFormData(ctx echo.Context) error {
	order := &model.OrderScalar{
		CreateOrderIp: ctx.RealIP(),
	}

	if err := (&OrderFormBinder{}).Bind(order, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Request data invalid")
	}

	if err := oApiV1.validate.Struct(order); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, oApiV1.getFirstValidationError(err))
	}

	nOrder, err := oApiV1.orderManager.Process(order)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.Render(http.StatusOK, "order.html", map[string]interface{}{
		"ProjectName": "Test project",
		"Years": oApiV1.orderManager.GetCardYears(),
		"Months": oApiV1.orderManager.GetCardMonths(),
		"Order": nOrder,
	})
}