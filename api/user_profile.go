package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type userProfileRoute struct {
	*Api
}

func (api *Api) initUserProfileRoutes() *Api {
	route := &userProfileRoute{Api: api}

	api.authUserRouteGroup.GET("/user_profile", route.getUserProfile)
	api.authUserRouteGroup.PATCH("/user_profile", route.setUserProfile)

	return api
}

func (r *userProfileRoute) getUserProfile(ctx echo.Context) error {
	if r.authUser.Id == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, errorMessageAccessDenied)
	}

	req := &grpc.GetUserProfileRequest{UserId: r.authUser.Id}
	rsp, err := r.billingService.GetUserProfile(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *userProfileRoute) setUserProfile(ctx echo.Context) error {
	req := &grpc.UserProfile{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.UserId = r.authUser.Id
	req.Email = &grpc.UserProfileEmail{
		Email: r.authUser.Email,
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.CreateOrUpdateUserProfile(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != http.StatusOK {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}
