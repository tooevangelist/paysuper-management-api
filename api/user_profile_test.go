package api

import (
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var userProfileRoutes = [][]string{
	{"/admin/api/v1/user_profile", http.MethodGet},
	{"/admin/api/v1/user_profile", http.MethodPatch},
}

type UserProfileTestSuite struct {
	suite.Suite
	router *userProfileRoute
	api    *Api
}

func Test_UserProfile(t *testing.T) {
	suite.Run(t, new(UserProfileTestSuite))
}

func (suite *UserProfileTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id:    "ffffffffffffffffffffffff",
			Email: "test@unit.test",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &userProfileRoute{Api: suite.api}

	err := suite.api.registerValidators()

	if err != nil {
		suite.FailNow("Validator registration failed", "%v", err)
	}
}

func (suite *UserProfileTestSuite) TearDownTest() {}

func (suite *UserProfileTestSuite) TestToken_InitCustomerRoutes_Ok() {
	api := suite.api.initUserProfileRoutes()
	assert.NotNil(suite.T(), api)

	routes := api.Http.Routes()
	routeCount := 0

	for _, v := range userProfileRoutes {
		for _, r := range routes {
			if v[0] != r.Path || v[1] != r.Method {
				continue
			}

			routeCount++
		}
	}

	assert.Len(suite.T(), userProfileRoutes, routeCount)
}

func (suite *UserProfileTestSuite) TestUserProfile_GetUserProfile_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/user_profile", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.getUserProfile(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *UserProfileTestSuite) TestUserProfile_GetUserProfile_UserIdNotFound_Error() {
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/user_profile", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.api.authUser.Id = ""

	err := suite.router.getUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusUnauthorized, httpErr.Code)
	assert.Equal(suite.T(), errorMessageAccessDenied, httpErr.Message)
}

func (suite *UserProfileTestSuite) TestUserProfile_GetUserProfile_BillingServerSystemError() {
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/user_profile", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.api.billingService = mock.NewBillingServerSystemErrorMock()

	err := suite.router.getUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *UserProfileTestSuite) TestUserProfile_GetUserProfile_BillingServerReturnError() {
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/user_profile", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.api.billingService = mock.NewBillingServerErrorMock()

	err := suite.router.getUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_Ok() {
	body := `{
		"personal": {"first_name": "unit test", "last_name": "test-unit", "position": "Software Developer"},
		"company": {
			"company_name": "Unit Test.-444", 
			"website": "http://localhost",
			"annual_income": {"from": 0, "to": 1000},
			"number_of_employees": {"from": 1, "to": 10},
			"kind_of_activity": "other"
		}
	}`

	req := httptest.NewRequest(http.MethodPatch, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.setUserProfile(ctx)
	assert.NoError(suite.T(), err)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_BindError() {
	body := `{"help": {"product_promotion_and_development": "unit test"}}`

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.setUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorRequestParamsIncorrect, httpErr.Message)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_ValidationError() {
	body := `{"help": {"product_promotion_and_development": false}}`

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.api.authUser.Id = ""

	err := suite.router.setUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	err1, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)

	assert.Regexp(suite.T(), "UserId", err1.Details)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_ValidationUserNameError() {
	body := `{"personal": {"first_name": "unit test♂", "last_name": "test-unit", "position": "Software Developer"}}`

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.setUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	err1, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)

	assert.Regexp(suite.T(), "Name", err1.Details)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_ValidationUserPositionError() {
	body := `{"personal": {"first_name": "unit test", "last_name": "test-unit", "position": "qwerty"}}`

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.setUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	err1, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)

	assert.Regexp(suite.T(), "Position", err1.Details)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_ValidationAnnualIncomeError() {
	body := `{
		"company": {
			"company_name": "Unit Test", 
			"website": "http://localhost", 
			"annual_income": {"from": 78898, "to": 9998},
			"number_of_employees": {"from": 1, "to": 10}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.setUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	err1, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)

	assert.Regexp(suite.T(), "AnnualIncome", err1.Details)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_ValidationNumberOfEmployeesError() {
	body := `{
		"company": {
			"company_name": "Unit Test", 
			"website": "http://localhost", 
			"annual_income": {"from": 0, "to": 1000},
			"number_of_employees": {"from": 23872, "to": 129}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.setUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	err1, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)

	assert.Regexp(suite.T(), "NumberOfEmployees", err1.Details)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_ValidationCompanyNameError() {
	body := `{
		"company": {
			"company_name": "Unit Test♂", 
			"website": "http://localhost", 
			"annual_income": {"from": 0, "to": 1000},
			"number_of_employees": {"from": 1, "to": 10}
		}
	}`

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.setUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	err1, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)

	assert.Regexp(suite.T(), "CompanyName", err1.Details)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_BillingServerSystemError() {
	body := `{"help": {"product_promotion_and_development": false}}`

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.api.billingService = mock.NewBillingServerSystemErrorMock()

	err := suite.router.setUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *UserProfileTestSuite) TestUserProfile_SetUserProfile_BillingServerReturnError() {
	body := `{"help": {"product_promotion_and_development": false}}`

	req := httptest.NewRequest(http.MethodPost, "/admin/api/v1/user_profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.api.billingService = mock.NewBillingServerErrorMock()

	err := suite.router.setUserProfile(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}
