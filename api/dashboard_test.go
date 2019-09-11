package api

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/config"
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

var dashboardRoutes = [][]string{
	{"/admin/api/v1/merchants/:id/dashboard/main", http.MethodGet},
	{"/admin/api/v1/merchants/:id/dashboard/revenue_dynamics", http.MethodGet},
	{"/admin/api/v1/merchants/:id/dashboard/base", http.MethodGet},
}

type DashboardTestSuite struct {
	suite.Suite
	router *dashboardRoute
	api    *Api
}

func Test_Dashboard(t *testing.T) {
	suite.Run(t, new(DashboardTestSuite))
}

func (suite *DashboardTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id:    "ffffffffffffffffffffffff",
			Email: "test@unit.test",
		},
		config: &config.Config{
			HttpScheme: "http",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &dashboardRoute{Api: suite.api}

	err := suite.api.registerValidators()

	if err != nil {
		suite.FailNow("Validator registration failed", "%v", err)
	}

	bs := &mock.BillingService{}
	bs.On("GetDashboardMainReport", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.GetDashboardMainResponse{Status: pkg.ResponseStatusOk, Item: &grpc.DashboardMainReport{}}, nil)
	bs.On("GetDashboardRevenueDynamicsReport", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(
			&grpc.GetDashboardRevenueDynamicsReportResponse{
				Status: pkg.ResponseStatusOk,
				Item:   []*grpc.DashboardRevenueDynamicReportItem{},
			},
			nil,
		)
	bs.On("GetDashboardBaseReport", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(
			&grpc.GetDashboardBaseReportResponse{
				Status: pkg.ResponseStatusOk,
				Item:   &grpc.DashboardBaseReports{},
			},
			nil,
		)
	suite.api.billingService = bs
}

func (suite *DashboardTestSuite) TearDownTest() {}

func (suite *DashboardTestSuite) TestDashboard_InitDashboardRoutes_Ok() {
	api := suite.api.initDashboardRoutes()
	assert.NotNil(suite.T(), api)

	routes := api.Http.Routes()
	routeCount := 0

	for _, v := range dashboardRoutes {
		for _, r := range routes {
			if v[0] != r.Path || v[1] != r.Method {
				continue
			}

			routeCount++
		}
	}

	assert.Len(suite.T(), dashboardRoutes, routeCount)
}

func (suite *DashboardTestSuite) TestDashboard_GetMainReports_Ok() {
	q := make(url.Values)
	q.Set("period", "current_month")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getMainReports(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *DashboardTestSuite) TestDashboard_GetMainReports_ValidationError() {
	req := httptest.NewRequest(http.MethodGet, "/?period=123", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getMainReports(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectPeriod, httpErr.Message)
}

func (suite *DashboardTestSuite) TestDashboard_GetMainReports_BillingServerSystemError() {
	req := httptest.NewRequest(http.MethodGet, "/?period=current_month", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	bs := &mock.BillingService{}
	bs.On("GetDashboardMainReport", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(nil, errors.New("some error"))
	suite.api.billingService = bs

	err := suite.router.getMainReports(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *DashboardTestSuite) TestDashboard_GetMainReports_BillingServerReturnError() {
	req := httptest.NewRequest(http.MethodGet, "/?period=current_month", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	bs := &mock.BillingService{}
	bs.On("GetDashboardMainReport", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(
			&grpc.GetDashboardMainResponse{
				Status:  pkg.ResponseStatusBadData,
				Message: &grpc.ResponseErrorMessage{Message: "some error"},
			},
			nil,
		)
	suite.api.billingService = bs

	err := suite.router.getMainReports(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "some error", msg.Message)
}

func (suite *DashboardTestSuite) TestDashboard_GetRevenueDynamicsReport_Ok() {
	q := make(url.Values)
	q.Set("period", "current_month")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getRevenueDynamicsReport(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *DashboardTestSuite) TestDashboard_GetRevenueDynamicsReport_ValidationError() {
	req := httptest.NewRequest(http.MethodGet, "/?period=123", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getRevenueDynamicsReport(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectPeriod, httpErr.Message)
}

func (suite *DashboardTestSuite) TestDashboard_GetRevenueDynamicsReport_BillingServerSystemError() {
	req := httptest.NewRequest(http.MethodGet, "/?period=current_month", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	bs := &mock.BillingService{}
	bs.On("GetDashboardRevenueDynamicsReport", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(nil, errors.New("some error"))
	suite.api.billingService = bs

	err := suite.router.getRevenueDynamicsReport(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *DashboardTestSuite) TestDashboard_GetRevenueDynamicsReport_BillingServerReturnError() {
	req := httptest.NewRequest(http.MethodGet, "/?period=current_month", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	bs := &mock.BillingService{}
	bs.On("GetDashboardRevenueDynamicsReport", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(
			&grpc.GetDashboardRevenueDynamicsReportResponse{
				Status:  pkg.ResponseStatusBadData,
				Message: &grpc.ResponseErrorMessage{Message: "some error"},
			},
			nil,
		)
	suite.api.billingService = bs

	err := suite.router.getRevenueDynamicsReport(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "some error", msg.Message)
}

func (suite *DashboardTestSuite) TestDashboard_GetBaseReports_Ok() {
	q := make(url.Values)
	q.Set("period", "current_month")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getBaseReports(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *DashboardTestSuite) TestDashboard_GetBaseReports_ValidationError() {
	req := httptest.NewRequest(http.MethodGet, "/?period=123", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getBaseReports(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectPeriod, httpErr.Message)
}

func (suite *DashboardTestSuite) TestDashboard_GetBaseReports_BillingServerSystemError() {
	req := httptest.NewRequest(http.MethodGet, "/?period=current_month", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	bs := &mock.BillingService{}
	bs.On("GetDashboardBaseReport", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(nil, errors.New("some error"))
	suite.api.billingService = bs

	err := suite.router.getBaseReports(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *DashboardTestSuite) TestDashboard_GetBaseReports_BillingServerReturnError() {
	req := httptest.NewRequest(http.MethodGet, "/?period=current_month", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/dashboard/main")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	bs := &mock.BillingService{}
	bs.On("GetDashboardBaseReport", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(
			&grpc.GetDashboardBaseReportResponse{
				Status:  pkg.ResponseStatusBadData,
				Message: &grpc.ResponseErrorMessage{Message: "some error"},
			},
			nil,
		)
	suite.api.billingService = bs

	err := suite.router.getBaseReports(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), "some error", msg.Message)
}
