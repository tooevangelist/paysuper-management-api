package api

import (
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type OrderTestSuite struct {
	suite.Suite
	router *orderRoute
	api    *Api
}

func Test_Order(t *testing.T) {
	suite.Run(t, new(OrderTestSuite))
}

func (suite *OrderTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &orderRoute{Api: suite.api}
}

func (suite *OrderTestSuite) TearDownTest() {}

func (suite *OrderTestSuite) TestOrder_GetRefund_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds/:refund_id")
	ctx.SetParamNames(requestParameterOrderId, requestParameterRefundId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), bson.NewObjectId().Hex())

	err := suite.router.getRefund(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OrderTestSuite) TestOrder_GetRefund_RefundIdEmpty_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds/:refund_id")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectRefundId, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_GetRefund_OrderIdEmpty_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds/:refund_id")
	ctx.SetParamNames(requestParameterRefundId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectOrderId, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_GetRefund_BillingServerError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds/:refund_id")
	ctx.SetParamNames(requestParameterOrderId, requestParameterRefundId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()

	err := suite.router.getRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_GetRefund_BillingServer_RefundNotFound_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds/:refund_id")
	ctx.SetParamNames(requestParameterOrderId, requestParameterRefundId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerErrorMock()

	err := suite.router.getRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusNotFound, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ListRefunds_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.listRefunds(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OrderTestSuite) TestOrder_ListRefunds_BindError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")

	err := suite.router.listRefunds(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectOrderId, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ListRefunds_BillingServerError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.listRefunds(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateRefund_Ok() {
	data := `{"amount": 10, "reason": "test"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.createRefund(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OrderTestSuite) TestOrder_CreateRefund_BindError() {
	data := `{"amount": 10, "reason": "test"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")

	err := suite.router.createRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectOrderId, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateRefund_ValidationError() {
	data := `{"amount": -10, "reason": "test"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.createRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Amount", httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateRefund_BillingServerError() {
	data := `{"amount": 10, "reason": "test"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.createRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateRefund_BillingServer_CreateError() {
	data := `{"amount": 10, "reason": "test"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err := suite.router.createRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}
