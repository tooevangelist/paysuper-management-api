package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"net/http"
)

type tokenRoute struct {
	*Api
}

func (api *Api) initTokenRoutes() *Api {
	route := &tokenRoute{Api: api}
	api.apiAuthProjectGroup.POST("/tokens", route.createToken)

	return api
}

func (r *tokenRoute) createToken(ctx echo.Context) error {
	req := &grpc.TokenRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	err = r.checkProjectAuthRequestSignature(ctx, req.Settings.ProjectId)

	if err != nil {
		return err
	}

	rsp, err := r.billingService.CreateToken(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"token": rsp.Token})
}
