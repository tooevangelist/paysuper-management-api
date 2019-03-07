package api

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo"
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

type OnboardingTestSuite struct {
	suite.Suite
	handler *onboardingRoute
}

func Test_Onboarding(t *testing.T) {
	suite.Run(t, new(OnboardingTestSuite))
}

func (suite *OnboardingTestSuite) SetupTest() {
	api := &Api{
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	err := api.validate.RegisterValidation("phone", api.PhoneValidator)
	assert.NoError(suite.T(), err)

	suite.handler = &onboardingRoute{Api: api}
}

func (suite *OnboardingTestSuite) TearDownTest() {}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchant_Ok() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchant/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	b, err := json.Marshal(mock.OnboardingMerchantMock)
	assert.NoError(suite.T(), err)

	err = suite.handler.getMerchant(ctx)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.Equal(suite.T(), string(b), rsp.Body.String())
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
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Name", httpErr.Message)
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

	changeStatusReq := &grpc.MerchantChangeStatusRequest{
		MerchantId: "",
		Status:     pkg.MerchantStatusAgreementRequested,
	}
	b, err = json.Marshal(changeStatusReq)
	assert.NoError(suite.T(), err)

	req = httptest.NewRequest(http.MethodPut, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp = httptest.NewRecorder()
	ctx = e.NewContext(req, rsp)

	err = suite.handler.changeMerchantStatus(ctx)

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "MerchantId", httpErr.Message)
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

	ctx.SetPath("/notification/:id")
	ctx.SetParamNames("id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

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

	err := suite.handler.getNotification(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorIdIsEmpty, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetNotification_BillingServerUnavailable_Error() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/notification/:id")
	ctx.SetParamNames("id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

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

	ctx.SetPath("/notification/:id//notification/:id/mark-as-read")
	ctx.SetParamNames("id")
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.markAsReadNotification(ctx)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}
