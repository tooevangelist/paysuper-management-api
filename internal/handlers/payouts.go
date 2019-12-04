package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	payoutsPath          = "/payout_documents"
	payoutsIdPath        = "/payout_documents/:payout_document_id"
	payoutsIdReportsPath = "/payout_documents/:payout_document_id/reports"
)

type PayoutDocumentsRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewPayoutDocumentsRoute(set common.HandlerSet, cfg *common.Config) *PayoutDocumentsRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PayoutDocumentsRoute"})
	return &PayoutDocumentsRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *PayoutDocumentsRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(payoutsPath, h.getPayoutDocumentsList)
	groups.AuthUser.GET(payoutsIdPath, h.getPayoutDocument)
	groups.AuthUser.GET(payoutsIdReportsPath, h.getPayoutRoyaltyReports)
	groups.AuthUser.POST(payoutsPath, h.createPayoutDocument)
	groups.SystemUser.POST(payoutsIdPath, h.updatePayoutDocument)

}

func (h *PayoutDocumentsRoute) getPayoutDocumentsList(ctx echo.Context) error {
	req := &grpc.GetPayoutDocumentsRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetPayoutDocuments(ctx.Request().Context(), req)
	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetPayoutDocuments")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Data)
}

func (h *PayoutDocumentsRoute) getPayoutDocument(ctx echo.Context) error {
	req := &grpc.GetPayoutDocumentRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetPayoutDocument(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetPayoutDocument")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PayoutDocumentsRoute) createPayoutDocument(ctx echo.Context) error {
	req := &grpc.CreatePayoutDocumentRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.Ip = ctx.RealIP()
	req.Initiator = pkg.RoyaltyReportChangeSourceMerchant

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreatePayoutDocument(ctx.Request().Context(), req)
	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "CreatePayoutDocument")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Items)
}

func (h *PayoutDocumentsRoute) updatePayoutDocument(ctx echo.Context) error {
	req := &grpc.UpdatePayoutDocumentRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}
	req.Ip = ctx.RealIP()

	res, err := h.dispatch.Services.Billing.UpdatePayoutDocument(ctx.Request().Context(), req)
	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "UpdatePayoutDocument")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PayoutDocumentsRoute) getPayoutRoyaltyReports(ctx echo.Context) error {
	req := &grpc.GetPayoutDocumentRequest{}
	req.PayoutDocumentId = ctx.Param(common.RequestParameterId)

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetPayoutDocumentRoyaltyReports(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetPayoutDocumentRoyaltyReports")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Data.Items)
}
