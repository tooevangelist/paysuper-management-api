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

type PaymentCostRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewPaymentCostRoute(set common.HandlerSet, cfg *common.Config) *PaymentCostRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PaymentCostRoute"})
	return &PaymentCostRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

const (
	paymentCostsChannelSystemPath        = "/payment_costs/channel/system"
	paymentCostsChannelSystemAllPath     = "/payment_costs/channel/system/all"
	paymentCostsChannelMerchantPath      = "/payment_costs/channel/merchant/:merchant_id"
	paymentCostsChannelMerchantAllPath   = "/payment_costs/channel/merchant/:merchant_id/all"
	paymentCostsChannelSystemIdPath      = "/payment_costs/channel/system/:id"
	paymentCostsChannelMerchantIdsPath   = "/payment_costs/channel/merchant/:merchant_id/:rate_id"
	paymentCostsMoneyBackAllPath         = "/payment_costs/money_back/system/all"
	paymentCostsMoneyBackMerchantPath    = "/payment_costs/money_back/merchant/:merchant_id"
	paymentCostsMoneyBackMerchantAllPath = "/payment_costs/money_back/merchant/:merchant_id/all"
	paymentCostsMoneyBackSystemPath      = "/payment_costs/money_back/system"
	paymentCostsMoneyBackSystemIdPath    = "/payment_costs/money_back/system/:id"
	paymentCostsMoneyBackMerchantIdsPath = "/payment_costs/money_back/merchant/:merchant_id/:rate_id"
)

func (h *PaymentCostRoute) Route(groups *common.Groups) {
	groups.SystemUser.GET(paymentCostsChannelSystemAllPath, h.getAllPaymentChannelCostSystem)
	groups.SystemUser.GET(paymentCostsChannelMerchantAllPath, h.getAllPaymentChannelCostMerchant) //надо править
	groups.SystemUser.GET(paymentCostsMoneyBackAllPath, h.getAllMoneyBackCostSystem)
	groups.SystemUser.GET(paymentCostsMoneyBackMerchantAllPath, h.getAllMoneyBackCostMerchant) //надо править

	groups.SystemUser.GET(paymentCostsChannelSystemPath, h.getPaymentChannelCostSystem)
	groups.SystemUser.GET(paymentCostsChannelMerchantPath, h.getPaymentChannelCostMerchant)
	groups.SystemUser.GET(paymentCostsMoneyBackSystemPath, h.getMoneyBackCostSystem)
	groups.SystemUser.GET(paymentCostsMoneyBackMerchantPath, h.getMoneyBackCostMerchant)

	groups.SystemUser.DELETE(paymentCostsChannelSystemIdPath, h.deletePaymentChannelCostSystem)
	groups.SystemUser.DELETE(paymentCostsChannelMerchantPath, h.deletePaymentChannelCostMerchant)
	groups.SystemUser.DELETE(paymentCostsMoneyBackSystemIdPath, h.deleteMoneyBackCostSystem)
	groups.SystemUser.DELETE(paymentCostsMoneyBackMerchantPath, h.deleteMoneyBackCostMerchant)

	groups.SystemUser.POST(paymentCostsChannelSystemPath, h.setPaymentChannelCostSystem)
	groups.SystemUser.POST(paymentCostsChannelMerchantPath, h.setPaymentChannelCostMerchant)
	groups.SystemUser.POST(paymentCostsMoneyBackSystemPath, h.setMoneyBackCostSystem)
	groups.SystemUser.POST(paymentCostsMoneyBackMerchantPath, h.setMoneyBackCostMerchant)

	groups.SystemUser.PUT(paymentCostsChannelSystemIdPath, h.setPaymentChannelCostSystem)
	groups.SystemUser.PUT(paymentCostsChannelMerchantIdsPath, h.setPaymentChannelCostMerchant)
	groups.SystemUser.PUT(paymentCostsMoneyBackSystemIdPath, h.setMoneyBackCostSystem)
	groups.SystemUser.PUT(paymentCostsMoneyBackMerchantIdsPath, h.setMoneyBackCostMerchant)
}

func (h *PaymentCostRoute) getPaymentChannelCostSystem(ctx echo.Context) error {
	req := &billing.PaymentChannelCostSystemRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaymentChannelCostSystem(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaymentChannelCostSystem", req)

		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) getPaymentChannelCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentChannelCostMerchantRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req.MerchantId = ctx.Param(common.RequestParameterMerchantId)
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPaymentChannelCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetPaymentChannelCostMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) getMoneyBackCostSystem(ctx echo.Context) error {
	req := &billing.MoneyBackCostSystemRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetMoneyBackCostSystem(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMoneyBackCostSystem", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) getMoneyBackCostMerchant(ctx echo.Context) error {
	req := &billing.MoneyBackCostMerchantRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req.MerchantId = ctx.Param(common.RequestParameterMerchantId)
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetMoneyBackCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMoneyBackCostMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) deletePaymentChannelCostSystem(ctx echo.Context) error {
	req := &billing.PaymentCostDeleteRequest{Id: ctx.Param(common.RequestParameterId)}
	err := h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeletePaymentChannelCostSystem(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "DeletePaymentChannelCostSystem", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *PaymentCostRoute) deletePaymentChannelCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentCostDeleteRequest{Id: ctx.Param(common.RequestParameterMerchantId)}
	err := h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeletePaymentChannelCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "DeletePaymentChannelCostMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *PaymentCostRoute) deleteMoneyBackCostSystem(ctx echo.Context) error {
	req := &billing.PaymentCostDeleteRequest{Id: ctx.Param(common.RequestParameterId)}
	err := h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeleteMoneyBackCostSystem(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "DeleteMoneyBackCostSystem", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *PaymentCostRoute) deleteMoneyBackCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentCostDeleteRequest{Id: ctx.Param(common.RequestParameterMerchantId)}
	err := h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeleteMoneyBackCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "DeleteMoneyBackCostMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *PaymentCostRoute) setPaymentChannelCostSystem(ctx echo.Context) error {
	req := &billing.PaymentChannelCostSystem{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	if pcId := ctx.Param(common.RequestParameterId); pcId != "" {
		req.Id = pcId
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.SetPaymentChannelCostSystem(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "SetPaymentChannelCostSystem", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) setPaymentChannelCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentChannelCostMerchant{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req.MerchantId = ctx.Param(common.RequestParameterMerchantId)

	if ctx.Request().Method == http.MethodPut {
		req.Id = ctx.Param(common.RequestParameterRateId)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.SetPaymentChannelCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "SetPaymentChannelCostMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) setMoneyBackCostSystem(ctx echo.Context) error {
	req := &billing.MoneyBackCostSystem{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	if pcId := ctx.Param(common.RequestParameterId); pcId != "" {
		req.Id = pcId
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.SetMoneyBackCostSystem(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "SetMoneyBackCostSystem", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) setMoneyBackCostMerchant(ctx echo.Context) error {
	req := &billing.MoneyBackCostMerchant{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
	}

	req.MerchantId = ctx.Param(common.RequestParameterMerchantId)

	if ctx.Request().Method == http.MethodPut {
		req.Id = ctx.Param(common.RequestParameterRateId)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.SetMoneyBackCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "SetMoneyBackCostMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) getAllPaymentChannelCostSystem(ctx echo.Context) error {
	res, err := h.dispatch.Services.Billing.GetAllPaymentChannelCostSystem(ctx.Request().Context(), &grpc.EmptyRequest{})

	if err != nil {
		h.L().Error(pkg.ErrorGrpcServiceCallFailed, logger.PairArgs("err", err.Error(), common.ErrorFieldService, pkg.ServiceName, common.ErrorFieldMethod, "GetAllPaymentChannelCostSystem"))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) getAllPaymentChannelCostMerchant(ctx echo.Context) error {
	req := &billing.PaymentChannelCostMerchantListRequest{MerchantId: ctx.Param(common.RequestParameterMerchantId)}
	err := h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetAllPaymentChannelCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetAllPaymentChannelCostMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) getAllMoneyBackCostSystem(ctx echo.Context) error {
	res, err := h.dispatch.Services.Billing.GetAllMoneyBackCostSystem(ctx.Request().Context(), &grpc.EmptyRequest{})

	if err != nil {
		h.L().Error(pkg.ErrorGrpcServiceCallFailed, logger.PairArgs("err", err.Error(), common.ErrorFieldService, pkg.ServiceName, common.ErrorFieldMethod, "GetAllMoneyBackCostSystem"))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *PaymentCostRoute) getAllMoneyBackCostMerchant(ctx echo.Context) error {
	req := &billing.MoneyBackCostMerchantListRequest{MerchantId: ctx.Param(common.RequestParameterMerchantId)}
	err := h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetAllMoneyBackCostMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetAllMoneyBackCostMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}
