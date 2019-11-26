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
	payoutsIdPath        = "/payout_documents/:id"
	payoutsIdReportsPath = "/payout_documents/:id/reports"
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
	req.PayoutDocumentId = ctx.Param(common.RequestParameterId)

	authUser := common.ExtractUserContext(ctx)
	merchantReq := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	merchant, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), merchantReq)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", merchantReq)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if merchant.Status != http.StatusOK {
		return echo.NewHTTPError(int(merchant.Status), merchant.Message)
	}

	req.MerchantId = merchant.Item.Id

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPayoutDocument(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPayoutDocuments", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
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
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}
	req.PayoutDocumentId = ctx.Param(common.RequestParameterId)
	req.Ip = ctx.RealIP()

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.UpdatePayoutDocument(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "UpdatePayoutDocument", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
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
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetPayoutDocumentSignUrl")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	if len(res.Data.Items) == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return ctx.JSON(http.StatusOK, res.Data.Items[0])
}
