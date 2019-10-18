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
	productsPath         = "/products"
	productsMerchantPath = "/products/merchant/:merchant_id"
	productsIdPath       = "/products/:id"
	productsPricesPath   = "/products/:id/prices"
)

type ProductRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewProductRoute(set common.HandlerSet, cfg *common.Config) *ProductRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "ProductRoute"})
	return &ProductRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *ProductRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(productsPath, h.getProductsList)
	groups.SystemUser.GET(productsMerchantPath, h.getProductsList)
	groups.AuthUser.POST(productsPath, h.createProduct)
	groups.AuthUser.GET(productsIdPath, h.getProduct)
	groups.AuthUser.PUT(productsIdPath, h.updateProduct)
	groups.AuthUser.DELETE(productsIdPath, h.deleteProduct)
	groups.AuthUser.GET(productsPricesPath, h.getProductPrices)    // TODO: Need test
	groups.AuthUser.PUT(productsPricesPath, h.updateProductPrices) // TODO: Need test
}

func (h *ProductRoute) getProductsList(ctx echo.Context) error {
	req := &grpc.ListProductsRequest{}
	err := (&common.ProductsGetProductsListBinder{
		LimitDefault:  h.cfg.LimitDefault,
		OffsetDefault: h.cfg.OffsetDefault,
	}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	reqCtx := ctx.Request().Context()
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ListProducts(reqCtx, req)
	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ListProducts", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *ProductRoute) getProduct(ctx echo.Context) error {

	req := &grpc.RequestProduct{
		Id: ctx.Param(common.RequestParameterId),
	}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetProduct(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *ProductRoute) deleteProduct(ctx echo.Context) error {
	req := &grpc.RequestProduct{
		Id: ctx.Param(common.RequestParameterId),
	}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	if _, err := h.dispatch.Services.Billing.DeleteProduct(ctx.Request().Context(), req); err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	return ctx.NoContent(http.StatusNoContent)
}

func (h *ProductRoute) createProduct(ctx echo.Context) error {
	return h.createOrUpdateProduct(ctx, &common.ProductsCreateProductBinder{})
}

func (h *ProductRoute) updateProduct(ctx echo.Context) error {
	return h.createOrUpdateProduct(ctx, &common.ProductsUpdateProductBinder{})
}

func (h *ProductRoute) createOrUpdateProduct(ctx echo.Context, binder echo.Binder) error {
	req := &grpc.Product{}

	if err := binder.Bind(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreateOrUpdateProduct(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *ProductRoute) getProductPrices(ctx echo.Context) error {

	req := &grpc.RequestProduct{
		Id: ctx.Param(common.RequestParameterId),
	}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetProductPrices(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageGetProductPrice)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *ProductRoute) updateProductPrices(ctx echo.Context) error {

	req := &grpc.UpdateProductPricesRequest{ProductId: ctx.Param(common.RequestParameterId)}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.UpdateProductPrices(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorMessageUpdateProductPrice)
	}

	return ctx.JSON(http.StatusOK, res)
}
