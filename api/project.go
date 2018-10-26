package api

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/labstack/echo"
	"net/http"
)

type ProjectApiV1 struct {
	*Api
	projectManager *manager.ProjectManager
	merchantManager *manager.MerchantManager
}

func (api *Api) InitProjectRoutes() *Api {
	pApiV1 := ProjectApiV1{
		Api:             api,
		projectManager: manager.InitProjectManager(api.database, api.logger),
		merchantManager: manager.InitMerchantManager(api.database, api.logger),
	}

	api.accessRouteGroup.GET("/project", pApiV1.getAll)
	api.accessRouteGroup.GET("/project/:id", pApiV1.get)
	api.accessRouteGroup.POST("/project", pApiV1.create)
	api.accessRouteGroup.PUT("/project", pApiV1.update)
	api.accessRouteGroup.DELETE("/project", pApiV1.delete)

	return api
}

func (pApiV1 *ProjectApiV1) get(ctx echo.Context) error {

}

func (pApiV1 *ProjectApiV1) getAll(ctx echo.Context) error {

}

func (pApiV1 *ProjectApiV1) create(ctx echo.Context) error {
	ps := &model.ProjectScalar{}

	if err := ctx.Bind(ps); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request param")
	}

	if err := pApiV1.validate.Struct(ps); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pApiV1.getFirstValidationError(err))
	}

	ps.Merchant = pApiV1.merchantManager.FindById(pApiV1.Merchant.Identifier)

	if ps.Merchant == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Merchant not found")
	}

	if p := pApiV1.projectManager.FindProjectsByMerchantIdAndName(ps.Merchant.Id, ps.Name); p != nil {
		return ctx.JSON(http.StatusCreated, p)
	}

	p, err := pApiV1.projectManager.Create(ps)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Project create failed")
	}

	return ctx.JSON(http.StatusCreated, p)
}


func (pApiV1 *ProjectApiV1) update(ctx echo.Context) error {
	//тоже проверять на уникальность
}

func (pApiV1 *ProjectApiV1) delete(ctx echo.Context) error {

}


