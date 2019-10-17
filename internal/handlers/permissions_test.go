package handlers

import (
	"errors"
	"github.com/labstack/echo/v4"
	billMock "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/paysuper/paysuper-management-api/internal/test"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type PermissionsTestSuite struct {
	suite.Suite
	router *PermissionsRoute
	caller *test.EchoReqResCaller
}

func Test_permissions(t *testing.T) {
	suite.Run(t, new(PermissionsTestSuite))
}

func (suite *PermissionsTestSuite) SetupTest() {
	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: mock.NewBillingServerOkMock(),
		Geo:     mock.NewGeoIpServiceTestOk(),
	}

	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		suite.router = NewPermissionsRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})

	if e != nil {
		panic(e)
	}
}

func (suite *PermissionsTestSuite) TestGetPermissions_Ok() {
	shouldBe := require.New(suite.T())

	billingService := &billMock.BillingService{}
	billingService.On("GetPermissionsForUser", mock2.Anything, mock2.Anything).Return(&grpc.GetPermissionsForUserResponse{
		Status: 200,
		Permissions: []*grpc.Permission{
			{Access: "read", Name: "get_something"},
		},
	}, nil)
	suite.router.dispatch.Services.Billing = billingService

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthProjectGroupPath + permissionsRoute).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.NoError(err)
	shouldBe.Equal(http.StatusOK, res.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *PermissionsTestSuite) TestGetPermissions_InternalError() {
	shouldBe := require.New(suite.T())

	billingService := &billMock.BillingService{}
	billingService.On("GetPermissionsForUser", mock2.Anything, mock2.Anything).Return(nil, errors.New("some error"))
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthProjectGroupPath + permissionsRoute).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	shouldBe.EqualValues(500, err.(*echo.HTTPError).Code)
}

func (suite *PermissionsTestSuite) TestGetPermissions_ServiceError() {
	shouldBe := require.New(suite.T())

	billingService := &billMock.BillingService{}
	billingService.On("GetPermissionsForUser", mock2.Anything, mock2.Anything).Return(&grpc.GetPermissionsForUserResponse{
		Status: 400,
		Permissions: nil,
		Message: &grpc.ResponseErrorMessage{
			Code: "sasd",
			Message: "asd",
		},
	}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthProjectGroupPath + permissionsRoute).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	shouldBe.EqualValues(400, err.(*echo.HTTPError).Code)
}