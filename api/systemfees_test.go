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

type SystemFeeTestSuite struct {
	suite.Suite
	router *systemFeeRoute
	api    *Api
}

func Test_SystemFee(t *testing.T) {
	suite.Run(t, new(SystemFeeTestSuite))
}

func (suite *SystemFeeTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &systemFeeRoute{Api: suite.api}
}

func (suite *SystemFeeTestSuite) TearDownTest() {}

func (suite *SystemFeeTestSuite) TestSystemFees_getSystemFeesList_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/systemfees", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/systemfees")
	err := suite.router.getSystemFeesList(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *SystemFeeTestSuite) TestSystemFees_addSystemFee_Ok() {
	bodyJson := `{ "method_id": "5be2d0b4b0b30d0007383ce6", "region": "EU", "card_brand": "MASTERCARD", 
		"fees": [ { "min_amounts": { "EUR": 0, "USD": 0 }, 
		"transaction_cost": { "percent": 1.15, "percent_currency": "EUR", "fix_amount": 0.2, "fix_currency": "EUR" }, 
		"authorization_fee": { "percent": 0, "percent_currency": "EUR", "fix_amount": 0.1, "fix_currency": "EUR" } } ], 
		"user_id": "5cb6e4aa68add437e8a8f0fa" }`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/systemfees", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/systemfees")
	err := suite.router.addSystemFee(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.Empty(suite.T(), rsp.Body.String())
	}
}

func (suite *SystemFeeTestSuite) TestSystemFees_addSystemFee_Fail() {
	bodyJson := `{ "method_id": "5be2d0b4b0b30d0007383ce6", "region": "EU", "card_brand": "MASTERCARD", 
		"fees": [], 
		"user_id": "5cb6e4aa68add437e8a8f0fa" }`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/systemfees", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/systemfees")
	err := suite.router.addSystemFee(ctx)

	assert.Error(suite.T(), err)
}
