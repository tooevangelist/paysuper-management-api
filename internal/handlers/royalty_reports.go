package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	reporterPkg "github.com/paysuper/paysuper-reporter/pkg"
	"net/http"
)

const (
	royaltyReportsPath                     = "/royalty_reports"
	royaltyReportsIdPath                   = "/royalty_reports/:report_id"
	royaltyReportsIdDownloadPath           = "/royalty_reports/:report_id/download"
	royaltyReportsTransactionsPath         = "/royalty_reports/:report_id/transactions"
	royaltyReportsTransactionsDownloadPath = "/royalty_reports/:report_id/transactions/download"
	royaltyReportsAcceptPath               = "/royalty_reports/:report_id/accept"
	royaltyReportsDeclinePath              = "/royalty_reports/:report_id/decline"
	royaltyReportsChangePath               = "/royalty_reports/:report_id/change"
)

type RoyaltyReportRequestFile struct {
	Id         string `json:"id" validate:"required,hexadecimal,len=24"`
	MerchantId string `json:"merchant_id" validate:"required,hexadecimal,len=24"`
}

type RoyaltyReportsRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewRoyaltyReportsRoute(set common.HandlerSet, cfg *common.Config) *RoyaltyReportsRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "RoyaltyReportsRoute"})
	return &RoyaltyReportsRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *RoyaltyReportsRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(royaltyReportsPath, h.getRoyaltyReportsList)
	groups.AuthUser.GET(royaltyReportsIdPath, h.getRoyaltyReport)
	groups.AuthUser.GET(royaltyReportsIdDownloadPath, h.downloadRoyaltyReport)
	groups.AuthUser.GET(royaltyReportsTransactionsPath, h.listRoyaltyReportOrders)
	groups.AuthUser.GET(royaltyReportsTransactionsDownloadPath, h.downloadRoyaltyReportOrders)
	groups.AuthUser.POST(royaltyReportsAcceptPath, h.merchantReviewRoyaltyReport)
	groups.AuthUser.POST(royaltyReportsDeclinePath, h.merchantDeclineRoyaltyReport)
	groups.SystemUser.POST(royaltyReportsChangePath, h.changeRoyaltyReport)
}

func (h *RoyaltyReportsRoute) getRoyaltyReportsList(ctx echo.Context) error {
	req := &grpc.ListRoyaltyReportsRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.NewValidationError(err.Error()))
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ListRoyaltyReports(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ListRoyaltyReports", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Data)
}

func (h *RoyaltyReportsRoute) getRoyaltyReport(ctx echo.Context) error {
	req := &grpc.GetRoyaltyReportRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetRoyaltyReport(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetRoyaltyReport")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *RoyaltyReportsRoute) downloadRoyaltyReport(ctx echo.Context) error {
	req := &common.ReportFileRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	req.ReportType = reporterPkg.ReportTypeRoyalty
	req.Params = map[string]interface{}{
		reporterPkg.ParamsFieldId: ctx.Param(common.RequestParameterReportId),
	}

	return h.dispatch.RequestReportFile(ctx, req)
}

func (h *RoyaltyReportsRoute) listRoyaltyReportOrders(ctx echo.Context) error {
	req := &grpc.ListRoyaltyReportOrdersRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.NewValidationError(err.Error()))
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ListRoyaltyReportOrders(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ListRoyaltyReportOrders", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Data)
}

func (h *RoyaltyReportsRoute) downloadRoyaltyReportOrders(ctx echo.Context) error {
	req := &common.ReportFileRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	req.ReportType = reporterPkg.ReportTypeRoyaltyTransactions
	req.Params = map[string]interface{}{
		reporterPkg.ParamsFieldId: ctx.Param(common.RequestParameterReportId),
	}

	return h.dispatch.RequestReportFile(ctx, req)
}

func (h *RoyaltyReportsRoute) merchantReviewRoyaltyReport(ctx echo.Context) error {
	req := &grpc.MerchantReviewRoyaltyReportRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.NewValidationError(err.Error()))
	}

	req.IsAccepted = true
	req.Ip = ctx.RealIP()

	res, err := h.dispatch.Services.Billing.MerchantReviewRoyaltyReport(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "MerchantReviewRoyaltyReport")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *RoyaltyReportsRoute) merchantDeclineRoyaltyReport(ctx echo.Context) error {
	req := &grpc.MerchantReviewRoyaltyReportRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.NewValidationError(err.Error()))
	}

	req.IsAccepted = false
	req.Ip = ctx.RealIP()

	res, err := h.dispatch.Services.Billing.MerchantReviewRoyaltyReport(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "MerchantReviewRoyaltyReport")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *RoyaltyReportsRoute) changeRoyaltyReport(ctx echo.Context) error {
	req := &grpc.ChangeRoyaltyReportRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.NewValidationError(err.Error()))
	}

	req.Ip = ctx.RealIP()

	res, err := h.dispatch.Services.Billing.ChangeRoyaltyReport(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ChangeRoyaltyReport")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}
