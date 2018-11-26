package api

import (
	"bytes"
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
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

func (cb *OrderFormBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	db := new(echo.DefaultBinder)

	if err = db.Bind(i, ctx); err != nil {
		return err
	}

	params, err := ctx.FormParams()
	addParams := make(map[string]interface{})
	rawParams := make(map[string]string)

	if err != nil {
		return err
	}

	o := i.(*model.OrderScalar)

	for key, value := range params {
		if _, ok := model.OrderReservedWords[key]; !ok {
			addParams[key] = value[0]
		}

		rawParams[key] = value[0]
	}

	o.Other = addParams
	o.RawRequestParams = rawParams

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

	i.(*model.OrderScalar).RawRequestBody = string(buf)

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

	if len(params) <= 0 {
		return errors.New(binderErrorQueryParamsIsEmpty)
	}

	if _, ok := params[model.OrderRevenueDynamicRequestFieldFrom]; !ok {
		return errors.New(binderErrorFromIsRequire)
	}

	if _, ok := params[model.OrderRevenueDynamicRequestFieldTo]; !ok {
		return errors.New(binderErrorToIsRequire)
	}

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

	t, err := strconv.ParseInt(params[model.OrderRevenueDynamicRequestFieldFrom][0], 10, 64)

	if err != nil {
		return err
	}

	s.From = time.Unix(t, 0)

	t, err = strconv.ParseInt(params[model.OrderRevenueDynamicRequestFieldTo][0], 10, 64)

	if err != nil {
		return err
	}

	s.To = time.Unix(t, 0)

	return
}
