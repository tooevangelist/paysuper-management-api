package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"net/http"
)

type zipCodeRoute onboardingRoute

func (api *Api) initZipCodeRoutes() *Api {
	route := &zipCodeRoute{Api: api}

	api.Http.GET("/api/v1/zip", route.checkZip)

	return api
}

func (r *zipCodeRoute) checkZip(ctx echo.Context) error {
	req := &grpc.FindByZipCodeRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	if req.Limit <= 0 {
		req.Limit = LimitDefault
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.FindByZipCode(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp)
}
