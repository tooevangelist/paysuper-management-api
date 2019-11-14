package handlers

import (
	"github.com/paysuper/paysuper-billing-server/pkg"
	billMock "github.com/paysuper/paysuper-billing-server/pkg/mocks"
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

type PaymentMinLimitSystemTestSuite struct {
	suite.Suite
	router  *PaymentMinLimitSystemRoute
	caller  *test.EchoReqResCaller
	somePDF []byte
}

func Test_PaymentMinLimitSystem(t *testing.T) {
	suite.Run(t, new(PaymentMinLimitSystemTestSuite))
}

func (suite *PaymentMinLimitSystemTestSuite) SetupTest() {
	user := &common.AuthUser{
		Id: "ffffffffffffffffffffffff",
	}

	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: mock.NewBillingServerOkMock(),
	}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		mw.Pre(test.PreAuthUserMiddleware(user))
		suite.router = NewPaymentMinLimitSystemRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *PaymentMinLimitSystemTestSuite) TearDownTest() {}

func (suite *PaymentMinLimitSystemTestSuite) TestPaymentMinLimitSystem_GetOperatingCompaniesList_Ok() {
	billingService := &billMock.BillingService{}
	billingService.On("GetOperatingCompaniesList", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.GetOperatingCompaniesListResponse{Status: pkg.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = billingService

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthUserGroupPath + paymentMinLimitSystemPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *PaymentMinLimitSystemTestSuite) TestPaymentMinLimitSystem_SetPaymentMinLimitSystem_Ok() {
	billingService := &billMock.BillingService{}
	billingService.On("SetPaymentMinLimitSystem", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.EmptyResponseWithStatus{Status: pkg.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = billingService

	body := `{"currency" : "RUB", "amount" : 100500}`

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.AuthUserGroupPath + paymentMinLimitSystemPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNoContent, res.Code)
	assert.Empty(suite.T(), res.Body.String())
}
