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

type PaylinkTestSuite struct {
	suite.Suite
	router *paylinkRoute
	api    *Api
}

func Test_Paylink(t *testing.T) {
	suite.Run(t, new(PaylinkTestSuite))
}

func (suite *PaylinkTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		paylinkService: mock.NewPaymentLinkOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &paylinkRoute{Api: suite.api}
}

func (suite *PaylinkTestSuite) TearDownTest() {}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinksList_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/paylinks/project/5c10ff51d5be4b0001bca600", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/paylinks/project/:" + requestParameterProjectId)
	ctx.SetParamNames(requestParameterProjectId)
	ctx.SetParamValues("5c10ff51d5be4b0001bca600")

	err := suite.router.getPaylinksList(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylink_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/paylinks/21784001599a47e5a69ac28f7af2ec22", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/paylinks/:" + requestParameterId)
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("21784001599a47e5a69ac28f7af2ec22")

	err := suite.router.getPaylink(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinkStat_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/paylinks/21784001599a47e5a69ac28f7af2ec22/stat", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/paylinks/:" + requestParameterId + "/stat")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("21784001599a47e5a69ac28f7af2ec22")

	err := suite.router.getPaylinkStat(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinkUrl_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/paylinks/21784001599a47e5a69ac28f7af2ec22/url", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/paylinks/:" + requestParameterId + "/url")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("21784001599a47e5a69ac28f7af2ec22")

	err := suite.router.getPaylinkUrl(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_deletePaylink_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/paylinks/21784001599a47e5a69ac28f7af2ec22", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/paylinks/:" + requestParameterId)
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("21784001599a47e5a69ac28f7af2ec22")

	err := suite.router.deletePaylink(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, rsp.Code)
		assert.Empty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_createPaylink_Ok() {
	bodyJson := `{"life_days": 7, "products": ["5c3c962781258d0001e65930"], "project_id": "5c8f6a914dad6a0001839408"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/paylinks", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/paylinks")

	err := suite.router.createPaylink(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_updatePaylink_Ok() {
	bodyJson := `{"life_days": 30, "products": ["5c3c962781258d0001e65930"], "project_id": "5c8f6a914dad6a0001839408"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/paylinks/21784001599a47e5a69ac28f7af2ec22", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/paylinks/:" + requestParameterId)
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("21784001599a47e5a69ac28f7af2ec22")

	err := suite.router.updatePaylink(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}
