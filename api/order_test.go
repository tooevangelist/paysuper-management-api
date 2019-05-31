package api

import (
	"bytes"
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-management-api/config"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
		config: &config.Config{
			Environment: "test",
		},
	}

	renderer := &Template{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("web/template/*.html")),
	}
	suite.api.Http.Renderer = renderer

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &orderRoute{Api: suite.api, projectManager: manager.InitProjectManager(nil, nil, mock.NewBillingServerOkMock())}

	err := suite.api.validate.RegisterValidation("uuid", suite.api.UuidValidator)
	assert.NoError(suite.T(), err, "Uuid validator registration failed")

	err = suite.api.validate.RegisterValidation("phone", suite.api.PhoneValidator)
	assert.NoError(suite.T(), err)
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
	ctx.SetParamValues(uuid.New().String(), bson.NewObjectId().Hex())

	err := suite.router.getRefund(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	refund := &billing.JsonRefund{}
	err = json.Unmarshal(rsp.Body.Bytes(), refund)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), refund.Id)
	assert.NotEmpty(suite.T(), refund.OrderId)
	assert.NotEmpty(suite.T(), refund.Currency)
	assert.Len(suite.T(), refund.Currency, 3)
}

func (suite *OrderTestSuite) TestOrder_GetRefund_RefundIdEmpty_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds/:refund_id")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.getRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "RefundId", httpErr.Message)
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
	assert.Regexp(suite.T(), "OrderId", httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_GetRefund_BillingServerError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds/:refund_id")
	ctx.SetParamNames(requestParameterOrderId, requestParameterRefundId)
	ctx.SetParamValues(uuid.New().String(), bson.NewObjectId().Hex())

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
	ctx.SetParamValues(uuid.New().String(), bson.NewObjectId().Hex())

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
	ctx.SetParamValues(uuid.New().String())

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
	assert.Regexp(suite.T(), "OrderId", httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ListRefunds_BillingServerError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

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
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.createRefund(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OrderTestSuite) TestOrder_CreateRefund_BindError() {
	data := `{"amount": "qwerty", "reason": "test"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/order/:order_id/refunds")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.createRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
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
	ctx.SetParamValues(uuid.New().String())

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
	ctx.SetParamValues(uuid.New().String())

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
	ctx.SetParamValues(uuid.New().String())

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err := suite.router.createRefund(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangeLanguage_Ok() {
	body := `{"lang": "en"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/language")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.changeLanguage(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OrderTestSuite) TestOrder_ChangeLanguage_OrderIdEmpty_Error() {
	body := `{"lang": "en"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.changeLanguage(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectOrderId, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangeLanguage_BindError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/language")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.changeLanguage(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangeLanguage_ValidationError() {
	body := `{"lang": "en"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/language")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues("some_value")

	err := suite.router.changeLanguage(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "OrderId", httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangeLanguage_BillingServerSystemError() {
	body := `{"lang": "en"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/language")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.changeLanguage(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangeLanguage_BillingServerErrorResult_Error() {
	body := `{"lang": "en"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/language")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err := suite.router.changeLanguage(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangePaymentAccount_Ok() {
	body := `{"method_id": "000000000000000000000000", "account": "4000000000000002"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/customer")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.changeCustomer(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OrderTestSuite) TestOrder_ChangePaymentAccount_OrderIdEmpty_Error() {
	body := `{"method_id": "000000000000000000000000", "account": "4000000000000002"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.changeCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectOrderId, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangePaymentAccount_BindError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/customer")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.changeCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangePaymentAccount_ValidationError() {
	body := `{"method_id": "some_value", "account": "4000000000000002"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/customer")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.changeCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "MethodId", httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangePaymentAccount_BillingServerSystemError() {
	body := `{"method_id": "000000000000000000000000", "account": "4000000000000002"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/customer")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.changeCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_ChangePaymentAccount_BillingServerErrorResult_Error() {
	body := `{"method_id": "000000000000000000000000", "account": "4000000000000002"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/customer")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err := suite.router.changeCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CalculateAmounts_Ok() {
	body := `{"country": "US", "city": "Washington", "zip": "98001"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/billing_address")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.processBillingAddress(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OrderTestSuite) TestOrder_CalculateAmounts_OrderIdEmpty_Error() {
	body := `{"country": "US", "city": "Washington", "zip": "98001"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.processBillingAddress(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectOrderId, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CalculateAmounts_BindError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/billing_address")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.processBillingAddress(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CalculateAmounts_ValidationError() {
	body := `{"country": "some_value", "city": "Washington", "zip": "98001"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/billing_address")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.processBillingAddress(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Country", httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CalculateAmounts_BillingServerSystemError() {
	body := `{"country": "US", "city": "Washington", "zip": "98001"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/billing_address")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.processBillingAddress(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CalculateAmounts_BillingServerErrorResult_Error() {
	body := `{"country": "US", "city": "Washington", "zip": "98001"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/api/v1/orders/:order_id/billing_address")
	ctx.SetParamNames(requestParameterOrderId)
	ctx.SetParamValues(uuid.New().String())

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err := suite.router.processBillingAddress(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateJson_Ok() {
	order := &billing.OrderCreateRequest{
		ProjectId:     bson.NewObjectId().Hex(),
		PaymentMethod: "BANKCARD",
		Currency:      "RUB",
		Amount:        100,
		Description:   "unit test",
		OrderId:       bson.NewObjectId().Hex(),
	}

	b, err := json.Marshal(order)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.router.createJson(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	var response *CreateOrderJsonProjectResponse
	err = json.Unmarshal(rsp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response.PaymentFormData)
}

func (suite *OrderTestSuite) TestOrder_CreateJson_WithUser_Ok() {
	order := &billing.OrderCreateRequest{
		ProjectId:     bson.NewObjectId().Hex(),
		PaymentMethod: "BANKCARD",
		Currency:      "RUB",
		Amount:        100,
		Description:   "unit test",
		OrderId:       bson.NewObjectId().Hex(),
		User: &billing.OrderUser{
			ExternalId:    bson.NewObjectId().Hex(),
			Ip:            "127.0.0.1",
			Locale:        "ru",
			Name:          "Unit Test",
			Email:         "test@unit.test",
			EmailVerified: true,
			Metadata:      map[string]string{"field1": "val1", "field2": "val2"},
		},
	}

	b, err := json.Marshal(order)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(HeaderXApiSignatureHeader, "signature")
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.router.createJson(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OrderTestSuite) TestOrder_CreateJson_BindError() {
	body := `{"project_id": "` + bson.NewObjectId().Hex() + `", "amount": "qwerty"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.createJson(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateJson_WithUser_EmptyRequestSignature_Error() {
	order := &billing.OrderCreateRequest{
		ProjectId:     bson.NewObjectId().Hex(),
		PaymentMethod: "BANKCARD",
		Currency:      "RUB",
		Amount:        100,
		Description:   "unit test",
		OrderId:       bson.NewObjectId().Hex(),
		User: &billing.OrderUser{
			ExternalId:    bson.NewObjectId().Hex(),
			Ip:            "127.0.0.1",
			Locale:        "ru",
			Name:          "Unit Test",
			Email:         "test@unit.test",
			EmailVerified: true,
			Metadata:      map[string]string{"field1": "val1", "field2": "val2"},
		},
	}

	b, err := json.Marshal(order)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.router.createJson(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorMessageSignatureHeaderIsEmpty, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateJson_WithUser_BillingServerSystemError() {
	order := &billing.OrderCreateRequest{
		ProjectId:     bson.NewObjectId().Hex(),
		PaymentMethod: "BANKCARD",
		Currency:      "RUB",
		Amount:        100,
		Description:   "unit test",
		OrderId:       bson.NewObjectId().Hex(),
		User: &billing.OrderUser{
			ExternalId:    bson.NewObjectId().Hex(),
			Ip:            "127.0.0.1",
			Locale:        "ru",
			Name:          "Unit Test",
			Email:         "test@unit.test",
			EmailVerified: true,
			Metadata:      map[string]string{"field1": "val1", "field2": "val2"},
		},
	}

	b, err := json.Marshal(order)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(HeaderXApiSignatureHeader, "signature")
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.router.createJson(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateJson_WithUser_BillingServerResultFail_Error() {
	order := &billing.OrderCreateRequest{
		ProjectId:     bson.NewObjectId().Hex(),
		PaymentMethod: "BANKCARD",
		Currency:      "RUB",
		Amount:        100,
		Description:   "unit test",
		OrderId:       bson.NewObjectId().Hex(),
		User: &billing.OrderUser{
			ExternalId:    bson.NewObjectId().Hex(),
			Ip:            "127.0.0.1",
			Locale:        "ru",
			Name:          "Unit Test",
			Email:         "test@unit.test",
			EmailVerified: true,
			Metadata:      map[string]string{"field1": "val1", "field2": "val2"},
		},
	}

	b, err := json.Marshal(order)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(HeaderXApiSignatureHeader, "signature")
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err = suite.router.createJson(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateJson_ValidationError() {
	body := `{"amount": -10}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.createJson(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Amount", httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateJson_OrderCreateError() {
	order := &billing.OrderCreateRequest{
		ProjectId:     mock.SomeMerchantId,
		PaymentMethod: "BANKCARD",
		Currency:      "RUB",
		Amount:        100,
		Description:   "unit test",
		OrderId:       bson.NewObjectId().Hex(),
	}

	b, err := json.Marshal(order)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.router.createJson(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateJson_PaymentFormJsonDataProcessError() {
	order := &billing.OrderCreateRequest{
		ProjectId:     mock.SomeMerchantId1,
		PaymentMethod: "BANKCARD",
		Currency:      "RUB",
		Amount:        100,
		Description:   "unit test",
		OrderId:       bson.NewObjectId().Hex(),
	}

	b, err := json.Marshal(order)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.router.createJson(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_CreateJson_ProductionEnvironment_Ok() {
	order := &billing.OrderCreateRequest{
		ProjectId:     bson.NewObjectId().Hex(),
		PaymentMethod: "BANKCARD",
		Currency:      "RUB",
		Amount:        100,
		Description:   "unit test",
		OrderId:       bson.NewObjectId().Hex(),
	}

	b, err := json.Marshal(order)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.router.config.Environment = EnvironmentProduction
	err = suite.router.createJson(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	var response *CreateOrderJsonProjectResponse
	err = json.Unmarshal(rsp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), response.PaymentFormData)
}

func (suite *OrderTestSuite) TestOrder_GetOrderForm_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/order/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(uuid.New().String())

	err := suite.router.getOrderForm(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
	assert.Equal(suite.T(), echo.MIMETextHTMLCharsetUTF8, rsp.Header().Get(echo.HeaderContentType))

	cookies := rsp.Result().Cookies()
	assert.True(suite.T(), len(cookies) == 1)
	assert.Equal(suite.T(), CustomerTokenCookiesName, cookies[0].Name)
	assert.True(suite.T(), cookies[0].HttpOnly)
}

func (suite *OrderTestSuite) TestOrder_GetOrderForm_TokenCookieExist_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	cookie := new(http.Cookie)
	cookie.Name = CustomerTokenCookiesName
	cookie.Value = bson.NewObjectId().Hex()
	cookie.Expires = time.Now().Add(time.Second * CustomerTokenCookiesLifetime)
	cookie.HttpOnly = true
	ctx.SetCookie(cookie)

	req.AddCookie(cookie)

	ctx.SetPath("/order/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.SomeMerchantId1)

	err := suite.router.getOrderForm(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
	assert.Equal(suite.T(), echo.MIMETextHTMLCharsetUTF8, rsp.Header().Get(echo.HeaderContentType))

	cookies := rsp.Result().Cookies()

	assert.True(suite.T(), len(cookies) > 1)
	assert.Equal(suite.T(), cookie.Name, cookies[0].Name)
	assert.Equal(suite.T(), cookie.Value, cookies[0].Value)
}

func (suite *OrderTestSuite) TestOrder_GetOrderForm_ParameterIdNotFound_Error() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.getOrderForm(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_GetOrderForm_BillingServerSystemError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/order/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.getOrderForm(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OrderTestSuite) TestOrder_GetOrders_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/order")

	err := suite.router.getOrders(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OrderTestSuite) TestOrder_GetOrders_BillingServerError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/order")

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.getOrders(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}
