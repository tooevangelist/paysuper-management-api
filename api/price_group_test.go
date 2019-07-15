package api

import (
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type PriceGroupTestSuite struct {
	suite.Suite
	router *PriceGroup
	api    *Api
}

func Test_PriceGroup(t *testing.T) {
	suite.Run(t, new(PriceGroupTestSuite))
}

func (suite *PriceGroupTestSuite) SetupTest() {
	suite.api = &Api{
		Http:     echo.New(),
		validate: validator.New(),
	}

	suite.router = &PriceGroup{Api: suite.api}
}

func (suite *PriceGroupTestSuite) TearDownTest() {}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getPriceGroupByCountry_BindError_RequiredCountry() {
	data := `{"variable": "test"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/country", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.getPriceGroupByCountry(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'Country' failed on the 'required' tag", msg.Details)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getPriceGroupByCountry_Error_BillingServer() {
	data := `{"country": "RU"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/country", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPriceGroupByCountry", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.api.billingService = billingService

	err := suite.router.getPriceGroupByCountry(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Regexp(suite.T(), errorMessagePriceGroupByCountry.Message, httpErr.Message)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getPriceGroupByCountry_Ok() {
	data := `{"country": "RU"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/country", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPriceGroupByCountry", mock2.Anything, mock2.Anything).Return(&billing.PriceGroup{}, nil)
	suite.api.billingService = billingService

	err := suite.router.getPriceGroupByCountry(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getCurrencyList_Error_BillingServer() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/currencies", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPriceGroupCurrencies", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.api.billingService = billingService

	err := suite.router.getCurrencyList(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Regexp(suite.T(), errorMessagePriceGroupCurrencyList.Message, httpErr.Message)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getCurrencyList_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/currencies", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPriceGroupCurrencies", mock2.Anything, mock2.Anything).Return(&grpc.PriceGroupCurrenciesResponse{}, nil)
	suite.api.billingService = billingService

	err := suite.router.getCurrencyList(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getCurrencyByRegion_BindError_RequiredRegion() {
	data := `{"variable": "test"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/region", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.getCurrencyByRegion(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'Region' failed on the 'required' tag", msg.Details)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getCurrencyByRegion_Error_BillingServer() {
	data := `{"region": "RUB"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/region", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPriceGroupCurrencyByRegion", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.api.billingService = billingService

	err := suite.router.getCurrencyByRegion(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Regexp(suite.T(), errorMessagePriceGroupCurrencyByRegion.Message, httpErr.Message)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getCurrencyByRegion_Ok() {
	data := `{"region": "RU"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/region", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPriceGroupCurrencyByRegion", mock2.Anything, mock2.Anything).Return(&grpc.PriceGroupCurrenciesResponse{}, nil)
	suite.api.billingService = billingService

	err := suite.router.getCurrencyByRegion(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getRecommendedPrice_BindError_RequiredAmount() {
	data := `{"variable": "test"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/region", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.getRecommendedPrice(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'Amount' failed on the 'required' tag", msg.Details)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getRecommendedPrice_Error_BillingServer() {
	data := `{"amount": 1}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/region", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPriceGroupRecommendedPrice", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.api.billingService = billingService

	err := suite.router.getRecommendedPrice(ctx)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Regexp(suite.T(), errorMessagePriceGroupRecommendedList.Message, httpErr.Message)
}

func (suite *PriceGroupTestSuite) TestPaymentMethod_getRecommendedPrice_Ok() {
	data := `{"amount": 1}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/price_group/region", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPriceGroupRecommendedPrice", mock2.Anything, mock2.Anything).Return(&grpc.PriceGroupRecommendedPriceResponse{}, nil)
	suite.api.billingService = billingService

	err := suite.router.getRecommendedPrice(ctx)
	assert.NoError(suite.T(), err)
}
