package api

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-management-api/database/model"
	"io/ioutil"
	"strconv"
	"time"
)

const (
	binderErrorQueryParamsIsEmpty   = "required params not found"
	binderErrorPeriodIsRequire      = "period is required field"
	binderErrorFromIsRequire        = "date from is required field"
	binderErrorToIsRequire          = "date to is required field"
	binderErrorUnknownRevenuePeriod = "unknown revenue period"
)

type OrderFormBinder struct{}
type OrderJsonBinder struct{}
type OrderRevenueDynamicRequestBinder struct{}
type OrderAccountingPaymentRequestBinder struct{}
type PaymentCreateProcessBinder struct{}

func (cb *OrderFormBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	db := new(echo.DefaultBinder)

	if err = db.Bind(i, ctx); err != nil {
		return err
	}

	params, err := ctx.FormParams()
	addParams := make(map[string]string)
	rawParams := make(map[string]string)

	if err != nil {
		return err
	}

	o := i.(*billing.OrderCreateRequest)

	for key, value := range params {
		if _, ok := model.OrderReservedWords[key]; !ok {
			addParams[key] = value[0]
		}

		rawParams[key] = value[0]
	}

	o.Other = addParams
	o.RawParams = rawParams

	return
}

func (cb *OrderJsonBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	var buf []byte

	if ctx.Request().Body != nil {
		buf, err = ioutil.ReadAll(ctx.Request().Body)
		rdr := ioutil.NopCloser(bytes.NewBuffer(buf))

		if err != nil {
			return err
		}

		ctx.Request().Body = rdr
	}

	db := new(echo.DefaultBinder)

	if err = db.Bind(i, ctx); err != nil {
		return err
	}

	i.(*billing.OrderCreateRequest).RawBody = string(buf)

	return
}

func (cb *OrderRevenueDynamicRequestBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	period := ctx.Param(model.OrderRevenueDynamicRequestFieldPeriod)

	if period == "" {
		return errors.New(binderErrorPeriodIsRequire)
	}

	if _, ok := model.RevenuePeriods[period]; !ok {
		return errors.New(binderErrorUnknownRevenuePeriod)
	}

	params := ctx.QueryParams()

	s := i.(*model.RevenueDynamicRequest)
	s.Period = period
	s.Project = []bson.ObjectId{}

	if projects, ok := params[model.OrderRevenueDynamicRequestFieldProject]; ok {
		for _, project := range projects {
			if bson.IsObjectIdHex(project) == false {
				return errors.New(model.ResponseMessageProjectIdIsInvalid)
			}

			s.Project = append(s.Project, bson.ObjectIdHex(project))
		}
	}

	s, err = OrderPrepareAccountingPaymentRequest(s, ctx)

	if err != nil {
		return err
	}

	return
}

func (cb *OrderAccountingPaymentRequestBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	rdr := i.(*model.RevenueDynamicRequest)
	rdr, err = OrderPrepareAccountingPaymentRequest(rdr, ctx)

	if err != nil {
		return err
	}

	return
}

func OrderPrepareAccountingPaymentRequest(rdr *model.RevenueDynamicRequest, ctx echo.Context) (*model.RevenueDynamicRequest, error) {
	params := ctx.QueryParams()

	if len(params) <= 0 {
		return nil, errors.New(binderErrorQueryParamsIsEmpty)
	}

	if _, ok := params[model.OrderRevenueDynamicRequestFieldFrom]; !ok {
		return nil, errors.New(binderErrorFromIsRequire)
	}

	if _, ok := params[model.OrderRevenueDynamicRequestFieldTo]; !ok {
		return nil, errors.New(binderErrorToIsRequire)
	}

	t, err := strconv.ParseInt(params[model.OrderRevenueDynamicRequestFieldFrom][0], 10, 64)

	if err != nil {
		return nil, err
	}

	rdr.From = time.Unix(t, 0)

	t, err = strconv.ParseInt(params[model.OrderRevenueDynamicRequestFieldTo][0], 10, 64)

	if err != nil {
		return nil, err
	}

	rdr.To = time.Unix(t, 0)

	return rdr, nil
}

func (cb *PaymentCreateProcessBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	db := new(echo.DefaultBinder)
	untypedData := make(map[string]interface{})

	if err = db.Bind(&untypedData, ctx); err != nil {
		return
	}

	data := i.(map[string]string)

	for k, v := range untypedData {
		switch sv := v.(type) {
		case bool:
			data[k] = "0"

			if sv == true {
				data[k] = "1"
			}
			break
		default:
			data[k] = fmt.Sprintf("%v", sv)
		}
	}

	return
}
