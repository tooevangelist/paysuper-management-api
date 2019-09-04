package api

import (
	"errors"
	"github.com/labstack/echo/v4"
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

type PaymentMethodTestSuite struct {
	suite.Suite
	router *PaymentMethodApiV1
	api    *Api
}

func Test_PaymentMethod(t *testing.T) {
	suite.Run(t, new(PaymentMethodTestSuite))
}

func (suite *PaymentMethodTestSuite) SetupTest() {
	suite.api = &Api{
		Http:     echo.New(),
		validate: validator.New(),
	}

	suite.router = &PaymentMethodApiV1{Api: suite.api}
}

func (suite *PaymentMethodTestSuite) TearDownTest() {}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_create_BindError_RequiredName() {
	data := `{"payment_system_id": "payment_system_id"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.create(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "Name", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_create_BindError_RequiredPaymentSystemId() {
	data := `{"name": "name"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.create(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "PaymentSystemId", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_create_BindError_Name() {
	data := `{"name": "!", "payment_system_id": "payment_system_id"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.create(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'Name' failed on the 'alphanum' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_create_BindError_PaymentSystemId() {
	data := `{"name": "name", "payment_system_id": "1"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.create(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'PaymentSystemId' failed on the 'len' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_Error_BillingServer() {
	data := `{"name": "name", "payment_system_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("CreateOrUpdatePaymentMethod", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.api.billingService = billingService

	err := suite.router.create(ctx)
	assert.Error(suite.T(), err)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_create_Ok() {
	data := `{"name": "name", "payment_system_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("CreateOrUpdatePaymentMethod", mock2.Anything, mock2.Anything).Return(&grpc.ChangePaymentMethodResponse{}, nil)
	suite.api.billingService = billingService

	err := suite.router.create(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_update_Ok() {
	data := `{"name": "name", "payment_system_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/payment_method", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("CreateOrUpdatePaymentMethod", mock2.Anything, mock2.Anything).Return(&grpc.ChangePaymentMethodResponse{}, nil)
	suite.api.billingService = billingService

	err := suite.router.update(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_getProductionSettings_BindError_RequiredPaymentMethodId() {
	data := `{"currency_a3": "rub"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.getProductionSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'PaymentMethodId' failed on the 'required' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_getProductionSettings_Error_BillingServer() {
	data := `{"currency_a3": "rub", "payment_method_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPaymentMethodProductionSettings", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.api.billingService = billingService

	err := suite.router.getProductionSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), errorUnknown.Message, httpErr.Message)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_getProductionSettings_Ok() {
	data := `{"currency_a3": "rub", "payment_method_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPaymentMethodProductionSettings", mock2.Anything, mock2.Anything).Return(&grpc.GetPaymentMethodSettingsResponse{}, nil)
	suite.api.billingService = billingService

	err := suite.router.getProductionSettings(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_getTestSettings_BindError_RequiredPaymentMethodId() {
	data := `{"currency_a3": "rub"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_method/1/test", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.getTestSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'PaymentMethodId' failed on the 'required' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_getTestSettings_Error_BillingServer() {
	data := `{"currency_a3": "rub", "payment_method_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_method/1/test", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPaymentMethodTestSettings", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.api.billingService = billingService

	err := suite.router.getTestSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), errorUnknown.Message, httpErr.Message)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_getTestSettings_Ok() {
	data := `{"currency_a3": "rub", "payment_method_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payment_method/1/test", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &mock.BillingService{}
	billingService.On("GetPaymentMethodTestSettings", mock2.Anything, mock2.Anything).Return(&grpc.GetPaymentMethodSettingsResponse{}, nil)
	suite.api.billingService = billingService

	err := suite.router.getTestSettings(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_createProductionSettings_BindError_RequiredParams() {
	data := `{"payment_method_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.createProductionSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'Params' failed on the 'required' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_createProductionSettings_BindError_RequiredPaymentMethodId() {
	data := `{"params": {}}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.createProductionSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'PaymentMethodId' failed on the 'required' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_updateProductionSettings_BindError_RequiredParams() {
	data := `{"payment_method_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.updateProductionSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'Params' failed on the 'required' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_updateProductionSettings_BindError_RequiredPaymentMethodId() {
	data := `{"params": {}}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.updateProductionSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'PaymentMethodId' failed on the 'required' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_createTestSettings_BindError_RequiredParams() {
	data := `{"payment_method_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.createTestSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'Params' failed on the 'required' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_createTestSettings_BindError_RequiredPaymentMethodId() {
	data := `{"params": {}}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.createTestSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'PaymentMethodId' failed on the 'required' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_updateTestSettings_BindError_RequiredParams() {
	data := `{"payment_method_id": "507f1f77bcf86cd799439011"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.updateTestSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'Params' failed on the 'required' tag", msg.Details)
}

func (suite *PaymentMethodTestSuite) TestPaymentMethod_updateTestSettings_BindError_RequiredPaymentMethodId() {
	data := `{"params": {}}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/payment_method/1/production", strings.NewReader(data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.updateTestSettings(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "field validation for 'PaymentMethodId' failed on the 'required' tag", msg.Details)
}
