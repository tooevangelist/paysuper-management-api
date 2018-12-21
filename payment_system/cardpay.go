package payment_system

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/payment_system/entity"
	"github.com/ProtocolONE/p1pay.api/payment_system/validator"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cardPayRequestFieldGrantType    = "grant_type"
	cardPayRequestFieldTerminalCode = "terminal_code"
	cardPayRequestFieldPassword     = "password"
	cardPayRequestFieldRefreshToken = "refresh_token"

	cardPayGrantTypePassword     = "password"
	cardPayGrantTypeRefreshToken = "refresh_token"

	cardPayActionAuthenticate  = "auth"
	cardPayActionRefresh       = "refresh"
	cardPayActionCreatePayment = "create_payment"

	cardPayDateFormat            = "2006-01-02T15:04:05Z"
	cardPayPaymentMethodBankCard = "BANKCARD"
	cardPayPaymentMethodWebMoney = "WEBMONEY"
	cardPayPaymentMethodQiwi     = "QIWI"
	cardPayPaymentMethodNeteller = "NETELLER"
	cardPayPaymentMethodAlipay   = "ALIPAY"
	cardPayPaymentMethodBitcoin  = "BITCOIN"
)

var paths = map[string]*Path{
	cardPayActionAuthenticate: {
		path:   "/api/auth/token",
		method: http.MethodPost,
	},
	cardPayActionRefresh: {
		path:   "/api/auth/token",
		method: http.MethodPost,
	},
	cardPayActionCreatePayment: {
		path:   "/api/payments",
		method: http.MethodPost,
	},
}

var tokens = map[string]*Token{}

type CardPay struct {
	*Settings
	*model.Order
	mu         sync.Mutex
	pmSettings map[string]string
}

type Token struct {
	TokenType              string `json:"token_type"`
	AccessToken            string `json:"access_token"`
	RefreshToken           string `json:"refresh_token"`
	AccessTokenExpire      int    `json:"expires_in"`
	RefreshTokenExpire     int    `json:"refresh_expires_in"`
	AccessTokenExpireTime  time.Time
	RefreshTokenExpireTime time.Time
}

func NewCardPayHandler(o *model.Order, settings *Settings) PaymentSystem {
	mSettings := make(map[string]string)
	iSettings := settings.Settings.(map[interface{}]interface{})

	for k, v := range iSettings {
		mSettings[k.(string)] = v.(string)
	}

	return &CardPay{Settings: settings, Order: o, pmSettings: mSettings}
}

func (cp *CardPay) auth(pmKey string) error {
	if token := cp.getToken(pmKey); token != nil {
		return nil
	}

	data := url.Values{
		cardPayRequestFieldGrantType:    []string{cardPayGrantTypePassword},
		cardPayRequestFieldTerminalCode: []string{cp.pmSettings[settingsFieldTerminalId]},
		cardPayRequestFieldPassword:     []string{cp.pmSettings[settingsFieldSecretWord]},
	}

	qUrl, err := cp.getUrl(cardPayActionAuthenticate)

	if err != nil {
		return err
	}

	client := cp.Settings.PaymentSystemSetting.GetLoggableHttpClient()
	req, err := http.NewRequest(paths[cardPayActionAuthenticate].method, qUrl, strings.NewReader(data.Encode()))

	if err != nil {
		return err
	}

	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Add(echo.HeaderContentLength, strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			return
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.New(paymentSystemErrorAuthenticateFailed)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := cp.setToken(b, pmKey); err != nil {
		return err
	}

	return nil
}

func (cp *CardPay) refresh(pmKey string) error {
	data := url.Values{
		cardPayRequestFieldGrantType:    []string{cardPayGrantTypeRefreshToken},
		cardPayRequestFieldTerminalCode: []string{cp.pmSettings[settingsFieldTerminalId]},
		cardPayRequestFieldRefreshToken: []string{tokens[pmKey].RefreshToken},
	}

	qUrl, err := cp.getUrl(cardPayActionRefresh)

	if err != nil {
		return err
	}

	client := cp.Settings.PaymentSystemSetting.GetLoggableHttpClient()
	req, err := http.NewRequest(paths[cardPayActionRefresh].method, qUrl, strings.NewReader(data.Encode()))

	if err != nil {
		return err
	}

	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	req.Header.Add(echo.HeaderContentLength, strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			return
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return errors.New(paymentSystemErrorAuthenticateFailed)
	}

	b, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if err := cp.setToken(b, pmKey); err != nil {
		return err
	}

	return nil
}

func (cp *CardPay) CreatePayment() *PaymentResponse {
	if err := cp.auth(cp.Order.PaymentMethod.Params.ExternalId); err != nil {
		return NewPaymentResponse(PaymentStatusErrorSystem, err.Error())
	}

	qUrl, err := cp.getUrl(cardPayActionCreatePayment)

	if err != nil {
		return NewPaymentResponse(PaymentStatusErrorSystem, err.Error())
	}

	cpo, err := cp.getCardPayOrder()

	if err != nil {
		return NewPaymentResponse(PaymentStatusErrorValidation, err.Error())
	}

	b, _ := json.Marshal(cpo)

	client := cp.Settings.PaymentSystemSetting.GetLoggableHttpClient()
	req, err := http.NewRequest(paths[cardPayActionCreatePayment].method, qUrl, bytes.NewBuffer(b))

	token := cp.getToken(cp.Order.PaymentMethod.Params.ExternalId)
	auth := strings.Title(token.TokenType) + " " + token.AccessToken

	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Add(echo.HeaderAuthorization, auth)

	resp, err := client.Do(req)

	if resp.StatusCode != http.StatusOK {
		return NewPaymentResponse(CreatePaymentStatusErrorPaymentSystem, paymentSystemErrorCreateRequestFailed)
	}

	if b, err = ioutil.ReadAll(resp.Body); err != nil {
		return NewPaymentResponse(CreatePaymentStatusErrorPaymentSystem, err.Error())
	}

	var cpResponse *entity.CardPayOrderResponse

	if err = json.Unmarshal(b, &cpResponse); err != nil {
		return NewPaymentResponse(CreatePaymentStatusErrorPaymentSystem, err.Error())
	}

	res := NewPaymentResponse(PaymentStatusOK, model.EmptyString)
	res.RedirectUrl = cpResponse.RedirectUrl

	return res
}

func (cp *CardPay) ProcessPayment(o *model.Order, opn *model.OrderPaymentNotification) *PaymentResponse {
	cpReq := opn.Request.(*entity.CardPayPaymentNotificationWebHookRequest)

	o.Status = model.OrderStatusPaymentSystemReject
	resp := NewPaymentResponse(PaymentStatusErrorValidation, model.EmptyString).SetOrder(o)

	if cp.checkNotificationRequestSignature(opn.RawRequest, cpReq.Signature) == false {
		return resp.SetError(paymentSystemErrorRequestSignatureIsInvalid)
	}

	var err error

	cpReq.CallbackTimeTime, err = time.Parse(cardPayDateFormat, cpReq.CallbackTime)

	if err != nil {
		return resp.SetError(paymentSystemErrorRequestTimeFieldIsInvalid)
	}

	if !cpReq.IsPaymentAllowedStatus() {
		return resp.SetError(paymentSystemErrorRequestStatusIsInvalid)
	}

	switch cpReq.PaymentMethod {
	case cardPayPaymentMethodBankCard:
		o.PaymentMethodPayerAccount = cpReq.CardAccount.MaskedPan
		o.PaymentMethodTxnParams = cpReq.GetBankCardTxnParams()
		break
	case cardPayPaymentMethodQiwi,
		cardPayPaymentMethodWebMoney,
		cardPayPaymentMethodNeteller,
		cardPayPaymentMethodAlipay:
		o.PaymentMethodPayerAccount = cpReq.EWalletAccount.Id
		o.PaymentMethodTxnParams = cpReq.GetEWalletTxnParams()
		break
	case cardPayPaymentMethodBitcoin:
		o.PaymentMethodPayerAccount = cpReq.CryptoCurrencyAccount.CryptoAddress
		o.PaymentMethodTxnParams = cpReq.GetCryptoCurrencyTxnParams()
		break
	default:
		return resp.SetError(paymentSystemErrorRequestPaymentMethodIsInvalid)
	}

	if cpReq.PaymentMethod != o.PaymentMethod.Params.ExternalId {
		return resp.SetError(paymentSystemErrorRequestPaymentMethodIsInvalid)
	}

	switch cpReq.PaymentData.Status {
	case entity.CardPayPaymentResponseStatusDeclined:
		o.Status = model.OrderStatusPaymentSystemDeclined
		break
	case entity.CardPayPaymentResponseStatusCancelled:
		o.Status = model.OrderStatusPaymentSystemCanceled
		break
	case entity.CardPayPaymentResponseStatusCompleted:
		o.Status = model.OrderStatusPaymentSystemComplete
		break
	default:
		return NewPaymentResponse(PaymentStatusTemporary, paymentSystemErrorRequestTemporarySkipped)
	}

	o.PaymentMethodTerminalId = cp.pmSettings[settingsFieldTerminalId]
	o.PaymentMethodOrderId = cpReq.PaymentData.Id
	o.PaymentMethodOrderClosedAt = &cpReq.CallbackTimeTime
	o.PaymentMethodIncomeAmount = cpReq.PaymentData.Amount
	o.PaymentMethodIncomeCurrencyA3 = cpReq.PaymentData.Currency

	res := NewPaymentResponse(PaymentStatusOK, model.EmptyString).SetOrder(o)

	return res
}

func (cp *CardPay) getUrl(action string) (string, error) {
	u, err := url.ParseRequestURI(cp.Url)

	if err != nil {
		return "", err
	}

	u.Path = paths[action].path

	return u.String(), nil
}

func (cp *CardPay) setToken(b []byte, pmKey string) error {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	var token *Token

	if err := json.Unmarshal(b, &token); err != nil {
		return err
	}

	token.AccessTokenExpireTime = time.Now().Add(time.Second * time.Duration(token.AccessTokenExpire))
	token.RefreshTokenExpireTime = time.Now().Add(time.Second * time.Duration(token.RefreshTokenExpire))

	tokens[pmKey] = token

	return nil
}

func (cp *CardPay) getToken(pmKey string) *Token {
	token, ok := tokens[pmKey]

	if !ok {
		return nil
	}

	tn := time.Now().Unix()

	if token.AccessTokenExpire > 0 && token.AccessTokenExpireTime.Unix() >= tn {
		return token
	}

	if token.RefreshTokenExpire <= 0 || token.RefreshTokenExpireTime.Unix() < tn {
		return nil
	}

	if err := cp.refresh(pmKey); err != nil {
		return nil
	}

	return tokens[pmKey]
}

func (cp *CardPay) getCardPayOrder() (*entity.CardPayOrder, error) {
	var err error

	o := &entity.CardPayOrder{
		Request: &entity.CardPayRequest{
			Id:   uuid.NewV4().String(),
			Time: time.Now().UTC().Format(cardPayDateFormat),
		},
		MerchantOrder: &entity.CardPayMerchantOrder{
			Id:          cp.Order.Id.Hex(),
			Description: cp.Order.Description,
			Items: []*entity.CardPayItem{
				{
					Name:        cp.Order.FixedPackage.Name,
					Description: cp.Order.FixedPackage.Name,
					Count:       1,
					Price:       cp.Order.FixedPackage.Price,
				},
			},
		},
		Description:   cp.Order.Description,
		PaymentMethod: cp.Order.PaymentMethod.Params.ExternalId,
		PaymentData: &entity.CardPayPaymentData{
			Currency: cp.Order.PaymentMethodOutcomeCurrency.CodeA3,
			Amount:   cp.Order.PaymentMethodOutcomeAmount,
		},
		Customer: &entity.CardPayCustomer{
			Email:   *cp.Order.PayerData.Email,
			Ip:      cp.Order.PayerData.Ip,
			Account: cp.Order.ProjectAccount,
		},
	}

	switch cp.Order.PaymentMethod.Params.ExternalId {
	case cardPayPaymentMethodBankCard:
		if o, err = cp.geBankCardCardPayOrder(o); err != nil {
			return nil, err
		}
		break
	case cardPayPaymentMethodQiwi,
		cardPayPaymentMethodWebMoney,
		cardPayPaymentMethodNeteller,
		cardPayPaymentMethodAlipay:
		if o, err = cp.getEWalletCardPayOrder(o); err != nil {
			return nil, err
		}
		break
	case cardPayPaymentMethodBitcoin:
		if o, err = cp.getCryptoCurrencyCardPayOrder(o); err != nil {
			return nil, err
		}
		break
	default:
		return nil, errors.New(paymentSystemErrorUnknownPaymentMethod)
	}

	return o, nil
}

func (cp *CardPay) geBankCardCardPayOrder(cpo *entity.CardPayOrder) (*entity.CardPayOrder, error) {
	v := &validator.BankCardValidator{
		Pan:    cp.Order.PaymentRequisites[entity.BankCardFieldPan],
		Cvv:    cp.Order.PaymentRequisites[entity.BankCardFieldCvv],
		Month:  cp.Order.PaymentRequisites[entity.BankCardFieldMonth],
		Year:   cp.Order.PaymentRequisites[entity.BankCardFieldYear],
		Holder: cp.Order.PaymentRequisites[entity.BankCardFieldHolder],
	}

	if err := v.Validate(); err != nil {
		return nil, err
	}

	if len(v.Year) < 3 {
		v.Year = strconv.Itoa(time.Now().UTC().Year())[:2] + v.Year
	}

	expire := v.Month + "/" + v.Year

	cpo.CardAccount = &entity.CardPayCardAccount{
		Card: &entity.CardPayBankCardAccount{
			Pan:        v.Pan,
			HolderName: v.Holder,
			Cvv:        v.Cvv,
			Expire:     expire,
		},
	}

	return cpo, nil
}

func (cp *CardPay) getEWalletCardPayOrder(cpo *entity.CardPayOrder) (*entity.CardPayOrder, error) {
	ewallet, ok := cp.Order.PaymentRequisites[entity.EWalletFieldIdentifier]

	if !ok || len(ewallet) <= 0 {
		return nil, errors.New(paymentSystemErrorEWalletIdentifierIsInvalid)
	}

	cpo.EWalletAccount = &entity.CardPayEWalletAccount{
		Id: cp.Order.PaymentRequisites[entity.EWalletFieldIdentifier],
	}

	return cpo, nil
}

func (cp *CardPay) getCryptoCurrencyCardPayOrder(cpo *entity.CardPayOrder) (*entity.CardPayOrder, error) {
	address, ok := cp.Order.PaymentRequisites[entity.CryptoFieldIdentifier]

	if !ok || len(address) <= 0 {
		return nil, errors.New(paymentSystemErrorCryptoCurrencyAddressIsInvalid)
	}

	cpo.CryptoCurrencyAccount = &entity.CardPayCryptoCurrencyAccount{
		RollbackAddress: address,
	}

	return cpo, nil
}

func (cp *CardPay) checkNotificationRequestSignature(reqRaw string, reqSign string) bool {
	return true

	h := sha512.New()
	h.Write([]byte(reqRaw + cp.pmSettings[settingsFieldCallbackSecretWord]))

	return hex.EncodeToString(h.Sum(nil)) == reqSign
}
