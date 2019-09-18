package api

import (
	"bytes"
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type ProjectTestSuite struct {
	suite.Suite
	router *projectRoute
	api    *Api
}

func Test_project(t *testing.T) {
	suite.Run(t, new(ProjectTestSuite))
}

func (suite *ProjectTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &projectRoute{Api: suite.api}
}

func (suite *ProjectTestSuite) TearDownTest() {}

func (suite *ProjectTestSuite) TestProject_CreateProject_Ok() {
	body := &billing.Project{
		MerchantId:         bson.NewObjectId().Hex(),
		Name:               map[string]string{"en": "A", "ru": "А"},
		CallbackCurrency:   "RUB",
		CallbackProtocol:   pkg.ProjectCallbackProtocolEmpty,
		LimitsCurrency:     "RUB",
		MinPaymentAmount:   0,
		MaxPaymentAmount:   15000,
		IsProductsCheckout: false,
	}

	b, err := json.Marshal(&body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err = suite.router.createProject(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *ProjectTestSuite) TestProject_CreateProject_BindError() {
	body := `{"name": "qwerty"}`

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.createProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	//assert.Equal(suite.T(), errorRequestParamsIncorrect, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_CreateProject_ValidationError() {
	body := &billing.Project{
		MerchantId:         bson.NewObjectId().Hex(),
		Name:               map[string]string{"en": "A", "ru": "А"},
		CallbackCurrency:   "RUB",
		CallbackProtocol:   pkg.ProjectCallbackProtocolEmpty,
		LimitsCurrency:     "RUB",
		MinPaymentAmount:   -100,
		MaxPaymentAmount:   15000,
		IsProductsCheckout: false,
	}

	b, err := json.Marshal(&body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err = suite.router.createProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), newValidationError("MinPaymentAmount"), httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_CreateProject_BillingServerError() {
	body := &billing.Project{
		MerchantId:         bson.NewObjectId().Hex(),
		Name:               map[string]string{"en": "A", "ru": "А"},
		CallbackCurrency:   "RUB",
		CallbackProtocol:   pkg.ProjectCallbackProtocolEmpty,
		LimitsCurrency:     "RUB",
		MinPaymentAmount:   100,
		MaxPaymentAmount:   15000,
		IsProductsCheckout: false,
	}

	b, err := json.Marshal(&body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.router.createProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_CreateProject_BillingServerResultError() {
	body := &billing.Project{
		MerchantId:         bson.NewObjectId().Hex(),
		Name:               map[string]string{"en": "A", "ru": "А"},
		CallbackCurrency:   "RUB",
		CallbackProtocol:   pkg.ProjectCallbackProtocolEmpty,
		LimitsCurrency:     "RUB",
		MinPaymentAmount:   100,
		MaxPaymentAmount:   15000,
		IsProductsCheckout: false,
	}

	b, err := json.Marshal(&body)
	assert.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err = suite.router.createProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_UpdateProject_Ok() {
	body := `{"min_payment_amount": 10}`

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.updateProject(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *ProjectTestSuite) TestProject_UpdateProject_BindError() {
	body := `{"name": "qwerty"}`

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.updateProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	//assert.Equal(suite.T(), errorRequestParamsIncorrect, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_UpdateProject_ValidationError() {
	body := `{"min_payment_amount": -10}`

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.updateProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), newValidationError("MinPaymentAmount"), httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_UpdateProject_BillingServerError() {
	body := `{"min_payment_amount": 10}`

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.SomeMerchantId)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.updateProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_UpdateProject_BillingServerResultError() {
	body := `{"min_payment_amount": 10}`

	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.SomeMerchantId)

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err := suite.router.updateProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_GetProject_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.getProject(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *ProjectTestSuite) TestProject_GetProject_ValidationError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.getProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), newValidationError("ProjectId"), httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_GetProject_BillingServerError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.getProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_GetProject_BillingServerResultError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err := suite.router.getProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_ListProjects_Ok() {
	q := make(url.Values)
	q.Set(requestParameterLimit, "-100")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.listProjects(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *ProjectTestSuite) TestProject_ListProjects_BindError() {
	q := make(url.Values)
	q.Set(requestParameterLimit, "qwerty")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.listProjects(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorRequestParamsIncorrect, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_ListProjects_ValidationError() {
	q := make(url.Values)
	q.Set(requestParameterOffset, "-100")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.listProjects(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), newValidationError("Offset"), httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_ListProjects_BillingServerError() {
	q := make(url.Values)
	q.Set(requestParameterOffset, "100")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.listProjects(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_DeleteProject_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.router.deleteProject(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *ProjectTestSuite) TestProject_DeleteProject_ValidateError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.router.deleteProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), newValidationError("ProjectId"), httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_DeleteProject_BillingServerError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.router.deleteProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *ProjectTestSuite) TestProject_DeleteProject_BillingServerResultError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/projects/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err := suite.router.deleteProject(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}
