package api

import (
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"net/http"
)

type productRoute struct {
	*Api
}

func (api *Api) InitProductRoutes() *Api {
	productApiV1 := productRoute{
		Api: api,
	}

	api.authUserRouteGroup.GET("/products", productApiV1.getProductsList)
	api.authUserRouteGroup.POST("/products", productApiV1.createProduct)
	api.authUserRouteGroup.GET("/products/:id", productApiV1.getProduct)
	api.authUserRouteGroup.PUT("/products/:id", productApiV1.updateProduct)
	api.authUserRouteGroup.DELETE("/products/:id", productApiV1.deleteProduct)
	api.authUserRouteGroup.GET("/products/:id/prices", productApiV1.getProductPrices)
	api.authUserRouteGroup.PUT("/products/:id/prices", productApiV1.updateProductPrices)

	return api
}

// @Description Get list of products for authenticated merchant
// @Example GET /admin/api/v1/products?name=car&sku=ru_0&project_id=5bdc39a95d1e1100019fb7df&offset=0&limit=10
func (r *productRoute) getProductsList(ctx echo.Context) error {
	req := &grpc.ListProductsRequest{}
	err := (&ProductsGetProductsListBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.S().Errorf("internal error", "err", err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.ListProducts(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Description Get product for authenticated merchant
// @Example GET /admin/api/v1/products/5c99288068add43f74be9c1d
func (r *productRoute) getProduct(ctx echo.Context) error {

	id := ctx.Param(requestParameterId)
	if id == "" || bson.IsObjectIdHex(id) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectProductId)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.S().Errorf("internal error", "err", err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	req := &grpc.RequestProduct{
		Id:         id,
		MerchantId: merchant.Item.Id,
	}

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Description Delete product for authenticated merchant
// @Example DELETE /admin/api/v1/products/5c99288068add43f74be9c1d
func (r *productRoute) deleteProduct(ctx echo.Context) error {
	id := ctx.Param(requestParameterId)
	if id == "" || bson.IsObjectIdHex(id) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectProductId)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.S().Errorf("internal error", "err", err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	req := &grpc.RequestProduct{
		Id:         id,
		MerchantId: merchant.Item.Id,
	}

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	_, err = r.billingService.DeleteProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	return ctx.NoContent(http.StatusNoContent)
}

// @Description Create new product for authenticated merchant
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"object": "product", "type": "simple_product", "sku": "ru_0_doom_2", "name": {"en": "Doom II"},
//          "default_currency": "USD", "enabled": true, "prices": [{"amount": 12.93, "currency": "USD"}],
//          "description": {"en": "Doom II description"}, "long_description": {}, "project_id": "5bdc39a95d1e1100019fb7df"}' \
//      https://api.paysuper.online/admin/api/v1/products
func (r *productRoute) createProduct(ctx echo.Context) error {
	return r.createOrUpdateProduct(ctx, &ProductsCreateProductBinder{})
}

// @Description Update existing product for authenticated merchant
// @Example curl -X PUT -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"object": "product", "type": "simple_product", "sku": "ru_0_doom_4", "name": {"en": "Doom IV"},
//          "default_currency": "USD", "enabled": true, "prices": [{"amount": 146.00, "currency": "USD"}],
//          "description": {"en": "Doom IV description"}, "long_description": {}, "project_id": "5bdc39a95d1e1100019fb7df"}' \
//      https://api.paysuper.online/admin/api/v1/products/5c99288068add43f74be9c1d
func (r *productRoute) updateProduct(ctx echo.Context) error {
	return r.createOrUpdateProduct(ctx, &ProductsUpdateProductBinder{})
}

func (r *productRoute) createOrUpdateProduct(ctx echo.Context, binder echo.Binder) error {
	req := &grpc.Product{}
	err := binder.Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	merchant, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})
	if err != nil || merchant.Item == nil {
		if err != nil {
			zap.S().Errorf("internal error", "err", err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	req.MerchantId = merchant.Item.Id

	err = r.validate.Struct(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.CreateOrUpdateProduct(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (r *productRoute) getProductPrices(ctx echo.Context) error {
	id := ctx.Param(requestParameterId)
	if id == "" || bson.IsObjectIdHex(id) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectProductId)
	}

	req := &grpc.RequestProduct{
		Id: id,
	}

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.GetProductPrices(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageGetProductPrice)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (r *productRoute) updateProductPrices(ctx echo.Context) error {
	id := ctx.Param(requestParameterId)
	if id == "" || bson.IsObjectIdHex(id) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectProductId)
	}

	req := &grpc.UpdateProductPricesRequest{}

	if err := r.validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	res, err := r.billingService.UpdateProductPrices(ctx.Request().Context(), req)
	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorMessageUpdateProductPrice)
	}

	return ctx.JSON(http.StatusOK, res)
}
