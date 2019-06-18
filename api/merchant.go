package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"net/http"
)

type MerchantApiV1 struct {
	*Api
	merchantManager *manager.MerchantManager
}

func (api *Api) InitMerchantRoutes() *Api {
	mApiV1 := MerchantApiV1{
		Api:             api,
		merchantManager: manager.InitMerchantManager(api.database, api.logger),
	}

	api.accessRouteGroup.GET("/merchant", mApiV1.get)
	api.accessRouteGroup.POST("/merchant", mApiV1.create)
	api.accessRouteGroup.PUT("/merchant", mApiV1.update)
	api.accessRouteGroup.DELETE("/merchant", mApiV1.delete)

	return api
}

// @Summary Get merchant
// @Description Get full data about merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Success 200 {object} model.Merchant "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/merchant/{id} [get]
func (mApiV1 *MerchantApiV1) get(ctx echo.Context) error {
	req := &grpc.GetMerchantByRequest{
		UserId: mApiV1.Merchant.Identifier,
	}

	res, err := mApiV1.billingService.GetMerchantBy(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if res == nil {
		return echo.NewHTTPError(http.StatusNotFound, errorMerchantNotFound)
	}

	return ctx.JSON(int(res.Status), res.Item)
}

// @Summary Create merchant
// @Description Create new merchant
// @Tags Merchant
// @Accept json
// @Produce json
// @Param data body model.MerchantScalar true "Creating merchant data"
// @Success 201 {object} model.Merchant "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/merchant [post]
func (mApiV1 *MerchantApiV1) create(ctx echo.Context) error {
	ms := &model.MerchantScalar{Id: mApiV1.Merchant.Identifier}

	err := ctx.Bind(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(manager.GetFirstValidationError(err)))
	}

	err = mApiV1.validate.Struct(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(manager.GetFirstValidationError(err)))
	}

	if ms.Email == nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorEmailFieldIsRequired)
	}

	m := mApiV1.merchantManager.FindById(ms.Id)

	if m != nil {
		return ctx.JSON(http.StatusCreated, m)
	}

	m1, err := mApiV1.merchantManager.Create(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorMerchantCreateFailed)
	}

	return ctx.JSON(http.StatusCreated, m1)
}

// @Summary Update merchant
// @Description Update merchant data
// @Tags Merchant
// @Accept json
// @Produce json
// @Param data body model.MerchantScalar true "Merchant object with new data"
// @Success 200 {object} model.Merchant "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/merchant [put]
func (mApiV1 *MerchantApiV1) update(ctx echo.Context) error {
	ms := &model.MerchantScalar{}

	err := ctx.Bind(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(manager.GetFirstValidationError(err)))
	}

	ms.Id = mApiV1.Merchant.Identifier

	err = mApiV1.validate.Struct(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(manager.GetFirstValidationError(err)))
	}

	m := mApiV1.merchantManager.FindById(ms.Id)

	if m == nil {
		return echo.NewHTTPError(http.StatusNotFound, errorMerchantNotFound)
	}

	m1, err := mApiV1.merchantManager.Update(m, ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorMerchantUpdateFailed)
	}

	return ctx.JSON(http.StatusOK, m1)
}

// @Summary Delete merchant
// @Description Mark merchant as deleted
// @Tags Merchant
// @Accept json
// @Produce json
// @Success 200 {string} string "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/merchant [delete]
func (mApiV1 *MerchantApiV1) delete(ctx echo.Context) error {
	m := mApiV1.merchantManager.FindById(mApiV1.Merchant.Identifier)

	if m == nil {
		return echo.NewHTTPError(http.StatusNotFound, errorMerchantNotFound)
	}

	err := mApiV1.merchantManager.Delete(m)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorMerchantDeleteFailed)
	}

	return ctx.NoContent(http.StatusOK)
}
