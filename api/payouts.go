package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type payoutDocumentsRoute struct {
	*Api
}

func (api *Api) initPayoutDocumentsRoutes() *Api {
	cApiV1 := payoutDocumentsRoute{
		Api: api,
	}

	api.authUserRouteGroup.GET("/payout_documents", cApiV1.getPayoutDocumentsList)
	api.authUserRouteGroup.GET("/payout_documents/:id", cApiV1.getPayoutDocument)
	api.authUserRouteGroup.POST("/payout_documents", cApiV1.createPayoutDocument)
	api.authUserRouteGroup.POST("/payout_documents/:id", cApiV1.updatePayoutDocument)

	return api
}

// Get payout documents list with filters and pagination
// GET /admin/api/v1/payout_documents?payout_document_id=5ced34d689fce60bf4440829
// GET /admin/api/v1/payout_documents?status=pending&merchant_id=5bdc39a95d1e1100019fb7df&limit=10&offset=0
// GET /admin/api/v1/payout_documents?status=pending&limit=10&offset=0
func (cApiV1 *payoutDocumentsRoute) getPayoutDocumentsList(ctx echo.Context) error {
	req := &grpc.GetPayoutDocumentsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cApiV1.getValidationError(err))
	}

	res, err := cApiV1.billingService.GetPayoutDocuments(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Data)
}

// Get payout document
// GET /admin/api/v1/payout_documents/5ced34d689fce60bf4440829
func (cApiV1 *payoutDocumentsRoute) getPayoutDocument(ctx echo.Context) error {
	req := &grpc.GetPayoutDocumentsRequest{}
	req.PayoutDocumentId = ctx.Param(requestParameterId)

	err := cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cApiV1.getValidationError(err))
	}

	res, err := cApiV1.billingService.GetPayoutDocuments(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	if len(res.Data.Items) == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return ctx.JSON(http.StatusOK, res.Data.Items[0])
}

// Create payout document
// POST /admin/api/v1/payout_documents
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{ "source_id": ["5bdc39a95d1e1100019fb7df", "5be2d0b4b0b30d0007383ce6"], "description": "royalty for june-july 2019", "merchant_id": "5bdc39a95d1e1100019fb7df" }' \
//      https://api.paysuper.online/admin/api/v1/payout_documents
func (cApiV1 *payoutDocumentsRoute) createPayoutDocument(ctx echo.Context) error {
	req := &grpc.CreatePayoutDocumentRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}

	req.Ip = ctx.RealIP()

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cApiV1.getValidationError(err))
	}

	res, err := cApiV1.billingService.CreatePayoutDocument(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Item)
}

// Update payout document by admin
// POST /admin/api/v1/payout_documents/5ced34d689fce60bf4440829
//
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"status": "failed", "failure_code": "account_closed"}' \
//      https://api.paysuper.online/admin/api/v1/payout_documents/5ced34d689fce60bf4440829
func (cApiV1 *payoutDocumentsRoute) updatePayoutDocument(ctx echo.Context) error {

	req := &grpc.UpdatePayoutDocumentRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(err.Error()))
	}
	req.PayoutDocumentId = ctx.Param(requestParameterId)
	req.Ip = ctx.RealIP()

	err = cApiV1.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, cApiV1.getValidationError(err))
	}

	res, err := cApiV1.billingService.UpdatePayoutDocument(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Item)
}
