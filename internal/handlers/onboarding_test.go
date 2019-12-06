package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	awsWrapperMocks "github.com/paysuper/paysuper-aws-manager/pkg/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg"
	billMock "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	billingMocks "github.com/paysuper/paysuper-billing-server/pkg/mocks"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/paysuper/paysuper-management-api/internal/test"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"
)

type OnboardingTestSuite struct {
	suite.Suite
	router  *OnboardingRoute
	caller  *test.EchoReqResCaller
	somePDF []byte
}

func Test_Onboarding(t *testing.T) {
	suite.Run(t, new(OnboardingTestSuite))
}

func (suite *OnboardingTestSuite) SetupTest() {
	user := &common.AuthUser{
		Id:         "ffffffffffffffffffffffff",
		Email:      "test@unit.test",
		MerchantId: "ffffffffffffffffffffffff",
	}

	var e error
	settings := test.DefaultSettings()
	srv := common.Services{
		Billing: mock.NewBillingServerOkMock(),
	}
	suite.caller, e = test.SetUp(settings, srv, func(set *test.TestSet, mw test.Middleware) common.Handlers {

		mw.Pre(test.PreAuthUserMiddleware(user))

		suite.somePDF, e = ioutil.ReadFile(set.Initial.WorkDir + "/test/test_pdf.pdf")
		if e != nil {
			panic(e)
		}

		downloadMockResultFn := func(
			ctx context.Context,
			filePath string,
			in *awsWrapper.DownloadInput,
			opts ...func(*s3manager.Downloader),
		) int64 {
			_, err := os.Stat(filePath)

			if err == nil {
				return 0
			}

			if !os.IsNotExist(err) {
				return 0
			}

			src, err := os.Open(set.Initial.WorkDir + "/test/test_pdf.pdf")
			if err != nil {
				return 0
			}
			defer src.Close()

			dst, err := os.Create(filePath)
			if err != nil {
				return 0
			}
			defer dst.Close()

			nBytes, err := io.Copy(dst, src)

			if err != nil {
				return 0
			}

			return nBytes
		}

		awsManagerMock := &awsWrapperMocks.AwsManagerInterface{}
		awsManagerMock.On("Upload", mock2.Anything, mock2.Anything, mock2.Anything).Return(&s3manager.UploadOutput{}, nil)
		awsManagerMock.On("Download", mock2.Anything, mock2.Anything, mock2.Anything, mock2.Anything).
			Return(downloadMockResultFn, nil)

		suite.router = NewOnboardingRoute(set.HandlerSet, set.Initial, awsManagerMock, set.GlobalConfig)
		return common.Handlers{
			suite.router,
		}
	})
	if e != nil {
		panic(e)
	}
}

func (suite *OnboardingTestSuite) TearDownTest() {}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchant_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)

	obj := &billing.Merchant{}
	err = json.Unmarshal(res.Body.Bytes(), obj)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.Equal(suite.T(), mock.OnboardingMerchantMock.Id, obj.Id)
	assert.Equal(suite.T(), mock.OnboardingMerchantMock.Company.City, obj.Company.City)
	assert.Equal(suite.T(), mock.OnboardingMerchantMock.Company.City, obj.Company.City)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchant_BillingServiceUnavailable_Error() {

	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchant_LogicError() {

	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantBy", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantResponse{Status: pkg.ResponseStatusBadData, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchant_EmptyId_Error() {

	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantBy", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantResponse{Status: pkg.ResponseStatusBadData, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "").
		Path(common.AuthUserGroupPath + merchantsIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusNotFound, httpErr.Code)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListMerchants_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)

	var m *grpc.MerchantListingResponse
	err = json.Unmarshal(res.Body.Bytes(), &m)
	assert.NoError(suite.T(), err)
	assert.EqualValues(suite.T(), 3, m.Count)
	assert.Equal(suite.T(), mock.OnboardingMerchantMock.Id, m.Items[0].Id)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListMerchants_BindingError() {

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParam(common.RequestParameterIsSigned, bson.NewObjectId().Hex()).
		Path(common.SystemUserGroupPath + merchantsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListMerchants_ValidationError() {

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParam(common.RequestParameterOffset, "-10").
		Path(common.SystemUserGroupPath + merchantsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.NewValidationError("Offset"), httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListMerchants_BillingServiceUnavailable_Error() {

	billingService := &billMock.BillingService{}
	billingService.On("ListMerchants", mock2.Anything, mock2.Anything).Return(nil, errors.New("error"))
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateNotification_Ok() {
	n := &grpc.NotificationRequest{
		MerchantId: bson.NewObjectId().Hex(),
		UserId:     "ffffffffffffffffffffffff",
		Title:      "Title",
		Message:    "Message",
	}

	b, err := json.Marshal(n)
	assert.NoError(suite.T(), err)

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdNotificationsPath).
		Init(test.ReqInitJSON()).
		BodyBytes(b).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusCreated, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateNotification_ValidationError() {
	n := &grpc.NotificationRequest{
		MerchantId: bson.NewObjectId().Hex(),
		UserId:     "ffffffffffffffffffffffff",
		Title:      "",
		Message:    "Message",
	}

	b, err := json.Marshal(n)
	assert.NoError(suite.T(), err)

	_, err = suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdNotificationsPath).
		Init(test.ReqInitJSON()).
		BodyBytes(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.NewValidationError("Title"), httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateNotification_BillingServerUnavailable_Error() {
	n := &grpc.NotificationRequest{
		MerchantId: bson.NewObjectId().Hex(),
		UserId:     "ffffffffffffffffffffffff",
		Title:      "Title",
		Message:    "Message",
	}

	b, err := json.Marshal(n)
	assert.NoError(suite.T(), err)

	billingService := &billMock.BillingService{}
	billingService.On("CreateNotification", mock2.Anything, mock2.Anything).
		Return(&grpc.CreateNotificationResponse{Status: pkg.ResponseStatusBadData, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err = suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdNotificationsPath).
		Init(test.ReqInitJSON()).
		BodyBytes(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetNotification_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterNotificationId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + merchantsNotificationsIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_GetNotification_EmptyId_Error() {

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthUserGroupPath + merchantsNotificationsIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorValidationFailed, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetNotification_BillingServerUnavailable_Error() {

	billingService := &billMock.BillingService{}
	billingService.On("GetNotification", mock2.Anything, mock2.Anything).
		Return(nil, mock.SomeError)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterNotificationId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + merchantsNotificationsIdPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListNotifications_Ok() {
	q := make(url.Values)
	q.Set(common.RequestParameterUserId, bson.NewObjectId().Hex())

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + merchantsNotificationsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_ListNotifications_BindError() {
	q := make(url.Values)
	q.Set(common.RequestParameterOffset, "some_invalid_value")

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + merchantsNotificationsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListNotifications_ValidationError() {
	q := make(url.Values)
	q.Set(common.RequestParameterUserId, "invalid_object_id")

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + merchantsNotificationsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), common.NewValidationError("UserId"), httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ListNotifications_BillingServerError() {
	q := make(url.Values)
	q.Set(common.RequestParameterUserId, bson.NewObjectId().Hex())

	billingService := &billMock.BillingService{}
	billingService.On("ListNotifications", mock2.Anything, mock2.Anything).Return(nil, mock.SomeError)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + merchantsNotificationsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_MarkAsReadNotification_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestParameterNotificationId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + merchantsNotificationsMarkReadPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_MarkAsReadNotification_EmptyId_Error() {

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsNotificationsMarkReadPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorValidationFailed, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_MarkAsReadNotification_BillingServer_Error() {

	billingService := &billMock.BillingService{}
	billingService.On("MarkNotificationAsRead", mock2.Anything, mock2.Anything).Return(nil, mock.SomeError)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestParameterNotificationId, bson.NewObjectId().Hex()).
		Path(common.AuthUserGroupPath + merchantsNotificationsMarkReadPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeMerchantStatus_BindError() {
	data := "<some string here wrong>"

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.SystemUserGroupPath + merchantsIdChangeStatusCompanyPath).
		BodyString(data).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_CreateNotification_BindError() {
	data := "<some string here wrong>"

	_, err := suite.caller.Builder().
		Method(http.MethodPost).
		Path(common.SystemUserGroupPath + merchantsIdNotificationsPath).
		BodyString(data).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeAgreement_Ok() {

	body := `{"has_merchant_signature": true, "agreement_sent_via_mail": true}`

	res, err := suite.caller.Builder().
		Method(http.MethodPatch).
		Path(common.AuthUserGroupPath + merchantsPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeAgreement_BindError() {
	body := `{"has_merchant_signature": "test", "agreement_sent_via_mail": 123}`

	_, err := suite.caller.Builder().
		Method(http.MethodPatch).
		Path(common.AuthUserGroupPath + merchantsPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_ChangeAgreement_BillingServerSystemError() {
	body := `{"has_merchant_signature": true, "agreement_sent_via_mail": true}`

	billingService := &billMock.BillingService{}
	billingService.On("ChangeMerchantData", mock2.Anything, mock2.Anything).Return(nil, errors.New("some error"))
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPatch).
		Path(common.AuthUserGroupPath + merchantsPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.AuthUserGroupPath + merchantsAgreementDocumentPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
	assert.Equal(suite.T(), agreementContentType, res.Header().Get(echo.HeaderContentType))
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_BillingServerSystemError() {

	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(nil, mock.SomeError)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthUserGroupPath + merchantsAgreementDocumentPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_BillingServerReturnError() {

	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantBy", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantResponse{Status: pkg.ResponseStatusBadData, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.AuthUserGroupPath + merchantsAgreementDocumentPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_AgreementNotGenerated_Error() {
	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(&grpc.GetMerchantResponse{
		Status:  pkg.ResponseStatusOk,
		Message: &grpc.ResponseErrorMessage{},
		Item: &billing.Merchant{
			Id:              "ffffffffffffffffffffffff",
			S3AgreementName: "",
		},
	}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.AuthUserGroupPath + merchantsAgreementDocumentPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusNotFound, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorMessageAgreementNotGenerated, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementDocument_AgreementFileNotExist_Error() {
	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(&grpc.GetMerchantResponse{
		Status:  pkg.ResponseStatusOk,
		Message: &grpc.ResponseErrorMessage{},
		Item: &billing.Merchant{
			Id:              "ffffffffffffffffffffffff",
			S3AgreementName: mock.SomeAgreementName,
		},
	}, nil)
	suite.router.dispatch.Services.Billing = billingService

	awsManagerMock := &awsWrapperMocks.AwsManagerInterface{}
	awsManagerMock.On("Upload", mock2.Anything, mock2.Anything, mock2.Anything).Return(&s3manager.UploadOutput{}, nil)
	awsManagerMock.On("Download", mock2.Anything, mock2.Anything, mock2.Anything, mock2.Anything).
		Return(int64(0), errors.New("some error"))
	suite.router.awsManager = awsManagerMock

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.AuthUserGroupPath + merchantsAgreementDocumentPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorAgreementFileNotExist, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_WithoutMerchantId_Ok() {
	company := &billing.MerchantCompanyInfo{
		Name:               mock.OnboardingMerchantMock.Company.Name,
		AlternativeName:    mock.OnboardingMerchantMock.Company.Name,
		Website:            "http://localhost",
		Country:            "RU",
		State:              "St.Petersburg",
		Zip:                "190000",
		City:               "St.Petersburg",
		Address:            "Nevskiy st. 1",
		RegistrationNumber: "1234567890",
	}
	b, err := json.Marshal(company)
	assert.NoError(suite.T(), err)

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyBytes(b).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())

	merchant := new(billing.Merchant)
	err = json.Unmarshal(res.Body.Bytes(), merchant)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), merchant.Company, company)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_WithMerchantId_Ok() {
	company := &billing.MerchantCompanyInfo{
		Name:               mock.OnboardingMerchantMock.Company.Name,
		AlternativeName:    mock.OnboardingMerchantMock.Company.Name,
		Website:            "http://localhost",
		Country:            "RU",
		State:              "St.Petersburg",
		Zip:                "190000",
		City:               "St.Petersburg",
		Address:            "Nevskiy st. 1",
		RegistrationNumber: "1234567890",
	}
	b, err := json.Marshal(company)
	assert.NoError(suite.T(), err)

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdCompanyPath).
		Init(test.ReqInitJSON()).
		BodyBytes(b).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())

	merchant := new(billing.Merchant)
	err = json.Unmarshal(res.Body.Bytes(), merchant)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), merchant.Company, company)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_BindError() {
	b := `{"name": 123}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_ValidationError_CompanyName() {
	b := `{"alternative_name": "123"}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectCompanyName.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectCompanyName.Message, msg.Message)
	assert.Regexp(suite.T(), "Name", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_ValidationError_CompanyAlternativeName() {
	b := `{"name": "123"}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectAlternativeName.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectAlternativeName.Message, msg.Message)
	assert.Regexp(suite.T(), "AlternativeName", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_ValidationError_CompanyWebsite() {
	b := `{"name": "123", "alternative_name": "123", "website": "123"}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectWebsite.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectWebsite.Message, msg.Message)
	assert.Regexp(suite.T(), "Website", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_ValidationError_CompanyCountry() {
	b := `{"name": "123", "alternative_name": "123", "website": "http://localhost"}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorIncorrectCountryIdentifier.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorIncorrectCountryIdentifier.Message, msg.Message)
	assert.Regexp(suite.T(), "Country", msg.Details)
}

/*
INFO After the #192483 State field is not required in Merchant on-boarding process. So this test should be
disabled.

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_ValidationError_CompanyState() {
	b := `{"name": "123", "alternative_name": "123", "website": "http://localhost", "country": "RU"}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectState.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectState.Message, msg.Message)
	assert.Regexp(suite.T(), "State", msg.Details)
}
*/

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_ValidationError_CompanyZip() {
	b := `{
        "name": "123",
        "alternative_name": "123",
        "website": "http://localhost",
        "country": "RU",
        "state": "St.Petersburg"
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectZip.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectZip.Message, msg.Message)
	assert.Regexp(suite.T(), "Zip", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_ValidationError_CompanyCity() {
	b := `{
        "name": "123",
        "alternative_name": "123",
        "website": "http://localhost",
        "country": "RU",
        "state": "St.Petersburg",
        "zip": "190000"
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectCity.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectCity.Message, msg.Message)
	assert.Regexp(suite.T(), "City", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_ValidationError_CompanyAddress() {
	b := `{
        "name": "123",
        "alternative_name": "123",
        "website": "http://localhost",
        "country": "RU",
        "state": "St.Petersburg",
        "zip": "190000",
        "city": "St.Petersburg"
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectAddress.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectAddress.Message, msg.Message)
	assert.Regexp(suite.T(), "Address", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_BillingServerSystemError() {
	b := `{
        "name": "123",
        "alternative_name": "123",
        "website": "http://localhost",
        "country": "RU",
        "state": "St.Petersburg",
        "zip": "190000",
        "city": "St.Petersburg",
        "address": "Nevskiy st. 1",
		"registration_number": "1234567890"
    }`

	billingService := &billMock.BillingService{}
	billingService.On("ChangeMerchant", mock2.Anything, mock2.Anything).Return(nil, mock.SomeError)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantCompany_BillingServerResultError() {
	b := `{
        "name": "123",
        "alternative_name": "123",
        "website": "http://localhost",
        "country": "RU",
        "state": "St.Petersburg",
        "zip": "190000",
        "city": "St.Petersburg",
        "address": "Nevskiy st. 1",
		"registration_number": "1234567890"
    }`

	billingService := &billMock.BillingService{}
	billingService.On("ChangeMerchant", mock2.Anything, mock2.Anything).
		Return(&grpc.ChangeMerchantResponse{Status: http.StatusBadRequest, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsCompanyPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_WithoutMerchantId_Ok() {
	contacts := &billing.MerchantContact{
		Authorized: &billing.MerchantContactAuthorized{
			Name:     "Unit Test",
			Email:    "test@unit.test",
			Phone:    "1234567890",
			Position: "CEO",
		},
		Technical: &billing.MerchantContactTechnical{
			Name:  "Unit Test",
			Email: "test@unit.test",
			Phone: "1234567890",
		},
	}
	b, err := json.Marshal(contacts)
	assert.NoError(suite.T(), err)

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyBytes(b).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())

	merchant := new(billing.Merchant)
	err = json.Unmarshal(res.Body.Bytes(), merchant)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), merchant.Contacts, contacts)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_WithMerchantId_Ok() {
	contacts := &billing.MerchantContact{
		Authorized: &billing.MerchantContactAuthorized{
			Name:     "Unit Test",
			Email:    "test@unit.test",
			Phone:    "1234567890",
			Position: "CEO",
		},
		Technical: &billing.MerchantContactTechnical{
			Name:  "Unit Test",
			Email: "test@unit.test",
			Phone: "1234567890",
		},
	}
	b, err := json.Marshal(contacts)
	assert.NoError(suite.T(), err)

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdContactsPath).
		Init(test.ReqInitJSON()).
		BodyBytes(b).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())

	merchant := new(billing.Merchant)
	err = json.Unmarshal(res.Body.Bytes(), merchant)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), merchant.Contacts, contacts)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_BindError() {
	b := `{"authorized": {"name": "Unit Test", "Email": "test@unit.test", "Phone": "1234567890"}, "technical": 1234}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_ValidationError_Authorized() {
	b := `{"technical": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"}}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageRequiredContactAuthorized.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageRequiredContactAuthorized.Message, msg.Message)
	assert.Regexp(suite.T(), "Authorized", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_ValidationError_Technical() {
	b := `{"authorized": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890", "position": "12345"}}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageRequiredContactTechnical.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageRequiredContactTechnical.Message, msg.Message)
	assert.Regexp(suite.T(), "Technical", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_ValidationError_AuthorizedName() {
	b := `{
        "authorized": {"email": "test@unit.test", "phone": "1234567890", "position": "12345"},
        "technical": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"}
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectName.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectName.Message, msg.Message)
	assert.Regexp(suite.T(), "Name", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_ValidationError_TechnicalName() {
	b := `{
        "authorized": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890", "position": "12345"},
        "technical": {"email": "test@unit.test", "phone": "1234567890"}
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectName.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectName.Message, msg.Message)
	assert.Regexp(suite.T(), "Name", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_ValidationError_AuthorizedEmail() {
	b := `{
        "authorized": {"name": "Unit Test", "phone": "1234567890", "position": "12345"},
        "technical": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"}
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorEmailFieldIncorrect.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorEmailFieldIncorrect.Message, msg.Message)
	assert.Regexp(suite.T(), "Email", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_ValidationError_TechnicalEmail() {
	b := `{
        "authorized": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890", "position": "12345"},
        "technical": {"name": "Unit Test", "phone": "1234567890"}
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorEmailFieldIncorrect.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorEmailFieldIncorrect.Message, msg.Message)
	assert.Regexp(suite.T(), "Email", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_ValidationError_AuthorizedPhone() {
	b := `{
        "authorized": {"name": "Unit Test", "email": "test@unit.test", "position": "12345"},
        "technical": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"}
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectPhone.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectPhone.Message, msg.Message)
	assert.Regexp(suite.T(), "Phone", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_ValidationError_TechnicalPhone() {
	b := `{
        "authorized": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890", "position": "12345"},
        "technical": {"name": "Unit Test", "email": "test@unit.test"}
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectPhone.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectPhone.Message, msg.Message)
	assert.Regexp(suite.T(), "Phone", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_ValidationError_AuthorizedPosition() {
	b := `{
        "authorized": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"},
        "technical": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"}
    }`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectPosition.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectPosition.Message, msg.Message)
	assert.Regexp(suite.T(), "Position", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_BillingServerSystemError() {
	b := `{
        "authorized": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890", "position": "unit test"},
        "technical": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"}
    }`

	billingService := &billMock.BillingService{}
	billingService.On("ChangeMerchant", mock2.Anything, mock2.Anything).Return(nil, mock.SomeError)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantContacts_BillingServerResultError() {
	b := `{
        "authorized": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890", "position": "unit test"},
        "technical": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"}
    }`

	billingService := &billMock.BillingService{}
	billingService.On("ChangeMerchant", mock2.Anything, mock2.Anything).
		Return(&grpc.ChangeMerchantResponse{Status: http.StatusBadRequest, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsContactsPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_WithoutMerchantId_Ok() {
	banking := &billing.MerchantBanking{
		Currency:             "RUB",
		Name:                 "Bank Name-Spb.",
		Address:              "St.Petersburg, Nevskiy st. 1",
		AccountNumber:        "SE1412345678901234567890",
		Swift:                "ALFARUMM",
		CorrespondentAccount: "408000000001",
	}
	b, err := json.Marshal(banking)
	assert.NoError(suite.T(), err)

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsBankingPath).
		Init(test.ReqInitJSON()).
		BodyBytes(b).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())

	merchant := new(billing.Merchant)
	err = json.Unmarshal(res.Body.Bytes(), merchant)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), merchant.Banking, banking)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_WithMerchantId_Ok() {
	banking := &billing.MerchantBanking{
		Currency:             "RUB",
		Name:                 "Bank Name-Spb.",
		Address:              "St.Petersburg, Nevskiy st. 1",
		AccountNumber:        "SE1412345678901234567890",
		Swift:                "ALFARUMM",
		CorrespondentAccount: "408000000001",
	}
	b, err := json.Marshal(banking)
	assert.NoError(suite.T(), err)

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdBankingPath).
		Init(test.ReqInitJSON()).
		BodyBytes(b).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())

	merchant := new(billing.Merchant)
	err = json.Unmarshal(res.Body.Bytes(), merchant)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), merchant.Banking, banking)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_BindError() {
	b := `{"name": 123}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsBankingPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_ValidationError_Name() {
	b := `{
		"currency": "RUB",
		"address": "St.Petersburg, Nevskiy st. 1",
		"account_number": "408000000001",
		"swift": "ALFARUMM",
		"correspondent_account": "408000000001"
	}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsBankingPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankName.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankName.Message, msg.Message)
	assert.Regexp(suite.T(), "Name", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_ValidationError_Address() {
	b := `{
		"currency": "RUB",
		"name": "Bank Name-Spb.",
		"account_number": "408000000001",
		"swift": "ALFARUMM",
		"correspondent_account": "408000000001"
	}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsBankingPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankAddress.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankAddress.Message, msg.Message)
	assert.Regexp(suite.T(), "Address", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_ValidationError_AccountNumber() {
	b := `{
		"currency": "RUB",
		"name": "Bank Name-Spb.",
		"address": "St.Petersburg, Nevskiy st. 1",
		"swift": "ALFARUMM",
		"correspondent_account": "408000000001"
	}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsBankingPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankAccountNumber.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankAccountNumber.Message, msg.Message)
	assert.Regexp(suite.T(), "AccountNumber", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_ValidationError_Swift() {
	b := `{
		"currency": "RUB",
		"name": "Bank Name-Spb.",
		"address": "St.Petersburg, Nevskiy st. 1",
		"account_number": "SE1412345678901234567890",
		"correspondent_account": "408000000001"
	}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsBankingPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankSwift.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankSwift.Message, msg.Message)
	assert.Regexp(suite.T(), "Swift", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_ValidationError_CorrespondentAccount() {
	b := `{
		"currency": "RUB",
		"name": "Bank Name-Spb.",
		"address": "St.Petersburg, Nevskiy st. 1",
		"account_number": "SE1412345678901234567890",
		"swift": "ALFARUMM",
		"correspondent_account": "408000000000000000000000000000000000000000000000000000000000000000000000000000000000001"
	}`

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsBankingPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankCorrespondentAccount.Code, msg.Code)
	assert.Equal(suite.T(), common.ErrorMessageIncorrectBankCorrespondentAccount.Message, msg.Message)
	assert.Regexp(suite.T(), "CorrespondentAccount", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_BillingServerSystemError() {
	b := `{
		"currency": "RUB",
		"name": "Bank Name-Spb.",
		"address": "St.Petersburg, Nevskiy st. 1",
		"account_number": "SE1412345678901234567890",
		"swift": "ALFARUMM",
		"correspondent_account": "408000000001"
	}`

	billingService := &billMock.BillingService{}
	billingService.On("ChangeMerchant", mock2.Anything, mock2.Anything).Return(nil, mock.SomeError)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsBankingPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetMerchantBanking_BillingServerResultError() {
	b := `{
		"currency": "RUB",
		"name": "Bank Name-Spb.",
		"address": "St.Petersburg, Nevskiy st. 1",
		"account_number": "SE1412345678901234567890",
		"swift": "ALFARUMM",
		"correspondent_account": "408000000001"
	}`

	billingService := &billMock.BillingService{}
	billingService.On("ChangeMerchant", mock2.Anything, mock2.Anything).
		Return(&grpc.ChangeMerchantResponse{Status: http.StatusBadRequest, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsBankingPath).
		Init(test.ReqInitJSON()).
		BodyString(b).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchantStatus_Ok() {

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthUserGroupPath + merchantsStatusCompanyPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchantStatus_BillingServerSystemError() {

	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantOnboardingCompleteData", mock2.Anything, mock2.Anything).Return(nil, mock.SomeError)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthUserGroupPath + merchantsStatusCompanyPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorInternal, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetMerchantStatus_BillingServerResultError() {

	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantOnboardingCompleteData", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantOnboardingCompleteDataResponse{Status: http.StatusBadRequest, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthUserGroupPath + merchantsStatusCompanyPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), mock.SomeError.Message, msg.Message)
	assert.Equal(suite.T(), mock.SomeError.Details, msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetTariffRates_Ok() {
	q := make(url.Values)
	q.Set("region", "russia_and_cis")
	q.Set("merchant_operations_type", "low-risk")
	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantTariffRates", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantTariffRatesResponse{Status: pkg.ResponseStatusOk, Items: &grpc.GetMerchantTariffRatesResponseItems{}}, nil)
	suite.router.dispatch.Services.Billing = billingService

	res, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + merchantsTariffsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_GetTariffRates_BindError() {
	q := make(url.Values)
	q.Set("region", "russia_and_cis")
	q.Set("min_amount", "qwerty")

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + merchantsTariffsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetTariffRates_ValidateError() {
	q := make(url.Values)
	q.Set("region", "777")
	q.Set("payout_currency", "USD")

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + merchantsTariffsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "Region", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetTariffRates_BillingServerError() {
	q := make(url.Values)
	q.Set("region", "russia_and_cis")
	q.Set("payout_currency", "USD")
	q.Set("merchant_operations_type", "low-risk")
	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantTariffRates", mock2.Anything, mock2.Anything).Return(nil, errors.New("some error"))
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + merchantsTariffsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetTariffRates_BillingServerResultError() {
	q := make(url.Values)
	q.Set("region", "russia_and_cis")
	q.Set("merchant_operations_type", "low-risk")
	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantTariffRates", mock2.Anything, mock2.Anything).
		Return(&grpc.GetMerchantTariffRatesResponse{Status: pkg.ResponseStatusBadData, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodGet).
		SetQueryParams(q).
		Path(common.AuthUserGroupPath + merchantsTariffsPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetTariffRates_Ok() {
	body := `{"home_region": "russia_and_cis", "merchant_operations_type": "low-risk"}`

	billingService := &billMock.BillingService{}
	billingService.On("SetMerchantTariffRates", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.CheckProjectRequestSignatureResponse{Status: pkg.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = billingService

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdTariffsPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.Empty(suite.T(), res.Body.String())
}

func (suite *OnboardingTestSuite) TestOnboarding_SetTariffRates_BindError() {
	body := `{"home_region": 2222}`

	_, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdTariffsPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorRequestParamsIncorrect, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetTariffRates_ValidationError() {
	body := `{"home_region": "north_america"}`

	_, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdTariffsPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)

	msg, ok := httpErr.Message.(*grpc.ResponseErrorMessage)
	assert.True(suite.T(), ok)
	assert.Regexp(suite.T(), "HomeRegion", msg.Details)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetTariffRates_BillingServerError() {
	body := `{"home_region": "russia_and_cis", "merchant_operations_type": "low-risk"}`

	billingService := &billMock.BillingService{}
	billingService.On("SetMerchantTariffRates", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(nil, errors.New("some error"))
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdTariffsPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), common.ErrorUnknown, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetTariffRates_BillingServerResultError() {
	body := `{"home_region": "russia_and_cis", "merchant_operations_type": "low-risk"}`

	billingService := &billMock.BillingService{}
	billingService.On("SetMerchantTariffRates", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.CheckProjectRequestSignatureResponse{Status: pkg.ResponseStatusBadData, Message: mock.SomeError}, nil)
	suite.router.dispatch.Services.Billing = billingService

	_, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, "ffffffffffffffffffffffff").
		Path(common.SystemUserGroupPath + merchantsIdTariffsPath).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *OnboardingTestSuite) TestOnboarding_GetAgreementData_Ok() {
	billingService := &billMock.BillingService{}
	billingService.On("GetMerchantBy", mock2.Anything, mock2.Anything).Return(&grpc.GetMerchantResponse{
		Status:  pkg.ResponseStatusOk,
		Message: &grpc.ResponseErrorMessage{},
		Item: &billing.Merchant{
			Id:              "ffffffffffffffffffffffff",
			S3AgreementName: mock.SomeAgreementName,
		},
	}, nil)
	suite.router.dispatch.Services.Billing = billingService

	rsp, err := suite.caller.Builder().
		Method(http.MethodGet).
		Path(common.AuthUserGroupPath + merchantsAgreementPath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

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

func (suite *OnboardingTestSuite) TestOnboarding_enableMerchantManualPayout_Ok() {
	billingService := &billingMocks.BillingService{}
	billingService.
		On("ChangeMerchantManualPayouts", mock2.Anything, mock2.Anything).
		Return(&grpc.ChangeMerchantManualPayoutsResponse{Status: http.StatusOK}, nil)
	suite.router.dispatch.Services.Billing = billingService

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsIdManualPayoutEnablePath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)

	obj := &billing.Merchant{}
	err = json.Unmarshal(res.Body.Bytes(), obj)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, res.Code)
}

func (suite *OnboardingTestSuite) TestOnboarding_disableMerchantManualPayout_Ok() {
	billingService := &billingMocks.BillingService{}
	billingService.
		On("ChangeMerchantManualPayouts", mock2.Anything, mock2.Anything).
		Return(&grpc.ChangeMerchantManualPayoutsResponse{Status: http.StatusOK}, nil)
	suite.router.dispatch.Services.Billing = billingService

	res, err := suite.caller.Builder().
		Method(http.MethodPut).
		Path(common.AuthUserGroupPath + merchantsIdManualPayoutDisablePath).
		Init(test.ReqInitJSON()).
		Exec(suite.T())

	assert.NoError(suite.T(), err)

	obj := &billing.Merchant{}
	err = json.Unmarshal(res.Body.Bytes(), obj)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), http.StatusOK, res.Code)
}

func (suite *OnboardingTestSuite) TestOnboarding_SetOperatingCompany_Ok() {
	body := `{"operating_company_id": "5dbc50d486616113a1cefe16"}`

	billingService := &billMock.BillingService{}
	billingService.On("SetMerchantOperatingCompany", mock2.Anything, mock2.Anything, mock2.Anything).
		Return(&grpc.SetMerchantOperatingCompanyResponse{Status: pkg.ResponseStatusOk}, nil)
	suite.router.dispatch.Services.Billing = billingService

	userOK := &common.AuthUser{
		MerchantId: mock.SomeMerchantId1,
	}

	customInit := func(request *http.Request, middleware test.Middleware) {
		middleware.Pre(test.PreAuthUserMiddleware(userOK))
	}

	res, err := suite.caller.Builder().
		Method(http.MethodPost).
		Params(":"+common.RequestParameterMerchantId, mock.SomeMerchantId1).
		Path(common.SystemUserGroupPath + merchantsIdSetOperatingCompanyPath).
		Init(customInit).
		Init(test.ReqInitJSON()).
		BodyString(body).
		Exec(suite.T())

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, res.Code)
	assert.NotEmpty(suite.T(), res.Body.String())
}
