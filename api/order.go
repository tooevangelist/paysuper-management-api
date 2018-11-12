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

	api.Http.GET("/order/:id", oApiV1.getOrderForm)
	api.Http.GET("/order/create", oApiV1.createFromFormData)
	api.Http.POST("/order/create", oApiV1.createFromFormData)
	api.Http.POST("/api/v1/order", oApiV1.createJson)

	return api
}

func (oApiV1 *OrderApiV1) createFromFormData(ctx echo.Context) error {
	order := &model.OrderScalar{
		CreateOrderIp: ctx.RealIP(),
		IsJsonRequest: false,
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

	rUrl := "/order/" + nOrder.Id.Hex()

	return ctx.Redirect(http.StatusFound, rUrl)
}

func (oApiV1 *OrderApiV1) createJson(ctx echo.Context) error {
	order := &model.OrderScalar{
		IsJsonRequest: true,
	}

	if err := (&OrderJsonBinder{}).Bind(order, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	if err := oApiV1.validate.Struct(order); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, oApiV1.getFirstValidationError(err))
	}

	nOrder, err := oApiV1.orderManager.Process(order)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	url := "https://" + ctx.Request().Host + "/order/" + nOrder.Id.Hex()

	ou := &model.OrderUrl{OrderUrl: url}

	return ctx.JSON(http.StatusOK, ou)
}

func (oApiV1 *OrderApiV1) getOrderForm(ctx echo.Context) error {
	id := ctx.Param("id")

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid order id")
	}

	o := oApiV1.orderManager.FindById(id)

	if o == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Order not found")
	}

	return ctx.Render(http.StatusOK, "order.html", map[string]interface{}{
		"ProjectName": "Test project",
		"Years":       oApiV1.orderManager.GetCardYears(),
		"Months":      oApiV1.orderManager.GetCardMonths(),
		"Order":       o,
	})
}
