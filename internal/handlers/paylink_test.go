package handlers

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/paysuper/paysuper-management-api/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/url"
	"testing"
)

type PaylinkTestSuite struct {
	suite.Suite
	router *PayLinkRoute
	caller *test.EchoReqResCaller
}

func Test_Paylink(t *testing.T) {
	suite.Run(t, new(PaylinkTestSuite))
}

func (suite *PaylinkTestSuite) SetupTest() {
	user := &common.AuthUser{
		Id: "ffffffffffffffffffffffff",
		MerchantId: "ffffffffffffffffffffffff",
	}

	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: mock.NewBillingServerOkMock(),
	}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		mw.Pre(test.PreAuthUserMiddleware(user))
		suite.router = NewPayLinkRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *PaylinkTestSuite) TearDownTest() {}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinksList_Merchant_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthUserGroupPath + paylinksPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylink_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + paylinksIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinkUrl_Ok() {

	q := make(url.Values)
	q.Add("utm_source", "google")
	q.Add("utm_medium", "cpc")
	q.Add("utm_campaign", "someval")

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + paylinksUrlPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_deletePaylink_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + paylinksIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusNoContent, res.Code)
		assert.Empty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_createPaylink_Ok() {
	bodyJson := `{"expires_at": 1572307200, "products": ["5c3c962781258d0001e65930"], "project_id": "5c8f6a914dad6a0001839408", 
					"merchant_id": "5c8f6a914dad6a0001839408", "products_type": "product", "name": "unit-test", "no_expiry_date": false}`

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + paylinksPath).
		Init(test.ReqInitJSON()).
		BodyString(bodyJson).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_updatePaylink_Ok() {
	bodyJson := `{"expires_at": 1572307200, "products": ["5c3c962781258d0001e65930"], "project_id": "5c8f6a914dad6a0001839408", 
			"merchant_id": "5c8f6a914dad6a0001839408", "products_type": "product", "name": "unit-test", "no_expiry_date": false}`

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + paylinksIdPath).
		Init(test.ReqInitJSON()).
		BodyString(bodyJson).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinkStatSummary_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + paylinksIdStatSummaryPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinkStatByCountry_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + paylinksIdStatCountryPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinkStatByReferrer_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + paylinksIdStatReferrerPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinkStatByDate_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + paylinksIdStatDatePath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}

func (suite *PaylinkTestSuite) TestPaylink_getPaylinkStatByUtm_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + paylinksIdStatUtmPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, res.Code)
		assert.NotEmpty(suite.T(), res.Body.String())
	}
}
