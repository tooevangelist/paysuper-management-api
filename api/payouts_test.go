package api

import (
	"github.com/globalsign/mgo/bson"
	"github.com/golang/protobuf/ptypes"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/stretchr/testify/assert"
	mock2 "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var payoutMock = &billing.PayoutDocument{
	Id:                 bson.NewObjectId().Hex(),
	SourceId:           []string{bson.NewObjectId().Hex(), bson.NewObjectId().Hex()},
	Transaction:        "",
	Amount:             100500,
	Currency:           "EUR",
	Status:             "pending",
	Description:        "royalty for june-july 2019",
	Destination:        &billing.MerchantBanking{},
	CreatedAt:          ptypes.TimestampNow(),
	UpdatedAt:          ptypes.TimestampNow(),
	ArrivalDate:        ptypes.TimestampNow(),
	FailureCode:        "",
	FailureMessage:     "",
	FailureTransaction: "",
	MerchantId:         bson.NewObjectId().Hex(),
	SignatureData: &billing.PayoutDocumentSignatureData{
		DetailsUrl:          "http://localhost",
		FilesUrl:            "http://localhost",
		SignatureRequestId:  bson.NewObjectId().Hex(),
		MerchantSignatureId: bson.NewObjectId().Hex(),
		PsSignatureId:       bson.NewObjectId().Hex(),
	},
	HasMerchantSignature:  false,
	HasPspSignature:       false,
	SignedDocumentFileUrl: "",
}

type PayoutDocumentsTestSuite struct {
	suite.Suite
	router *payoutDocumentsRoute
	api    *Api
}

func Test_PayoutDocuments(t *testing.T) {
	suite.Run(t, new(PayoutDocumentsTestSuite))
}

func (suite *PayoutDocumentsTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
		authUser: &AuthUser{
			Id: "ffffffffffffffffffffffff",
		},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	suite.router = &payoutDocumentsRoute{Api: suite.api}

	billingService := &mock.BillingService{}

	billingService.On("GetPayoutDocuments", mock2.Anything, mock2.Anything).
		Return(&grpc.GetPayoutDocumentsResponse{
			Status: http.StatusOK,
			Data: &grpc.PayoutDocumentsPaginate{
				Count: 1,
				Items: []*billing.PayoutDocument{payoutMock},
			},
		}, nil)

	billingService.On("CreatePayoutDocument", mock2.Anything, mock2.Anything).
		Return(&grpc.PayoutDocumentResponse{
			Status: http.StatusOK,
			Item:   payoutMock,
		}, nil)

	billingService.On("UpdatePayoutDocument", mock2.Anything, mock2.Anything).
		Return(&grpc.PayoutDocumentResponse{
			Status: http.StatusOK,
			Item:   payoutMock,
		}, nil)

	suite.api.billingService = billingService
}

func (suite *PayoutDocumentsTestSuite) TearDownTest() {}

func (suite *PayoutDocumentsTestSuite) TestPayoutDocuments_getPayoutDocumentsList() {
	e := echo.New()
	q := make(url.Values)
	q.Set("status", "pending")
	q.Set("merchant_id", "5bdc39a95d1e1100019fb7df")
	q.Set("limit", "10")
	q.Set("offset", "10")
	req := httptest.NewRequest(http.MethodGet, "/payout_documents?"+q.Encode(), nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payout_documents?" + q.Encode())

	err := suite.router.getPayoutDocumentsList(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PayoutDocumentsTestSuite) TestPayoutDocuments_getPayoutDocument() {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/payout_documents/5ced34d689fce60bf4440829", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payout_documents")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("5ced34d689fce60bf4440829")

	err := suite.router.getPayoutDocument(ctx)

	if assert.NoError(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusOK, rsp.Code)
		assert.NotEmpty(suite.T(), rsp.Body.String())
	}
}

func (suite *PayoutDocumentsTestSuite) TestPayoutDocuments_createPayoutDocument() {
	bodyJson := `{"description": "royalty for june-july 2019", "merchant_id": "5bdc39a95d1e1100019fb7df"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payout_documents", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payout_documents")

	err := suite.router.createPayoutDocument(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}

func (suite *PayoutDocumentsTestSuite) TestPayoutDocuments_updatePayoutDocument() {
	bodyJson := `{"status": "failed", "failure_code": "account_closed"}`

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/payout_documents/5ced34d689fce60bf4440829", strings.NewReader(bodyJson))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	ctx.SetPath("/payout_documents")
	ctx.SetParamNames(requestParameterId)
	ctx.SetParamValues("5ced34d689fce60bf4440829")

	err := suite.router.updatePayoutDocument(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.NotEmpty(suite.T(), rsp.Body.String())
}
