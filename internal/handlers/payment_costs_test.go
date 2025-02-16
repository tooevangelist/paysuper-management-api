package handlers

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/paysuper/paysuper-management-api/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type PaymentCostTestSuite struct {
	suite.Suite
	router *PaymentCostRoute
	caller *test.EchoReqResCaller
}

func Test_PaymentCost(t *testing.T) {
	suite.Run(t, new(PaymentCostTestSuite))
}

func (suite *PaymentCostTestSuite) SetupTest() {
	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: mock.NewBillingServerOkMock(),
	}
	user := &common.AuthUser{
		Id:    "ffffffffffffffffffffffff",
		Email: "test@unit.test",
		MerchantId: "ffffffffffffffffffffffff",
	}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		mw.Pre(test.PreAuthUserMiddleware(user))
		suite.router = NewPaymentCostRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *PaymentCostTestSuite) TearDownTest() {}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostSystem_GetAll() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, "").
		Path(common.SystemUserGroupPath + paymentCostsChannelSystemAllPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostSystem_Get() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, "").
		SetQueryParam("name", "VISA").
		SetQueryParam("region", "CIS").
		SetQueryParam("country", "AZ").
		SetQueryParam("mcc_code", "1234").
		SetQueryParam("operating_company_id", "5dbc50d486616113a1cefe16").
		Path(common.SystemUserGroupPath + paymentCostsChannelSystemPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostSystem_Add() {
	bodyJson := `{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 0.1, "fix_amount": 2.34, "fix_amount_currency": "USD"}`

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterId, "").
		Path(common.SystemUserGroupPath + paymentCostsChannelSystemPath).
		Init(test.ReqInitJSON()).
		BodyString(bodyJson).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostSystem_Delete() {

	res, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Params(":"+common.RequestParameterId, "5be2d0b4b0b30d0007383ce6").
		Path(common.SystemUserGroupPath + paymentCostsChannelSystemIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, res.Code)
		assert.Empty(suite.T(), res.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostMerchant_GetAll() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + paymentCostsMoneyBackMerchantAllPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostMerchant_Get() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		SetQueryParam("name", "VISA").
		SetQueryParam("region", "CIS").
		SetQueryParam("country", "AZ").
		SetQueryParam("payout_currency", "USD").
		SetQueryParam("amount", "100").
		SetQueryParam("mcc_code", "1234").
		Path(common.SystemUserGroupPath + paymentCostsChannelMerchantPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostMerchant_Add() {
	bodyJson := `{"name": "VISA", "region": "CIS", "country": "AZ", "min_amount": 0.75, "method_percent": 0.0101, 
                  "method_fix_amount": 2.34, "ps_percent": 0.00035, "ps_fixed_fee": 2, "ps_fixed_fee_currency": "EUR", 
                  "payout_currency": "USD", "method_fix_amount_currency": "EUR"}`

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + paymentCostsChannelMerchantPath).
		Init(test.ReqInitJSON()).
		BodyString(bodyJson).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostMerchant_Delete() {

	res, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + paymentCostsChannelMerchantPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, res.Code)
		assert.Empty(suite.T(), res.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostSystem_GetAll() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + paymentCostsMoneyBackAllPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostSystem_Get() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		SetQueryParam("name", "VISA").
		SetQueryParam("region", "CIS").
		SetQueryParam("country", "AZ").
		SetQueryParam("payout_currency", "USD").
		SetQueryParam("days", "10").
		SetQueryParam("undo_reason", "chargeback").
		SetQueryParam("payment_stage", "1").
		SetQueryParam("mcc_code", "1234").
		SetQueryParam("operating_company_id", "5dbc50d486616113a1cefe16").
		Path(common.SystemUserGroupPath + paymentCostsMoneyBackSystemPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostSystem_Add() {
	bodyJson := `{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 0.0101, "fix_amount": 2.34, 
				  "fix_amount_currency": "EUR",  
                  "payout_currency": "USD", "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1}`

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + paymentCostsMoneyBackSystemPath).
		Init(test.ReqInitJSON()).
		BodyString(bodyJson).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostSystem_Delete() {

	res, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + paymentCostsMoneyBackSystemIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, res.Code)
		assert.Empty(suite.T(), res.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostMerchant_GetAll() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + paymentCostsMoneyBackMerchantAllPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostMerchant_Get() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, bson.NewObjectId().Hex()).
		SetQueryParam("name", "VISA").
		SetQueryParam("region", "CIS").
		SetQueryParam("country", "AZ").
		SetQueryParam("payout_currency", "USD").
		SetQueryParam("days", "10").
		SetQueryParam("undo_reason", "chargeback").
		SetQueryParam("payment_stage", "1").
		SetQueryParam("mcc_code", "1234").
		Path(common.SystemUserGroupPath + paymentCostsMoneyBackMerchantPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostMerchant_Add() {
	bodyJson := `{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 0.0101, "fix_amount": 2.34, "fix_amount_currency": "USD",
                  "payout_currency": "USD", "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1, 
                  "is_paid_by_merchant": true}`

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + paymentCostsMoneyBackMerchantPath).
		Init(test.ReqInitJSON()).
		BodyString(bodyJson).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostMerchant_Delete() {

	res, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Params(":"+common.RequestParameterMerchantId, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + paymentCostsMoneyBackMerchantPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, res.Code)
		assert.Empty(suite.T(), res.Body.String())
	}
}
