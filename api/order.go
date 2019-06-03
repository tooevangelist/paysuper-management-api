package api

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/micro/go-micro"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/paysuper/paysuper-payment-link/proto"
	"net/http"
	"time"
)

const (
	orderFormTemplateName  = "order.html"
	orderInlineFormUrlMask = "%s://%s/order/%s"
	errorTemplateName      = "error.html"
)

type orderRoute struct {
	*Api
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
	route := &orderRoute{
		Api:            api,
		projectManager: manager.InitProjectManager(api.database, api.logger, api.billingService),
	}

	api.Http.GET("/order/:id", route.getOrderForm)
	api.Http.GET("/paylink/:id", route.getOrderForPaylink)
	api.Http.GET("/order/create", route.createFromFormData)
	api.Http.POST("/order/create", route.createFromFormData)

	api.Http.POST("/api/v1/order", route.createJson)

	api.Http.POST("/api/v1/payment", route.processCreatePayment)

	api.authUserRouteGroup.GET("/order", route.getOrders)
	api.authUserRouteGroup.GET("/order/:id", route.getOrderJson)

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

	var order *billing.Order

	// If request contain prepared order identifier than try to get order by this identifier
	if req.PspOrderUuid != "" {
		req1 := &grpc.IsOrderCanBePayingRequest{
			OrderId:   req.PspOrderUuid,
			ProjectId: req.ProjectId,
		}
		rsp1, err := r.billingService.IsOrderCanBePaying(context.TODO(), req1)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
		}

		if rsp1.Status != pkg.ResponseStatusOk {
			return echo.NewHTTPError(int(rsp1.Status), rsp1.Message)
		}

		order = rsp1.Item
	} else {
		order, err = r.billingService.OrderCreateProcess(context.TODO(), req)

		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
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

// Get full object with order data
// Route GET /api/v1/s/order/{id}
func (r *orderRoute) getOrderJson(ctx echo.Context) error {
	id := ctx.Param(requestParameterId)

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorIdIsEmpty)
	}

	if _, err := uuid.Parse(id); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorIdIsEmpty)
	}

	req := &grpc.GetOrderRequest{
		Id: id,
	}
	rsp, err := r.billingService.GetOrder(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errorMessageOrdersNotFound)
	}

	if rsp == nil {
		return echo.NewHTTPError(http.StatusNotFound, model.ResponseMessageNotFound)
	}

	return ctx.JSON(http.StatusOK, rsp)
}

// Get orders list
// Route GET /api/v1/s/order
func (r *orderRoute) getOrders(ctx echo.Context) error {
	req := &grpc.ListOrdersRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	if req.Limit <= 0 {
		req.Limit = LimitDefault
	}

	if req.Offset <= 0 {
		req.Offset = OffsetDefault
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.FindAllOrders(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errorMessageOrdersNotFound)
	}

	return ctx.JSON(http.StatusOK, rsp)
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
