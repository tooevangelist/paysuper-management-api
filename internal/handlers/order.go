package handlers

import (
	"context"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	u "github.com/PuerkitoBio/purell"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	orderIdPath              = "/order/:order_id"
	paylinkIdPath            = "/paylink/:id"
	orderCreatePath          = "/order/create"
	orderPath                = "/order"
	paymentPath              = "/payment"
	orderRefundsPath         = "/order/:order_id/refunds"
	orderRefundsIdsPath      = "/order/:order_id/refunds/:refund_id"
	orderReplaceCodePath     = "/order/:order_id/replace_code"
	orderLanguagePath        = "/orders/:order_id/language"
	orderCustomerPath        = "/orders/:order_id/customer"
	orderBillingAddressPath  = "/orders/:order_id/billing_address"
	orderNotifySalesPath     = "/orders/:order_id/notify_sale"
	orderNotifyNewRegionPath = "/orders/:order_id/notify_new_region"
	orderPlatformPath        = "/orders/:order_id/platform"
	orderReceiptPath         = "/orders/receipt/:receipt_id/:order_id"
)

const (
	errorTemplateName = "error.html"
)

type CreateOrderJsonProjectResponse struct {
	Id              string                    `json:"id"`
	PaymentFormUrl  string                    `json:"payment_form_url"`
	PaymentFormData *grpc.PaymentFormJsonData `json:"payment_form_data,omitempty"`
}

type OrderListRefundsBinder struct {
	dispatch common.HandlerSet
	provider.LMT
	cfg common.Config
}

func (b *OrderListRefundsBinder) Bind(i interface{}, ctx echo.Context) error {
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)

	if err != nil {
		return err
	}

	structure := i.(*grpc.ListRefundsRequest)
	structure.OrderId = ctx.Param(common.RequestParameterOrderId)

	if structure.Limit <= 0 {
		structure.Limit = b.cfg.LimitDefault
	}

	return nil
}

type OrderRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewOrderRoute(set common.HandlerSet, cfg *common.Config) *OrderRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "OrderRoute"})
	return &OrderRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *OrderRoute) Route(groups *common.Groups) {
	groups.Common.GET(orderIdPath, h.getPaymentFormData)
	groups.Common.GET(paylinkIdPath, h.getOrderForPaylink)    // TODO: Need a test
	groups.Common.GET(orderCreatePath, h.createFromFormData)  // TODO: Need a test
	groups.Common.POST(orderCreatePath, h.createFromFormData) // TODO: Need a test
	groups.Common.POST(orderPath, h.createJson)               // TODO: Need a test
	groups.Common.POST(paymentPath, h.processCreatePayment)   // TODO: Need a test

	groups.Common.PATCH(orderLanguagePath, h.changeLanguage)
	groups.Common.PATCH(orderCustomerPath, h.changeCustomer)
	groups.Common.POST(orderBillingAddressPath, h.processBillingAddress)
	groups.Common.POST(orderNotifySalesPath, h.notifySale)
	groups.Common.POST(orderNotifyNewRegionPath, h.notifyNewRegion)
	groups.Common.POST(orderPlatformPath, h.changePlatform)

	groups.Common.GET(orderReceiptPath, h.getReceipt)

	groups.AuthUser.GET(orderPath, h.listOrdersPublic)
	groups.AuthUser.GET(orderIdPath, h.getOrderPublic) // TODO: Need a test

	groups.AuthUser.GET(orderRefundsPath, h.listRefunds)
	groups.AuthUser.GET(orderRefundsIdsPath, h.getRefund)
	groups.AuthUser.POST(orderRefundsPath, h.createRefund)
	groups.SystemUser.PUT(orderReplaceCodePath, h.replaceCode)
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
func (h *OrderRoute) createFromFormData(ctx echo.Context) error {
	req := &billing.OrderCreateRequest{
		PayerIp: ctx.RealIP(),
		IsJson:  false,
	}

	if err := (&common.OrderFormBinder{}).Bind(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req.IssuerUrl = ctx.Request().Header.Get(common.HeaderReferer)

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	orderResponse, err := h.dispatch.Services.Billing.OrderCreateProcess(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "OrderCreateProcess")
	}

	if orderResponse.Status != http.StatusOK {
		return echo.NewHTTPError(int(orderResponse.Status), orderResponse.Message)
	}

	rUrl := "/order/" + orderResponse.Item.Id

	return ctx.Redirect(http.StatusFound, rUrl)
}

// Create order from json request.
// Order can be create:
// 1) By project host2host request with sending user (customer) information.
// 2) By payment form client request with sending prepare created user (customer) identification token.
// 3) By payment form client request without anything user identification information.
func (h *OrderRoute) createJson(ctx echo.Context) error {
	req := &billing.OrderCreateRequest{}
	err := (&common.OrderJsonBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	// If request contain user object then paysuper must check request signature
	if req.User != nil {
		httpErr := common.CheckProjectAuthRequestSignature(h.dispatch, ctx, req.ProjectId)

		if httpErr != nil {
			return httpErr
		}
	}

	ctxReq := ctx.Request().Context()
	req.IssuerUrl = ctx.Request().Header.Get(common.HeaderReferer)

	var (
		order         *billing.Order
		orderResponse *grpc.OrderCreateProcessResponse
	)

	// If request contain prepared order identifier than try to get order by this identifier
	if req.PspOrderUuid != "" {
		req1 := &grpc.IsOrderCanBePayingRequest{
			OrderId:   req.PspOrderUuid,
			ProjectId: req.ProjectId,
		}
		rsp1, err := h.dispatch.Services.Billing.IsOrderCanBePaying(ctxReq, req1)

		if err != nil {
			return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "IsOrderCanBePaying")
		}

		if rsp1.Status != pkg.ResponseStatusOk {
			return echo.NewHTTPError(int(rsp1.Status), rsp1.Message)
		}

		order = rsp1.Item
	} else {
		orderResponse, err = h.dispatch.Services.Billing.OrderCreateProcess(ctxReq, req)

		if err != nil {
			return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "OrderCreateProcess")
		}

		if orderResponse.Status != http.StatusOK {
			return echo.NewHTTPError(int(orderResponse.Status), orderResponse.Message)
		}

		order = orderResponse.Item
	}

	response := &CreateOrderJsonProjectResponse{
		Id:             order.Uuid,
		PaymentFormUrl: h.cfg.OrderInlineFormUrlMask + "?order_id=" + order.Uuid,
	}

	return ctx.JSON(http.StatusOK, response)
}

func (h *OrderRoute) getPaymentFormData(ctx echo.Context) error {
	id := ctx.Param(common.RequestParameterOrderId)

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	cookie, err := ctx.Cookie(common.CustomerTokenCookiesName)

	req := &grpc.PaymentFormJsonDataRequest{
		OrderId: id,
		Scheme:  h.cfg.HttpScheme,
		Host:    ctx.Request().Host,
		Locale:  ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		Ip:      ctx.RealIP(),
		Referer: ctx.Request().Header.Get(common.HeaderReferer),
	}

	h.L().Info("debug", logger.PairArgs("X-Real-IP", ctx.Request().Header.Get(echo.HeaderXRealIP)))
	h.L().Info("debug", logger.PairArgs("X-Forwarded-For", ctx.Request().Header.Get(echo.HeaderXForwardedFor)))
	h.L().Info("debug", logger.PairArgs("IP Echo", ctx.RealIP()))

	if err == nil && cookie != nil && cookie.Value != "" {
		req.Cookie = cookie.Value
	}

	res, err := h.dispatch.Services.Billing.PaymentFormJsonDataProcess(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentFormJsonDataProcess")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// Create order from payment link and redirect to order payment form
func (h *OrderRoute) getOrderForPaylink(ctx echo.Context) error {
	paylinkId := ctx.Param(common.RequestParameterId)
	ctxReq := ctx.Request().Context()

	go func() {
		req := &grpc.PaylinkRequestById{Id: paylinkId}
		// call with background context to prevent request abandoning when redirect will bw returned in responce below
		_, err := h.dispatch.Services.Billing.IncrPaylinkVisits(context.Background(), req)

		if err != nil {
			common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "IncrPaylinkVisits", req)
		}
	}()

	qParams := ctx.QueryParams()

	oReq := &billing.OrderCreateByPaylink{
		PaylinkId:   paylinkId,
		PayerIp:     ctx.RealIP(),
		IssuerUrl:   ctx.Request().Header.Get(common.HeaderReferer),
		UtmSource:   qParams.Get(common.QueryParameterNameUtmSource),
		UtmMedium:   qParams.Get(common.QueryParameterNameUtmMedium),
		UtmCampaign: qParams.Get(common.QueryParameterNameUtmCampaign),
		IsEmbedded:  false,
	}

	orderResponse, err := h.dispatch.Services.Billing.OrderCreateByPaylink(ctxReq, oReq)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "OrderCreateByPaylink", oReq)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}

	if orderResponse.Status != http.StatusOK {
		return echo.NewHTTPError(int(orderResponse.Status), orderResponse.Message)
	}

	qParams.Set("order_id", orderResponse.Item.Uuid)

	inlineFormRedirectUrl := h.cfg.OrderInlineFormUrlMask + "?" + qParams.Encode()

	inlineFormRedirectUrl, err = u.NormalizeURLString(inlineFormRedirectUrl, u.FlagsUsuallySafeGreedy|u.FlagRemoveDuplicateSlashes)
	if err != nil {
		h.L().Error("NormalizeURLString failed", logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	return ctx.Redirect(http.StatusFound, inlineFormRedirectUrl)
}

// @Description Get order by id
// @Example curl -X GET -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  https://api.paysuper.online/admin/api/v1/order/%order_id_here%
func (h *OrderRoute) getOrderPublic(ctx echo.Context) error {
	req := &grpc.GetOrderRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetOrderPublic(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetOrderPublic")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Get orders list
// @Example curl -X GET -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  https://api.paysuper.online/admin/api/v1/order?project[]=%project_identifier_here%
func (h *OrderRoute) listOrdersPublic(ctx echo.Context) error {
	req := &grpc.ListOrdersRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if req.Limit <= 0 {
		req.Limit = h.cfg.LimitDefault
	}

	if req.Offset <= 0 {
		req.Offset = h.cfg.OffsetDefault
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.FindAllOrdersPublic(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "FindAllOrdersPublic")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// Create payment by order
// route POST /api/v1/payment
func (h *OrderRoute) processCreatePayment(ctx echo.Context) error {
	data := make(map[string]string)
	err := (&common.PaymentCreateProcessBinder{}).Bind(data, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req := &grpc.PaymentCreateRequest{
		Data:           data,
		AcceptLanguage: ctx.Request().Header.Get(common.HeaderAcceptLanguage),
		UserAgent:      ctx.Request().Header.Get(common.HeaderUserAgent),
		Ip:             ctx.RealIP(),
	}
	res, err := h.dispatch.Services.Billing.PaymentCreateProcess(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentCreateProcess")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	body := map[string]interface{}{
		"redirect_url":  res.RedirectUrl,
		"need_redirect": res.NeedRedirect,
	}

	return ctx.JSON(http.StatusOK, body)
}

func (h *OrderRoute) getRefund(ctx echo.Context) error {
	req := &grpc.GetRefundRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetRefund(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetRefund")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OrderRoute) listRefunds(ctx echo.Context) error {
	req := &grpc.ListRefundsRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ListRefunds(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ListRefunds")
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *OrderRoute) replaceCode(ctx echo.Context) error {
	req := &grpc.ChangeCodeInOrderRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.OrderId = ctx.Param("order_id")
	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res := &grpc.ChangeCodeInOrderResponse{}

	res, err := h.dispatch.Services.Billing.ChangeCodeInOrder(ctx.Request().Context(), req)
	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ChangeCodeInOrder")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Order)
}

func (h *OrderRoute) createRefund(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	req := &grpc.CreateRefundRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.OrderId = ctx.Param(common.RequestParameterOrderId)
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	req.CreatorId = authUser.Id
	res, err := h.dispatch.Services.Billing.CreateRefund(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "CreateRefund")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusCreated, res.Item)
}

func (h *OrderRoute) changeLanguage(ctx echo.Context) error {
	orderId := ctx.Param(common.RequestParameterOrderId)

	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorIncorrectOrderId)
	}

	req := &grpc.PaymentFormUserChangeLangRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.AcceptLanguage = ctx.Request().Header.Get(common.HeaderAcceptLanguage)
	req.UserAgent = ctx.Request().Header.Get(common.HeaderUserAgent)
	req.Ip = ctx.RealIP()
	req.OrderId = orderId
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.PaymentFormLanguageChanged(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentFormLanguageChanged")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OrderRoute) changeCustomer(ctx echo.Context) error {
	orderId := ctx.Param(common.RequestParameterOrderId)

	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorIncorrectOrderId)
	}

	req := &grpc.PaymentFormUserChangePaymentAccountRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.AcceptLanguage = ctx.Request().Header.Get(common.HeaderAcceptLanguage)
	req.UserAgent = ctx.Request().Header.Get(common.HeaderUserAgent)
	req.Ip = ctx.RealIP()
	req.OrderId = orderId
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.PaymentFormPaymentAccountChanged(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentFormPaymentAccountChanged")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OrderRoute) processBillingAddress(ctx echo.Context) error {
	orderId := ctx.Param(common.RequestParameterOrderId)

	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorIncorrectOrderId)
	}

	req := &grpc.ProcessBillingAddressRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.OrderId = orderId
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ProcessBillingAddress(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ProcessBillingAddress")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

/*
Switching sales notifications for order customer
POST /api/v1/orders/:order_id/notify_sale
@Param [email] string
@Param enableNotification string true|false
*/
func (h *OrderRoute) notifySale(ctx echo.Context) error {
	orderId := ctx.Param(common.RequestParameterOrderId)

	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorIncorrectOrderId)
	}

	req := &grpc.SetUserNotifyRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.OrderUuid = orderId
	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	_, err = h.dispatch.Services.Billing.SetUserNotifySales(ctx.Request().Context(), req)
	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "SetUserNotifySales")
	}

	return ctx.NoContent(http.StatusNoContent)
}

/*
Switching notifications customer about new regions allowed to make payments
POST /api/v1/orders/:order_uuid/notify_new_region
@Param [email] string
@Param enableNotification string true|false
*/
func (h *OrderRoute) notifyNewRegion(ctx echo.Context) error {
	orderUuid := ctx.Param(common.RequestParameterOrderId)

	if orderUuid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorIncorrectOrderId)
	}

	req := &grpc.SetUserNotifyRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.OrderUuid = orderUuid
	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	_, err = h.dispatch.Services.Billing.SetUserNotifyNewRegion(ctx.Request().Context(), req)
	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "SetUserNotifyNewRegion")
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *OrderRoute) changePlatform(ctx echo.Context) error {
	orderId := ctx.Param(common.RequestParameterOrderId)

	if orderId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorIncorrectOrderId)
	}

	req := &grpc.PaymentFormUserChangePlatformRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.OrderId = orderId
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.PaymentFormPlatformChanged(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "PaymentFormPlatformChanged")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
}
func (h *OrderRoute) getReceipt(ctx echo.Context) error {
	orderId := ctx.Param(common.RequestParameterOrderId)

	if _, err := uuid.Parse(orderId); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	receiptId := ctx.Param(common.RequestParameterReceiptId)

	if _, err := uuid.Parse(receiptId); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req := &grpc.OrderReceiptRequest{OrderId: orderId, ReceiptId: receiptId}
	res, err := h.dispatch.Services.Billing.OrderReceipt(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "OrderReceipt")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Receipt)
}
