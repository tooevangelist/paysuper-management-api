package api

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
)

func (api *Api) LimitOffsetSortMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().Method != http.MethodGet {
			return next(ctx)
		}

		limit, err := strconv.Atoi(ctx.QueryParam(model.QueryParameterNameLimit))

		if err != nil {
			limit = model.DefaultLimit
		}

		offset, err := strconv.Atoi(ctx.QueryParam(model.QueryParameterNameOffset))

		if err != nil {
			offset = model.DefaultOffset
		}

		qParams := ctx.QueryParams()

		sort := model.DefaultSort

		if s, ok := qParams[model.QueryParameterNameSort]; ok {
			sort = s
		}

		api.limit = limit
		api.offset = offset
		api.sort = sort

		return next(ctx)
	}
}
