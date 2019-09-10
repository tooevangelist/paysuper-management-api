package api

import (
	"github.com/labstack/echo/v4"
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

	return api
}

// Get current merchant balance
// GET /admin/api/v1/balance
func (cApiV1 *balanceRoute) getMerchantBalance(ctx echo.Context) error {
	req := &grpc.GetMerchantBalanceRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	merchant, err := cApiV1.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: cApiV1.authUser.Id})
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.S().Errorf("internal error", "err", err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	req.MerchantId = merchant.Item.Id

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cApiV1.getValidationError(err))
	}

	res, err := cApiV1.billingService.GetMerchantBalance(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Item)
}
