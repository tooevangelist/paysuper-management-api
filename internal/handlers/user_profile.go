package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"net/http"
)

const (
	userProfilePath             = "/user/profile"
	userCommonProfilePath       = "/user/profile/common"
	userProfilePathId           = "/user/profile/:id"
	userProfilePathFeedback     = "/user/feedback"
	userProfileConfirmEmailPath = "/user/confirm_email"
)

type UserProfileRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewUserProfileRoute(set common.HandlerSet, cfg *common.Config) *UserProfileRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "UserProfileRoute"})
	return &UserProfileRoute{
		dispatch: set,
		LMT:      &set.AwareSet,
		cfg:      *cfg,
	}
}

func (h *UserProfileRoute) Route(groups *common.Groups) {
	groups.AuthProject.GET(userProfilePath, h.getUserProfile)
	groups.AuthProject.GET(userCommonProfilePath, h.getUserCommonProfile)
	groups.SystemUser.GET(userProfilePathId, h.getUserProfile)
	groups.AuthProject.PATCH(userProfilePath, h.setUserProfile)
	groups.AuthProject.POST(userProfilePathFeedback, h.createFeedback)
	groups.Common.PUT(userProfileConfirmEmailPath, h.confirmEmail)
}

func (h *UserProfileRoute) getUserProfile(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	req := &grpc.GetUserProfileRequest{
		UserId:    authUser.Id,
		ProfileId: ctx.Param(common.RequestParameterId),
	}
	err := h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetUserProfile(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetUserProfile", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *UserProfileRoute) getUserCommonProfile(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	req := &grpc.CommonUserProfileRequest{UserId: authUser.Id}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	res, err := h.dispatch.Services.Billing.GetCommonUserProfile(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetCommonUserProfile")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Profile)
}

func (h *UserProfileRoute) setUserProfile(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	req := &grpc.UserProfile{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.UserId = authUser.Id
	req.Email = &grpc.UserProfileEmail{
		Email: authUser.Email,
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreateOrUpdateUserProfile(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.WithFields(logger.Fields{"err": err.Error()}))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *UserProfileRoute) confirmEmail(ctx echo.Context) error {
	req := &grpc.ConfirmUserEmailRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	res, err := h.dispatch.Services.Billing.ConfirmUserEmail(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ConfirmUserEmail")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	req2 := &grpc.OnboardingRequest{
		User: &billing.MerchantUser{
			ProfileId:        res.Profile.Id,
			Id:               res.Profile.UserId,
			Email:            res.Profile.Email.Email,
			RegistrationDate: res.Profile.CreatedAt,
		},
	}

	if err = h.dispatch.Validate.Struct(req2); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res2, err := h.dispatch.Services.Billing.ChangeMerchant(ctx.Request().Context(), req2)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ChangeMerchant")
	}

	if res2.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
}

func (h *UserProfileRoute) createFeedback(ctx echo.Context) error {

	authUser := common.ExtractUserContext(ctx)
	if authUser.Id == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, common.ErrorMessageAccessDenied)
	}

	req := &grpc.CreatePageReviewRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.UserId = authUser.Id
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreatePageReview(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
}
