package api

import (
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

type ProductTestSuite struct {
	suite.Suite
	router *productRoute
	api    *Api
}

func Test_Product(t *testing.T) {
	suite.Run(t, new(ProductTestSuite))
}

func (suite *ProductTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &productRoute{Api: suite.api}
}

func (suite *ProductTestSuite) TearDownTest() {}

func (suite *ProductTestSuite) TestProduct_getProductsList_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/product", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/product")

	err := suite.router.getProductsList(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *ProductTestSuite) TestProduct_getProduct_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/product/5c99391568add439ccf0ffaf", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/product/:id")
	ctx.SetParamNames("id")
	ctx.SetParamValues("5c99391568add439ccf0ffaf")

	err := suite.router.getProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *ProductTestSuite) TestProduct_deleteProduct_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/product/5c99391568add439ccf0ffaf", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/product/:id")
	ctx.SetParamNames("id")
	ctx.SetParamValues("5c99391568add439ccf0ffaf")

	err := suite.router.deleteProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNoContent, rsp.Code)
	assert.Empty(suite.T(), rsp.Body.String())
}

func (suite *ProductTestSuite) TestProduct_createProduct_Ok() {
	bodyJson := `{"object": "product", "type": "simple_product", "sku": "ru_0_doom_2", "name": "Doom II", 
        "default_currency": "USD", "enabled": true, "prices": [{"amount": 12.93, "currency": "USD"}], 
        "description": "Doom II description", "long_description": ""}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/product", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/product")

	err := suite.router.createProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *ProductTestSuite) TestProduct_updateProduct_Ok() {
	bodyJson := `{"object": "product", "type": "simple_product", "sku": "ru_0_doom_4", "name": "Doom IV", 
        "default_currency": "USD", "enabled": true, "prices": [{"amount": 112.93, "currency": "USD"}], 
        "description": "Doom IV description", "long_description": ""}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/product/5c99391568add439ccf0ffaf", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/product/:id")
	ctx.SetParamNames("id")
	ctx.SetParamValues("5c99391568add439ccf0ffaf")

	err := suite.router.updateProduct(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}
