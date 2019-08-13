package api

import (
	"github.com/ProtocolONE/geoip-service/pkg/proto"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

type keyProductRoute struct {
	*Api
}

const internalErrorTemplate = "internal error"

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
	api.authUserRouteGroup.GET("/platforms", keyProductApiV1.getPlatformsList)

	api.publicRouteGroup.GET("/key-products/:key_product_id", keyProductApiV1.getKeyProduct)

	return api
}

// @Description Remove platform from product
// @Example DELETE /admin/api/v1/key-products/:key_product_id/platforms/:platform_id
func (r *keyProductRoute) removePlatformForKeyProduct(ctx echo.Context) error {
	req := &grpc.RemovePlatformRequest{}
	req.KeyProductId = ctx.Param("key_product_id")
	if req.KeyProductId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorMessageKeyProductIdInvalid)
	}

	req.PlatformId = ctx.Param("platform_id")
	if req.PlatformId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorMessagePlatformIdInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	if merchant.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(http.StatusBadRequest, merchant.Message)
	}
	req.MerchantId = merchant.Item.Id

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.DeletePlatformFromProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(internalErrorTemplate, "err", err.Error())
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
		return echo.NewHTTPError(http.StatusBadRequest, ErrorMessageKeyProductIdInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	if merchant.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(http.StatusBadRequest, merchant.Message)
	}
	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.UpdatePlatformPrices(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(internalErrorTemplate, "err", err.Error())
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
		return echo.NewHTTPError(http.StatusBadRequest, ErrorMessageKeyProductIdInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	if merchant.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(http.StatusBadRequest, merchant.Message)
	}
	req.MerchantId = merchant.Item.Id

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.PublishKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(internalErrorTemplate, "err", err.Error())
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

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetPlatforms(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(internalErrorTemplate, "err", err.Error())
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
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.Id = ctx.Param("key_product_id")
	if req.Id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorMessageKeyProductIdInvalid)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	if merchant.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(http.StatusBadRequest, merchant.Message)
	}
	req.MerchantId = merchant.Item.Id

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.CreateOrUpdateKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(internalErrorTemplate, "err", err.Error())
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
	req := &grpc.RequestKeyProductMerchant{}
	req.Id = ctx.Param("key_product_id")

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	if merchant.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(http.StatusBadRequest, merchant.Message)
	}
	req.MerchantId = merchant.Item.Id

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(internalErrorTemplate, "err", err.Error())
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
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	if merchant.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(http.StatusBadRequest, merchant.Message)
	}
	req.MerchantId = merchant.Item.Id

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.CreateOrUpdateKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(internalErrorTemplate, "err", err.Error())
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
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	if req.Limit > LimitMax {
		req.Limit = LimitMax
	}

	if req.Limit <= 0 {
		req.Limit = LimitDefault
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}
	if merchant.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(http.StatusBadRequest, merchant.Message)
	}

	req.MerchantId = merchant.Item.Id

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetKeyProducts(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Description Get product with platforms list and their prices
// @Example GET /api/v1/key-products/:key_product_id?country=RUS&currency=EUR
func (r *keyProductRoute) getKeyProduct(ctx echo.Context) error {
	req := &grpc.GetKeyProductInfoRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.KeyProductId = ctx.Param("key_product_id")

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	if req.Currency == "" && req.Country == "" {
		rsp, err := r.geoService.GetIpData(ctx.Request().Context(), &proto.GeoIpDataRequest{IP: ctx.RealIP()})
		if err != nil {
			zap.S().Error(internalErrorTemplate, "err", err.Error())
		} else {
			req.Country = rsp.Country.IsoCode
		}
	}

	if req.Language == "" {
		req.Language, _ = r.getCountryFromAcceptLanguage(ctx.Request().Header.Get("Accept-Language"))
	}

	res, err := r.billingService.GetKeyProductInfo(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Error(internalErrorTemplate, "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.KeyProduct)
}

func (r *keyProductRoute) getCountryFromAcceptLanguage(acceptLanguage string) (string, string) {
	it := strings.Split(acceptLanguage, ",")

	if len(it) == 0 {
		return "", ""
	}

	if strings.Index(it[0], "-") == -1 {
		return "", ""
	}

	it = strings.Split(it[0], "-")

	return strings.ToLower(it[0]), strings.ToUpper(it[1])
}