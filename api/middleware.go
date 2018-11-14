package api

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func (api *Api) LimitOffsetMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().Method != http.MethodGet {
			return next(ctx)
		}

		limit, err := strconv.Atoi(ctx.QueryParam("limit"))

		if err != nil {
			limit = model.DefaultLimit
		}

		offset, err := strconv.Atoi(ctx.QueryParam("offset"))

		if err != nil {
			offset = model.DefaultOffset
		}

		api.limit = limit
		api.offset = offset

		return next(ctx)
	}
}
