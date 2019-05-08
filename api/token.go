package api

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type tokenRoute struct {
	*Api
}

func (api *Api) initTokenRoutes() (*Api, error) {
	route := &tokenRoute{Api: api}
	api.apiAuthProjectGroup.POST("/tokens", route.createToken)

	return api, nil
}

func (r *tokenRoute) createToken(ctx echo.Context) error {
	req := &grpc.TokenRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	err = r.checkProjectAuthRequestSignature(ctx, req.Settings.ProjectId)

	if err != nil {
		return err
	}

	rsp, err := r.billingService.CreateToken(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"token": rsp.Token})
}
