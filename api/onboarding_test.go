package api

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var onboardingRoutes = [][]string{
	{"/admin/api/v1/merchants", http.MethodGet},
	{"/admin/api/v1/merchants/:id", http.MethodGet},
	{"/admin/api/v1/merchants", http.MethodPost},
	{"/admin/api/v1/merchants", http.MethodPut},
	{"/admin/api/v1/merchants/:id/change-status", http.MethodPut},
	{"/admin/api/v1/merchants/:merchant_id/notifications", http.MethodPost},
	{"/admin/api/v1/merchants/:merchant_id/notifications/:notification_id", http.MethodGet},
	{"/admin/api/v1/merchants/:merchant_id/notifications/:notification_id", http.MethodGet},
	{"/admin/api/v1/merchants/:merchant_id/notifications/:notification_id/mark-as-read", http.MethodPut},
	{"/admin/api/v1/merchants/:merchant_id/methods/:method_id", http.MethodGet},
	{"/admin/api/v1/merchants/:merchant_id/methods", http.MethodGet},
	{"/admin/api/v1/merchants/:merchant_id/methods", http.MethodPut},
}

type OnboardingTestSuite struct {
	suite.Suite
	handler *onboardingRoute
	api     *Api
}

func Test_Onboarding(t *testing.T) {
	suite.Run(t, new(OnboardingTestSuite))
}

func (suite *OnboardingTestSuite) SetupTest() {
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
	err := suite.api.validate.RegisterValidation("phone", suite.api.PhoneValidator)
	assert.NoError(suite.T(), err)

	suite.handler = &onboardingRoute{Api: suite.api}
}

func (suite *OnboardingTestSuite) TearDownTest() {}

func (suite *OnboardingTestSuite) TestOnboarding_InitRoutes_Ok() {
	api := suite.api.initOnboardingRoutes()
	assert.NotNil(suite.T(), api)

	routes := api.Http.Routes()
	routeCount := 0

	for _, v := range onboardingRoutes {
		for _, r := range routes {
			if v[0] != r.Path || v[1] != r.Method {
				continue
			}

			routeCount++
		}
	}

	assert.Len(suite.T(), onboardingRoutes, routeCount)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchant_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchant/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.getMerchant(ctx)
	assert.NoError(suite.T(), err)

	obj := &billing.Merchant{}
	err = json.Unmarshal(rsp.Body.Bytes(), obj)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.Equal(suite.T(), mock.OnboardingMerchantMock.Id, obj.Id)
	assert.Equal(suite.T(), mock.OnboardingMerchantMock.City, obj.City)
	assert.Equal(suite.T(), mock.OnboardingMerchantMock.City, obj.City)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchant_BillingServiceUnavailable_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchant/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerSystemErrorMock()

	err := suite.handler.getMerchant(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchant_LogicError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchant/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()

	err := suite.handler.getMerchant(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchant_EmptyId_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchant/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("")

	suite.handler.billingService = mock.NewBillingServerErrorMock()

	err := suite.handler.getMerchant(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIdIsEmpty, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListMerchants_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.handler.listMerchants(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)

	var m []*billing.Merchant
	err = json.Unmarshal(rsp.Body.Bytes(), &m)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), m, 3)
	assert.Equal(suite.T(), mock.OnboardingMerchantMock.Id, m[0].Id)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListMerchants_BindingError() {
	e := echo.New()

	q := make(url.Values)
	q.Set(requestParameterIsSigned, bson.NewObjectId().Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.handler.listMerchants(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListMerchants_BillingServiceUnavailable_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.handler.billingService = mock.NewBillingServerSystemErrorMock()

	err := suite.handler.listMerchants(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateMerchant_Ok() {
	merchant := &grpc.OnboardingRequest{
		Name:               mock.OnboardingMerchantMock.Name,
		AlternativeName:    "",
		Website:            "https://unit.test",
		Country:            "RU",
		State:              "St.Petersburg",
		Zip:                "190000",
		City:               "St.Petersburg",
		Address:            "",
		AddressAdditional:  "",
		RegistrationNumber: "",
		TaxId:              "",
		Contacts: &billing.MerchantContact{
			Authorized: &billing.MerchantContactAuthorized{
				Name:     "Unit Test",
				Email:    "test@unit.test",
				Phone:    "1234567890",
				Position: "Unit Test",
			},
			Technical: &billing.MerchantContactTechnical{
				Name:  "Unit Test",
				Email: "test@unit.test",
				Phone: "1234567890",
			},
		},
		Banking: &grpc.OnboardingBanking{
			Currency:      "RUB",
			Name:          "Bank name",
			Address:       "Unknown",
			AccountNumber: "1234567890",
			Swift:         "TEST",
			Details:       "",
		},
	}

	b, err := json.Marshal(merchant)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.handler.changeMerchant(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	var merchantRsp *billing.Merchant
	err = json.Unmarshal(rsp.Body.Bytes(), &merchantRsp)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), len(merchantRsp.Id) > 0)
	assert.Equal(suite.T(), merchant.Name, merchantRsp.Name)
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateMerchant_ValidationError() {
	merchant := &grpc.OnboardingRequest{
		AlternativeName:    "",
		Website:            "https://unit.test",
		Country:            "RU",
		State:              "St.Petersburg",
		Zip:                "190000",
		City:               "St.Petersburg",
		Address:            "",
		AddressAdditional:  "",
		RegistrationNumber: "",
		TaxId:              "",
		Contacts: &billing.MerchantContact{
			Authorized: &billing.MerchantContactAuthorized{
				Name:     "Unit Test",
				Email:    "test@unit.test",
				Phone:    "1234567890",
				Position: "Unit Test",
			},
			Technical: &billing.MerchantContactTechnical{
				Name:  "Unit Test",
				Email: "test@unit.test",
				Phone: "1234567890",
			},
		},
	}

	b, err := json.Marshal(merchant)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.handler.changeMerchant(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Banking", httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateMerchant_BillingServiceUnavailable_Error() {
	merchant := &grpc.OnboardingRequest{
		Name:               mock.OnboardingMerchantMock.Name,
		AlternativeName:    "",
		Website:            "https://unit.test",
		Country:            "RU",
		State:              "St.Petersburg",
		Zip:                "190000",
		City:               "St.Petersburg",
		Address:            "",
		AddressAdditional:  "",
		RegistrationNumber: "",
		TaxId:              "",
		Contacts: &billing.MerchantContact{
			Authorized: &billing.MerchantContactAuthorized{
				Name:     "Unit Test",
				Email:    "test@unit.test",
				Phone:    "1234567890",
				Position: "Unit Test",
			},
			Technical: &billing.MerchantContactTechnical{
				Name:  "Unit Test",
				Email: "test@unit.test",
				Phone: "1234567890",
			},
		},
		Banking: &grpc.OnboardingBanking{
			Currency:      "RUB",
			Name:          "Bank name",
			Address:       "Unknown",
			AccountNumber: "1234567890",
			Swift:         "TEST",
			Details:       "",
		},
	}

	b, err := json.Marshal(merchant)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err = suite.handler.changeMerchant(ctx)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_UpdateMerchant_Ok() {
	merchant := &grpc.OnboardingRequest{
		Name:               mock.OnboardingMerchantMock.Name,
		AlternativeName:    "",
		Website:            "https://unit.test",
		Country:            "RU",
		State:              "St.Petersburg",
		Zip:                "190000",
		City:               "St.Petersburg",
		Address:            "",
		AddressAdditional:  "",
		RegistrationNumber: "",
		TaxId:              "",
		Contacts: &billing.MerchantContact{
			Authorized: &billing.MerchantContactAuthorized{
				Name:     "Unit Test",
				Email:    "test@unit.test",
				Phone:    "1234567890",
				Position: "Unit Test",
			},
			Technical: &billing.MerchantContactTechnical{
				Name:  "Unit Test",
				Email: "test@unit.test",
				Phone: "1234567890",
			},
		},
		Banking: &grpc.OnboardingBanking{
			Currency:      "RUB",
			Name:          "Bank name",
			Address:       "Unknown",
			AccountNumber: "1234567890",
			Swift:         "TEST",
			Details:       "",
		},
	}

	b, err := json.Marshal(merchant)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.handler.changeMerchant(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)

	var merchantRsp *billing.Merchant
	err = json.Unmarshal(rsp.Body.Bytes(), &merchantRsp)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), len(merchantRsp.Id) > 0)

	merchant.Id = merchantRsp.Id
	merchant.Name = "New merchant name"

	b, err = json.Marshal(merchant)
	assert.NoError(suite.T(), err)

	req = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp = httptest.NewRecorder()
	ctx = e.NewContext(req, rsp)

	err = suite.handler.changeMerchant(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	var merchantRsp1 *billing.Merchant
	err = json.Unmarshal(rsp.Body.Bytes(), &merchantRsp1)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), merchantRsp.Id, merchantRsp1.Id)
	assert.NotEqual(suite.T(), merchantRsp.Name, merchantRsp1.Name)
	assert.Equal(suite.T(), merchant.Name, merchantRsp1.Name)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeMerchantStatus_Ok() {
	merchant := &grpc.OnboardingRequest{
		Name:               mock.OnboardingMerchantMock.Name,
		AlternativeName:    "",
		Website:            "https://unit.test",
		Country:            "RU",
		State:              "St.Petersburg",
		Zip:                "190000",
		City:               "St.Petersburg",
		Address:            "",
		AddressAdditional:  "",
		RegistrationNumber: "",
		TaxId:              "",
		Contacts: &billing.MerchantContact{
			Authorized: &billing.MerchantContactAuthorized{
				Name:     "Unit Test",
				Email:    "test@unit.test",
				Phone:    "1234567890",
				Position: "Unit Test",
			},
			Technical: &billing.MerchantContactTechnical{
				Name:  "Unit Test",
				Email: "test@unit.test",
				Phone: "1234567890",
			},
		},
		Banking: &grpc.OnboardingBanking{
			Currency:      "RUB",
			Name:          "Bank name",
			Address:       "Unknown",
			AccountNumber: "1234567890",
			Swift:         "TEST",
			Details:       "",
		},
	}

	b, err := json.Marshal(merchant)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.handler.changeMerchant(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	var mRsp *billing.Merchant
	err = json.Unmarshal(rsp.Body.Bytes(), &mRsp)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), len(mRsp.Id) > 0)
	assert.Equal(suite.T(), pkg.MerchantStatusDraft, mRsp.Status)

	changeStatusReq := &grpc.MerchantChangeStatusRequest{
		MerchantId: mRsp.Id,
		Status:     pkg.MerchantStatusAgreementRequested,
	}
	b, err = json.Marshal(changeStatusReq)
	assert.NoError(suite.T(), err)

	req = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp = httptest.NewRecorder()
	ctx = e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:id/change-status")
	ctx.SetParamNames("id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err = suite.handler.changeMerchantStatus(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeMerchantStatus_ValidationError() {
	merchant := &grpc.OnboardingRequest{
		Name:               mock.OnboardingMerchantMock.Name,
		AlternativeName:    "",
		Website:            "https://unit.test",
		Country:            "RU",
		State:              "St.Petersburg",
		Zip:                "190000",
		City:               "St.Petersburg",
		Address:            "",
		AddressAdditional:  "",
		RegistrationNumber: "",
		TaxId:              "",
		Contacts: &billing.MerchantContact{
			Authorized: &billing.MerchantContactAuthorized{
				Name:     "Unit Test",
				Email:    "test@unit.test",
				Phone:    "1234567890",
				Position: "Unit Test",
			},
			Technical: &billing.MerchantContactTechnical{
				Name:  "Unit Test",
				Email: "test@unit.test",
				Phone: "1234567890",
			},
		},
		Banking: &grpc.OnboardingBanking{
			Currency:      "RUB",
			Name:          "Bank name",
			Address:       "Unknown",
			AccountNumber: "1234567890",
			Swift:         "TEST",
			Details:       "",
		},
	}

	b, err := json.Marshal(merchant)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.handler.changeMerchant(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	var mRsp *billing.Merchant
	err = json.Unmarshal(rsp.Body.Bytes(), &mRsp)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), len(mRsp.Id) > 0)
	assert.Equal(suite.T(), pkg.MerchantStatusDraft, mRsp.Status)

	changeStatusReq := &grpc.MerchantChangeStatusRequest{}
	b, err = json.Marshal(changeStatusReq)
	assert.NoError(suite.T(), err)

	req = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp = httptest.NewRecorder()
	ctx = e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:id/change-status")
	ctx.SetParamNames("id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err = suite.handler.changeMerchantStatus(ctx)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Status", httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeMerchantStatus_BillingServerUnavailable_Error() {
	merchant := &grpc.OnboardingRequest{
		Name:               mock.OnboardingMerchantMock.Name,
		AlternativeName:    "",
		Website:            "https://unit.test",
		Country:            "RU",
		State:              "St.Petersburg",
		Zip:                "190000",
		City:               "St.Petersburg",
		Address:            "",
		AddressAdditional:  "",
		RegistrationNumber: "",
		TaxId:              "",
		Contacts: &billing.MerchantContact{
			Authorized: &billing.MerchantContactAuthorized{
				Name:     "Unit Test",
				Email:    "test@unit.test",
				Phone:    "1234567890",
				Position: "Unit Test",
			},
			Technical: &billing.MerchantContactTechnical{
				Name:  "Unit Test",
				Email: "test@unit.test",
				Phone: "1234567890",
			},
		},
		Banking: &grpc.OnboardingBanking{
			Currency:      "RUB",
			Name:          "Bank name",
			Address:       "Unknown",
			AccountNumber: "1234567890",
			Swift:         "TEST",
			Details:       "",
		},
	}

	b, err := json.Marshal(merchant)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.handler.changeMerchant(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	var mRsp *billing.Merchant
	err = json.Unmarshal(rsp.Body.Bytes(), &mRsp)
	assert.NoError(suite.T(), err)

	assert.True(suite.T(), len(mRsp.Id) > 0)
	assert.Equal(suite.T(), pkg.MerchantStatusDraft, mRsp.Status)

	changeStatusReq := &grpc.MerchantChangeStatusRequest{
		MerchantId: mRsp.Id,
		Status:     pkg.MerchantStatusAgreementRequested,
	}
	b, err = json.Marshal(changeStatusReq)
	assert.NoError(suite.T(), err)

	req = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp = httptest.NewRecorder()
	ctx = e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:id/change-status")
	ctx.SetParamNames("id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err = suite.handler.changeMerchantStatus(ctx)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateNotification_Ok() {
	n := &grpc.NotificationRequest{
		MerchantId: bson.NewObjectId().Hex(),
		UserId:     suite.handler.authUser.Id,
		Title:      "Title",
		Message:    "Message",
	}

	b, err := json.Marshal(n)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err = suite.handler.createNotification(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateNotification_ValidationError() {
	n := &grpc.NotificationRequest{
		MerchantId: bson.NewObjectId().Hex(),
		UserId:     suite.handler.authUser.Id,
		Title:      "",
		Message:    "Message",
	}

	b, err := json.Marshal(n)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err = suite.handler.createNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Title", httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateNotification_BillingServerUnavailable_Error() {
	n := &grpc.NotificationRequest{
		MerchantId: bson.NewObjectId().Hex(),
		UserId:     suite.handler.authUser.Id,
		Title:      "Title",
		Message:    "Message",
	}

	b, err := json.Marshal(n)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err = suite.handler.createNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetNotification_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications/:notification_id")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterNotificationId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), bson.NewObjectId().Hex())

	err := suite.handler.getNotification(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_GetNotification_EmptyId_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications/:notification_id")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.getNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectNotificationId, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetNotification_BillingServerUnavailable_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications/:notification_id")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterNotificationId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err := suite.handler.getNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusNotFound, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListNotifications_Ok() {
	e := echo.New()

	q := make(url.Values)
	q.Set(requestParameterMerchantId, bson.NewObjectId().Hex())
	q.Set(requestParameterUserId, bson.NewObjectId().Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.handler.listNotifications(ctx)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_ListNotifications_ValidationError() {
	e := echo.New()

	q := make(url.Values)
	q.Set(requestParameterMerchantId, bson.NewObjectId().Hex())
	q.Set(requestParameterUserId, "invalid_object_id")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.handler.listNotifications(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectUserId, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListNotifications_BillingServerError() {
	e := echo.New()

	q := make(url.Values)
	q.Set(requestParameterMerchantId, bson.NewObjectId().Hex())
	q.Set(requestParameterUserId, bson.NewObjectId().Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err := suite.handler.listNotifications(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_MarkAsReadNotification_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications/:notification_id/mark-as-read")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterNotificationId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), bson.NewObjectId().Hex())

	err := suite.handler.markAsReadNotification(ctx)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_MarkAsReadNotification_EmptyId_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications/:notification_id/mark-as-read")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.markAsReadNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectNotificationId, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_MarkAsReadNotification_BillingServer_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications/:notification_id/mark-as-read")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterNotificationId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err := suite.handler.markAsReadNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetPaymentMethod_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchant/:merchant_id/payment-method/:payment_method_id")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterPaymentMethodId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), bson.NewObjectId().Hex())

	err := suite.handler.getPaymentMethod(ctx)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_GetPaymentMethod_ValidationError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchant/:merchant_id/payment-method/:payment_method_id")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.getPaymentMethod(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectPaymentMethodId, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetPaymentMethod_BillingServer_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchant/:merchant_id/payment-method/:payment_method_id")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterPaymentMethodId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err := suite.handler.getPaymentMethod(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListPaymentMethods_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/methods")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.listPaymentMethods(ctx)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_ListPaymentMethods_ValidationError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/methods")

	err := suite.handler.listPaymentMethods(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectMerchantId, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListPaymentMethods_BillingServer_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/methods")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()

	err := suite.handler.listPaymentMethods(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangePaymentMethod_Ok() {
	pm := &grpc.MerchantPaymentMethodRequest{
		PaymentMethod: &billing.MerchantPaymentMethodIdentification{
			Id:   bson.NewObjectId().Hex(),
			Name: "Unit test",
		},
		Commission: &billing.MerchantPaymentMethodCommissions{
			Fee: 3,
			PerTransaction: &billing.MerchantPaymentMethodPerTransactionCommission{
				Fee:      4,
				Currency: "USD",
			},
		},
		Integration: &billing.MerchantPaymentMethodIntegration{
			TerminalId:       "1234567890",
			TerminalPassword: "0987654321",
		},
		IsActive: true,
	}

	b, err := json.Marshal(pm)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/methods")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err = suite.handler.changePaymentMethod(ctx)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangePaymentMethod_BindingError() {
	pm := &grpc.MerchantPaymentMethodRequest{
		PaymentMethod: &billing.MerchantPaymentMethodIdentification{
			Id:   bson.NewObjectId().Hex(),
			Name: "Unit test",
		},
		Commission: &billing.MerchantPaymentMethodCommissions{
			Fee: 3,
			PerTransaction: &billing.MerchantPaymentMethodPerTransactionCommission{
				Fee:      4,
				Currency: "USD",
			},
		},
		Integration: &billing.MerchantPaymentMethodIntegration{
			TerminalId:       "1234567890",
			TerminalPassword: "0987654321",
		},
		IsActive: true,
	}

	b, err := json.Marshal(pm)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/methods")

	err = suite.handler.changePaymentMethod(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectMerchantId, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangePaymentMethod_ValidationError() {
	pm := &grpc.MerchantPaymentMethodRequest{
		PaymentMethod: &billing.MerchantPaymentMethodIdentification{
			Name: "Unit test",
		},
		Commission: &billing.MerchantPaymentMethodCommissions{
			Fee: 3,
			PerTransaction: &billing.MerchantPaymentMethodPerTransactionCommission{
				Fee:      4,
				Currency: "USD",
			},
		},
		Integration: &billing.MerchantPaymentMethodIntegration{
			TerminalId:       "1234567890",
			TerminalPassword: "0987654321",
		},
		IsActive: true,
	}

	b, err := json.Marshal(pm)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/methods")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err = suite.handler.changePaymentMethod(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Id", httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangePaymentMethod_BillingServer_Error() {
	pm := &grpc.MerchantPaymentMethodRequest{
		PaymentMethod: &billing.MerchantPaymentMethodIdentification{
			Id:   bson.NewObjectId().Hex(),
			Name: "Unit test",
		},
		Commission: &billing.MerchantPaymentMethodCommissions{
			Fee: 3,
			PerTransaction: &billing.MerchantPaymentMethodPerTransactionCommission{
				Fee:      4,
				Currency: "USD",
			},
		},
		Integration: &billing.MerchantPaymentMethodIntegration{
			TerminalId:       "1234567890",
			TerminalPassword: "0987654321",
		},
		IsActive: true,
	}

	b, err := json.Marshal(pm)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/methods")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.handler.changePaymentMethod(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangePaymentMethod_BillingServerErrorResponse_Error() {
	pm := &grpc.MerchantPaymentMethodRequest{
		PaymentMethod: &billing.MerchantPaymentMethodIdentification{
			Id:   bson.NewObjectId().Hex(),
			Name: "Unit test",
		},
		Commission: &billing.MerchantPaymentMethodCommissions{
			Fee: 3,
			PerTransaction: &billing.MerchantPaymentMethodPerTransactionCommission{
				Fee:      4,
				Currency: "USD",
			},
		},
		Integration: &billing.MerchantPaymentMethodIntegration{
			TerminalId:       "1234567890",
			TerminalPassword: "0987654321",
		},
		IsActive: true,
	}

	b, err := json.Marshal(pm)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/methods")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err = suite.handler.changePaymentMethod(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeMerchantStatus_BindError() {
	data := &grpc.MerchantChangeStatusRequest{
		Status:  pkg.MerchantStatusAgreementSigning,
		Message: "some message",
	}

	b, err := json.Marshal(data)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:id/change-status")

	err = suite.handler.changeMerchantStatus(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectMerchantId, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateNotification_BindError() {
	data := &grpc.NotificationRequest{
		Title:   "Title",
		Message: "Message",
	}

	b, err := json.Marshal(data)
	assert.NoError(suite.T(), err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications")

	err = suite.handler.createNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectMerchantId, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetNotification_IncorrectMerchant_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications/:notification_id")
	ctx.SetParamNames(requestParameterNotificationId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.getNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectMerchantId, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_MarkAsReadNotification_IncorrectMerchant_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications/:notification_id/mark-as-read")
	ctx.SetParamNames(requestParameterNotificationId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.markAsReadNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIncorrectMerchantId, httpErr.Message)
}
