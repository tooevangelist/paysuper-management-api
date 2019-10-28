package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	pricingRecommendedConversionPath = "/pricing/recommended/conversion"
	pricingRecommendedSteamPath      = "/pricing/recommended/steam"
	pricingRecommendedTablePath      = "/pricing/recommended/table"
)

type Pricing struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewPricingRoute(set common.HandlerSet, cfg *common.Config) *Pricing {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "PriceGroup"})
	return &Pricing{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *Pricing) Route(groups *common.Groups) {
	groups.AuthProject.GET(pricingRecommendedConversionPath, h.getRecommendedByConversion)
	groups.AuthProject.GET(pricingRecommendedSteamPath, h.getRecommendedBySteam)
	groups.AuthProject.GET(pricingRecommendedTablePath, h.getRecommendedTable)
}

// Get recommended prices by currency conversion
// GET /api/v1/pricing/recommended/conversion
func (h *Pricing) getRecommendedByConversion(ctx echo.Context) error {
	req := &grpc.RecommendedPriceRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetRecommendedPriceByConversion(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessagePriceGroupRecommendedList)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get recommended prices by price groups
// GET /api/v1/pricing/recommended/steam
func (h *Pricing) getRecommendedBySteam(ctx echo.Context) error {
	req := &grpc.RecommendedPriceRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetRecommendedPriceByPriceGroup(ctx.Request().Context(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessagePriceGroupRecommendedList)
	}

	return ctx.JSON(http.StatusOK, res)
}

// Get recommended prices
// GET /api/v1/pricing/recommended/table
func (h *Pricing) getRecommendedTable(ctx echo.Context) error {
	req := &RecommendedTableRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res := &RecommendedTableResponse{
		Ranges: []*PriceTableRange{
			{From: 0, To: 0.99},
			{From: 0.99, To: 1.99},
			{From: 1.99, To: 2.99},
			{From: 2.99, To: 3.99},
			{From: 3.99, To: 4.99},
			{From: 4.99, To: 5.99},
			{From: 5.99, To: 6.99},
			{From: 6.99, To: 7.99},
			{From: 7.99, To: 8.99},
			{From: 8.99, To: 9.99},
			{From: 9.99, To: 10.99},
			{From: 10.99, To: 11.99},
			{From: 11.99, To: 12.99},
			{From: 12.99, To: 13.99},
			{From: 13.99, To: 14.99},
			{From: 14.99, To: 15.99},
			{From: 15.99, To: 16.99},
			{From: 16.99, To: 17.99},
			{From: 17.99, To: 18.99},
			{From: 18.99, To: 19.99},
			{From: 19.99, To: 24.99},
			{From: 24.99, To: 29.99},
			{From: 29.99, To: 34.99},
			{From: 34.99, To: 39.99},
			{From: 39.99, To: 44.99},
			{From: 44.99, To: 49.99},
			{From: 49.99, To: 54.99},
			{From: 54.99, To: 59.99},
			{From: 59.99, To: 64.99},
			{From: 64.99, To: 69.99},
			{From: 69.99, To: 74.99},
			{From: 74.99, To: 79.99},
			{From: 79.99, To: 84.99},
			{From: 84.99, To: 89.99},
			{From: 89.99, To: 99.99},
			{From: 99.99, To: 119.99},
			{From: 119.99, To: 129.99},
			{From: 129.99, To: 149.99},
			{From: 149.99, To: 199.99},
		},
	}

	return ctx.JSON(http.StatusOK, res)
}

type RecommendedTableRequest struct {
	Currency string `json:"currency" validate:"required,alpha,len=3"`
}

type RecommendedTableResponse struct {
	Ranges []*PriceTableRange `json:"ranges"`
}

type PriceTableRange struct {
	From float64 `json:"from"`
	To   float64 `json:"to"`
}
