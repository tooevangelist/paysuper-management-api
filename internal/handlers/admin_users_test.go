package handlers

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/internal/test"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type AdminUsersTestSuite struct {
	suite.Suite
	router *AdminUsersRoute
	caller *test.EchoReqResCaller
}

func Test_AdminUsers(t *testing.T) {
	suite.Run(t, new(AdminUsersTestSuite))
}

func (suite *AdminUsersTestSuite) SetupTest() {
	user := &common.AuthUser{
		Id:    "ffffffffffffffffffffffff",
		Email: "test@unit.test",
	}
	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: &mocks.BillingService{},
	}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		mw.Pre(test.PreAuthUserMiddleware(user))
		suite.router = NewAdminUsersRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *AdminUsersTestSuite) TearDownTest() {}

func (suite *AdminUsersTestSuite) TestAdminUsers_GetList_InternalError() {
	shouldBe := require.New(suite.T())

	billingService := suite.router.dispatch.Services.Billing.(*mocks.BillingService)
	billingService.On("GetAdminUsers", mock2.Anything, mock2.Anything).Return(nil, errors.New("some error"))

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.SystemUserGroupPath + users).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	hErr, ok := err.(*echo.HTTPError)
	shouldBe.True(ok)
	shouldBe.Equal(http.StatusInternalServerError, hErr.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *AdminUsersTestSuite) TestAdminUsers_GetList_ServiceError() {
	shouldBe := require.New(suite.T())

	billingService := suite.router.dispatch.Services.Billing.(*mocks.BillingService)
	billingService.On("GetAdminUsers", mock2.Anything, mock2.Anything).Return(&grpc.GetAdminUsersResponse{
		Status:  400,
		Message: &grpc.ResponseErrorMessage{Message: "some error"},
		Users:   nil,
	}, nil)

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.SystemUserGroupPath + users).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	hErr, ok := err.(*echo.HTTPError)
	shouldBe.True(ok)
	shouldBe.Equal(http.StatusBadRequest, hErr.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *AdminUsersTestSuite) TestAdminUsers_GetList_Ok() {
	shouldBe := require.New(suite.T())

	billingService := suite.router.dispatch.Services.Billing.(*mocks.BillingService)
	billingService.On("GetAdminUsers", mock2.Anything, mock2.Anything).Return(&grpc.GetAdminUsersResponse{
		Status: 200,
		Users: []*billing.UserRole{
			{Id: bson.NewObjectId().Hex(), Role: "some_role"},
		},
	}, nil)

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.SystemUserGroupPath + users).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.NoError(err)
	shouldBe.Equal(http.StatusOK, res.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *AdminUsersTestSuite) TestAdminChangeRole_InternalError() {
	shouldBe := require.New(suite.T())

	billingService := suite.router.dispatch.Services.Billing.(*mocks.BillingService)
	billingService.On("ChangeRoleForAdminUser", mock2.Anything, mock2.Anything).Return(nil, errors.New("some error"))

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestRoleId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + adminUserRole).
		Init(test.ReqInitJSON()).
		BodyString(`{"role": "some_role"}`).
		Exec(suite.T())

	shouldBe.Error(err)
	hErr, ok := err.(*echo.HTTPError)
	shouldBe.True(ok)
	shouldBe.Equal(http.StatusInternalServerError, hErr.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *AdminUsersTestSuite) TestAdminChangeRole_ValidationError() {
	shouldBe := require.New(suite.T())

	billingService := suite.router.dispatch.Services.Billing.(*mocks.BillingService)
	billingService.On("ChangeRoleForAdminUser", mock2.Anything, mock2.Anything).Return(&grpc.EmptyResponseWithStatus{
		Status:  400,
		Message: &grpc.ResponseErrorMessage{Message: "some error"},
	}, nil)

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestRoleId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + adminUserRole).
		BodyString(`{"no_role": "some_role"}`).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	hErr, ok := err.(*echo.HTTPError)
	shouldBe.True(ok)
	shouldBe.Equal(http.StatusBadRequest, hErr.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *AdminUsersTestSuite) TestAdminChangeRole_Error() {
	shouldBe := require.New(suite.T())

	billingService := suite.router.dispatch.Services.Billing.(*mocks.BillingService)
	billingService.On("ChangeRoleForAdminUser", mock2.Anything, mock2.Anything).Return(&grpc.EmptyResponseWithStatus{
		Status:  400,
		Message: &grpc.ResponseErrorMessage{Message: "some error"},
	}, nil)

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestRoleId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + adminUserRole).
		BodyString(`{"role": "some_role"}`).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	hErr, ok := err.(*echo.HTTPError)
	shouldBe.True(ok)
	shouldBe.Equal(http.StatusBadRequest, hErr.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *AdminUsersTestSuite) TestAdminChangeRole_EmptyBodyError() {
	shouldBe := require.New(suite.T())

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestRoleId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + adminUserRole).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	hErr, ok := err.(*echo.HTTPError)
	shouldBe.True(ok)
	shouldBe.Equal(http.StatusBadRequest, hErr.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *AdminUsersTestSuite) TestAdminChangeRole_Ok() {
	shouldBe := require.New(suite.T())

	billingService := suite.router.dispatch.Services.Billing.(*mocks.BillingService)
	billingService.On("ChangeRoleForAdminUser", mock2.Anything, mock2.Anything).Return(&grpc.EmptyResponseWithStatus{
		Status: 200,
	}, nil)

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestRoleId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + adminUserRole).
		Init(test.ReqInitJSON()).
		BodyString(`{"role": "some_role"}`).
		Exec(suite.T())

	shouldBe.NoError(err)
	shouldBe.Equal(http.StatusOK, res.Code)
	shouldBe.Empty(res.Body.String())
}
