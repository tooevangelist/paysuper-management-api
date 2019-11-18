package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const testMerchantWebhook = "/webhook/testing"

type WebHookRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewWebHookRoute(set common.HandlerSet, cfg *common.Config) *WebHookRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "KeyProductRoute"})
	return &WebHookRoute{
		dispatch: set,
		cfg:      *cfg,
		LMT:      &set.AwareSet,
	}
}

func (h *WebHookRoute) Route(groups *common.Groups) {
	groups.AuthUser.POST(testMerchantWebhook, h.sendWebhookTest)
}

func (h *WebHookRoute) sendWebhookTest(ctx echo.Context) error {
	req := &billing.OrderCreateRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	if len(req.TestingCase) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	res, err := h.dispatch.Services.Billing.SendWebhookToMerchant(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "SendWebhookToMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}
