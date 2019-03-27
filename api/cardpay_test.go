package api

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-management-api/internal/mock"
	"github.com/paysuper/paysuper-management-api/payment_system/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type CardPayTestSuite struct {
	suite.Suite
	router *CardPayWebHook
	api    *Api
}

func Test_CardPayTestSuite(t *testing.T) {
	suite.Run(t, new(CardPayTestSuite))
}

func (suite *CardPayTestSuite) SetupTest() {
	suite.api = &Api{
		Http:           echo.New(),
		validate:       validator.New(),
		billingService: mock.NewBillingServerOkMock(),
	}

	suite.api.webhookRouteGroup = suite.api.Http.Group(apiWebHookGroupPath)
	suite.router = &CardPayWebHook{Api: suite.api}
}

func (suite *CardPayTestSuite) TearDownTest() {}

func (suite *CardPayTestSuite) TestCardPay_RefundCallback_Ok() {
	refundReq := &billing.CardPayRefundCallback{
		MerchantOrder: &billing.CardPayMerchantOrder{
			Id: bson.NewObjectId().Hex(),
		},
		PaymentMethod: "BANKCARD",
		PaymentData: &billing.CardPayRefundCallbackPaymentData{
			Id:              bson.NewObjectId().Hex(),
			RemainingAmount: 0,
		},
		RefundData: &billing.CardPayRefundCallbackRefundData{
			Amount:   100,
			Created:  time.Now().Format("2006-01-02T15:04:05Z"),
			Id:       bson.NewObjectId().Hex(),
			Currency: "RUB",
			Status:   pkg.CardPayPaymentResponseStatusCompleted,
			AuthCode: bson.NewObjectId().Hex(),
			Is_3D:    true,
			Rrn:      bson.NewObjectId().Hex(),
		},
		CallbackTime: time.Now().Format("2006-01-02T15:04:05Z"),
		Customer: &billing.CardPayCustomer{
			Email: "test@unut.test",
			Id:    "test@unut.test",
		},
	}

	b, err := json.Marshal(refundReq)
	assert.NoError(suite.T(), err)

	hash := sha512.New()
	hash.Write([]byte(string(b) + "secret_key"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(entity.CardPayPaymentResponseHeaderSignature, hex.EncodeToString(hash.Sum(nil)))

	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.router.refundCallback(ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusOK, rsp.Code)
	assert.Empty(suite.T(), rsp.Body.String())
}

func (suite *CardPayTestSuite) TestCardPay_RefundCallback_BindError() {
	refundReq := `{"payment_method": 11111}`

	hash := sha512.New()
	hash.Write([]byte(refundReq + "secret_key"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(refundReq))

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(entity.CardPayPaymentResponseHeaderSignature, hex.EncodeToString(hash.Sum(nil)))

	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err := suite.router.refundCallback(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Equal(suite.T(), errorQueryParamsIncorrect, httpErr.Message)
}

func (suite *CardPayTestSuite) TestCardPay_RefundCallback_ValidationError() {
	refundReq := &billing.CardPayRefundCallback{
		MerchantOrder: &billing.CardPayMerchantOrder{
			Id: bson.NewObjectId().Hex(),
		},
		PaymentMethod: "BANKCARD",
		RefundData: &billing.CardPayRefundCallbackRefundData{
			Amount:   100,
			Created:  time.Now().Format("2006-01-02T15:04:05Z"),
			Id:       bson.NewObjectId().Hex(),
			Currency: "RUB",
			Status:   pkg.CardPayPaymentResponseStatusCompleted,
			AuthCode: bson.NewObjectId().Hex(),
			Is_3D:    true,
			Rrn:      bson.NewObjectId().Hex(),
		},
		CallbackTime: time.Now().Format("2006-01-02T15:04:05Z"),
		Customer: &billing.CardPayCustomer{
			Email: "test@unut.test",
			Id:    "test@unut.test",
		},
	}

	b, err := json.Marshal(refundReq)
	assert.NoError(suite.T(), err)

	hash := sha512.New()
	hash.Write([]byte(string(b) + "secret_key"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(entity.CardPayPaymentResponseHeaderSignature, hex.EncodeToString(hash.Sum(nil)))

	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	err = suite.router.refundCallback(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusBadRequest, httpErr.Code)
	assert.Regexp(suite.T(), "PaymentData", httpErr.Message)
}

func (suite *CardPayTestSuite) TestCardPay_RefundCallback_BillingServerSystemError() {
	refundReq := &billing.CardPayRefundCallback{
		MerchantOrder: &billing.CardPayMerchantOrder{
			Id: bson.NewObjectId().Hex(),
		},
		PaymentMethod: "BANKCARD",
		PaymentData: &billing.CardPayRefundCallbackPaymentData{
			Id:              bson.NewObjectId().Hex(),
			RemainingAmount: 0,
		},
		RefundData: &billing.CardPayRefundCallbackRefundData{
			Amount:   100,
			Created:  time.Now().Format("2006-01-02T15:04:05Z"),
			Id:       bson.NewObjectId().Hex(),
			Currency: "RUB",
			Status:   pkg.CardPayPaymentResponseStatusCompleted,
			AuthCode: bson.NewObjectId().Hex(),
			Is_3D:    true,
			Rrn:      bson.NewObjectId().Hex(),
		},
		CallbackTime: time.Now().Format("2006-01-02T15:04:05Z"),
		Customer: &billing.CardPayCustomer{
			Email: "test@unut.test",
			Id:    "test@unut.test",
		},
	}

	b, err := json.Marshal(refundReq)
	assert.NoError(suite.T(), err)

	hash := sha512.New()
	hash.Write([]byte(string(b) + "secret_key"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(entity.CardPayPaymentResponseHeaderSignature, hex.EncodeToString(hash.Sum(nil)))

	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerSystemErrorMock()
	err = suite.router.refundCallback(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusInternalServerError, httpErr.Code)
	assert.Equal(suite.T(), errorUnknown, httpErr.Message)
}

func (suite *CardPayTestSuite) TestCardPay_RefundCallback_BillingServer_Error() {
	refundReq := &billing.CardPayRefundCallback{
		MerchantOrder: &billing.CardPayMerchantOrder{
			Id: bson.NewObjectId().Hex(),
		},
		PaymentMethod: "BANKCARD",
		PaymentData: &billing.CardPayRefundCallbackPaymentData{
			Id:              bson.NewObjectId().Hex(),
			RemainingAmount: 0,
		},
		RefundData: &billing.CardPayRefundCallbackRefundData{
			Amount:   100,
			Created:  time.Now().Format("2006-01-02T15:04:05Z"),
			Id:       bson.NewObjectId().Hex(),
			Currency: "RUB",
			Status:   pkg.CardPayPaymentResponseStatusCompleted,
			AuthCode: bson.NewObjectId().Hex(),
			Is_3D:    true,
			Rrn:      bson.NewObjectId().Hex(),
		},
		CallbackTime: time.Now().Format("2006-01-02T15:04:05Z"),
		Customer: &billing.CardPayCustomer{
			Email: "test@unut.test",
			Id:    "test@unut.test",
		},
	}

	b, err := json.Marshal(refundReq)
	assert.NoError(suite.T(), err)

	hash := sha512.New()
	hash.Write([]byte(string(b) + "secret_key"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(entity.CardPayPaymentResponseHeaderSignature, hex.EncodeToString(hash.Sum(nil)))

	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerErrorMock()
	err = suite.router.refundCallback(ctx)
	assert.Error(suite.T(), err)

	httpErr, ok := err.(*echo.HTTPError)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), http.StatusNotFound, httpErr.Code)
	assert.Equal(suite.T(), mock.SomeError, httpErr.Message)
}

func (suite *CardPayTestSuite) TestCardPay_RefundCallback_BillingServerTemporary_Ok() {
	refundReq := &billing.CardPayRefundCallback{
		MerchantOrder: &billing.CardPayMerchantOrder{
			Id: bson.NewObjectId().Hex(),
		},
		PaymentMethod: "BANKCARD",
		PaymentData: &billing.CardPayRefundCallbackPaymentData{
			Id:              bson.NewObjectId().Hex(),
			RemainingAmount: 0,
		},
		RefundData: &billing.CardPayRefundCallbackRefundData{
			Amount:   100,
			Created:  time.Now().Format("2006-01-02T15:04:05Z"),
			Id:       bson.NewObjectId().Hex(),
			Currency: "RUB",
			Status:   pkg.CardPayPaymentResponseStatusCompleted,
			AuthCode: bson.NewObjectId().Hex(),
			Is_3D:    true,
			Rrn:      bson.NewObjectId().Hex(),
		},
		CallbackTime: time.Now().Format("2006-01-02T15:04:05Z"),
		Customer: &billing.CardPayCustomer{
			Email: "test@unut.test",
			Id:    "test@unut.test",
		},
	}

	b, err := json.Marshal(refundReq)
	assert.NoError(suite.T(), err)

	hash := sha512.New()
	hash.Write([]byte(string(b) + "secret_key"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(entity.CardPayPaymentResponseHeaderSignature, hex.EncodeToString(hash.Sum(nil)))

	rsp := httptest.NewRecorder()
	ctx := e.NewContext(req, rsp)

	suite.router.billingService = mock.NewBillingServerOkTemporaryMock()
	err = suite.router.refundCallback(ctx)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), rsp.Body.String())

	var body map[string]string
	err = json.Unmarshal(rsp.Body.Bytes(), &body)
	assert.NoError(suite.T(), err)

	v, ok := body["message"]
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), mock.SomeError, v)
}
