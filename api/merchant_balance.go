package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"net/http"
)

type balanceRoute struct {
	*Api
}

func (api *Api) initBalanceRoutes() *Api {
	cApiV1 := balanceRoute{
		Api: api,
	}

	api.authUserRouteGroup.GET("/balance", cApiV1.getMerchantBalance)
	api.authUserRouteGroup.GET("/balance/:merchant_id", cApiV1.getMerchantBalance)

	return api
}

// Get merchant balance
// GET /admin/api/v1/balance - for current merchant
// GET /admin/api/v1/balance/:merchant_id - for any merchant by it's id
func (cApiV1 *balanceRoute) getMerchantBalance(ctx echo.Context) error {
	req := &grpc.GetMerchantBalanceRequest{}

	merchantId := ctx.Param(requestParameterMerchantId)

	if merchantId == "" {
		mReq := &grpc.GetMerchantByRequest{UserId: cApiV1.authUser.Id}
		merchant, err := cApiV1.billingService.GetMerchantBy(ctx.Request().Context(), mReq)
		if err != nil {
			zap.L().Error(
				pkg.ErrorGrpcServiceCallFailed,
				zap.Error(err),
				zap.String(ErrorFieldService, pkg.ServiceName),
				zap.String(ErrorFieldMethod, "GetMerchantBy"),
				zap.Any(ErrorFieldRequest, mReq),
			)
			return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
		}
		if merchant.Status != http.StatusOK {
			return echo.NewHTTPError(int(merchant.Status), merchant.Message)
		}
		merchantId = merchant.Item.Id
	}

	req.MerchantId = merchantId

	err := cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cApiV1.getValidationError(err))
	}

	res, err := cApiV1.billingService.GetMerchantBalance(ctx.Request().Context(), req)
	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetMerchantBalance"),
			zap.Any(ErrorFieldRequest, req),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Item)
}
