package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

const (
	confirmEmailUrlMask = "%s://%s/confirm_email"
)

type userProfileRoute struct {
	*Api
}

func (api *Api) initUserProfileRoutes() *Api {
	route := &userProfileRoute{Api: api}

	api.authUserRouteGroup.GET("/user_profile", route.getUserProfile)
	api.authUserRouteGroup.PATCH("/user_profile", route.setUserProfile)
	api.authUserRouteGroup.GET("/user_profile/send_confirm_email", route.sendConfirmEmail)
	api.authUserRouteGroup.POST("/page_reviews", route.createPageReview)
	api.Http.GET("/api/v1/confirm_email", route.confirmEmail)

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

func (r *userProfileRoute) sendConfirmEmail(ctx echo.Context) error {
	if r.authUser.Id == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, errorMessageAccessDenied)
	}

	req := &grpc.SendConfirmEmailToUserRequest{
		UserId: r.authUser.Id,
		Url:    fmt.Sprintf(confirmEmailUrlMask, r.config.HttpScheme, ctx.Request().Host),
	}
	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.SendConfirmEmailToUser(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != http.StatusOK {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.NoContent(http.StatusOK)
}

func (r *userProfileRoute) confirmEmail(ctx echo.Context) error {
	token := ctx.QueryParam(requestParameterToken)

	if token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req := &grpc.ConfirmUserEmailRequest{Token: token}
	rsp, err := r.billingService.ConfirmUserEmail(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != http.StatusOK {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.NoContent(http.StatusOK)
}

func (r *userProfileRoute) createPageReview(ctx echo.Context) error {
	if r.authUser.Id == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, errorMessageAccessDenied)
	}

	req := &grpc.CreatePageReviewRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.UserId = r.authUser.Id
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.CreatePageReview(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != http.StatusOK {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.NoContent(http.StatusOK)
}
