package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"testing"
)

type BalanceTestSuite struct {
	suite.Suite
	router *balanceRoute
	api    *Api
}

func Test_Balance(t *testing.T) {
	suite.Run(t, new(BalanceTestSuite))
}

func (suite *BalanceTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &balanceRoute{Api: suite.api}
}

func (suite *BalanceTestSuite) TearDownTest() {}

func (suite *BalanceTestSuite) TestBalance_getBalanceDashboard() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/balance", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("balance")
	err := suite.router.getMerchantBalance(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}
