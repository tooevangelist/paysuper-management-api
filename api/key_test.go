package api

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"testing"
)

type KeyTestSuite struct {
	suite.Suite
	router *keyRoute
	api    *Api
}

func Test_key(t *testing.T) {
	suite.Run(t, new(KeyTestSuite))
}

func (suite *KeyTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
		geoService: mock.NewGeoIpServiceTestOk(),
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.api.apiAuthProjectGroup = suite.api.Http.Group(apiAuthProjectGroupPath)
	suite.router = &keyRoute{Api: suite.api}
}

func (suite *KeyTestSuite) TearDownTest() {}

func (suite *KeyTestSuite) Test_GetKeyById_Ok() {
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/keys/:key_id")
	ctx.SetParamNames("key_id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &mock.BillingService{}
	billingService.On("GetKeyByID", mock2.Anything, mock2.Anything).Return(&grpc.GetKeyForOrderRequestResponse{
		Status: 200,
		Key: &billing.Key{
			Id: bson.NewObjectId().Hex(),
			Code: "XXXX-YYYY-ZZZZ",
		},
	}, nil)

	suite.api.billingService = billingService

	err := suite.router.getKeyInfo(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyTestSuite) Test_GetKeyById_InternalError() {
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/keys/:key_id")
	ctx.SetParamNames("key_id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &mock.BillingService{}
	billingService.On("GetKeyByID", mock2.Anything, mock2.Anything).Return(nil, errors.New("some error"))

	suite.api.billingService = billingService

	err := suite.router.getKeyInfo(ctx)
	assert.NotNil(suite.T(), err)
	httpErr := err.(*echo.HTTPError)
	assert.EqualValues(suite.T(), 500, httpErr.Code)
}

func (suite *KeyTestSuite) Test_GetKeyById_ServiceError() {
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/keys/:key_id")
	ctx.SetParamNames("key_id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &mock.BillingService{}
	billingService.On("GetKeyByID", mock2.Anything, mock2.Anything).Return(&grpc.GetKeyForOrderRequestResponse{
		Status: 404,
	}, nil)

	suite.api.billingService = billingService

	err := suite.router.getKeyInfo(ctx)
	assert.NotNil(suite.T(), err)
	httpErr := err.(*echo.HTTPError)
	assert.EqualValues(suite.T(), 404, httpErr.Code)
}

func (suite *KeyTestSuite) Test_GetKeyById_ValidationError() {
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/keys/:key_id")
	ctx.SetParamNames("key_id")
	ctx.SetParamValues("")

	billingService := &mock.BillingService{}
	billingService.On("GetKeyByID", mock2.Anything, mock2.Anything).Return(&grpc.GetKeyForOrderRequestResponse{
		Status: 200,
		Key: &billing.Key{
			Id: bson.NewObjectId().Hex(),
			Code: "XXXX-YYYY-ZZZZ",
		},
	}, nil)

	suite.api.billingService = billingService

	err := suite.router.getKeyInfo(ctx)
	assert.NotNil(suite.T(), err)
	httpErr := err.(*echo.HTTPError)
	assert.EqualValues(suite.T(), 400, httpErr.Code)
}