package payment_system

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/payment_system/entity"
	"github.com/ProtocolONE/p1pay.api/payment_system/validator"
	"github.com/labstack/echo"
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
	cardPayPaymentMethodBankCard = "BANK_CARD"
	cardPayPaymentMethodWebMoney = "WEBMONEY"
	cardPayPaymentMethodQiwi     = "QIWI"
	cardPayPaymentMethodNeteller = "NETELLER"
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
	mu sync.Mutex
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
	return &CardPay{Settings: settings, Order: o}
}

func (cp *CardPay) Auth(pmKey string) error {
	if token := cp.getToken(pmKey); token != nil {
		return nil
	}

	data := url.Values{
		cardPayRequestFieldGrantType:    []string{cardPayGrantTypePassword},
		cardPayRequestFieldTerminalCode: []string{"15985"},
		cardPayRequestFieldPassword:     []string{"A1tph4I6BD0f"},
	}

	qUrl, err := cp.getUrl(cardPayActionAuthenticate)

	if err != nil {
		return err
	}

	client := &http.Client{}
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
		cardPayRequestFieldTerminalCode: []string{"15985"},
		cardPayRequestFieldRefreshToken: []string{tokens[pmKey].RefreshToken},
	}

	qUrl, err := cp.getUrl(cardPayActionRefresh)

	if err != nil {
		return err
	}

	client := &http.Client{}
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

func (cp *CardPay) CreatePayment() error {
	if err := cp.Auth(cp.Order.PaymentMethod.Params.ExternalId); err != nil {
		return err
	}

	qUrl, err := cp.getUrl(cardPayActionCreatePayment)

	if err != nil {
		return err
	}

	cpo, err := cp.getCardPayOrder()

	if err != nil {
		return err
	}

	b, _ := json.Marshal(cpo)

	client := GetLoggableHttpClient()
	req, err := http.NewRequest(paths[cardPayActionCreatePayment].method, qUrl, bytes.NewBuffer(b))

	token := cp.getToken(cp.Order.PaymentMethod.Params.ExternalId)
	auth := strings.Title(token.TokenType) + " " + token.AccessToken

	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Add(echo.HeaderAuthorization, auth)

	resp, err := client.Do(req)

	if resp.StatusCode != http.StatusOK {
		return errors.New(paymentSystemErrorCreateRequestFailed)
	}

	return nil
}

func (cp *CardPay) ProcessPayment() error {
	return nil
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
			Id:   cp.Order.Id.Hex(),
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
	case cardPayPaymentMethodQiwi:
	case cardPayPaymentMethodWebMoney:
	case cardPayPaymentMethodNeteller:
		if o, err = cp.geEWalletCardPayOrder(o); err != nil {
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
		Pan:    cp.Order.PaymentRequisites[bankCardFieldPan],
		Cvv:    cp.Order.PaymentRequisites[bankCardFieldCvv],
		Month:  cp.Order.PaymentRequisites[bankCardFieldMonth],
		Year:   cp.Order.PaymentRequisites[bankCardFieldYear],
		Holder: cp.Order.PaymentRequisites[bankCardFieldHolder],
	}

	if err := v.Validate(); err != nil {
		return nil, err
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

func (cp *CardPay) geEWalletCardPayOrder(cpo *entity.CardPayOrder) (*entity.CardPayOrder, error) {
	ewallet, ok := cp.Order.PaymentRequisites[eWalletFieldIdentifier]

	if !ok || len(ewallet) <= 0 {
		return nil, errors.New(paymentSystemErrorEWalletIdentifierIsInvalid)
	}

	cpo.EWalletAccount = &entity.CardPayEWalletAccount{
		Id: cp.Order.PaymentRequisites[eWalletFieldIdentifier],
	}

	return cpo, nil
}
