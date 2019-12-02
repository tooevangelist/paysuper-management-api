package handlers

import (
	"errors"
	"github.com/labstack/echo/v4"
	billingMocks "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/paysuper/paysuper-management-api/internal/test"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type PricingTestSuite struct {
	suite.Suite
	router *Pricing
	caller *test.EchoReqResCaller
}

func Test_Pricing(t *testing.T) {
	suite.Run(t, new(PricingTestSuite))
}

func (suite *PricingTestSuite) SetupTest() {
	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: mock.NewBillingServerOkMock(),
	}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		suite.router = NewPricingRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *PricingTestSuite) TearDownTest() {}

func (suite *PricingTestSuite) TestPricing_getRecommendedPrice_BindError_RequiredAmount() {
	data := `{"variable": "test"}`

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.NoAuthGroupPath + pricingRecommendedSteamPath).
		Init(test.ReqInitJSON()).
		BodyString(data).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'Amount' failed on the 'required' tag", msg.Details)
}

func (suite *PricingTestSuite) TestPricing_getRecommendedPrice_Error_BillingServer() {
	data := `{"amount": 1, "currency": "USD"}`

	billingService := &billingMocks.BillingService{}
	billingService.On("GetRecommendedPriceByPriceGroup", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.NoAuthGroupPath + pricingRecommendedSteamPath).
		Init(test.ReqInitJSON()).
		BodyString(data).
		Exec(suite.T())

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Regexp(suite.T(), common.ErrorMessagePriceGroupRecommendedList.Message, httpErr.Message)
}

func (suite *PricingTestSuite) TestPricing_getRecommendedPrice_Ok() {
	data := `{"amount": 1, "currency": "USD"}`

	billingService := &billingMocks.BillingService{}
	billingService.On("GetRecommendedPriceByPriceGroup", mock2.Anything, mock2.Anything).Return(&grpc.RecommendedPriceResponse{}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.NoAuthGroupPath + pricingRecommendedSteamPath).
		Init(test.ReqInitJSON()).
		BodyString(data).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
}
