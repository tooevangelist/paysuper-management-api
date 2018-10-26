package api

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/labstack/echo"
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

	return api
}

func (mApiV1 *MerchantApiV1) get(ctx echo.Context) error {
	m := mApiV1.merchantManager.FindById(mApiV1.Merchant.Identifier)

	if m == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Merchant not found")
	}

	return ctx.JSON(http.StatusCreated, m)
}

func (mApiV1 *MerchantApiV1) create(ctx echo.Context) error {
	ms := &model.MerchantScalar{Id: mApiV1.Merchant.Identifier}

	err := ctx.Bind(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}

	err = mApiV1.validate.Struct(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, mApiV1.getFirstValidationError(err))
	}

	m := mApiV1.merchantManager.FindById(ms.Id)

	if m != nil {
		return ctx.JSON(http.StatusCreated, m)
	}

	m, err = mApiV1.merchantManager.Create(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Merchant create failed")
	}

	return ctx.JSON(http.StatusCreated, m)
}

func (mApiV1 *MerchantApiV1) update(ctx echo.Context) error {
	ms := &model.MerchantScalar{}

	err := ctx.Bind(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param: "+err.Error())
	}

	ms.Id = mApiV1.Merchant.Identifier

	err = mApiV1.validate.Struct(ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, mApiV1.getFirstValidationError(err))
	}

	m := mApiV1.merchantManager.FindById(ms.Id)

	if m == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Merchant not found")
	}

	m, err = mApiV1.merchantManager.Update(m, ms)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Merchant update failed")
	}

	return ctx.JSON(http.StatusOK, m)
}

func (mApiV1 *MerchantApiV1) delete(ctx echo.Context) error {
	m := mApiV1.merchantManager.FindById(mApiV1.Merchant.Identifier)

	if m == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Merchant not found")
	}

	err := mApiV1.merchantManager.Delete(m)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Merchant delete failed")
	}

	return ctx.NoContent(http.StatusOK)
}