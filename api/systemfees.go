package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type systemFeeRoute struct {
	*Api
}

func (api *Api) InitSystemFeeRoutes() *Api {
	systemFeeApiV1 := &systemFeeRoute{
		Api: api,
	}

	api.authUserRouteGroup.GET("/systemfees", systemFeeApiV1.getSystemFeesList)
	api.authUserRouteGroup.POST("/systemfees", systemFeeApiV1.addSystemFee)
	return api
}

// @Description Get list of actual system fees
// @Example GET /admin/api/v1/systemfees
func (r *systemFeeRoute) getSystemFeesList(ctx echo.Context) error {
	systemFees, err := r.billingService.GetActualSystemFeesList(ctx.Request().Context(), &grpc.EmptyRequest{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	return ctx.JSON(http.StatusOK, systemFees)
}

// @Description Add new actual system fee
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{ "method_id": "5be2d0b4b0b30d0007383ce6", "region": "EU", "card_brand": "MASTERCARD",
//      "fees": [ { "min_amounts": { "EUR": 0, "USD": 0 },
//      "transaction_cost": { "percent": 1.15, "percent_currency": "EUR", "fix_amount": 0.2, "fix_currency": "EUR" },
//      "authorization_fee": { "percent": 0, "percent_currency": "EUR", "fix_amount": 0.1, "fix_currency": "EUR" } } ],
//      "user_id": "5cb6e4aa68add437e8a8f0fa" }' \
//      https://api.paysuper.online/admin/api/v1/systemfees
func (r *systemFeeRoute) addSystemFee(ctx echo.Context) error {
	req := &billing.AddSystemFeesRequest{}

	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}
	req.UserId = r.authUser.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	_, err = r.billingService.AddSystemFees(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.NoContent(http.StatusOK)
}
