package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type paymentCostRoute struct {
	*Api
}

func (api *Api) InitPaymentCostRoutes() *Api {
	paymentCostApiV1 := &paymentCostRoute{
		Api: api,
	}

	api.authUserRouteGroup.GET("/payment_costs/channel/system/all", paymentCostApiV1.getAllPaymentChannelCostSystem)
	api.authUserRouteGroup.GET("/payment_costs/channel/merchant/all", paymentCostApiV1.getAllPaymentChannelCostMerchant)
	api.authUserRouteGroup.GET("/payment_costs/money_back/system/all", paymentCostApiV1.getAllMoneyBackCostSystem)
	api.authUserRouteGroup.GET("/payment_costs/money_back/merchant/all", paymentCostApiV1.getAllMoneyBackCostMerchant)

	api.authUserRouteGroup.GET("/payment_costs/channel/system", paymentCostApiV1.getPaymentChannelCostSystem)
	api.authUserRouteGroup.GET("/payment_costs/channel/merchant", paymentCostApiV1.getPaymentChannelCostMerchant)
	api.authUserRouteGroup.GET("/payment_costs/money_back/system", paymentCostApiV1.getMoneyBackCostSystem)
	api.authUserRouteGroup.GET("/payment_costs/money_back/merchant", paymentCostApiV1.getMoneyBackCostMerchant)

	api.authUserRouteGroup.DELETE("/payment_costs/channel/system/:id", paymentCostApiV1.deletePaymentChannelCostSystem)
	api.authUserRouteGroup.DELETE("/payment_costs/channel/merchant/:id", paymentCostApiV1.deletePaymentChannelCostMerchant)
	api.authUserRouteGroup.DELETE("/payment_costs/money_back/system/:id", paymentCostApiV1.deleteMoneyBackCostSystem)
	api.authUserRouteGroup.DELETE("/payment_costs/money_back/merchant/:id", paymentCostApiV1.deleteMoneyBackCostMerchant)

	api.authUserRouteGroup.POST("/payment_costs/channel/system", paymentCostApiV1.setPaymentChannelCostSystem)
	api.authUserRouteGroup.POST("/payment_costs/channel/merchant", paymentCostApiV1.setPaymentChannelCostMerchant)
	api.authUserRouteGroup.POST("/payment_costs/money_back/system", paymentCostApiV1.setMoneyBackCostSystem)
	api.authUserRouteGroup.POST("/payment_costs/money_back/merchant", paymentCostApiV1.setMoneyBackCostMerchant)

	api.authUserRouteGroup.PUT("/payment_costs/channel/system/:id", paymentCostApiV1.setPaymentChannelCostSystem)
	api.authUserRouteGroup.PUT("/payment_costs/channel/merchant/:id", paymentCostApiV1.setPaymentChannelCostMerchant)
	api.authUserRouteGroup.PUT("/payment_costs/money_back/system/:id", paymentCostApiV1.setMoneyBackCostSystem)
	api.authUserRouteGroup.PUT("/payment_costs/money_back/merchant/:id", paymentCostApiV1.setMoneyBackCostMerchant)

	return api
}

// @Description Get PaymentChannelCostSystem
// @Example GET /admin/api/v1/payment_costs/channel/system?name=VISA&region=CIS&country=AZ
func (r *paymentCostRoute) getPaymentChannelCostSystem(ctx echo.Context) error {

	req := &billing.PaymentChannelCostSystemRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetPaymentChannelCostSystem(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description Get PaymentChannelCostMerchant
// @Example GET /admin/api/v1/payment_costs/channel/system?name=VISA&region=CIS&country=AZ&payoutCurrency=USD&amount=100
func (r *paymentCostRoute) getPaymentChannelCostMerchant(ctx echo.Context) error {

	req := &billing.PaymentChannelCostMerchantRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorIncorrectMerchantId)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetPaymentChannelCostMerchant(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description Get MoneyBackCostSystem
// @Example GET /admin/api/v1/payment_costs/money_back/system?name=VISA&region=CIS&country=AZ&payoutCurrency=USD&days=10&undoReason=chargeback&paymentStage=1
func (r *paymentCostRoute) getMoneyBackCostSystem(ctx echo.Context) error {

	req := &billing.MoneyBackCostSystemRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}
	res, err := r.billingService.GetMoneyBackCostSystem(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description Get MoneyBackCostSystem
// @Example GET /admin/api/v1/payment_costs/money_back/merchant?name=VISA&region=CIS&country=AZ&payoutCurrency=USD&days=10&undoReason=chargeback&paymentStage=1
func (r *paymentCostRoute) getMoneyBackCostMerchant(ctx echo.Context) error {

	req := &billing.MoneyBackCostMerchantRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorIncorrectMerchantId)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetMoneyBackCostMerchant(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description Delete PaymentChannelCostSystem
// @Example DELETE /admin/api/v1/payment_costs/channel/system/5be2d0b4b0b30d0007383ce6
func (r *paymentCostRoute) deletePaymentChannelCostSystem(ctx echo.Context) error {
	pcId := ctx.Param(requestParameterId)

	req := &billing.PaymentCostDeleteRequest{
		Id: pcId,
	}
	err := r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeletePaymentChannelCostSystem(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.NoContent(http.StatusNoContent)
}

// @Description Delete PaymentCostDeleteRequest
// @Example DELETE /admin/api/v1/payment_costs/channel/merchant/5be2d0b4b0b30d0007383ce6
func (r *paymentCostRoute) deletePaymentChannelCostMerchant(ctx echo.Context) error {
	pcId := ctx.Param(requestParameterId)

	req := &billing.PaymentCostDeleteRequest{
		Id: pcId,
	}
	err := r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeletePaymentChannelCostMerchant(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.NoContent(http.StatusNoContent)
}

// @Description Delete PaymentCostDeleteRequest
// @Example DELETE /admin/api/v1/payment_costs/money_back/system/5be2d0b4b0b30d0007383ce6
func (r *paymentCostRoute) deleteMoneyBackCostSystem(ctx echo.Context) error {
	pcId := ctx.Param(requestParameterId)

	req := &billing.PaymentCostDeleteRequest{
		Id: pcId,
	}
	err := r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeleteMoneyBackCostSystem(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.NoContent(http.StatusNoContent)
}

// @Description Delete PaymentCostDeleteRequest
// @Example DELETE /admin/api/v1/payment_costs/money_back/merchant/5be2d0b4b0b30d0007383ce6
func (r *paymentCostRoute) deleteMoneyBackCostMerchant(ctx echo.Context) error {
	pcId := ctx.Param(requestParameterId)

	req := &billing.PaymentCostDeleteRequest{
		Id: pcId,
	}
	err := r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeleteMoneyBackCostMerchant(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.NoContent(http.StatusNoContent)
}

// @Description create/update PaymentChannelCostSystem
// @Example POST /admin/api/v1/payment_costs/channel/system
// @Example PUT /admin/api/v1/payment_costs/channel/system/5be2d0b4b0b30d0007383ce6
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 1.01, "fix_amount": 2.34, "fix_amount_currency": "USD"}' \
//      https://api.paysuper.online/admin/api/v1/payment_costs/channel/system
func (r *paymentCostRoute) setPaymentChannelCostSystem(ctx echo.Context) error {
	req := &billing.PaymentChannelCostSystem{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	if pcId := ctx.Param(requestParameterId); pcId != "" {
		req.Id = pcId
	}

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.SetPaymentChannelCostSystem(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description create/update PaymentChannelCostMerchant
// @Example POST /admin/api/v1/payment_costs/channel/merchant
// @Example PUT /admin/api/v1/payment_costs/channel/merchant/5be2d0b4b0b30d0007383ce6
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"name": "VISA", "region": "CIS", "country": "AZ", "min_amount": 0.75, "method_percent": 1.01, "method_fix_amount": 2.34
//      "ps_percent": 3.5, "ps_fixed_fee": 2, "ps_fixed_fee_currency": "EUR", "payout_currency": "USD"}' \
//      https://api.paysuper.online/admin/api/v1/payment_costs/channel/merchant
func (r *paymentCostRoute) setPaymentChannelCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentChannelCostMerchant{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	if pcId := ctx.Param(requestParameterId); pcId != "" {
		req.Id = pcId
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorIncorrectMerchantId)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.SetPaymentChannelCostMerchant(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description create/update MoneyBackCostSystem
// @Example POST /admin/api/v1/payment_costs/money_back/system
// @Example PUT /admin/api/v1/payment_costs/money_back/system/5be2d0b4b0b30d0007383ce6
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 1.01, "fix_amount": 2.34, "payout_currency": "USD"
//      "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1}' \
//      https://api.paysuper.online/admin/api/v1/payment_costs/money_back/system
func (r *paymentCostRoute) setMoneyBackCostSystem(ctx echo.Context) error {
	req := &billing.MoneyBackCostSystem{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	if pcId := ctx.Param(requestParameterId); pcId != "" {
		req.Id = pcId
	}

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.SetMoneyBackCostSystem(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description create/update MoneyBackCostMerchant
// @Example POST /admin/api/v1/payment_costs/money_back/merchant
// @Example PUT /admin/api/v1/payment_costs/money_back/merchant/5be2d0b4b0b30d0007383ce6
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 1.01, "fix_amount": 2.34, "payout_currency": "USD",
////      "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1, "is_paid_by_merchant": true}' \
//      https://api.paysuper.online/admin/api/v1/payment_costs/money_back/merchant
func (r *paymentCostRoute) setMoneyBackCostMerchant(ctx echo.Context) error {
	req := &billing.MoneyBackCostMerchant{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	if pcId := ctx.Param(requestParameterId); pcId != "" {
		req.Id = pcId
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorIncorrectMerchantId)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.SetMoneyBackCostMerchant(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description Get All PaymentChannelCostSystem
// @Example GET /admin/api/v1/payment_costs/channel/system/all
func (r *paymentCostRoute) getAllPaymentChannelCostSystem(ctx echo.Context) error {
	res, err := r.billingService.GetAllPaymentChannelCostSystem(ctx.Request().Context(), &grpc.EmptyRequest{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description Get All PaymentChannelCostMerchant
// @Example GET /admin/api/v1/payment_costs/channel/merchant/all
func (r *paymentCostRoute) getAllPaymentChannelCostMerchant(ctx echo.Context) error {
	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorIncorrectMerchantId)
	}

	req := &billing.PaymentChannelCostMerchantListRequest{
		MerchantId: merchant.Item.Id,
	}

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetAllPaymentChannelCostMerchant(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description Get All PaymentChannelCostSystem
// @Example GET /admin/api/v1/payment_costs/money_back/system/all
func (r *paymentCostRoute) getAllMoneyBackCostSystem(ctx echo.Context) error {
	res, err := r.billingService.GetAllMoneyBackCostSystem(ctx.Request().Context(), &grpc.EmptyRequest{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}

// @Description Get All PaymentChannelCostMerchant
// @Example GET /admin/api/v1/payment_costs/money_back/merchant/all
func (r *paymentCostRoute) getAllMoneyBackCostMerchant(ctx echo.Context) error {
	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorIncorrectMerchantId)
	}

	req := &billing.MoneyBackCostMerchantListRequest{
		MerchantId: merchant.Item.Id,
	}

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetAllMoneyBackCostMerchant(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res)
}
