package api

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type CurrencyTestSuite struct {
	suite.Suite
	router *CurrencyApiV1
	api    *Api
}

func Test_Currency(t *testing.T) {
	suite.Run(t, new(CurrencyTestSuite))
}

func (suite *CurrencyTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &CurrencyApiV1{Api: suite.api}
}

func (suite *CurrencyTestSuite) TearDownTest() {}

func (suite *CurrencyTestSuite) TestCurrency_ListCurrency_Ok() {
	e := echo.New()

	q := make(url.Values)
	q.Set(requestParameterLimit, "2")
	q.Set(requestParameterOffset, "1")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/currency?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.get(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *CurrencyTestSuite) TestCurrency_GetByName_BindingError() {
	e := echo.New()

	q := make(url.Values)
	q.Set("name", "1@#")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/currency/name?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.getByName(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), "field validation for 'CurrencyCode' failed on the 'alpha' tag", httpErr.Message)
}

func (suite *CurrencyTestSuite) TestCurrency_GetByName_NotFound() {
	e := echo.New()

	q := make(url.Values)
	q.Set("name", "unit")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/currency/name?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	bs := &mock.BillingService{}
	bs.On("GetCurrency", context.Background(), mock2.Anything).Return(nil, errors.New(""))
	suite.router.billingService = bs

	err := suite.router.getByName(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusNotFound, httpErr.Code)
	assert.Equal(suite.T(), "Currency not found", httpErr.Message)
}

func (suite *CurrencyTestSuite) TestCurrency_GetByName_Ok() {
	e := echo.New()

	q := make(url.Values)
	q.Set("name", "unit")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/currency/name?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	bs := &mock.BillingService{}
	bs.On("GetCurrency", context.Background(), mock2.Anything).Return(&billing.Currency{}, nil)
	suite.router.billingService = bs

	err := suite.router.getByName(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *CurrencyTestSuite) TestCurrency_GetById_BindingError() {
	e := echo.New()

	q := make(url.Values)
	q.Set("id", "a")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/currency?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.getById(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), newValidationError("incorrect currency identifier"), httpErr.Message)
}

func (suite *CurrencyTestSuite) TestCurrency_GetById_NotFound() {
	e := echo.New()

	q := make(url.Values)
	q.Set("id", "123")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/currency?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	bs := &mock.BillingService{}
	bs.On("GetCurrency", context.Background(), mock2.Anything).Return(nil, errors.New(""))
	suite.router.billingService = bs

	err := suite.router.getById(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusNotFound, httpErr.Code)
	assert.Equal(suite.T(), "Currency not found", httpErr.Message)
}

func (suite *CurrencyTestSuite) TestCurrency_GetById_Ok() {
	e := echo.New()

	q := make(url.Values)
	q.Set("id", "123")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/currency?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	bs := &mock.BillingService{}
	bs.On("GetCurrency", context.Background(), mock2.Anything).Return(&billing.Currency{}, nil)
	suite.router.billingService = bs

	err := suite.router.getById(ctx)
	assert.NoError(suite.T(), err)
}
