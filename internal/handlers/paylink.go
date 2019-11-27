package handlers

import (
	"fmt"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	u "github.com/PuerkitoBio/purell"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/paylink"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	paylinksPath               = "/paylinks"
	paylinksIdPath             = "/paylinks/:id"
	paylinksUrlPath            = "/paylinks/:id/url"
	paylinksIdStatSummaryPath  = "/paylinks/:id/dashboard/summary"
	paylinksIdStatCountryPath  = "/paylinks/:id/dashboard/country"
	paylinksIdStatReferrerPath = "/paylinks/:id/dashboard/referrer"
	paylinksIdStatDatePath     = "/paylinks/:id/dashboard/date"
	paylinksIdStatUtmPath      = "/paylinks/:id/dashboard/utm"

	paylinkUrlMask = "%s://%s/%s"
)

type PayLinkRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewPayLinkRoute(set common.HandlerSet, cfg *common.Config) *PayLinkRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PayLinkRoute"})
	return &PayLinkRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *PayLinkRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(paylinksPath, h.getPaylinksList)
	groups.AuthUser.GET(paylinksIdPath, h.getPaylink)
	groups.AuthUser.GET(paylinksUrlPath, h.getPaylinkUrl)
	groups.AuthUser.DELETE(paylinksIdPath, h.deletePaylink)
	groups.AuthUser.POST(paylinksPath, h.createPaylink)
	groups.AuthUser.PUT(paylinksIdPath, h.updatePaylink)
	groups.AuthUser.GET(paylinksIdStatSummaryPath, h.getPaylinkStatSummary)
	groups.AuthUser.GET(paylinksIdStatCountryPath, h.getPaylinkStatByCountry)
	groups.AuthUser.GET(paylinksIdStatReferrerPath, h.getPaylinkStatByReferrer)
	groups.AuthUser.GET(paylinksIdStatDatePath, h.getPaylinkStatByDate)
	groups.AuthUser.GET(paylinksIdStatUtmPath, h.getPaylinkStatByUtm)
}

func (h *PayLinkRoute) getPaylinksList(ctx echo.Context) error {
	req := &grpc.GetPaylinksRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.ProjectId = ""

	if req.Limit == 0 {
		req.Limit = int64(h.cfg.LimitDefault)
	}

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinks(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinks", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Data)
}

func (h *PayLinkRoute) getPaylink(ctx echo.Context) error {
	req := &grpc.PaylinkRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	err := h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylink(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylink", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PayLinkRoute) getPaylinkUrl(ctx echo.Context) error {
	req := &grpc.GetPaylinkURLRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

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

	req.Id = ctx.Param(common.RequestParameterId)
	req.MerchantId = merchant.Item.Id
	req.UrlMask = pkg.PaylinkUrlDefaultMask

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkURL(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkURL", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	url := fmt.Sprintf(paylinkUrlMask, h.cfg.HttpScheme, ctx.Request().Host, res.Url)

	url, err = u.NormalizeURLString(url, u.FlagsUsuallySafeGreedy|u.FlagRemoveDuplicateSlashes)
	if err != nil {
		h.L().Error("NormalizeURLString failed", logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, url)
}

func (h *PayLinkRoute) deletePaylink(ctx echo.Context) error {
	req := &grpc.PaylinkRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeletePaylink(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "DeletePaylink", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *PayLinkRoute) createPaylink(ctx echo.Context) error {
	return h.createOrUpdatePaylink(ctx, "")
}

func (h *PayLinkRoute) updatePaylink(ctx echo.Context) error {
	return h.createOrUpdatePaylink(ctx, ctx.Param(common.RequestParameterId))
}

func (h *PayLinkRoute) createOrUpdatePaylink(ctx echo.Context, paylinkId string) error {
	req := &paylink.CreatePaylinkRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.Id = paylinkId

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreateOrUpdatePaylink(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "CreateOrUpdatePaylink", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PayLinkRoute) getPaylinkStatSummary(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatTotal(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatTotal", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PayLinkRoute) getPaylinkStatByCountry(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatByCountry(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatByCountry", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PayLinkRoute) getPaylinkStatByReferrer(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatByReferrer(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatByReferrer", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PayLinkRoute) getPaylinkStatByDate(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatByDate(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatByDate", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PayLinkRoute) getPaylinkStatByUtm(ctx echo.Context) error {
	req := &grpc.GetPaylinkStatCommonRequest{}
	err := ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaylinkStatByUtm(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaylinkStatByUtm", req)
		return ctx.Render(http.StatusBadRequest, errorTemplateName, map[string]interface{}{})
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}
