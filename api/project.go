package api

import (
	"github.com/labstack/echo"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"net/http"
)

type ProjectApiV1 struct {
	*Api
	projectManager  *manager.ProjectManager
	merchantManager *manager.MerchantManager
}

func (api *Api) InitProjectRoutes() *Api {
	pApiV1 := ProjectApiV1{
		Api:             api,
		projectManager:  manager.InitProjectManager(api.database, api.logger),
		merchantManager: manager.InitMerchantManager(api.database, api.logger),
	}

	api.accessRouteGroup.GET("/project", pApiV1.getAll)
	api.accessRouteGroup.GET("/project/:id", pApiV1.get)
	api.accessRouteGroup.POST("/project", pApiV1.create)
	api.accessRouteGroup.PUT("/project/:id", pApiV1.update)
	api.accessRouteGroup.DELETE("/project/:id", pApiV1.delete)
	api.accessRouteGroup.GET("/project/filters", pApiV1.getFiltersProjects)

	api.Http.GET("/api/v1/project/package/:region/:project_id", pApiV1.getFixedPackage)

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
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
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
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
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

	if p.Merchant.ExternalId != pApiV1.Merchant.Identifier {
		return echo.NewHTTPError(http.StatusForbidden, model.ResponseMessageAccessDenied)
	}

	err := pApiV1.projectManager.Delete(p)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Project delete failed")
	}

	return ctx.NoContent(http.StatusOK)
}

// @Summary Get project's fixed packages
// @Description Get list of project's fixed packages by filters
// @Tags Project
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param region path string true "2 symbols region code by ISO 3166-1"
// @Param project_id query string true "project unique identifier"
// @Param id query array false "array of identifiers of fixed packages"
// @Param name query array false "array of names of fixed packages"
// @Success 200 {object} model.FilteredFixedPackage "OK"
// @Failure 400 {object} model.Error "Invalid request data"
// @Failure 404 {object} model.Error "Not found"
// @Failure 500 {object} model.Error "Object with error message"
// @Router /api/v1/project/package/{region}/{project_id} [get]
func (pApiV1 *ProjectApiV1) getFixedPackage(ctx echo.Context) error {
	filters := &model.FixedPackageFilters{
		Region:    ctx.Param(model.ApiRequestParameterRegion),
		ProjectId: ctx.Param(model.ApiRequestParameterProjectId),
	}

	if err := ctx.Bind(filters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := pApiV1.validate.Struct(filters); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, manager.GetFirstValidationError(err))
	}

	p := pApiV1.projectManager.FindProjectById(filters.ProjectId)

	if p == nil {
		return echo.NewHTTPError(http.StatusNotFound, model.ResponseMessageNotFound)
	}

	fps := pApiV1.projectManager.FindFixedPackage(filters)

	if fps == nil {
		return ctx.JSON(http.StatusOK, []string{})
	}

	return ctx.JSON(http.StatusOK, fps)
}

func (pApiV1 *ProjectApiV1) getFiltersProjects(ctx echo.Context) error {
	p := pApiV1.projectManager.FindProjectsMainData(pApiV1.Merchant.Identifier)

	return ctx.JSON(http.StatusOK, p)
}
