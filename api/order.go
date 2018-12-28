package api

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/ProtocolONE/p1pay.api/payment_system"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/micro/go-micro"
	"net/http"
	"net/url"
)

const (
	orderFormTemplateName = "order.html"
)

type OrderApiV1 struct {
	*Api
	orderManager   *manager.OrderManager
	projectManager *manager.ProjectManager
	publisher      micro.Publisher
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
		Api: api,
		orderManager: manager.InitOrderManager(
			api.database,
			api.logger,
			api.geoDbReader,
			api.pspAccountingCurrencyA3,
			api.paymentSystemsSettings,
			api.publisher,
			api.centrifugoSecret,
		),
		projectManager: manager.InitProjectManager(api.database, api.logger),
	}

	api.Http.GET("/order/:id", oApiV1.getOrderForm)
	api.Http.GET("/order/create", oApiV1.createFromFormData)
	api.Http.POST("/order/create", oApiV1.createFromFormData)
	api.Http.POST("/api/v1/order", oApiV1.createJson)

	api.Http.POST("/api/v1/payment", oApiV1.processCreatePayment)

	api.accessRouteGroup.GET("/order", oApiV1.getOrders)
	api.accessRouteGroup.GET("/order/:id", oApiV1.getOrderJson)
	api.accessRouteGroup.GET("/order/revenue_dynamic/:period", oApiV1.getRevenueDynamic)
	api.accessRouteGroup.GET("/order/accounting_payment", oApiV1.getAccountingPaymentCalculation)

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
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
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
// @Success 200 {object} model.JsonOrderCreateResponse "Object which contain data to render payment form"
// @Failure 400 {object} model.Error "Object with error message"
// @Failure 500 {object} model.Error "Object with error message"
// @Router /api/v1/order [post]
func (oApiV1 *OrderApiV1) createJson(ctx echo.Context) error {
	order := &model.OrderScalar{
		CreateOrderIp: ctx.RealIP(),
		IsJsonRequest: true,
	}

	if err := (&OrderJsonBinder{}).Bind(order, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	if err := oApiV1.validate.Struct(order); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
	}

	nOrder, err := oApiV1.orderManager.Process(order)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	oh := &manager.OrderHttp{
		Host:   ctx.Request().Host,
		Scheme: oApiV1.httpScheme,
	}

	jo, err := oApiV1.orderManager.JsonOrderCreatePostProcess(nOrder, oh)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, jo)
}

func (oApiV1 *OrderApiV1) getOrderForm(ctx echo.Context) error {
	id := ctx.Param(model.OrderFilterFieldId)

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageInvalidRequestData)
	}

	o := oApiV1.orderManager.FindById(id)

	if o == nil {
		return echo.NewHTTPError(http.StatusNotFound, model.ResponseMessageNotFound)
	}

	oh := &manager.OrderHttp{
		Host:   ctx.Request().Host,
		Scheme: oApiV1.httpScheme,
	}

	jo, err := oApiV1.orderManager.JsonOrderCreatePostProcess(o, oh)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.Render(http.StatusOK, orderFormTemplateName, map[string]interface{}{"Order": jo})
}

// @Summary Get order data
// @Description Get full object with order data
// @Tags Payment Order
// @Accept json
// @Produce json
// @Param id path string true "Order unique identifier"
// @Success 200 {object} model.Order "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 403 {object} model.Error "Access denied"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Object with error message"
// @Router /api/v1/s/order/{id} [get]
func (oApiV1 *OrderApiV1) getOrderJson(ctx echo.Context) error {
	id := ctx.Param("id")

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageInvalidRequestData)
	}

	p, merchant, err := oApiV1.projectManager.FilterProjects(oApiV1.Merchant.Identifier, []bson.ObjectId{})

	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err)
	}

	params := &manager.FindAll{
		Values:   url.Values{"id": []string{id}},
		Projects: p,
		Merchant: merchant,
		Limit:    oApiV1.GetParams.limit,
		Offset:   oApiV1.GetParams.offset,
	}

	pOrders, err := oApiV1.orderManager.FindAll(params)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	if pOrders.Count == 0 {
		return echo.NewHTTPError(http.StatusNotFound, model.ResponseMessageNotFound)
	}

	return ctx.JSON(http.StatusOK, pOrders.Items[0])
}

// @Summary Get orders
// @Description Get orders list
// @Tags Payment Order
// @Accept json
// @Produce json
// @Param id query string false "order unique identifier"
// @Param project query array false "list of projects to get orders filtered by they"
// @Param payment_method query array false "list of payment methods to get orders filtered by they"
// @Param country query array false "list of payer countries to get orders filtered by they"
// @Param status query array false "list of orders statuses to get orders filtered by they"
// @Param account query string false "payer account on the any side of payment process. for example it may be account in project, account in payment system, payer email and etc"
// @Param pm_date_from query integer false "start date when payment was closed to get orders filtered by they"
// @Param pm_date_to query integer false "end date when payment was closed to get orders filtered by they"
// @Param project_date_from query integer false "start date when payment was created to get orders filtered by they"
// @Param project_date_to query integer false "end date when payment was closed in project to get orders filtered by they"
// @Param limit query integer false "maximum number of returning orders. default value is 100"
// @Param offset query integer false "offset from which you want to return the list of orders. default value is 0"
// @Param sort query array false "fields list for sorting"
// @Success 200 {object} model.OrderPaginate "OK"
// @Failure 404 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Object with error message"
// @Router /api/v1/s/order [get]
func (oApiV1 *OrderApiV1) getOrders(ctx echo.Context) error {
	values := ctx.QueryParams()

	var fp []bson.ObjectId

	if fProjects, ok := values[model.OrderFilterFieldProjects]; ok {
		for _, project := range fProjects {
			if bson.IsObjectIdHex(project) == false {
				return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageProjectIdIsInvalid)
			}

			fp = append(fp, bson.ObjectIdHex(project))
		}
	}

	p, merchant, err := oApiV1.projectManager.FilterProjects(oApiV1.Merchant.Identifier, fp)

	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err)
	}

	params := &manager.FindAll{
		Values:   values,
		Projects: p,
		Merchant: merchant,
		Limit:    oApiV1.GetParams.limit,
		Offset:   oApiV1.GetParams.offset,
		SortBy:   oApiV1.GetParams.sort,
	}

	pOrders, err := oApiV1.orderManager.FindAll(params)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return ctx.JSON(http.StatusOK, pOrders)
}

// @Summary Create payment
// @Description Create payment by order
// @Tags Payment Order
// @Accept json
// @Produce json
// @Param data body model.OrderCreatePaymentRequest true "data to create payment"
// @Success 200 {object} payment_system.PaymentResponse "contain url to redirect user"
// @Failure 400 {object} payment_system.PaymentResponse "contain error description about data validation error"
// @Failure 402 {object} payment_system.PaymentResponse "contain error description about error on payment system side"
// @Failure 500 {object} payment_system.PaymentResponse "contain error description about error on PSP (P1) side"
// @Router /api/v1/payment [post]
func (oApiV1 *OrderApiV1) processCreatePayment(ctx echo.Context) error {
	data := make(map[string]string)

	if err := ctx.Bind(&data); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": model.ResponseMessageInvalidRequestData})
	}

	resp := oApiV1.orderManager.ProcessCreatePayment(data, oApiV1.PaymentSystemConfig)

	var httpStatus int

	switch resp.Status {
	case payment_system.PaymentStatusErrorValidation:
		httpStatus = http.StatusBadRequest
		break
	case payment_system.PaymentStatusErrorSystem:
		httpStatus = http.StatusInternalServerError
		break
	case payment_system.CreatePaymentStatusErrorPaymentSystem:
		httpStatus = http.StatusPaymentRequired
		break
	default:
		httpStatus = http.StatusOK
	}

	return ctx.JSON(httpStatus, resp)
}

// @Summary Get revenue dynamics
// @Description Get revenue dynamics by merchant or project
// @Tags Payment Order
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param period path string true "period to group revenue dynamics data. now allowed next values: hour, day, week, month, year"
// @Param from query int true "period start in unix timestamp"
// @Param to query int true "period end in unix timestamp"
// @Param project query array false "list of projects to calculate dynamics of revenue"
// @Success 200 {object} model.RevenueDynamicResult "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 403 {object} model.Error "Access denied"
// @Failure 500 {object} model.Error "Object with error message"
// @Router /api/v1/s/order/revenue_dynamic/{period} [get]
func (oApiV1 *OrderApiV1) getRevenueDynamic(ctx echo.Context) error {
	rdr := &model.RevenueDynamicRequest{}

	if err := (&OrderRevenueDynamicRequestBinder{}).Bind(rdr, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	pMap, _, err := oApiV1.projectManager.FilterProjects(oApiV1.Merchant.Identifier, rdr.Project)

	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	if len(rdr.Project) <= 0 {
		rdr.SetProjectsFromMap(pMap)
	}

	res, err := oApiV1.orderManager.GetRevenueDynamic(rdr)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Summary Get accounting payment amounts by accounting period of merchant
// @Description accounting payment by accounting period of merchant
// @Tags Payment Order
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param from query int true "period start in unix timestamp"
// @Param to query int true "period end in unix timestamp"
// @Success 200 {object} model.AccountingPayment "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 403 {object} model.Error "Access denied"
// @Failure 500 {object} model.Error "Object with error message"
// @Router /api/v1/s/order/accounting_payment [get]
func (oApiV1 *OrderApiV1) getAccountingPaymentCalculation(ctx echo.Context) error {
	rdr := &model.RevenueDynamicRequest{}

	if err := (&OrderAccountingPaymentRequestBinder{}).Bind(rdr, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	res, err := oApiV1.orderManager.GetAccountingPayment(rdr, oApiV1.Merchant.Identifier)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, res)
}
