package handlers

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	billMock "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/internal/test"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
	"time"
)

type RecurringTestSuite struct {
	suite.Suite
	router *RecurringRoute
	caller *test.EchoReqResCaller
}

func Test_Recurring(t *testing.T) {
	suite.Run(t, new(RecurringTestSuite))
}

func (suite *RecurringTestSuite) SetupTest() {
	var e error
	settings := test.DefaultSettings()

	bs := &billMock.BillingService{}
	bs.On("DeleteSavedCard", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.EmptyResponseWithStatus{Status: pkg.ResponseStatusOk}, nil)
	srv := common.Services{
		Billing: bs,
	}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {
		suite.router = NewRecurringRoute(set.HandlerSet, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *RecurringTestSuite) TearDownTest() {}

func (suite *RecurringTestSuite) TestRecurring_RemoveSavedCard_Ok() {
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = bson.NewObjectId().Hex()
	cookie.Expires = time.Now().Add(suite.router.cfg.CustomerTokenCookiesLifetime)
	cookie.HttpOnly = true

	body := []byte(`{"id": "ffffffffffffffffffffffff"}`)
	rsp, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Path(common.NoAuthGroupPath + removeSavedCardPath).
		AddCookie(cookie).
		Init(test.ReqInitJSON()).
		BodyBytes(body).
		Exec(suite.T())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.Empty(suite.T(), rsp.Body.String())
}

func (suite *RecurringTestSuite) TestRecurring_RemoveSavedCard_RequestBindingError() {
	body := []byte(`{"id": "ffffffffffffffffffffffff", "cookie": 123}`)
	_, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Path(common.NoAuthGroupPath + removeSavedCardPath).
		Init(test.ReqInitJSON()).
		BodyBytes(body).
		Exec(suite.T())
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *RecurringTestSuite) TestRecurring_RemoveSavedCard_ValidateError() {
	body := []byte(`{"id": "ffffffffffffffffffffffff"}`)
	_, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Path(common.NoAuthGroupPath + removeSavedCardPath).
		Init(test.ReqInitJSON()).
		BodyBytes(body).
		Exec(suite.T())
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "Cookie", msg.Details)
}

func (suite *RecurringTestSuite) TestRecurring_RemoveSavedCard_BillingServerSystemError() {
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = bson.NewObjectId().Hex()
	cookie.Expires = time.Now().Add(suite.router.cfg.CustomerTokenCookiesLifetime)
	cookie.HttpOnly = true

	bs := &billMock.BillingService{}
	bs.On("DeleteSavedCard", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(nil, errors.New("some error"))
	suite.router.dispatch.Services.Billing = bs

	body := []byte(`{"id": "ffffffffffffffffffffffff"}`)
	_, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Path(common.NoAuthGroupPath + removeSavedCardPath).
		AddCookie(cookie).
		Init(test.ReqInitJSON()).
		BodyBytes(body).
		Exec(suite.T())
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
}

func (suite *RecurringTestSuite) TestRecurring_RemoveSavedCard_BillingServerResultError() {
	cookie := new(http.Cookie)
	cookie.Name = common.CustomerTokenCookiesName
	cookie.Value = bson.NewObjectId().Hex()
	cookie.Expires = time.Now().Add(suite.router.cfg.CustomerTokenCookiesLifetime)
	cookie.HttpOnly = true

	errMsg := &grpc.ResponseErrorMessage{
		Code:    "000001",
		Message: "some error",
	}

	bs := &billMock.BillingService{}
	bs.On("DeleteSavedCard", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(
			&grpc.EmptyResponseWithStatus{
				Status:  pkg.ResponseStatusNotFound,
				Message: errMsg,
			},
			nil,
		)
	suite.router.dispatch.Services.Billing = bs

	body := []byte(`{"id": "ffffffffffffffffffffffff"}`)
	_, err := suite.caller.Builder().
		Method(http.MethodDelete).
		Path(common.NoAuthGroupPath + removeSavedCardPath).
		AddCookie(cookie).
		Init(test.ReqInitJSON()).
		BodyBytes(body).
		Exec(suite.T())
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusNotFound, httpErr.Code)
	assert.Equal(suite.T(), errMsg, httpErr.Message)
}
