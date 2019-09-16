package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"net/http"
)

type keyRoute struct {
	*Api
}

func (api *Api) initKeyRoutes() *Api {
	keyApiV1 := keyRoute{
		Api: api,
	}

	api.authUserRouteGroup.GET("/keys/:key_id", keyApiV1.getKeyInfo)

	return api
}

func (r *keyRoute) getKeyInfo(ctx echo.Context) error {
	req := &grpc.KeyForOrderRequest{
		KeyId: ctx.Param("key_id"),
	}

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetKeyByID(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(InternalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Key)
}
