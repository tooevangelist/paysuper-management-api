package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type PaymentCostTestSuite struct {
	suite.Suite
	router *paymentCostRoute
	api    *Api
}

func Test_PaymentCost(t *testing.T) {
	suite.Run(t, new(PaymentCostTestSuite))
}

func (suite *PaymentCostTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &paymentCostRoute{Api: suite.api}
}

func (suite *PaymentCostTestSuite) TearDownTest() {}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostSystem_GetAll() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_costs/channel/system/all", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/channel/system/all")
	err := suite.router.getAllPaymentChannelCostSystem(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostSystem_Get() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_costs/channel/system?name=VISA&region=CIS&country=AZ", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/channel/system?name=VISA&region=CIS&country=AZ")
	err := suite.router.getPaymentChannelCostSystem(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostSystem_Add() {
	bodyJson := `{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 0.0101, "fix_amount": 2.34, "fix_amount_currency": "USD"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_costs/channel/system", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/channel/system")
	err := suite.router.setPaymentChannelCostSystem(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostSystem_Delete() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_costs/channel/system/5be2d0b4b0b30d0007383ce6", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/channel/system/:" + requestParameterId)
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("5be2d0b4b0b30d0007383ce6")

	err := suite.router.deletePaymentChannelCostSystem(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, rsp.Code)
		assert.Empty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostMerchant_GetAll() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_costs/channel/merchant/all", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/channel/merchant/all")
	err := suite.router.getAllPaymentChannelCostMerchant(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostMerchant_Get() {
	path := "/payment_costs/channel/merchant?name=VISA&region=CIS&country=AZ&payoutCurrency=USD&amount=100"
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath(path)
	err := suite.router.getPaymentChannelCostMerchant(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostMerchant_Add() {
	bodyJson := `{"name": "VISA", "region": "CIS", "country": "AZ", "min_amount": 0.75, "method_percent": 0.0101, 
                  "method_fix_amount": 2.34, "ps_percent": 0.00035, "ps_fixed_fee": 2, "ps_fixed_fee_currency": "EUR", 
                  "payout_currency": "USD", "method_fix_amount_currency": "EUR"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_costs/channel/merchant", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/channel/merchant")
	err := suite.router.setPaymentChannelCostMerchant(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_PaymentChannelCostMerchant_Delete() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_costs/channel/merchant/5be2d0b4b0b30d0007383ce6", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/channel/merchant/:" + requestParameterId)
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("5be2d0b4b0b30d0007383ce6")

	err := suite.router.deletePaymentChannelCostMerchant(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, rsp.Code)
		assert.Empty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostSystem_GetAll() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_costs/money_back/system/all", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/money_back/system/all")
	err := suite.router.getAllMoneyBackCostSystem(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostSystem_Get() {
	path := "/payment_costs/money_back/system?name=VISA&region=CIS&country=AZ&payoutCurrency=USD&days=10&undoReason=chargeback&paymentStage=1"
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath(path)
	err := suite.router.getMoneyBackCostSystem(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostSystem_Add() {
	bodyJson := `{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 0.0101, "fix_amount": 2.34, 
                  "payout_currency": "USD", "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_costs/money_back/system", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/money_back/system")
	err := suite.router.setMoneyBackCostSystem(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostSystem_Delete() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_costs/money_back/system/5be2d0b4b0b30d0007383ce6", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/money_back/system/:" + requestParameterId)
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("5be2d0b4b0b30d0007383ce6")

	err := suite.router.deleteMoneyBackCostSystem(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, rsp.Code)
		assert.Empty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostMerchant_GetAll() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_costs/money_back/merchant/all", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/money_back/merchant/all")
	err := suite.router.getAllMoneyBackCostMerchant(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostMerchant_Get() {
	path := "/payment_costs/money_back/merchant?name=VISA&region=CIS&country=AZ&payoutCurrency=USD&days=10&undoReason=chargeback&paymentStage=1"
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath(path)
	err := suite.router.getMoneyBackCostMerchant(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostMerchant_Add() {
	bodyJson := `{"name": "VISA", "region": "CIS", "country": "AZ", "percent": 0.0101, "fix_amount": 2.34, "fix_amount_currency": "USD",
                  "payout_currency": "USD", "undo_reason": "chargeback", "days_from": 0, "payment_stage": 1, 
                  "is_paid_by_merchant": true}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_costs/money_back/merchant", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/money_back/merchant")
	err := suite.router.setMoneyBackCostMerchant(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaymentCostTestSuite) TestPaymentCosts_MoneyBackCostMerchant_Delete() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_costs/money_back/merchant/5be2d0b4b0b30d0007383ce6", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payment_costs/money_back/merchant/:" + requestParameterId)
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("5be2d0b4b0b30d0007383ce6")

	err := suite.router.deleteMoneyBackCostMerchant(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, rsp.Code)
		assert.Empty(suite.T(), rsp.Body.String())
	}
}
