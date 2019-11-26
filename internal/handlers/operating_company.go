package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	operatingCompanyPath   = "/operating_company"
	operatingCompanyIdPath = "/operating_company/:id"
)

type OperatingCompanyRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewOperatingCompanyRoute(set common.HandlerSet, cfg *common.Config) *OperatingCompanyRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "OperatingCompanyRoute"})
	return &OperatingCompanyRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *OperatingCompanyRoute) Route(groups *common.Groups) {
	groups.SystemUser.GET(operatingCompanyPath, h.getOperatingCompanyList)
	groups.SystemUser.GET(operatingCompanyIdPath, h.getOperatingCompany)
	groups.SystemUser.POST(operatingCompanyPath, h.addOperatingCompany)
	groups.SystemUser.POST(operatingCompanyIdPath, h.updateOperatingCompany)

}

func (h *OperatingCompanyRoute) getOperatingCompanyList(ctx echo.Context) error {
	req := &grpc.EmptyRequest{}

	res, err := h.dispatch.Services.Billing.GetOperatingCompaniesList(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetOperatingCompaniesList", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Items)
}

func (h *OperatingCompanyRoute) getOperatingCompany(ctx echo.Context) error {
	req := &grpc.GetOperatingCompanyRequest{
		Id: ctx.Param(common.RequestParameterId),
	}

	res, err := h.dispatch.Services.Billing.GetOperatingCompany(ctx.Request().Context(), req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetOperatingCompaniesList", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}
	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}
	return ctx.JSON(http.StatusOK, res.Company)
}

func (h *OperatingCompanyRoute) addOperatingCompany(ctx echo.Context) error {
	return h.addOrUpdateOperatingCompany(ctx, "")
}

func (h *OperatingCompanyRoute) updateOperatingCompany(ctx echo.Context) error {
	return h.addOrUpdateOperatingCompany(ctx, ctx.Param(common.RequestParameterId))
}

func (h *OperatingCompanyRoute) addOrUpdateOperatingCompany(ctx echo.Context, operatingCompanyId string) error {
	req := &billing.OperatingCompany{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.Id = operatingCompanyId

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.AddOperatingCompany(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "AddOperatingCompany", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}
