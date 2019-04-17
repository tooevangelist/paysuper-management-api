package api

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var customersRoutes = [][]string{
	{"/api/v1/customers", http.MethodPost},
}

type CustomerTestSuite struct {
	suite.Suite
	router *customerRoute
	api    *Api
}

func Test_Customer(t *testing.T) {
	suite.Run(t, new(CustomerTestSuite))
}

func (suite *CustomerTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.apiAuthProjectGroup = suite.api.Http.Group(apiAuthProjectGroupPath)
	suite.api.apiAuthProjectGroup.Use(suite.api.RawBodyMiddleware)
	suite.api.apiAuthProjectGroup.Use(suite.api.RequestSignatureMiddleware)

	err := suite.api.validate.RegisterValidation("phone", suite.api.PhoneValidator)
	assert.NoError(suite.T(), err)

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &customerRoute{Api: suite.api}
}

func (suite *CustomerTestSuite) TearDownTest() {}

func (suite *CustomerTestSuite) TestCustomer_InitCustomerRoutes_Ok() {
	api, err := suite.api.initCustomerRoutes()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), api)

	routes := api.Http.Routes()
	routeCount := 0

	for _, v := range customersRoutes {
		for _, r := range routes {
			if v[0] != r.Path || v[1] != r.Method {
				continue
			}

			routeCount++
		}
	}

	assert.Len(suite.T(), customersRoutes, routeCount)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_Ok() {
	body := &billing.Customer{
		ProjectId:     bson.NewObjectId().Hex(),
		ExternalId:    bson.NewObjectId().Hex(),
		Email:         "test@unit.test",
		EmailVerified: true,
		Phone:         "9123456789",
		Ip:            "127.0.0.1",
		Locale:        "ru",
		Metadata: map[string]string{
			"field1": "value1",
			"field2": "value2",
		},
		Address: &billing.OrderBillingAddress{
			Country:    "US",
			City:       "New York",
			PostalCode: "000000",
			State:      "CA",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err = suite.router.createCustomer(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_BindError() {
	body := `{"email_verified": "qwerty", "metadata": "qwerty"}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.createCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_ValidationError() {
	body := &billing.Customer{
		ExternalId:    bson.NewObjectId().Hex(),
		Email:         "test@unit.test",
		EmailVerified: true,
		Phone:         "9123456789",
		Ip:            "127.0.0.1",
		Locale:        "ru",
		Metadata: map[string]string{
			"field1": "value1",
			"field2": "value2",
		},
		Address: &billing.OrderBillingAddress{
			Country:    "US",
			City:       "New York",
			PostalCode: "000000",
			State:      "CA",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err = suite.router.createCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "ProjectId", httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_CheckProjectRequestSignature_System_Error() {
	body := &billing.Customer{
		ProjectId:     bson.NewObjectId().Hex(),
		ExternalId:    bson.NewObjectId().Hex(),
		Email:         "test@unit.test",
		EmailVerified: true,
		Phone:         "9123456789",
		Ip:            "127.0.0.1",
		Locale:        "ru",
		Metadata: map[string]string{
			"field1": "value1",
			"field2": "value2",
		},
		Address: &billing.OrderBillingAddress{
			Country:    "US",
			City:       "New York",
			PostalCode: "000000",
			State:      "CA",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.router.createCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_CheckProjectRequestSignature_ResultError() {
	body := &billing.Customer{
		ProjectId:     bson.NewObjectId().Hex(),
		ExternalId:    bson.NewObjectId().Hex(),
		Email:         "test@unit.test",
		EmailVerified: true,
		Phone:         "9123456789",
		Ip:            "127.0.0.1",
		Locale:        "ru",
		Metadata: map[string]string{
			"field1": "value1",
			"field2": "value2",
		},
		Address: &billing.OrderBillingAddress{
			Country:    "US",
			City:       "New York",
			PostalCode: "000000",
			State:      "CA",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err = suite.router.createCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_ChangeCustomer_System_Error() {
	body := &billing.Customer{
		ProjectId:     mock.SomeMerchantId,
		ExternalId:    bson.NewObjectId().Hex(),
		Email:         "test@unit.test",
		EmailVerified: true,
		Phone:         "9123456789",
		Ip:            "127.0.0.1",
		Locale:        "ru",
		Metadata: map[string]string{
			"field1": "value1",
			"field2": "value2",
		},
		Address: &billing.OrderBillingAddress{
			Country:    "US",
			City:       "New York",
			PostalCode: "000000",
			State:      "CA",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.router.createCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_ChangeCustomer_ResultError() {
	body := &billing.Customer{
		ProjectId:     mock.SomeMerchantId,
		ExternalId:    bson.NewObjectId().Hex(),
		Email:         "test@unit.test",
		EmailVerified: true,
		Phone:         "9123456789",
		Ip:            "127.0.0.1",
		Locale:        "ru",
		Metadata: map[string]string{
			"field1": "value1",
			"field2": "value2",
		},
		Address: &billing.OrderBillingAddress{
			Country:    "US",
			City:       "New York",
			PostalCode: "000000",
			State:      "CA",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err = suite.router.createCustomer(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}
