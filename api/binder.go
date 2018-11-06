package api

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/labstack/echo"
)

type OrderFormBinder struct{}

func (cb *OrderFormBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	db := new(echo.DefaultBinder)

	if err = db.Bind(i, ctx); err != nil {
		return err
	}

	params, err := ctx.FormParams()
	addParams := make(map[string]string)

	if err != nil {
		return err
	}

	for key, value := range params {
		if _, ok := model.OrderReservedWords[key]; !ok {
			addParams[key] = value[0]
		}
	}

	i.(*model.OrderScalar).Other = addParams

	return
}