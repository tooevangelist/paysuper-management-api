package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/labstack/echo"
	"github.com/micro/go-micro/client"
	"github.com/paysuper/paysuper-tax-service/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

type TaxServiceMock struct {
	mock.Mock
}

type TaxesTestSuite struct {
	suite.Suite
	handler *taxesRoute
	api     *Api
}

func Test_Taxes(t *testing.T) {
	suite.Run(t, new(TaxesTestSuite))
}

func (suite *TaxesTestSuite) SetupTest() {
	suite.api = &Api{
		Http:       echo.New(),
		validate:   validator.New(),
		taxService: createNewTaxServiceMock(),
		authUser:   &AuthUser{Id: "ffffffffffffffffffffffff"},
	}

	suite.api.authUserRouteGroup = suite.api.Http.Group(apiAuthUserGroupPath)
	err := suite.api.validate.RegisterValidation("phone", suite.api.PhoneValidator)
	assert.NoError(suite.T(), err)

	suite.handler = &taxesRoute{Api: suite.api}
}

func (suite *TaxesTestSuite) Test_Routes() {
	shouldHaveRoutes := [][]string{
		{"/admin/api/v1/taxes", http.MethodGet},
		{"/admin/api/v1/taxes", http.MethodPost},
		{"/admin/api/v1/taxes/:id", http.MethodDelete},
	}

	api := suite.api.initTaxesRoutes()

	routeCount := 0

	routes := api.Http.Routes()
	for _, v := range shouldHaveRoutes {
		for _, r := range routes {
			if v[0] != r.Path || v[1] != r.Method {
				continue
			}

			routeCount++
		}
	}

	assert.Equal(suite.T(), len(shouldHaveRoutes), routeCount)
}

func (suite *TaxesTestSuite) Test_GetRates() {
	rates := getRates(suite.T(), suite.handler, "RU", "", "", "", 0, 0)
	assert.Len(suite.T(), rates, LimitDefault)
	assert.Equal(suite.T(), "RU", rates[0].Country)

	rates = getRates(suite.T(), suite.handler, "", "City", "", "", 0, 0)
	assert.Len(suite.T(), rates, LimitDefault)
	assert.Equal(suite.T(), "City", rates[0].City)

	rates = getRates(suite.T(), suite.handler, "", "", "00001", "", 0, 0)
	assert.Len(suite.T(), rates, LimitDefault)
	assert.Equal(suite.T(), "00001", rates[0].Zip)

	rates = getRates(suite.T(), suite.handler, "", "", "", "NY", 0, 0)
	assert.Len(suite.T(), rates, LimitDefault)
	assert.Equal(suite.T(), "NY", rates[0].State)

	rates = getRates(suite.T(), suite.handler, "", "", "", "NY", 1, 0)
	assert.Len(suite.T(), rates, 1)
	assert.Equal(suite.T(), "NY", rates[0].State)

	rates = getRates(suite.T(), suite.handler, "", "", "", "NY", 1, 2)
	assert.Len(suite.T(), rates, 1)
	assert.Equal(suite.T(), "NY", rates[0].State)
	assert.EqualValues(suite.T(), 2, rates[0].Id)
}

func getRates(t *testing.T, handler *taxesRoute, country, city, zip, state string, limit, offset int) []*tax_service.TaxRate {
	t.Helper()

	q := make(url.Values)
	q.Set("city", city)
	q.Set("country", country)
	q.Set("zip", zip)
	q.Set("state", state)
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)

	e := echo.New()
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := e.NewContext(req, rec)

	ctx.SetPath("/taxes")
	err := handler.getTaxes(ctx)

	assert.NoError(t, err)

	var response []*tax_service.TaxRate
	err = json.Unmarshal(rec.Body.Bytes(), &response)

	return response
}

func (suite *TaxesTestSuite) Test_GetRatesError() {
	req := httptest.NewRequest(http.MethodGet, "/?city=fail", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	e := echo.New()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/taxes")

	err := suite.handler.getTaxes(ctx)

	if assert.Error(suite.T(), err) {
		assert.Equal(suite.T(), http.StatusInternalServerError, err.(*echo.HTTPError).Code)
	}
}

func (suite *TaxesTestSuite) Test_CreateTax() {
	obj := &tax_service.TaxRate{
		Zip:  "00001",
		City: "City",
	}

	b, _ := json.Marshal(obj)

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := e.NewContext(req, rec)

	ctx.SetPath("/taxes")
	if assert.NoError(suite.T(), suite.handler.setTax(ctx)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
		assert.Equal(suite.T(), b, rec.Body.Bytes())
	}
}

func (suite *TaxesTestSuite) Test_CreateTaxWithError() {
	testCreateTaxWithError(suite.T(), suite.handler, nil, http.StatusBadRequest)
	testCreateTaxWithError(suite.T(), suite.handler, &tax_service.TaxRate{Id: 1}, http.StatusInternalServerError)
}

func testCreateTaxWithError(t *testing.T, handler *taxesRoute, obj *tax_service.TaxRate, code int) {
	t.Helper()

	var req *http.Request
	if obj == nil {
		req = httptest.NewRequest(http.MethodPost, "/", nil)
	} else {
		b, _ := json.Marshal(obj)
		req = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	}

	e := echo.New()

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	ctx := e.NewContext(req, rec)

	ctx.SetPath("/taxes")
	err := handler.setTax(ctx)

	if assert.Error(t, err) {
		assert.Equal(t, code, err.(*echo.HTTPError).Code)
	}
}

func (suite *TaxesTestSuite) Test_DeleteTax() {
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	e := echo.New()
	ctx := e.NewContext(req, rec)

	ctx.SetPath("/taxes/:id")
	ctx.SetParamNames("id")
	ctx.SetParamValues("1")

	if assert.NoError(suite.T(), suite.handler.deleteTax(ctx)) {
		assert.Equal(suite.T(), http.StatusOK, rec.Code)
	}
}

func (suite *TaxesTestSuite) Test_DeleteTaxWithInvalidId() {
	testDeleteTaxWithError(suite.T(), suite.handler, "", http.StatusBadRequest)
	testDeleteTaxWithError(suite.T(), suite.handler, "string", http.StatusBadRequest)
	testDeleteTaxWithError(suite.T(), suite.handler, "0.1", http.StatusBadRequest)
	testDeleteTaxWithError(suite.T(), suite.handler, "0", http.StatusInternalServerError)
}

func testDeleteTaxWithError(t *testing.T, handler *taxesRoute, id string, code int) {
	t.Helper()

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()

	e := echo.New()
	ctx := e.NewContext(req, rec)
	ctx.SetPath("/taxes/:id")
	ctx.SetParamNames("id")
	ctx.SetParamValues(id)

	err := handler.deleteTax(ctx)

	if assert.Error(t, err) {
		assert.Equal(t, code, err.(*echo.HTTPError).Code)
	}
}

func createNewTaxServiceMock() tax_service.TaxService {
	return &TaxServiceMock{}
}

func (ts *TaxServiceMock) GetRate(ctx context.Context, in *tax_service.GetRateRequest, opts ...client.CallOption) (*tax_service.GetRateResponse, error) {
	panic("this method is not implemented in mock")
}

func (ts *TaxServiceMock) GetRates(ctx context.Context, in *tax_service.GetRatesRequest, opts ...client.CallOption) (*tax_service.GetRatesResponse, error) {
	if in.City == "fail" {
		return nil, errors.New("Invalid request")
	}
	res := &tax_service.GetRatesResponse{}
	for i := 0; i < int(in.Limit); i++ {
		res.Rates = append(
			res.Rates,
			&tax_service.TaxRate{
				Id:      uint32(i) + uint32(in.Offset),
				City:    in.City,
				Zip:     in.Zip,
				Country: in.Country,
				Rate:    0.1,
				State:   in.State,
			},
		)
	}

	return res, nil
}

func (ts *TaxServiceMock) CreateOrUpdate(ctx context.Context, in *tax_service.TaxRate, opts ...client.CallOption) (*tax_service.TaxRate, error) {
	if in.Id == 1 {
		return nil, errors.New("Invalid request")
	}

	return in, nil
}

func (ts *TaxServiceMock) DeleteRateById(ctx context.Context, in *tax_service.DeleteRateRequest, opts ...client.CallOption) (*tax_service.DeleteRateResponse, error) {
	if in.Id == 1 {
		return &tax_service.DeleteRateResponse{}, nil
	}

	return nil, errors.New("Invalid request")
}
