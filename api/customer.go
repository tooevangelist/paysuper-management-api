package api

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
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

	req1 := &grpc.CheckProjectRequestSignatureRequest{
		Body:      r.rawBody,
		ProjectId: req.ProjectId,
		Signature: r.reqSignature,
	}
	rsp1, err := r.billingService.CheckProjectRequestSignature(context.TODO(), req1)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp1.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp1.Status), rsp1.Message)
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
