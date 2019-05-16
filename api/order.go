package api

import (
	"context"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/micro/go-micro"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/paysuper/paysuper-payment-link/proto"
	"net/http"
	"net/url"
	"time"
)

const (
	orderFormTemplateName  = "order.html"
	orderInlineFormUrlMask = "%s://%s/order/%s"
	errorTemplateName      = "error.html"
)

type orderRoute struct {
	*Api
	orderManager   *manager.OrderManager
	projectManager *manager.ProjectManager
	publisher      micro.Publisher
}

type CreateOrderJsonProjectResponse struct {
	Id              string                            `json:"id"`
	PaymentFormUrl  string                            `json:"payment_form_url"`
	PaymentFormData *grpc.PaymentFormJsonDataResponse `json:"payment_form_data,omitempty"`
}

type OrderListRefundsBinder struct{}

func (b *OrderListRefundsBinder) Bind(i interface{}, ctx echo.Context) error {
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)

	if err != nil {
		return err
	}

	structure := i.(*grpc.ListRefundsRequest)
	structure.OrderId = ctx.Param(requestParameterOrderId)

	if structure.Limit <= 0 {
		structure.Limit = LimitDefault
	}

	return nil
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
	route := orderRoute{
		Api: api,
		orderManager: manager.InitOrderManager(
			api.database,
			api.logger,
			api.notifierPub,
			api.repository,
			api.geoService,
		),
		projectManager: manager.InitProjectManager(api.database, api.logger, api.billingService),
	}

	api.Http.GET("/order/:id", route.getOrderForm)
	api.Http.GET("/paylink/:id", route.getOrderForPaylink)
	api.Http.GET("/order/create", route.createFromFormData)
	api.Http.POST("/order/create", route.createFromFormData)

	api.Http.POST("/api/v1/order", route.createJson)

	api.Http.POST("/api/v1/payment", route.processCreatePayment)

	api.authUserRouteGroup.GET("/order", route.getOrders)

	api.accessRouteGroup.GET("/order/:id", route.getOrderJson)
	api.accessRouteGroup.GET("/order/revenue_dynamic/:period", route.getRevenueDynamic)
	api.accessRouteGroup.GET("/order/accounting_payment", route.getAccountingPaymentCalculation)

	api.authUserRouteGroup.GET("/order/:order_id/refunds", route.listRefunds)
	api.authUserRouteGroup.GET("/order/:order_id/refunds/:refund_id", route.getRefund)
	api.authUserRouteGroup.POST("/order/:order_id/refunds", route.createRefund)

	api.Http.PATCH("/api/v1/orders/:order_id/language", route.changeLanguage)
	api.Http.PATCH("/api/v1/orders/:order_id/customer", route.changeCustomer)
	api.Http.POST("/api/v1/orders/:order_id/billing_address", route.processBillingAddress)

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
func (r *orderRoute) createFromFormData(ctx echo.Context) error {
	req := &billing.OrderCreateRequest{
		PayerIp: ctx.RealIP(),
		IsJson:  false,
	}

	if err := (&OrderFormBinder{}).Bind(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Request data invalid")
	}

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
	}

	order, err := r.billingService.OrderCreateProcess(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	rUrl := "/order/" + order.Id

	return ctx.Redirect(http.StatusFound, rUrl)
}

// Create order from json request.
// Order can be create:
// 1) By project host2host request with sending user (customer) information.
// 2) By payment form client request with sending prepare created user (customer) identification token.
// 3) By payment form client request without anything user identification information.
func (r *orderRoute) createJson(ctx echo.Context) error {
	req := &billing.OrderCreateRequest{}
	err := (&OrderJsonBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
	}

	// If request contain user object then paysuper must check request signature
	if req.User != nil {
		err = r.checkProjectAuthRequestSignature(ctx, req.ProjectId)

		if err != nil {
			return err
		}
	}

	order, err := r.billingService.OrderCreateProcess(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	response := &CreateOrderJsonProjectResponse{
		Id:             order.Uuid,
		PaymentFormUrl: fmt.Sprintf(pkg.OrderInlineFormUrlMask, r.httpScheme, ctx.Request().Host, order.Uuid),
	}

	// If not production environment then return data to payment form
	if r.isProductionEnvironment() != true {
		req2 := &grpc.PaymentFormJsonDataRequest{
			OrderId: order.Uuid,
			Scheme:  r.httpScheme,
			Host:    ctx.Request().Host,
			Locale:  ctx.Request().Header.Get(HeaderAcceptLanguage),
			Ip:      ctx.RealIP(),
		}
		rsp2, err := r.billingService.PaymentFormJsonDataProcess(context.TODO(), req2)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		response.PaymentFormData = rsp2
	}

	return ctx.JSON(http.StatusOK, response)
}

func (r *orderRoute) getOrderForm(ctx echo.Context) error {
	id := ctx.Param(requestParameterId)

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	cookie, err := ctx.Cookie(CustomerTokenCookiesName)

	req := &grpc.PaymentFormJsonDataRequest{
		OrderId: id,
		Scheme:  r.httpScheme,
		Host:    ctx.Request().Host,
		Locale:  ctx.Request().Header.Get(HeaderAcceptLanguage),
		Ip:      ctx.RealIP(),
	}

	if err == nil && cookie != nil && cookie.Value != "" {
		req.Cookie = cookie.Value
	}

	rsp, err := r.billingService.PaymentFormJsonDataProcess(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if rsp.Cookie != "" && rsp.Cookie != req.Cookie {
		cookie := new(http.Cookie)
		cookie.Name = CustomerTokenCookiesName
		cookie.Value = rsp.Cookie
		cookie.Expires = time.Now().Add(time.Second * CustomerTokenCookiesLifetime)
		cookie.HttpOnly = true
		ctx.SetCookie(cookie)
	}

	return ctx.Render(http.StatusOK, orderFormTemplateName, map[string]interface{}{"Order": rsp})
}

// Create order from payment link and redirect to order payment form
func (r *orderRoute) getOrderForPaylink(ctx echo.Context) error {

	paylinkId := ctx.Param(requestParameterId)

	req := &paylink.PaylinkRequest{
		Id: paylinkId,
	}

	err := r.validate.Struct(req)
	if err != nil {
		r.logError("Cannot validate request", []interface{}{"error", err.Error(), "request", req})
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}

	pl, err := r.paylinkService.GetPaylink(context.Background(), req)
	if err != nil {
		return ctx.Render(http.StatusNotFound, errorTemplateName, map[string]interface{}{})
	}

	oReq := &billing.OrderCreateRequest{
		ProjectId: pl.ProjectId,
		PayerIp:   ctx.RealIP(),
		Products:  pl.Products,
		PrivateMetadata: map[string]string{
			"PaylinkId": paylinkId,
		},
	}
	params := ctx.QueryParams()
	if v, ok := params[requestParameterUtmSource]; ok {
		oReq.PrivateMetadata[requestParameterUtmSource] = v[0]
	}
	if v, ok := params[requestParameterUtmMedium]; ok {
		oReq.PrivateMetadata[requestParameterUtmMedium] = v[0]
	}
	if v, ok := params[requestParameterUtmCampaign]; ok {
		oReq.PrivateMetadata[requestParameterUtmCampaign] = v[0]
	}

	order, err := r.billingService.OrderCreateProcess(context.Background(), oReq)
	if err != nil {
		r.logError("Cannot create order for paylink", []interface{}{"error", err.Error(), "request", req})
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}

	inlineFormRedirectUrl := fmt.Sprintf(orderInlineFormUrlMask, r.httpScheme, ctx.Request().Host, order.Uuid)
	qs := ctx.QueryString()
	if qs != "" {
		inlineFormRedirectUrl += "?" + qs
	}

	go func() {
		_, err := r.paylinkService.IncrPaylinkVisits(context.Background(), &paylink.PaylinkRequest{
			Id: paylinkId,
		})
		if err != nil {
			r.logError("Cannot update paylink stat", []interface{}{"error", err.Error(), "request", req})
		}
	}()
	return ctx.Redirect(http.StatusFound, inlineFormRedirectUrl)
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
func (r *orderRoute) getOrderJson(ctx echo.Context) error {
	id := ctx.Param("id")

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageInvalidRequestData)
	}

	p, merchant, err := r.projectManager.FilterProjects(r.Merchant.Identifier, []bson.ObjectId{})

	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err)
	}

	params := &manager.FindAll{
		Values:   url.Values{"id": []string{id}},
		Projects: p,
		Merchant: merchant,
		Limit:    r.GetParams.limit,
		Offset:   r.GetParams.offset,
	}

	pOrders, err := r.orderManager.FindAll(params)

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
// @Param quick_filter query string false "string for full text search in quick filter"
// @Param limit query integer false "maximum number of returning orders. default value is 100"
// @Param offset query integer false "offset from which you want to return the list of orders. default value is 0"
// @Param sort query array false "fields list for sorting"
// @Success 200 {object} model.OrderPaginate "OK"
// @Failure 404 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Object with error message"
// @Router /api/v1/s/order [get]
func (r *orderRoute) getOrders(ctx echo.Context) error {
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

	rsp, err := r.billingService.GetMerchantBy(context.TODO(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	p, _, err := r.projectManager.FilterProjects(rsp.Item.Id, fp)

	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	params := &manager.FindAll{
		Values:   values,
		Projects: p,
		Merchant: rsp.Item,
		Limit:    r.GetParams.limit,
		Offset:   r.GetParams.offset,
		SortBy:   r.GetParams.sort,
	}

	pOrders, err := r.orderManager.FindAll(params)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, pOrders)
}

// Create payment by order
// route POST /api/v1/payment
func (r *orderRoute) processCreatePayment(ctx echo.Context) error {
	data := make(map[string]string)
	err := (&PaymentCreateProcessBinder{}).Bind(data, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageInvalidRequestData)
	}

	req := &grpc.PaymentCreateRequest{
		Data:           data,
		AcceptLanguage: ctx.Request().Header.Get(HeaderAcceptLanguage),
		UserAgent:      ctx.Request().Header.Get(HeaderUserAgent),
		Ip:             ctx.RealIP(),
	}
	rsp, err := r.billingService.PaymentCreateProcess(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, model.ResponseMessageUnknownError)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	body := map[string]interface{}{
		"redirect_url":  rsp.RedirectUrl,
		"need_redirect": rsp.NeedRedirect,
	}

	return ctx.JSON(http.StatusOK, body)
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
func (r *orderRoute) getRevenueDynamic(ctx echo.Context) error {
	rdr := &model.RevenueDynamicRequest{}

	if err := (&OrderRevenueDynamicRequestBinder{}).Bind(rdr, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	pMap, _, err := r.projectManager.FilterProjects(r.Merchant.Identifier, rdr.Project)

	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	if len(rdr.Project) <= 0 {
		rdr.SetProjectsFromMap(pMap)
	}

	res, err := r.orderManager.GetRevenueDynamic(rdr)

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
func (r *orderRoute) getAccountingPaymentCalculation(ctx echo.Context) error {
	rdr := &model.RevenueDynamicRequest{}

	if err := (&OrderAccountingPaymentRequestBinder{}).Bind(rdr, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	res, err := r.orderManager.GetAccountingPayment(rdr, r.Merchant.Identifier)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, res)
}

func (r *orderRoute) getRefund(ctx echo.Context) error {
	req := &grpc.GetRefundRequest{
		OrderId:  ctx.Param(requestParameterOrderId),
		RefundId: ctx.Param(requestParameterRefundId),
	}

	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.GetRefund(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *orderRoute) listRefunds(ctx echo.Context) error {
	req := &grpc.ListRefundsRequest{}
	err := (&OrderListRefundsBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ListRefunds(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *orderRoute) createRefund(ctx echo.Context) error {
	req := &grpc.CreateRefundRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	req.OrderId = ctx.Param(requestParameterOrderId)
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	req.CreatorId = r.authUser.Id
	rsp, err := r.billingService.CreateRefund(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusCreated, rsp.Item)
}

func (r *orderRoute) changeLanguage(ctx echo.Context) error {
	orderId := ctx.Param(requestParameterOrderId)

	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectOrderId)
	}

	req := &grpc.PaymentFormUserChangeLangRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	req.AcceptLanguage = ctx.Request().Header.Get(HeaderAcceptLanguage)
	req.UserAgent = ctx.Request().Header.Get(HeaderUserAgent)
	req.Ip = ctx.RealIP()
	req.OrderId = orderId
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.PaymentFormLanguageChanged(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *orderRoute) changeCustomer(ctx echo.Context) error {
	orderId := ctx.Param(requestParameterOrderId)

	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectOrderId)
	}

	req := &grpc.PaymentFormUserChangePaymentAccountRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	req.AcceptLanguage = ctx.Request().Header.Get(HeaderAcceptLanguage)
	req.UserAgent = ctx.Request().Header.Get(HeaderUserAgent)
	req.Ip = ctx.RealIP()
	req.OrderId = orderId
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.PaymentFormPaymentAccountChanged(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *orderRoute) processBillingAddress(ctx echo.Context) error {
	orderId := ctx.Param(requestParameterOrderId)

	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectOrderId)
	}

	req := &grpc.ProcessBillingAddressRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	req.OrderId = orderId
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ProcessBillingAddress(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}
