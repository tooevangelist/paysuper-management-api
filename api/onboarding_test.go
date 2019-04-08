package api

import (
	"bytes"
	"encoding/json"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/config"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
	{"/admin/api/v1/merchants/:merchant_id/methods/:method_id", http.MethodPut},
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
	s3Cfg := config.S3{}
	err := envconfig.Process("", &s3Cfg)
	assert.NoError(suite.T(), err)

	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id:    "ffffffffffffffffffffffff",
			Email: "test@unit.test",
		},
		config: &config.Config{
			S3: s3Cfg,
		},
	}

	renderer := &Template{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("../web/template/*.html")),
	}
	suite.api.Http.Renderer = renderer

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	err = suite.api.validate.RegisterValidation("phone", suite.api.PhoneValidator)
	assert.NoError(suite.T(), err)

	mClt, err := minio.New(
		suite.api.config.S3.Endpoint,
		suite.api.config.S3.AccessKeyId,
		suite.api.config.S3.SecretKey,
		suite.api.config.S3.Secure,
	)
	assert.NoError(suite.T(), err)

	err = mClt.MakeBucket(suite.api.config.S3.BucketName, suite.api.config.S3.Region)
	assert.NoError(suite.T(), err)

	suite.handler = &onboardingRoute{
		Api:  suite.api,
		mClt: mClt,
	}
}

func (suite *OnboardingTestSuite) TearDownTest() {}

func (suite *OnboardingTestSuite) TestOnboarding_InitRoutes_Ok() {
	api, err := suite.api.initOnboardingRoutes()
	assert.NoError(suite.T(), err)
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

	var m *grpc.MerchantListingResponse
	err = json.Unmarshal(rsp.Body.Bytes(), &m)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int32(3), m.Count)
	assert.Equal(suite.T(), mock.OnboardingMerchantMock.Id, m.Items[0].Id)
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

	req = httptest.NewRequest(http.MethodPut, "/", strings.NewReader(`{"status": 33}`))
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
	q.Set(requestParameterUserId, bson.NewObjectId().Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.listNotifications(ctx)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_ListNotifications_ValidationError() {
	e := echo.New()

	q := make(url.Values)
	q.Set(requestParameterUserId, "invalid_object_id")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

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
	q.Set(requestParameterUserId, bson.NewObjectId().Hex())

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/merchants/:merchant_id/notifications")
	ctx.SetParamNames(requestParameterMerchantId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

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

	ctx.SetPath("/merchants/:merchant_id/methods/:method_id")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterPaymentMethodId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), pm.PaymentMethod.Id)

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

	ctx.SetPath("/merchants/:merchant_id/methods/:method_id")

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
			Id:   bson.NewObjectId().Hex(),
			Name: "Unit test",
		},
		Commission: &billing.MerchantPaymentMethodCommissions{
			Fee: -1,
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

	ctx.SetPath("/merchants/:merchant_id/methods/:method_id")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterPaymentMethodId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), pm.PaymentMethod.Id)

	err = suite.handler.changePaymentMethod(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "Fee", httpErr.Message)
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

	ctx.SetPath("/merchants/:merchant_id/methods/:method_id")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterPaymentMethodId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), pm.PaymentMethod.Id)

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

	ctx.SetPath("/merchants/:merchant_id/methods/:method_id")
	ctx.SetParamNames(requestParameterMerchantId, requestParameterPaymentMethodId)
	ctx.SetParamValues(bson.NewObjectId().Hex(), pm.PaymentMethod.Id)

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

func (suite *OnboardingTestSuite) TestOnboarding_ChangeAgreement_Ok() {
	body := `{"has_merchant_signature": true, "agreement_sent_via_mail": true}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement-sign")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.changeAgreement(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeAgreement_BindError() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.changeAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeAgreement_ValidationError() {
	body := `{"has_merchant_signature": true, "agreement_type": 3}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.changeAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "AgreementType", httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeAgreement_BillingServerSystemError() {
	body := `{"has_merchant_signature": true, "agreement_sent_via_mail": true}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.SomeMerchantId)

	err := suite.handler.changeAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeAgreement_BillingServerReturnError() {
	body := `{"has_merchant_signature": true, "agreement_sent_via_mail": true}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement-sign")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err := suite.handler.changeAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GenerateAgreement_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.generateAgreement(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	data := &OnboardingFileData{}
	err = json.Unmarshal(rsp.Body.Bytes(), data)
	assert.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), data.Url)
	assert.NotNil(suite.T(), data.Metadata)
	assert.NotEmpty(suite.T(), data.Metadata.Name)
	assert.NotEmpty(suite.T(), data.Metadata.Extension)
	assert.NotEmpty(suite.T(), data.Metadata.ContentType)
	assert.True(suite.T(), data.Metadata.Size > 0)
}

func (suite *OnboardingTestSuite) TestOnboarding_GenerateAgreement_MerchantIdInvalid_Error() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.handler.generateAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GenerateAgreement_BillingServerSystemError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.handler.generateAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GenerateAgreement_BillingServerResultError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err := suite.handler.generateAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GenerateAgreement_SetMerchantS3AgreementRequest_Error() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.SomeMerchantId)

	err := suite.handler.generateAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GenerateAgreement_AgreementExist_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.OnboardingMerchantMock.Id)

	buf := new(bytes.Buffer)
	data := map[string]interface{}{"Merchant": mock.OnboardingMerchantMock}
	err := ctx.Echo().Renderer.Render(buf, agreementPageTemplateName, data, ctx)
	assert.NoError(suite.T(), err)

	pdf, err := wkhtmltopdf.NewPDFGenerator()
	assert.NoError(suite.T(), err)

	pdf.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(buf.String())))
	err = pdf.Create()
	assert.NoError(suite.T(), err)

	filePath := os.TempDir() + string(os.PathSeparator) + mock.SomeAgreementName
	err = pdf.WriteFile(filePath)
	assert.NoError(suite.T(), err)

	_, err = suite.handler.mClt.FPutObject(suite.api.config.S3.BucketName, mock.SomeAgreementName, filePath, minio.PutObjectOptions{ContentType: agreementContentType})
	assert.NoError(suite.T(), err)

	err = suite.handler.generateAgreement(ctx)
	assert.NoError(suite.T(), err)

	fData := &OnboardingFileData{}
	err = json.Unmarshal(rsp.Body.Bytes(), fData)
	assert.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), fData.Url)
	assert.NotNil(suite.T(), fData.Metadata)
	assert.NotEmpty(suite.T(), fData.Metadata.Name)
	assert.NotEmpty(suite.T(), fData.Metadata.Extension)
	assert.NotEmpty(suite.T(), fData.Metadata.ContentType)
	assert.True(suite.T(), fData.Metadata.Size > 0)

	err = suite.handler.mClt.RemoveObject(suite.api.config.S3.BucketName, mock.SomeAgreementName)
	assert.NoError(suite.T(), err)
}

func (suite *OnboardingTestSuite) TestOnboarding_GenerateAgreement_AgreementExist_Error() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.SomeMerchantId2)

	err := suite.handler.generateAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), "The specified key does not exist.", httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_Ok() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.SomeMerchantId1)

	buf := new(bytes.Buffer)
	data := map[string]interface{}{"Merchant": mock.OnboardingMerchantMock}
	err := ctx.Echo().Renderer.Render(buf, agreementPageTemplateName, data, ctx)
	assert.NoError(suite.T(), err)

	pdf, err := wkhtmltopdf.NewPDFGenerator()
	assert.NoError(suite.T(), err)

	pdf.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(buf.String())))
	err = pdf.Create()
	assert.NoError(suite.T(), err)

	filePath := os.TempDir() + string(os.PathSeparator) + mock.SomeAgreementName1
	err = pdf.WriteFile(filePath)
	assert.NoError(suite.T(), err)

	_, err = suite.handler.mClt.FPutObject(suite.api.config.S3.BucketName, mock.SomeAgreementName1, filePath, minio.PutObjectOptions{ContentType: agreementContentType})
	assert.NoError(suite.T(), err)

	err = suite.handler.getAgreementDocument(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
	assert.Equal(suite.T(), agreementContentType, rsp.Header().Get(echo.HeaderContentType))

	err = suite.handler.mClt.RemoveObject(suite.api.config.S3.BucketName, mock.SomeAgreementName1)
	assert.NoError(suite.T(), err)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_MerchantIdIncorrect_Error() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")

	err := suite.handler.getAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_BillingServerSystemError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.handler.getAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_BillingServerReturnError() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err := suite.handler.getAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_AgreementNotGenerated_Error() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.getAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorMessageAgreementNotGenerated, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_AgreementFileNotExist_Error() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.SomeMerchantId2)

	err := suite.handler.getAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), "The specified key does not exist.", httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_UploadAgreementDocument_Ok() {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	buf := new(bytes.Buffer)
	data := map[string]interface{}{"Merchant": mock.OnboardingMerchantMock}
	err := ctx.Echo().Renderer.Render(buf, agreementPageTemplateName, data, ctx)
	assert.NoError(suite.T(), err)

	pdf, err := wkhtmltopdf.NewPDFGenerator()
	assert.NoError(suite.T(), err)

	pdf.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(buf.String())))
	err = pdf.Create()
	assert.NoError(suite.T(), err)

	filePath := os.TempDir() + string(os.PathSeparator) + mock.SomeAgreementName1
	err = pdf.WriteFile(filePath)
	assert.NoError(suite.T(), err)

	params := map[string]string{}
	req1, err := suite.newFileUploadRequest("/", params, requestParameterFile, filePath)
	assert.NoError(suite.T(), err)

	rsp1 := httptest.NewRecorder()
	ctx1 := suite.api.Http.NewContext(req1, rsp1)

	ctx1.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx1.SetParamNames(requestParameterId)
	ctx1.SetParamValues(bson.NewObjectId().Hex())

	err = suite.handler.uploadAgreementDocument(ctx1)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), rsp1.Body.String())

	fData := &OnboardingFileData{}
	err = json.Unmarshal(rsp1.Body.Bytes(), fData)
	assert.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), fData.Url)
	assert.NotNil(suite.T(), fData.Metadata)
	assert.NotEmpty(suite.T(), fData.Metadata.Name)
	assert.NotEmpty(suite.T(), fData.Metadata.Extension)
	assert.NotEmpty(suite.T(), fData.Metadata.ContentType)
	assert.True(suite.T(), fData.Metadata.Size > 0)
}

func (suite *OnboardingTestSuite) TestOnboarding_UploadAgreementDocument_MerchantIdInvalid_Error() {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	err := suite.handler.uploadAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_UploadAgreementDocument_BillingServerSystemError() {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerSystemErrorMock()
	err := suite.handler.uploadAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_UploadAgreementDocument_BillingServerResultError() {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	suite.handler.billingService = mock.NewBillingServerErrorMock()
	err := suite.handler.uploadAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_UploadAgreementDocument_NotMultipartForm_Error() {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err := suite.handler.uploadAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), "no multipart boundary param in Content-Type", httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_UploadAgreementDocument_UploadFileValidationError() {
	img := image.NewRGBA(image.Rect(0, 0, 100, 50))
	img.Set(2, 3, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	fPath := os.TempDir() + string(os.PathSeparator) + "out.png"
	f, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE, 0600)
	assert.NoError(suite.T(), err)

	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()

	err = png.Encode(f, img)
	assert.NoError(suite.T(), err)

	params := map[string]string{}
	req, err := suite.newFileUploadRequest("/", params, requestParameterFile, fPath)
	assert.NoError(suite.T(), err)

	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(bson.NewObjectId().Hex())

	err = suite.handler.uploadAgreementDocument(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorMessageAgreementContentType, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_UploadAgreementDocument_SetMerchantS3AgreementRequest_Error() {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	buf := new(bytes.Buffer)
	data := map[string]interface{}{"Merchant": mock.OnboardingMerchantMock}
	err := ctx.Echo().Renderer.Render(buf, agreementPageTemplateName, data, ctx)
	assert.NoError(suite.T(), err)

	pdf, err := wkhtmltopdf.NewPDFGenerator()
	assert.NoError(suite.T(), err)

	pdf.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(buf.String())))
	err = pdf.Create()
	assert.NoError(suite.T(), err)

	filePath := os.TempDir() + string(os.PathSeparator) + mock.SomeAgreementName1
	err = pdf.WriteFile(filePath)
	assert.NoError(suite.T(), err)

	params := map[string]string{}
	req1, err := suite.newFileUploadRequest("/", params, requestParameterFile, filePath)
	assert.NoError(suite.T(), err)

	rsp1 := httptest.NewRecorder()
	ctx1 := suite.api.Http.NewContext(req1, rsp1)

	ctx1.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx1.SetParamNames(requestParameterId)
	ctx1.SetParamValues(mock.SomeMerchantId)

	err = suite.handler.uploadAgreementDocument(ctx1)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_UploadAgreementDocument_S3UploadError() {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEMultipartForm)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	buf := new(bytes.Buffer)
	data := map[string]interface{}{"Merchant": mock.OnboardingMerchantMock}
	err := ctx.Echo().Renderer.Render(buf, agreementPageTemplateName, data, ctx)
	assert.NoError(suite.T(), err)

	pdf, err := wkhtmltopdf.NewPDFGenerator()
	assert.NoError(suite.T(), err)

	pdf.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(buf.String())))
	err = pdf.Create()
	assert.NoError(suite.T(), err)

	filePath := os.TempDir() + string(os.PathSeparator) + mock.SomeAgreementName1
	err = pdf.WriteFile(filePath)
	assert.NoError(suite.T(), err)

	params := map[string]string{}
	req1, err := suite.newFileUploadRequest("/", params, requestParameterFile, filePath)
	assert.NoError(suite.T(), err)

	rsp1 := httptest.NewRecorder()
	ctx1 := suite.api.Http.NewContext(req1, rsp1)

	ctx1.SetPath("/admin/api/v1/merchants/:id/agreement/document")
	ctx1.SetParamNames(requestParameterId)
	ctx1.SetParamValues(bson.NewObjectId().Hex())

	suite.api.config.S3.BucketName = "fake_bucket"
	err = suite.handler.uploadAgreementDocument(ctx1)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), "unexpected EOF", httpErr.Message)
}

func (suite *OnboardingTestSuite) newFileUploadRequest(
	uri string,
	params map[string]string,
	paramName, path string,
) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))

	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()

	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(http.MethodPost, uri, body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())

	return req, err
}

func (suite *OnboardingTestSuite) TestOnboarding_GenerateAgreement_MerchantDataNotChecked_Error() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := suite.api.Http.NewContext(req, rsp)

	ctx.SetPath("/admin/api/v1/merchants/:id/agreement")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues(mock.SomeMerchantId3)

	err := suite.handler.generateAgreement(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorMessageAgreementCanNotBeGenerate, httpErr.Message)
}
