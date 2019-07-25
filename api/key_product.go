package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"net/http"
)

type keyProductRoute struct {
	*Api
}

const internalErrorTemplate = "internal error"

var (
	KeyProductIdInvalid = newManagementApiResponseError("ka000001", "key product id is invalid")
	PlatformIdInvalid   = newManagementApiResponseError("ka000002", "platform id is invalid")
)

func (api *Api) InitKeyProductRoutes() *Api {
	keyProductApiV1 := keyProductRoute{
		Api: api,
	}

	api.authUserRouteGroup.GET("/key-products", keyProductApiV1.getKeyProductList)
	api.authUserRouteGroup.POST("/key-products", keyProductApiV1.createKeyProduct)
	api.authUserRouteGroup.GET("/key-products/:key_product_id", keyProductApiV1.getKeyProductById)
	api.authUserRouteGroup.PUT("/key-products/:key_product_id", keyProductApiV1.changeKeyProduct)
	api.authUserRouteGroup.POST("/key-products/:key_product_id/publish", keyProductApiV1.publishKeyProduct)
	api.authUserRouteGroup.POST("/key-products/:key_product_id/platforms", keyProductApiV1.changePlatformPricesForKeyProduct)
	api.authUserRouteGroup.DELETE("/key-products/:key_product_id/platforms/:platform_id", keyProductApiV1.removePlatformForKeyProduct)
	api.authUserRouteGroup.PUT("/platforms", keyProductApiV1.getPlatformsList)

	return api
}

// @Description Remove platform from product
// @Example DELETE /admin/api/v1/key-products/:key_product_id/platforms/:platform_id
func (r *keyProductRoute) removePlatformForKeyProduct(ctx echo.Context) error {
	req := &grpc.RemovePlatformRequest{}
	req.KeyProductId = ctx.Param("key_product_id")
	if req.KeyProductId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, KeyProductIdInvalid)
	}

	req.PlatformId = ctx.Param("platform_id")
	if req.PlatformId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, PlatformIdInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeletePlatformFromProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Description Change prices for specified platform and key product
// @Example POST /admin/api/v1/key-products/:key_product_id/platforms
func (r *keyProductRoute) changePlatformPricesForKeyProduct(ctx echo.Context) error {
	req := &grpc.AddOrUpdatePlatformPricesRequest{}
	req.KeyProductId = ctx.Param("key_product_id")
	if req.KeyProductId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, KeyProductIdInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.UpdatePlatformPrices(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Description Publishes product
// @Example POST /admin/api/v1/key-products/:key_product_id/publish
func (r *keyProductRoute) publishKeyProduct(ctx echo.Context) error {
	req := &grpc.PublishKeyProductRequest{}
	req.KeyProductId = ctx.Param("key_product_id")
	if req.KeyProductId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, KeyProductIdInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.PublishKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Product)
}

// @Description Get available platform list
// @Example GET /admin/api/v1/platforms
func (r *keyProductRoute) getPlatformsList(ctx echo.Context) error {
	req := &grpc.ListPlatformsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetPlatforms(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Description Create new key product for authenticated merchant
// @Example PUT /admin/api/v1/key-products/:key_product_id
func (r *keyProductRoute) changeKeyProduct(ctx echo.Context) error {
	req := &grpc.CreateOrUpdateKeyProductRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.Id = ctx.Param("key_product_id")
	if req.Id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, KeyProductIdInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.CreateOrUpdateKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Product)
}

// @Description Gets key product by id
// @Example POST /admin/api/v1/key-products/:key_product_id
func (r *keyProductRoute) getKeyProductById(ctx echo.Context) error {
	req := &grpc.RequestKeyProduct{}
	req.Id = ctx.Param("key_product_id")

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Product)
}

// @Description Create new key product for authenticated merchant
// @Example POST /admin/api/v1/key-products
func (r *keyProductRoute) createKeyProduct(ctx echo.Context) error {
	req := &grpc.CreateOrUpdateKeyProductRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.CreateOrUpdateKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusCreated, res.Product)
}

// @Description Get list of key products for authenticated merchant
// @Example GET /admin/api/v1/key-products?name=car&project_id=5bdc39a95d1e1100019fb7df&offset=0&limit=10
func (r *keyProductRoute) getKeyProductList(ctx echo.Context) error {
	req := &grpc.ListKeyProductsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	if req.Limit <= 0 {
		req.Limit = LimitDefault
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetKeyProducts(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}
