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
	api.accessRouteGroup.PUT("/project/:id", pApiV1.update)
	api.accessRouteGroup.DELETE("/project/:id", pApiV1.delete)

	return api
}

// @Summary Get project
// @Description "Get data about project"
// @Tags Project
// @Accept json
// @Produce json
// @Param data path string true "Project identifier"
// @Success 200 {object} model.Project "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 404 {object} model.Error "Project not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/project/{id} [get]
func (pApiV1 *ProjectApiV1) get(ctx echo.Context) error {
	id := ctx.Param("id")

	p := pApiV1.projectManager.FindProjectById(id)

	if p == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Project not found")
	}

	return ctx.JSON(http.StatusOK, p)
}

// @Summary List projects
// @Description Get list of project for authenticated merchant
// @Tags Project
// @Accept json
// @Produce json
// @Success 200 {array} model.Project "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/project [get]
func (pApiV1 *ProjectApiV1) getAll(ctx echo.Context) error {
	p := pApiV1.projectManager.FindProjectsByMerchantId(pApiV1.Merchant.Identifier, pApiV1.limit, pApiV1.offset)

	return ctx.JSON(http.StatusOK, p)
}

// @Summary Create project
// @Description Create new project for authenticated merchant
// @Tags Project
// @Accept json
// @Produce json
// @Param data body model.ProjectScalar true "Creating project data"
// @Success 201 {object} model.Project "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/project [post]
func (pApiV1 *ProjectApiV1) create(ctx echo.Context) error {
	ps := &model.ProjectScalar{}

	if err := ctx.Bind(ps); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Bad request")
	}

	if err := pApiV1.validate.Struct(ps); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pApiV1.getFirstValidationError(err))
	}

	ps.Merchant = pApiV1.merchantManager.FindById(pApiV1.Merchant.Identifier)

	if ps.Merchant == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Merchant not found")
	}

	if ps.Merchant.Status < model.MerchantStatusCompleted {
		return echo.NewHTTPError(http.StatusBadRequest, "Merchant data not set")
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

// @Summary Update project
// @Description Update project for authenticated merchant
// @Tags Project
// @Accept json
// @Produce json
// @Param data body model.ProjectScalar true "Project object with new data"
// @Param id path string true "Project identifier"
// @Success 200 {object} model.Project "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 403 {object} model.Error "Access denied"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/project/{id} [put]
func (pApiV1 *ProjectApiV1) update(ctx echo.Context) error {
	p := pApiV1.projectManager.FindProjectById(ctx.Param("id"))

	if p == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Project not found")
	}

	if p.Merchant.ExternalId != pApiV1.Merchant.Identifier {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	ps := &model.ProjectScalar{}

	if err := ctx.Bind(ps); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	if err := pApiV1.validate.Struct(ps); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, pApiV1.getFirstValidationError(err))
	}

	pf := pApiV1.projectManager.FindProjectsByMerchantIdAndName(p.Merchant.Id, ps.Name)

	if pf != nil && pf.Id != p.Id {
		return echo.NewHTTPError(http.StatusBadRequest, "Project with received name already exist")
	}

	p, err := pApiV1.projectManager.Update(p, ps)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Project update failed")
	}

	return ctx.JSON(http.StatusOK, p)
}

// @Summary Delete project
// @Description Delete project for authenticated merchant
// @Tags Project
// @Accept json
// @Produce json
// @Param id path string true "Project identifier"
// @Success 200 {string} string "OK"
// @Failure 401 {object} model.Error "Unauthorized"
// @Failure 403 {object} model.Error "Access denied"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Some unknown error"
// @Router /api/v1/s/project/{id} [delete]
func (pApiV1 *ProjectApiV1) delete(ctx echo.Context) error {
	p := pApiV1.projectManager.FindProjectById(ctx.Param("id"))

	if p == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Project not found")
	}

	if p.Merchant.Id.String() != pApiV1.Merchant.Identifier {
		return echo.NewHTTPError(http.StatusForbidden, "Access denied")
	}

	err := pApiV1.projectManager.Delete(p)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Project delete failed")
	}

	return ctx.NoContent(http.StatusOK)
}


