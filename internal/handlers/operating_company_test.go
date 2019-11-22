package handlers

import (
	"github.com/globalsign/mgo/bson"
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

type OperatingCompanyTestSuite struct {
	suite.Suite
	router  *OperatingCompanyRoute
	caller  *test.EchoReqResCaller
	somePDF []byte
}

func Test_OperatingCompany(t *testing.T) {
	suite.Run(t, new(OperatingCompanyTestSuite))
}

func (suite *OperatingCompanyTestSuite) SetupTest() {
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
		suite.router = NewOperatingCompanyRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *OperatingCompanyTestSuite) TearDownTest() {}

func (suite *OperatingCompanyTestSuite) TestOperatingCompany_GetOperatingCompaniesList_Ok() {
	billingService := &billMock.BillingService{}
	billingService.On("GetOperatingCompaniesList", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.GetOperatingCompaniesListResponse{Status: pkg.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = billingService

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.SystemUserGroupPath + operatingCompanyPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OperatingCompanyTestSuite) TestOperatingCompany_GetOperatingCompany_Ok() {
	billingService := &billMock.BillingService{}
	billingService.On("GetOperatingCompany", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.GetOperatingCompanyResponse{Status: pkg.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = billingService

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.SystemUserGroupPath+operatingCompanyIdPath).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OperatingCompanyTestSuite) TestOperatingCompany_AddOperatingCompany_Ok() {
	billingService := &billMock.BillingService{}
	billingService.On("AddOperatingCompany", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.EmptyResponseWithStatus{Status: pkg.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = billingService

	body := `{"name" : "Paysuper", "country" : "CY", 
			  "registration_number" : "some number", "vat_number" : "some vat number", "address" : "Cyprus", 
			  "registration_date" : "17 April 2019", "email": "test@test.com",
			  "vat_address" : "Cyprus", "signatory_name" : "Vassiliy Poupkine", "signatory_position" : "CEO", 
			  "banking_details" : "bank details including bank, bank address, account number, swift/ bic, intermediary bank"}`

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.SystemUserGroupPath + operatingCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNoContent, res.Code)
	assert.Empty(suite.T(), res.Body.String())
}

func (suite *OperatingCompanyTestSuite) TestOperatingCompany_UpdateOperatingCompany_Ok() {
	billingService := &billMock.BillingService{}
	billingService.On("AddOperatingCompany", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.EmptyResponseWithStatus{Status: pkg.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = billingService

	body := `{"name" : "Paysuper", "country" : "CY", 
			  "registration_number" : "some number", "vat_number" : "some vat number", "address" : "Cyprus",
			  "registration_date" : "17 April 2019", "email": "test@test.com",
			  "vat_address" : "Cyprus", "signatory_name" : "Vassiliy Poupkine", "signatory_position" : "CEO", 
			  "banking_details" : "bank details including bank, bank address, account number, swift/ bic, intermediary bank"}`

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.SystemUserGroupPath+operatingCompanyIdPath).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNoContent, res.Code)
	assert.Empty(suite.T(), res.Body.String())
}
