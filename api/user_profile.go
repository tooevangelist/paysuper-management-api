package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"net/http"
)

type userProfileRoute struct {
	*Api
}

func (api *Api) initUserProfileRoutes() *Api {
	route := &userProfileRoute{Api: api}

	api.authUserRouteGroup.GET("/user/profile", route.getUserProfile)
	api.authUserRouteGroup.GET("/user/profile/:id", route.getUserProfile)
	api.authUserRouteGroup.PATCH("/user/profile", route.setUserProfile)
	api.Http.PUT("/api/v1/user/confirm_email", route.confirmEmail)
	api.authUserRouteGroup.POST("/user/feedback", route.createFeedback)

	return api
}

// @Description Get user profile
// @Example curl -X GET 'Authorization: Bearer %access_token_here%' \
//  https://api.paysuper.online/admin/api/v1/user/profile
//
// @Example curl -X GET 'Authorization: Bearer %access_token_here%' \
//  https://api.paysuper.online/admin/api/v1/user/profile/ffffffffffffffffffffffff
func (r *userProfileRoute) getUserProfile(ctx echo.Context) error {
	req := &grpc.GetUserProfileRequest{
		UserId:    r.authUser.Id,
		ProfileId: ctx.Param(requestParameterId),
	}
	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.GetUserProfile(ctx.Request().Context(), req)

	if err != nil {
		zap.L().Error(
			pkg.ErrorGrpcServiceCallFailed,
			zap.Error(err),
			zap.String(ErrorFieldService, pkg.ServiceName),
			zap.String(ErrorFieldMethod, "GetUserProfile"),
			zap.Any(ErrorFieldRequest, req),
		)

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
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != http.StatusOK {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *userProfileRoute) confirmEmail(ctx echo.Context) error {
	req := &grpc.ConfirmUserEmailRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	rsp, err := r.billingService.ConfirmUserEmail(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != http.StatusOK {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.NoContent(http.StatusOK)
}

func (r *userProfileRoute) createFeedback(ctx echo.Context) error {
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
