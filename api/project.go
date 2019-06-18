package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type projectRoute onboardingRoute

func (api *Api) InitProjectRoutes() *Api {
	route := &projectRoute{Api: api}

	api.authUserRouteGroup.GET("/projects", route.listProjects)
	api.authUserRouteGroup.GET("/projects/:id", route.getProject)
	api.authUserRouteGroup.POST("/projects", route.createProject)
	api.authUserRouteGroup.PATCH("/projects/:id", route.updateProject)
	api.authUserRouteGroup.DELETE("/projects/:id", route.deleteProject)

	return api
}

func (r *projectRoute) createProject(ctx echo.Context) error {
	req := &billing.Project{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(r.getValidationError(err)))
	}

	rsp, err := r.billingService.ChangeProject(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusCreated, rsp)
}

func (r *projectRoute) updateProject(ctx echo.Context) error {
	req := &billing.Project{}
	binder := &ChangeProjectRequestBinder{Api: r.Api}
	err := binder.Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(r.getValidationError(err)))
	}

	rsp, err := r.billingService.ChangeProject(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *projectRoute) getProject(ctx echo.Context) error {
	req := &grpc.GetProjectRequest{
		ProjectId: ctx.Param(requestParameterId),
	}

	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(r.getValidationError(err)))
	}

	rsp, err := r.billingService.GetProject(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *projectRoute) listProjects(ctx echo.Context) error {
	req := &grpc.ListProjectsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	if req.Limit <= 0 {
		req.Limit = LimitDefault
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(r.getValidationError(err)))
	}

	rsp, err := r.billingService.ListProjects(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *projectRoute) deleteProject(ctx echo.Context) error {
	req := &grpc.GetProjectRequest{
		ProjectId: ctx.Param(requestParameterId),
	}

	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, newValidationError(r.getValidationError(err)))
	}

	rsp, err := r.billingService.DeleteProject(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp)
}
