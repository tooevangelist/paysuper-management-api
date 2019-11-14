package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/paysuper/paysuper-management-api/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type ProductTestSuite struct {
	suite.Suite
	router *ProductRoute
	caller *test.EchoReqResCaller
}

func Test_Product(t *testing.T) {
	suite.Run(t, new(ProductTestSuite))
}

func (suite *ProductTestSuite) SetupTest() {
	user := &common.AuthUser{
		Id: "ffffffffffffffffffffffff",
		MerchantId: "ffffffffffffffffffffffff",
		Role: "owner",
	}

	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: mock.NewBillingServerOkMock(),
	}

	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		mw.Pre(test.PreAuthUserMiddleware(user))
		suite.router = NewProductRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *ProductTestSuite) TearDownTest() {}

func (suite *ProductTestSuite) TestProduct_getProductsList_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParam("limit", "100").
		Path(common.AuthUserGroupPath + productsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *ProductTestSuite) TestProduct_getProduct_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestProductId, "5c99391568add439ccf0ffaf").
		Path(common.AuthUserGroupPath + productsIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *ProductTestSuite) TestProduct_deleteProduct_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Params(":"+common.RequestProductId, "5c99391568add439ccf0ffaf").
		Path(common.AuthUserGroupPath + productsIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNoContent, res.Code)
	assert.Empty(suite.T(), res.Body.String())
}

func (suite *ProductTestSuite) TestProduct_createProduct_Ok() {
	bodyJson := `{"object": "product", "billing_type":"real", "pricing": "manual", "type": "simple_product", "sku": "ru_0_doom_2", "name": {"en": "Doom II"}, 
        "default_currency": "USD", "enabled": true, "prices": [{"amount": 12.93, "currency": "USD", "region": "USD"}], 
        "description":  {"en": "Doom II description"}, "long_description": {}, "project_id": "5bdc39a95d1e1100019fb7df"}`

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.AuthUserGroupPath + productsPath).
		BodyString(bodyJson).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *ProductTestSuite) TestProduct_updateProduct_Ok() {

	bodyJson := `{"object": "product", "billing_type":"real", "pricing": "manual", "type": "simple_product", "sku": "ru_0_doom_4", "name":  {"en": "Doom IV"}, 
        "default_currency": "USD", "enabled": true, "prices": [{"amount": 112.93, "currency": "USD", "region": "russia"}], 
        "description":  {"en": "Doom IV description"}, "long_description": {}, "project_id": "5bdc39a95d1e1100019fb7df"}`

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestProductId, "5c99391568add439ccf0ffaf").
		Path(common.AuthUserGroupPath + productsIdPath).
		BodyString(bodyJson).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *ProductTestSuite) TestProduct_updateProductNonVirtualPrice_Error() {
	bodyJson := `{"object": "product", "billing_type":"real", "pricing": "manual", "type": "simple_product", "sku": "ru_0_doom_4", "name":  {"en": "Doom IV"}, 
        "default_currency": "USD", "enabled": true, "prices": [{"amount": 112.93, "is_virtual_currency": false}], 
        "description":  {"en": "Doom IV description"}, "long_description": {}, "project_id": "5bdc39a95d1e1100019fb7df"}`

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestParameterId, "5c99391568add439ccf0ffaf").
		Path(common.AuthUserGroupPath + productsIdPath).
		BodyString(bodyJson).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)
	hErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), 400, hErr.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *ProductTestSuite) TestProduct_updateProductVirtualPrice_Ok() {
	bodyJson := `{"object": "product", "billing_type":"real", "pricing": "manual", "type": "simple_product", "sku": "ru_0_doom_4", "name":  {"en": "Doom IV"}, 
        "default_currency": "USD", "enabled": true, "prices": [{"amount": 112.93, "is_virtual_currency": true}], 
        "description":  {"en": "Doom IV description"}, "long_description": {}, "project_id": "5bdc39a95d1e1100019fb7df"}`

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestParameterId, "5c99391568add439ccf0ffaf").
		Path(common.AuthUserGroupPath + productsIdPath).
		BodyString(bodyJson).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}