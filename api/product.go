package api

import (
	"context"
	"github.com/labstack/echo"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
	"strconv"
)

type productRoute struct {
	*Api
}

func (api *Api) InitProductRoutes() *Api {
	productApiV1 := productRoute{
		Api: api,
	}

	api.accessRouteGroup.GET("/product", productApiV1.getProductsList)
	api.accessRouteGroup.POST("/product", productApiV1.createProduct)
	api.accessRouteGroup.GET("/product/:id", productApiV1.getProduct)
	api.accessRouteGroup.PUT("/product/:id", productApiV1.updateProduct)
	api.accessRouteGroup.DELETE("/product/:id", productApiV1.deleteProduct)

	return api
}

// @Summary List products
// @Description Get list of products for authenticated merchant
// @Tags Product
// @Produce json
// @Success 200 {array} model.Project "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/product [get]
// @Example GET /api/v1/s/product?name=car&sku=ru_0&offset=0&limit=10
func (r *productRoute) getProductsList(ctx echo.Context) error {
	limit := int32(LimitDefault)
	offset := int32(OffsetDefault)

	params := ctx.QueryParams()

	if v, ok := params[requestParameterLimit]; ok {
		if i, err := strconv.ParseInt(v[0], 10, 32); err == nil {
			limit = int32(i)
		}
	}

	if v, ok := params[requestParameterOffset]; ok {
		if i, err := strconv.ParseInt(v[0], 10, 32); err == nil {
			offset = int32(i)
		}
	}

	req := &grpc.ListProductsRequest{
		Limit:  limit,
		Offset: offset,
	}

	if v, ok := params[requestParameterName]; ok {
		if v[0] != "" {
			req.Name = v[0]
		}
	}

	if v, ok := params[requestParameterSku]; ok {
		if v[0] != "" {
			req.Sku = v[0]
		}
	}

	res, err := r.billingService.ListProducts(context.TODO(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Request data invalid")
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Summary Get Product
// @Description Get product for authenticated merchant
// @Tags Product
// @Produce json
// @Success 200 {array} model.Project "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/product/:id [get]
// @Example GET /api/v1/s/product/5c99288068add43f74be9c1d
func (r *productRoute) getProduct(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Request data invalid")
	}

	req := &grpc.RequestProductById{
		Id: id,
	}

	res, err := r.billingService.GetProduct(context.TODO(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Requested product not found")
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Summary Delete Product
// @Description Delete product for authenticated merchant
// @Tags Product
// @Produce json
// @Success 200 {array} model.Project "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/product/:id [delete]
// @Example DELETE /api/v1/s/product/5c99288068add43f74be9c1d
func (r *productRoute) deleteProduct(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Request data invalid")
	}

	req := &grpc.RequestProductById{
		Id: id,
	}

	_, err := r.billingService.DeleteProduct(context.TODO(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Requested product not found")
	}

	return ctx.NoContent(http.StatusNoContent)
}

// @Summary Create product
// @Description Create product for authenticated merchant
// @Tags Product
// @Accept json
// @Produce json
// @Success 200 {array} model.Project "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/product [post]
// @Example curl -X POST -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"object": "product", "type": "simple_product", "sku": "ru_0_doom_2", "name": "Doom II",
//          "default_currency": "USD", "enabled": true, "prices": [{"amount": 12.93, "currency": "USD"}],
//          "description": "Doom II description", "long_description": ""}' \
//      https://api.paysuper.online/api/v1/s/product
func (r *productRoute) createProduct(ctx echo.Context) error {
	req := &grpc.Product{}
	err := ctx.Bind(req)

	req.Id = ""

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}

	res, err := r.billingService.CreateOrUpdateProduct(context.TODO(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, res)
}

// @Summary Update existing product
// @Description Update existing product for authenticated merchant
// @Tags Product
// @Accept json
// @Produce json
// @Success 200 {array} model.Project "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/product [put]
// @Example curl -X PUT -H "Accept: application/json" -H "Content-Type: application/json" \
//      -H "Authorization: Bearer %access_token_here%" \
//      -d '{"object": "product", "type": "simple_product", "sku": "ru_0_doom_4", "name": "Doom IV",
//          "default_currency": "USD", "enabled": true, "prices": [{"amount": 146.00, "currency": "USD"}],
//          "description": "Doom IV description", "long_description": ""}' \
//      https://api.paysuper.online/api/v1/s/product/5c99288068add43f74be9c1d
func (r *productRoute) updateProduct(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Request data invalid")
	}

	req := &grpc.Product{}
	err := ctx.Bind(req)

	req.Id = id

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}

	res, err := r.billingService.CreateOrUpdateProduct(context.TODO(), req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, res)
}
