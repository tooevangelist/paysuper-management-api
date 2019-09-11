package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
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
	api.authUserRouteGroup.GET("/payment_costs/channel/merchant/:id/all", paymentCostApiV1.getAllPaymentChannelCostMerchant) //надо править
	api.authUserRouteGroup.GET("/payment_costs/money_back/system/all", paymentCostApiV1.getAllMoneyBackCostSystem)
	api.authUserRouteGroup.GET("/payment_costs/money_back/merchant/:id/all", paymentCostApiV1.getAllMoneyBackCostMerchant) //надо править

	api.authUserRouteGroup.GET("/payment_costs/channel/system", paymentCostApiV1.getPaymentChannelCostSystem)
	api.authUserRouteGroup.GET("/payment_costs/channel/merchant/:id", paymentCostApiV1.getPaymentChannelCostMerchant)
	api.authUserRouteGroup.GET("/payment_costs/money_back/system", paymentCostApiV1.getMoneyBackCostSystem)
	api.authUserRouteGroup.GET("/payment_costs/money_back/merchant/:id", paymentCostApiV1.getMoneyBackCostMerchant)

	api.authUserRouteGroup.DELETE("/payment_costs/channel/system/:id", paymentCostApiV1.deletePaymentChannelCostSystem)
	api.authUserRouteGroup.DELETE("/payment_costs/channel/merchant/:id", paymentCostApiV1.deletePaymentChannelCostMerchant)
	api.authUserRouteGroup.DELETE("/payment_costs/money_back/system/:id", paymentCostApiV1.deleteMoneyBackCostSystem)
	api.authUserRouteGroup.DELETE("/payment_costs/money_back/merchant/:id", paymentCostApiV1.deleteMoneyBackCostMerchant)

	api.authUserRouteGroup.POST("/payment_costs/channel/system", paymentCostApiV1.setPaymentChannelCostSystem)
	api.authUserRouteGroup.POST("/payment_costs/channel/merchant/:id", paymentCostApiV1.setPaymentChannelCostMerchant)
	api.authUserRouteGroup.POST("/payment_costs/money_back/system", paymentCostApiV1.setMoneyBackCostSystem)
	api.authUserRouteGroup.POST("/payment_costs/money_back/merchant/:id", paymentCostApiV1.setMoneyBackCostMerchant)

	api.authUserRouteGroup.PUT("/payment_costs/channel/system/:id", paymentCostApiV1.setPaymentChannelCostSystem)
	api.authUserRouteGroup.PUT("/payment_costs/channel/merchant/:merchant_id/:rate_id", paymentCostApiV1.setPaymentChannelCostMerchant)
	api.authUserRouteGroup.PUT("/payment_costs/money_back/system/:id", paymentCostApiV1.setMoneyBackCostSystem)
	api.authUserRouteGroup.PUT("/payment_costs/money_back/merchant/:merchant_id/:rate_id", paymentCostApiV1.setMoneyBackCostMerchant)

	return api
}

// @Description Get system costs for payments operations
// @Example curl -X GET -H "Authorization: Bearer %access_token_here%"  \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/channel/system?name=VISA&region=CIS&country=AZ
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
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetPaymentChannelCostSystem"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Get merchant costs for payment operations
// @Example curl -X GET -H "Authorization: Bearer %access_token_here%"  \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/channel/merchant/ffffffffffffffffffffffff?name=VISA&region=CIS&country=AZ&payout_currency=USD&amount=100
func (r *paymentCostRoute) getPaymentChannelCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentChannelCostMerchantRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	req.MerchantId = ctx.Param(requestParameterId)
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetPaymentChannelCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetPaymentChannelCostMerchant"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Get system costs for money back operations
// @Example curl -X GET -H "Authorization: Bearer %access_token_here%"  \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/system?name=VISA&region=CIS&country=AZ&payout_currency=USD&days=10&undo_reason=chargeback&payment_stage=1
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
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetMoneyBackCostSystem"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Get merchant costs for money back operations
// @Example curl -X GET -H "Authorization: Bearer %access_token_here%"  \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/merchant/ffffffffffffffffffffffff??name=VISA&region=CIS&country=AZ&payout_currency=USD&days=10&undo_reason=chargeback&payment_stage=1
func (r *paymentCostRoute) getMoneyBackCostMerchant(ctx echo.Context) error {
	req := &billing.MoneyBackCostMerchantRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	req.MerchantId = ctx.Param(requestParameterId)
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetMoneyBackCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetMoneyBackCostMerchant"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Delete system costs for payment operations
// @Example curl -X DELETE -H "Authorization: Bearer %access_token_here%" -H "Content-Type: application/json" \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/channel/system/ffffffffffffffffffffffff
func (r *paymentCostRoute) deletePaymentChannelCostSystem(ctx echo.Context) error {
	req := &billing.PaymentCostDeleteRequest{Id: ctx.Param(requestParameterId)}
	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeletePaymentChannelCostSystem(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "DeletePaymentChannelCostSystem"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// @Description Delete merchant costs for payment operations
// @Example curl -X DELETE -H "Authorization: Bearer %access_token_here%" -H "Content-Type: application/json" \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/channel/merchant/ffffffffffffffffffffffff
func (r *paymentCostRoute) deletePaymentChannelCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentCostDeleteRequest{Id: ctx.Param(requestParameterId)}
	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeletePaymentChannelCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "DeletePaymentChannelCostMerchant"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// @Description Delete system costs for money back operations
// @Example curl -X DELETE -H "Authorization: Bearer %access_token_here%" -H "Content-Type: application/json" \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/system/ffffffffffffffffffffffff
func (r *paymentCostRoute) deleteMoneyBackCostSystem(ctx echo.Context) error {
	req := &billing.PaymentCostDeleteRequest{Id: ctx.Param(requestParameterId)}
	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeleteMoneyBackCostSystem(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "DeleteMoneyBackCostSystem"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// @Description Delete merchant costs for money back operations
// @Example curl -X DELETE -H "Authorization: Bearer %access_token_here%" -H "Content-Type: application/json" \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/merchant/ffffffffffffffffffffffff
func (r *paymentCostRoute) deleteMoneyBackCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentCostDeleteRequest{Id: ctx.Param(requestParameterId)}
	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeleteMoneyBackCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "DeleteMoneyBackCostMerchant"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// @Description Create and update system costs for payments operations
// @Example curl -X POST -H "Authorization: Bearer %access_token_here%" -H "Content-Type: application/json" \
//      -d '{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 1.01, "fix_amount": 2.34,
//      "fix_amount_currency": "USD"}' \
//      https://api.paysuper.online/admin/api/v1/payment_costs/channel/system
//
// @Example curl -X PUT -H "Authorization: Bearer %access_token_here%" -H "Content-Type: application/json" \
//      -d '{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 1.01, "fix_amount": 2.34,
//      "fix_amount_currency": "USD"}' \
//      https://api.paysuper.online/admin/api/v1/payment_costs/channel/system/ffffffffffffffffffffffff
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
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "SetPaymentChannelCostSystem"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Create and update merchant costs for payments operations
//  @Example curl -X POST -H "Authorization: Bearer %access_token_here%" -H "Content-Type: application/json" \
//      -d '{"name": "VISA", "region": "CIS", "country": "AZ", "min_amount": 0.75, "method_percent": 1.01,
// 			"method_fix_amount": 2.34, "ps_percent": 3.5, "ps_fixed_fee": 2, "ps_fixed_fee_currency": "EUR",
// 			"payout_currency": "USD"}' \
//      https://api.paysuper.online/admin/api/v1/payment_costs/channel/merchant/ffffffffffffffffffffffff
//
// @Example curl -X PUT -H "Authorization: Bearer %access_token_here%" -H "Content-Type: application/json" \
//      -d '{"name": "VISA", "region": "CIS", "country": "AZ", "min_amount": 0.75, "method_percent": 1.01,
//      "method_fix_amount": 2.34, "ps_percent": 3.5, "ps_fixed_fee": 2, "ps_fixed_fee_currency": "EUR",
//      "payout_currency": "USD"}' \
//      https://api.paysuper.online/admin/api/v1/payment_costs/channel/merchant/ffffffffffffffffffffffff/aaaaaaaaaaaaaaaaaaaaaaaa
func (r *paymentCostRoute) setPaymentChannelCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentChannelCostMerchant{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	req.MerchantId = ctx.Param(requestParameterId)

	if ctx.Request().Method == http.MethodPut {
		req.Id = ctx.Param(requestParameterRateId)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.SetPaymentChannelCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "SetPaymentChannelCostMerchant"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Create and update system costs for money back operations
// @Example curl -X POST -H 'Authorization: Bearer %access_token_here%' -H "Content-Type: application/json" \
//		-d '{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 1.01, "fix_amount": 2.34,
//		"payout_currency": "USD", "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1}' \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/system
//
// @Example curl -X PUT -H 'Authorization: Bearer %access_token_here%' -H "Content-Type: application/json" \
//		-d '{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 1.01, "fix_amount": 2.34,
//		"payout_currency": "USD", "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1}' \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/system/ffffffffffffffffffffffff
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
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "SetMoneyBackCostSystem"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Create and update merchant costs for money back operations
// @Example curl -X POST -H 'Authorization: Bearer %access_token_here%' -H "Content-Type: application/json" \
//		-d '{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 1.01, "fix_amount": 2.34,
//		"payout_currency": "USD", "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1,
//		"is_paid_by_merchant": true}' \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/merchant/ffffffffffffffffffffffff
//
// @Example curl -X PUT -H 'Authorization: Bearer %access_token_here%' -H "Content-Type: application/json" \
//		-d '{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 1.01, "fix_amount": 2.34,
//		"payout_currency": "USD", "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1,
//		"is_paid_by_merchant": true}' \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/merchant/ffffffffffffffffffffffff/aaaaaaaaaaaaaaaaaaaaaaaa
func (r *paymentCostRoute) setMoneyBackCostMerchant(ctx echo.Context) error {
	req := &billing.MoneyBackCostMerchant{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestDataInvalid)
	}

	req.MerchantId = ctx.Param(requestParameterId)

	if ctx.Request().Method == http.MethodPut {
		req.Id = ctx.Param(requestParameterRateId)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.SetMoneyBackCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "SetMoneyBackCostMerchant"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Get all system costs for payments
// @Example @Example curl -X GET -H 'Authorization: Bearer %access_token_here%' \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/channel/system/all
func (r *paymentCostRoute) getAllPaymentChannelCostSystem(ctx echo.Context) error {
	res, err := r.billingService.GetAllPaymentChannelCostSystem(ctx.Request().Context(), &grpc.EmptyRequest{})

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetAllPaymentChannelCostSystem"),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Get all merchant costs for payments operations
// @Example @Example curl -X GET -H 'Authorization: Bearer %access_token_here%' \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/channel/merchant/ffffffffffffffffffffffff/all
func (r *paymentCostRoute) getAllPaymentChannelCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentChannelCostMerchantListRequest{MerchantId: ctx.Param(requestParameterId)}
	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetAllPaymentChannelCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetAllPaymentChannelCostMerchant"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Get all system costs for money back operations
// @Example @Example curl -X GET -H 'Authorization: Bearer %access_token_here%' \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/system/all
func (r *paymentCostRoute) getAllMoneyBackCostSystem(ctx echo.Context) error {
	res, err := r.billingService.GetAllMoneyBackCostSystem(ctx.Request().Context(), &grpc.EmptyRequest{})

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetAllMoneyBackCostSystem"),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

// @Description Get all merchant costs for money back operations
// @Example @Example curl -X GET -H 'Authorization: Bearer %access_token_here%' \
// 		https://api.paysuper.online/admin/api/v1/payment_costs/money_back/merchant/ffffffffffffffffffffffff/all
func (r *paymentCostRoute) getAllMoneyBackCostMerchant(ctx echo.Context) error {
	req := &billing.MoneyBackCostMerchantListRequest{MerchantId: ctx.Param(requestParameterId)}
	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetAllMoneyBackCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetAllMoneyBackCostMerchant"),
			zap.Any(ErrorFieldRequest, req),
		)

		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}
