package payment_system

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/labstack/echo"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httputil"
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

	cardPayErrorAuthenticateFailed = "authentication failed"
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

func Init() {

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
		return errors.New(cardPayErrorAuthenticateFailed)
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
		return errors.New(cardPayErrorAuthenticateFailed)
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

func (cp *CardPay) CreatePayment(o *model.Order) error {
	qUrl, err := cp.getUrl(cardPayActionCreatePayment)

	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	cpo := &model.CardPayOrder{
		Request: &model.CardPayRequest{
			Id:   o.Id.Hex(),
			Time: time.Now().Format("2006-01-02T15:04:05Z"),
		},
		MerchantOrder: &model.CardPayMerchantOrder{
			Id: o.Id.Hex(),
			Description: "Test",
			Items: []*model.CardPayItem {
				{
					Name: o.FixedPackage.Name,
					Description: o.FixedPackage.Name,
					Count: 1,
					Price: o.FixedPackage.Price,
				},
			},
		},
		Description:     "Test description",
		PaymentMethod:   "BANK_CARD",
		PaymentData: &model.CardPayPaymentData{
			Currency: o.PaymentMethodOutcomeCurrency.CodeA3,
			Amount:   math.Floor(o.PaymentMethodOutcomeAmount*100)/100,
		},
		CardAccount: &model.CardPayCardAccount{
			Card: &model.CardPayBankCardAccount{
				Pan:        "4000000000000002",
				HolderName: "Mr. Card Holder",
				Cvv:        "123",
				Expire:     "12/2019",
			},
		},
		Customer: &model.CardPayCustomer{
			Email:   *o.PayerData.Email,
			Ip:      o.PayerData.Ip,
			Account: o.ProjectAccount,
		},
	}
	bytesRepresentation, err := json.Marshal(cpo)
	if err != nil {
		log.Fatalln(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(paths[cardPayActionCreatePayment].method, qUrl, bytes.NewBuffer(bytesRepresentation))

	token := cp.getToken("cards")
	auth := strings.Title(token.TokenType) + " " + token.AccessToken

	req.Header.Add(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Add(echo.HeaderAuthorization, auth)

	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("%s\n\n", dump)

	resp, err := client.Do(req)

	dump, _ = httputil.DumpResponse(resp, true)
	fmt.Printf("%s\n\n", dump)

	if resp.StatusCode != http.StatusOK {
		return errors.New("not 200")
	}

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
