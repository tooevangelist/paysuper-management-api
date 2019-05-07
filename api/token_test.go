package api

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
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
	router *tokenRoute
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

	err := suite.api.validate.RegisterValidation("phone", suite.api.PhoneValidator)
	assert.NoError(suite.T(), err)

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &tokenRoute{Api: suite.api}
}

func (suite *CustomerTestSuite) TearDownTest() {}

func (suite *CustomerTestSuite) TestCustomer_InitCustomerRoutes_Ok() {
	api, err := suite.api.initTokenRoutes()
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
	body := &grpc.TokenRequest{
		User: &billing.TokenUser{
			Id: bson.NewObjectId().Hex(),
			Email: &billing.TokenUserEmailValue{
				Value: "test@unit.test",
			},
			Phone: &billing.TokenUserPhoneValue{
				Value: "1234567890",
			},
			Name: &billing.TokenUserValue{
				Value: "Unit Test",
			},
			Ip: &billing.TokenUserIpValue{
				Value: "127.0.0.1",
			},
			Locale: &billing.TokenUserLocaleValue{
				Value: "ru",
			},
			Address: &billing.OrderBillingAddress{
				Country:    "RU",
				City:       "St.Petersburg",
				PostalCode: "190000",
				State:      "SPE",
			},
		},
		Settings: &billing.TokenSettings{
			OrderId:     bson.NewObjectId().Hex(),
			ProjectId:   bson.NewObjectId().Hex(),
			Currency:    "RUB",
			Amount:      100,
			Description: "test payment",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(HeaderXApiSignatureHeader, "signature")
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err = suite.router.createToken(ctx)
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

	err := suite.router.createToken(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_ValidationError() {
	body := &grpc.TokenRequest{
		User: &billing.TokenUser{
			Id: bson.NewObjectId().Hex(),
			Email: &billing.TokenUserEmailValue{
				Value: "test@unit.test",
			},
			Phone: &billing.TokenUserPhoneValue{
				Value: "1234567890",
			},
			Name: &billing.TokenUserValue{
				Value: "Unit Test",
			},
			Ip: &billing.TokenUserIpValue{
				Value: "127.0.0.1",
			},
			Locale: &billing.TokenUserLocaleValue{
				Value: "ru",
			},
			Address: &billing.OrderBillingAddress{
				Country:    "RU",
				City:       "St.Petersburg",
				PostalCode: "190000",
				State:      "SPE",
			},
		},
		Settings: &billing.TokenSettings{
			OrderId:     bson.NewObjectId().Hex(),
			ProjectId:   bson.NewObjectId().Hex(),
			Currency:    "RUB",
			Amount:      -100,
			Description: "test payment",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err = suite.router.createToken(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Amount", httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_CheckProjectRequestSignature_System_Error() {
	body := &grpc.TokenRequest{
		User: &billing.TokenUser{
			Id: bson.NewObjectId().Hex(),
			Email: &billing.TokenUserEmailValue{
				Value: "test@unit.test",
			},
			Phone: &billing.TokenUserPhoneValue{
				Value: "1234567890",
			},
			Name: &billing.TokenUserValue{
				Value: "Unit Test",
			},
			Ip: &billing.TokenUserIpValue{
				Value: "127.0.0.1",
			},
			Locale: &billing.TokenUserLocaleValue{
				Value: "ru",
			},
			Address: &billing.OrderBillingAddress{
				Country:    "RU",
				City:       "St.Petersburg",
				PostalCode: "190000",
				State:      "SPE",
			},
		},
		Settings: &billing.TokenSettings{
			OrderId:     bson.NewObjectId().Hex(),
			ProjectId:   bson.NewObjectId().Hex(),
			Currency:    "RUB",
			Amount:      100,
			Description: "test payment",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(HeaderXApiSignatureHeader, "signature")
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.router.createToken(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_CheckProjectRequestSignature_ResultError() {
	body := &grpc.TokenRequest{
		User: &billing.TokenUser{
			Id: bson.NewObjectId().Hex(),
			Email: &billing.TokenUserEmailValue{
				Value: "test@unit.test",
			},
			Phone: &billing.TokenUserPhoneValue{
				Value: "1234567890",
			},
			Name: &billing.TokenUserValue{
				Value: "Unit Test",
			},
			Ip: &billing.TokenUserIpValue{
				Value: "127.0.0.1",
			},
			Locale: &billing.TokenUserLocaleValue{
				Value: "ru",
			},
			Address: &billing.OrderBillingAddress{
				Country:    "RU",
				City:       "St.Petersburg",
				PostalCode: "190000",
				State:      "SPE",
			},
		},
		Settings: &billing.TokenSettings{
			OrderId:     bson.NewObjectId().Hex(),
			ProjectId:   bson.NewObjectId().Hex(),
			Currency:    "RUB",
			Amount:      100,
			Description: "test payment",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(HeaderXApiSignatureHeader, "signature")
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err = suite.router.createToken(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_ChangeCustomer_System_Error() {
	body := &grpc.TokenRequest{
		User: &billing.TokenUser{
			Id: bson.NewObjectId().Hex(),
			Email: &billing.TokenUserEmailValue{
				Value: "test@unit.test",
			},
			Phone: &billing.TokenUserPhoneValue{
				Value: "1234567890",
			},
			Name: &billing.TokenUserValue{
				Value: "Unit Test",
			},
			Ip: &billing.TokenUserIpValue{
				Value: "127.0.0.1",
			},
			Locale: &billing.TokenUserLocaleValue{
				Value: "ru",
			},
			Address: &billing.OrderBillingAddress{
				Country:    "RU",
				City:       "St.Petersburg",
				PostalCode: "190000",
				State:      "SPE",
			},
		},
		Settings: &billing.TokenSettings{
			OrderId:     bson.NewObjectId().Hex(),
			ProjectId:   bson.NewObjectId().Hex(),
			Currency:    "RUB",
			Amount:      100,
			Description: "test payment",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(HeaderXApiSignatureHeader, "signature")
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.router.createToken(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *CustomerTestSuite) TestCustomer_CreateCustomer_ChangeCustomer_ResultError() {
	body := &grpc.TokenRequest{
		User: &billing.TokenUser{
			Id: bson.NewObjectId().Hex(),
			Email: &billing.TokenUserEmailValue{
				Value: "test@unit.test",
			},
			Phone: &billing.TokenUserPhoneValue{
				Value: "1234567890",
			},
			Name: &billing.TokenUserValue{
				Value: "Unit Test",
			},
			Ip: &billing.TokenUserIpValue{
				Value: "127.0.0.1",
			},
			Locale: &billing.TokenUserLocaleValue{
				Value: "ru",
			},
			Address: &billing.OrderBillingAddress{
				Country:    "RU",
				City:       "St.Petersburg",
				PostalCode: "190000",
				State:      "SPE",
			},
		},
		Settings: &billing.TokenSettings{
			OrderId:     bson.NewObjectId().Hex(),
			ProjectId:   bson.NewObjectId().Hex(),
			Currency:    "RUB",
			Amount:      100,
			Description: "test payment",
		},
	}

	b, err := json.Marshal(body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(HeaderXApiSignatureHeader, "signature")
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err = suite.router.createToken(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}
