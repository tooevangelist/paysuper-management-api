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

// @Summary Create order with HTML form
// @Description Create a payment order use GET or POST HTML form
// @Tags Payment Order
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Produce html
// @Param PP_PROJECT_ID query string true "Project unique identifier in Protocol One payment solution"
// @Param PP_AMOUNT query float64 true "Order amount"
// @Param PP_ACCOUNT query string true "User unique account in project"
// @Param PP_ORDER_ID query string false "Unique order identifier in project. This field not required, BUT we're recommend send this field always"
// @Param PP_PAYMENT_METHOD query string false "Payment method identifier in Protocol One payment solution"
// @Param PP_DESCRIPTION query string false "Order description. If this field not send in request, then we're create standard order description"
// @Param PP_CURRENCY query string false "Order currency by ISO 4217 (3 chars). If this field send, then we're process amount in this currency"
// @Param PP_REGION query string false "User (payer) region code by ISO 3166-1 (2 chars) for check project packages. If this field not send, then user region will be get from user ip"
// @Param PP_PAYER_EMAIL query string false "User (payer) email"
// @Param PP_PAYER_PHONE query string false "User (payer) phone"
// @Param PP_URL_VERIFY query string false "URL for payment data verification request to project. This field can be send if it allowed in project admin panel"
// @Param PP_URL_NOTIFY query string false "URL for payment notification request to project. This field can be send if it allowed in project admin panel"
// @Param PP_URL_SUCCESS query string false "URL for redirect user after successfully completed payment. This field can be send if it allowed in project admin panel"
// @Param PP_URL_FAIL query string false "URL for redirect user after failed payment. This field can be send if it allowed in project admin panel"
// @Param PP_SIGNATURE query string false "Signature of request to verify that the data has not been changed. This field not required, BUT we're recommend send this field always"
// @Param Other query string false "Any fields on the project side that do not match the names of the reserved fields"
// @Success 302 {string} html "Redirect user to form entering payment requisites"
// @Failure 400 {string} html "Redirect user to page with error description"
// @Failure 500 {string} html "Redirect user to page with error description"
// @Router /order/create [get]
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

// @Summary Create order with HTML form
// @Description Create a payment order use GET or POST HTML form
// @Tags Payment Order
// @Accept multipart/form-data
// @Accept application/x-www-form-urlencoded
// @Produce html
// @Param PP_PROJECT_ID query string true "Project unique identifier in Protocol One payment solution"
// @Param PP_AMOUNT query float64 true "Order amount"
// @Param PP_ACCOUNT query string true "User unique account in project"
// @Param PP_ORDER_ID query string false "Unique order identifier in project. This field not required, BUT we're recommend send this field always"
// @Param PP_PAYMENT_METHOD query string false "Payment method identifier in Protocol One payment solution"
// @Param PP_DESCRIPTION query string false "Order description. If this field not send in request, then we're create standard order description"
// @Param PP_CURRENCY query string false "Order currency by ISO 4217 (3 chars). If this field send, then we're process amount in this currency"
// @Param PP_REGION query string false "User (payer) region code by ISO 3166-1 (2 chars) for check project packages. If this field not send, then user region will be get from user ip"
// @Param PP_PAYER_EMAIL query string false "User (payer) email"
// @Param PP_PAYER_PHONE query string false "User (payer) phone"
// @Param PP_URL_VERIFY query string false "URL for payment data verification request to project. This field can be send if it allowed in project admin panel"
// @Param PP_URL_NOTIFY query string false "URL for payment notification request to project. This field can be send if it allowed in project admin panel"
// @Param PP_URL_SUCCESS query string false "URL for redirect user after successfully completed payment. This field can be send if it allowed in project admin panel"
// @Param PP_URL_FAIL query string false "URL for redirect user after failed payment. This field can be send if it allowed in project admin panel"
// @Param PP_SIGNATURE query string false "Signature of request to verify that the data has not been changed. This field not required, BUT we're recommend send this field always"
// @Param Other query string false "Any fields on the project side that do not match the names of the reserved fields"
// @Success 302 {string} html "Redirect user to form entering payment requisites"
// @Failure 400 {string} html "Redirect user to page with error description"
// @Failure 500 {string} html "Redirect user to page with error description"
// @Router /order/create [post]
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

// @Summary Create order with json request
// @Description Create a payment order use POST JSON request
// @Tags Payment Order
// @Accept json
// @Produce json
// @Param data body model.OrderScalar true "Order create data"
// @Success 200 {object} model.OrderUrl "Object with url to form entering payment requisites"
// @Failure 400 {object} model.Error "Object with error message"
// @Failure 500 {object} model.Error "Object with error message"
// @Router /api/v1/order [post]
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