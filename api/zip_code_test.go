package api

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var zipCodeRoutes = [][]string{
	{"/api/v1/zip", http.MethodGet},
}

type ZipCodeTestSuite struct {
	suite.Suite
	router *zipCodeRoute
	api    *Api
}

func Test_ZipCode(t *testing.T) {
	suite.Run(t, new(ZipCodeTestSuite))
}

func (suite *ZipCodeTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
	}

	suite.router = &zipCodeRoute{Api: suite.api}
}

func (suite *ZipCodeTestSuite) TearDownTest() {}

func (suite *ZipCodeTestSuite) TestOrder_InitRoutes_Ok() {
	api := suite.api.initZipCodeRoutes()
	assert.NotNil(suite.T(), api)

	routes := api.Http.Routes()
	routeCount := 0

	for _, v := range zipCodeRoutes {
		for _, r := range routes {
			if v[0] != r.Path || v[1] != r.Method {
				continue
			}

			routeCount++
		}
	}

	assert.Len(suite.T(), zipCodeRoutes, routeCount)
}

func (suite *ZipCodeTestSuite) TestOrder_CheckZip_Ok() {
	q := make(url.Values)
	q.Set("country", "US")
	q.Set("zip", "98")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/zip")

	err := suite.router.checkZip(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	data := &grpc.FindByZipCodeResponse{}
	err = json.Unmarshal(rsp.Body.Bytes(), data)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int32(1), data.Count)
	assert.Len(suite.T(), data.Items, 1)
}

func (suite *ZipCodeTestSuite) TestOrder_CheckZip_BindError() {
	q := make(url.Values)
	q.Set("country", "US")
	q.Set("zip", "98")
	q.Set("limit", "qwerty")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/zip")

	err := suite.router.checkZip(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *ZipCodeTestSuite) TestOrder_CheckZip_ValidateError() {
	q := make(url.Values)
	q.Set("zip", "98")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/zip")

	err := suite.router.checkZip(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Country", httpErr.Message)
}

func (suite *ZipCodeTestSuite) TestOrder_CheckZip_BillingServerError() {
	q := make(url.Values)
	q.Set("country", "US")
	q.Set("zip", "98")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/zip")

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err := suite.router.checkZip(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}
