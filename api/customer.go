package api

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"net/http"
)

type customerRoute struct {
	*Api
}

func (api *Api) initCustomerRoutes() (*Api, error) {
	route := &customerRoute{Api: api}
	api.apiAuthProjectGroup.POST("/customers", route.createCustomer)

	return api, nil
}

func (r *customerRoute) createCustomer(ctx echo.Context) error {
	req := &billing.Customer{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	err = r.checkProjectAuthRequestSignature(ctx, req.ProjectId)

	if err != nil {
		return err
	}

	rsp, err := r.billingService.ChangeCustomer(context.TODO(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, map[string]string{"token": rsp.Item.Token})
}
