package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	billMock "github.com/paysuper/paysuper-billing-server/pkg/mocks"
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

type KeyProductTestSuite struct {
	suite.Suite
	router *keyProductRoute
	api    *Api
}

func Test_keyProduct(t *testing.T) {
	suite.Run(t, new(KeyProductTestSuite))
}

func (suite *KeyProductTestSuite) SetupTest() {
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
	suite.router = &keyProductRoute{Api: suite.api}
}

func (suite *KeyProductTestSuite) TearDownTest() {}

func (suite *KeyProductTestSuite) TestProject_RemovePlatform_Ok() {
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id/platforms/:platform_id")
	ctx.SetParamNames("key_product_id", "platform_id")
	ctx.SetParamValues(bson.NewObjectId().Hex(), "steam")

	err := suite.router.removePlatformForKeyProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyProductTestSuite) TestProject_PublishKeyProduct_Ok() {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id/publish")
	ctx.SetParamNames("key_product_id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.publishKeyProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyProductTestSuite) TestProject_GetPlatformList_Error() {
	req := httptest.NewRequest(http.MethodGet, "/platforms?limit=qwe", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/platforms?limit=qwe")

	err := suite.router.getPlatformsList(ctx)
	assert.Error(suite.T(), err)
}

func (suite *KeyProductTestSuite) TestProject_GetPlatformList_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/platforms?limit=10", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/platforms?limit=10")

	err := suite.router.getPlatformsList(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/platforms", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp = httptest.NewRecorder()
	ctx = suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/platforms?limit")

	err = suite.router.getPlatformsList(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/platforms?limit=300000", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp = httptest.NewRecorder()
	ctx = suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/platforms?limit=300000")

	err = suite.router.getPlatformsList(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyProductTestSuite) TestProject_GetListKeyProduct_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.getKeyProductList(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyProductTestSuite) TestProject_GetKeyProduct_ValidationError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id")
	ctx.SetParamNames("key_product_id")
	ctx.SetParamValues("")

	err := suite.router.getKeyProductById(ctx)
	assert.Error(suite.T(), err)
	e, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), 400, e.Code)
	assert.NotEmpty(suite.T(), e.Message)
}

func (suite *KeyProductTestSuite) TestProject_GetKeyProduct_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id")
	ctx.SetParamNames("key_product_id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getKeyProductById(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyProductTestSuite) TestProject_CreateKeyProduct_Ok() {
	body := &grpc.CreateOrUpdateKeyProductRequest{
		MerchantId:      bson.NewObjectId().Hex(),
		Name:            map[string]string{"en": "A", "ru": "А"},
		Description:     map[string]string{"en": "A", "ru": "А"},
		DefaultCurrency: "RUB",
		ProjectId:       bson.NewObjectId().Hex(),
		Object:          "bla-bla-bla",
		Sku:             "some_sku",
	}

	b, err := json.Marshal(&body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err = suite.router.createKeyProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyProductTestSuite) TestProject_ChangeKeyProduct_ValidationError() {
	body := &grpc.CreateOrUpdateKeyProductRequest{
		Id:              bson.NewObjectId().Hex(),
		MerchantId:      bson.NewObjectId().Hex(),
		DefaultCurrency: "RUB",
		ProjectId:       bson.NewObjectId().Hex(),
		Sku:             "some_sku",
	}

	b, err := json.Marshal(&body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id")
	ctx.SetParamNames("key_product_id")
	ctx.SetParamValues(body.Id)

	err = suite.router.changeKeyProduct(ctx)
	assert.Error(suite.T(), err)
	e, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), 400, e.Code)
	assert.NotEmpty(suite.T(), e.Message)
}

func (suite *KeyProductTestSuite) TestProject_ChangeKeyProduct_Ok() {
	body := &grpc.CreateOrUpdateKeyProductRequest{
		Id:              bson.NewObjectId().Hex(),
		MerchantId:      bson.NewObjectId().Hex(),
		Name:            map[string]string{"en": "A", "ru": "А"},
		Description:     map[string]string{"en": "A", "ru": "А"},
		DefaultCurrency: "RUB",
		ProjectId:       bson.NewObjectId().Hex(),
		Object:          "bla-bla-bla",
		Sku:             "some_sku",
	}

	b, err := json.Marshal(&body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id")
	ctx.SetParamNames("key_product_id")
	ctx.SetParamValues(body.Id)

	err = suite.router.changeKeyProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyProductTestSuite) TestProject_CreateKeyProduct_ValidationError() {
	body := &grpc.CreateOrUpdateKeyProductRequest{
		MerchantId:      bson.NewObjectId().Hex(),
		Description:     map[string]string{"en": "A", "ru": "А"},
		DefaultCurrency: "RUB",
		ProjectId:       bson.NewObjectId().Hex(),
		Sku:             "some_sku",
	}

	b, err := json.Marshal(&body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err = suite.router.createKeyProduct(ctx)

	assert.Error(suite.T(), err)
	e, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), 400, e.Code)
	assert.NotEmpty(suite.T(), e.Message)
}

func (suite *KeyProductTestSuite) TestProject_getKeyProduct_ValidationError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/key-products/1", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	billingService := &billMock.BillingService{}
	billingService.On("GetKeyProduct", mock2.Anything, mock2.Anything).Return(&grpc.KeyProductResponse{}, nil)
	suite.api.billingService = billingService

	err := suite.router.getKeyProduct(ctx)
	assert.Error(suite.T(), err)
	err2, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, err2.Code)
	assert.NotEmpty(suite.T(), err2.Message)
}

func (suite *KeyProductTestSuite) TestProject_getKeyProduct_BillingServer() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/key-products/1", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id")
	ctx.SetParamNames("key_product_id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &billMock.BillingService{}
	billingService.On("GetKeyProductInfo", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.api.billingService = billingService

	err := suite.router.getKeyProduct(ctx)
	assert.Error(suite.T(), err)
	err2, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, err2.Code)
	assert.NotEmpty(suite.T(), err2.Message)
}

func (suite *KeyProductTestSuite) TestProject_getKeyProductWithCurrency_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/key-products/1?currency=USD", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id")
	ctx.SetParamNames("key_product_id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &billMock.BillingService{}
	billingService.On("GetKeyProductInfo", mock2.Anything, mock2.Anything).Return(&grpc.GetKeyProductInfoResponse{
		Status: 200,
		KeyProduct: &grpc.KeyProductInfo{
			LongDescription: "Description",
			Name:            "Name",
			Platforms: []*grpc.PlatformPriceInfo{
				{Name: "Steam", Id: "steam", Price: &grpc.ProductPriceInfo{Currency: "USD", Amount: 10, Region: "USD"}},
			},
		},
	}, nil)
	suite.api.billingService = billingService

	err := suite.router.getKeyProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyProductTestSuite) TestProject_getKeyProductWithCountry_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/key-products/1?country=RUS", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id")
	ctx.SetParamNames("key_product_id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &billMock.BillingService{}
	billingService.On("GetKeyProductInfo", mock2.Anything, mock2.Anything).Return(&grpc.GetKeyProductInfoResponse{
		Status: 200,
		KeyProduct: &grpc.KeyProductInfo{
			LongDescription: "Description",
			Name:            "Name",
			Platforms: []*grpc.PlatformPriceInfo{
				{Name: "Steam", Id: "steam", Price: &grpc.ProductPriceInfo{Currency: "USD", Amount: 10, Region: "USD"}},
			},
		},
	}, nil)
	suite.api.billingService = billingService

	err := suite.router.getKeyProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *KeyProductTestSuite) Test_Init_Ok() {
	assert.NotNil(suite.T(), suite.api.initKeyProductRoutes())
}

func (suite *KeyProductTestSuite) TestProject_getKeyProduct_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/key-products/1", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/key-products/:key_product_id")
	ctx.SetParamNames("key_product_id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	billingService := &billMock.BillingService{}
	billingService.On("GetKeyProductInfo", mock2.Anything, mock2.Anything).Return(&grpc.GetKeyProductInfoResponse{
		Status: 200,
		KeyProduct: &grpc.KeyProductInfo{
			LongDescription: "Description",
			Name:            "Name",
			Platforms: []*grpc.PlatformPriceInfo{
				{Name: "Steam", Id: "steam", Price: &grpc.ProductPriceInfo{Currency: "USD", Amount: 10, Region: "USD"}},
			},
		},
	}, nil)
	suite.api.billingService = billingService

	err := suite.router.getKeyProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}
