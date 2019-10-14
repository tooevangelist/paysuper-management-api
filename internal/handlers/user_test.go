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

type UserTestSuite struct {
	suite.Suite
	router *UserRoute
	caller *test.EchoReqResCaller
}

func Test_user(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}

func (suite *UserTestSuite) SetupTest() {
	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: &billMock.BillingService{},
		Geo:     mock.NewGeoIpServiceTestOk(),
	}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		suite.router = NewUserRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *UserTestSuite) TearDownTest() {}

func (suite *UserTestSuite) TestUser_getMerchantList_Ok() {
	shouldBe := require.New(suite.T())

	billingService := suite.router.dispatch.Services.Billing.(*billMock.BillingService)
	billingService.On("GetMerchantsForUser", mock2.Anything, mock2.Anything).Return(&grpc.GetMerchantsForUserResponse{
		Status: 200,
		Message: nil,
		Merchants: []*grpc.MerchantForUserInfo{
			{
				Id: "some id",
				Name: "Some name",
			},
		},
	}, nil)


	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.AuthUserGroupPath + userMerchantsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	hErr, ok := err.(*echo.HTTPError)
	shouldBe.True(ok)
	shouldBe.Equal(http.StatusBadRequest, hErr.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *UserTestSuite) TestUser_getMerchantList_ServiceError() {
	shouldBe := require.New(suite.T())

	billingService := suite.router.dispatch.Services.Billing.(*billMock.BillingService)
	billingService.On("GetMerchantsForUser", mock2.Anything, mock2.Anything).Return(&grpc.GetMerchantsForUserResponse{
		Status: 400,
		Message: &grpc.ResponseErrorMessage{Message: "some error"},
		Merchants: nil,
	}, nil)


	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.AuthUserGroupPath + userMerchantsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	hErr, ok := err.(*echo.HTTPError)
	shouldBe.True(ok)
	shouldBe.Equal(http.StatusBadRequest, hErr.Code)
	shouldBe.NotEmpty(res.Body.String())
}

func (suite *UserTestSuite) TestUser_getMerchantList_InternalError() {
	shouldBe := require.New(suite.T())
	billingService := suite.router.dispatch.Services.Billing.(*billMock.BillingService)
	billingService.On("GetMerchantsForUser", mock2.Anything, mock2.Anything).Return(nil, errors.New("some error"))

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.AuthUserGroupPath + userMerchantsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	shouldBe.Error(err)
	hErr, ok := err.(*echo.HTTPError)
	shouldBe.True(ok)
	shouldBe.Equal(http.StatusInternalServerError, hErr.Code)
	shouldBe.NotEmpty(res.Body.String())
}

