package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type VatReportsTestSuite struct {
	suite.Suite
	router *vatReportsRoute
	api    *Api
}

func Test_VatReports(t *testing.T) {
	suite.Run(t, new(VatReportsTestSuite))
}

func (suite *VatReportsTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &vatReportsRoute{Api: suite.api}
}

func (suite *VatReportsTestSuite) TearDownTest() {}

func (suite *VatReportsTestSuite) TestVatReports_getVatReportsDashboard() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/vat_reports", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("vat_reports")
	err := suite.router.getVatReportsDashboard(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *VatReportsTestSuite) TestVatReports_getVatReportsForCountry() {
	e := echo.New()
	q := make(url.Values)
	q.Set("limit", "100")
	q.Set("offset", "200")
	req := httptest.NewRequest(http.MethodGet, "/vat_reports/country/ru?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/vat_reports/country/:" + requestParameterCountry + "?" + q.Encode())
	ctx.SetParamNames(requestParameterCountry)
	ctx.SetParamValues("ru")

	err := suite.router.getVatReportsForCountry(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *VatReportsTestSuite) TestVatReports_getVatReportTransactions() {
	e := echo.New()
	q := make(url.Values)
	q.Set("limit", "100")
	q.Set("offset", "200")
	req := httptest.NewRequest(http.MethodGet, "/vat_reports/details/5ced34d689fce60bf4440829?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/vat_reports/details/:" + requestParameterId + "?" + q.Encode())
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("5ced34d689fce60bf4440829")

	err := suite.router.getVatReportTransactions(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *VatReportsTestSuite) TestVatReports_updateVatReportStatus() {
	bodyJson := `{"status": "paid"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/vat_reports/status/5ced34d689fce60bf4440829?", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/vat_reports/status/:" + requestParameterId)
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("5ced34d689fce60bf4440829")

	err := suite.router.updateVatReportStatus(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, rsp.Code)
	}
}
