package api

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"net/http"
	"strconv"
)

func (api *Api) LimitOffsetSortMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		if ctx.Request().Method != http.MethodGet {
			return next(ctx)
		}

		limit, err := strconv.Atoi(ctx.QueryParam(QueryParameterNameLimit))

		if err != nil {
			limit = LimitDefault
		}

		offset, err := strconv.Atoi(ctx.QueryParam(QueryParameterNameOffset))

		if err != nil {
			offset = OffsetDefault
		}

		qParams := ctx.QueryParams()

		sort := DefaultSort

		if s, ok := qParams[QueryParameterNameSort]; ok {
			sort = s
		}

		api.limit = int32(limit)
		api.offset = int32(offset)
		api.sort = sort

		return next(ctx)
	}
}

func (api *Api) RawBodyMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		buf, _ := ioutil.ReadAll(ctx.Request().Body)
		rdr := ioutil.NopCloser(bytes.NewBuffer(buf))

		ctx.Request().Body = rdr
		api.rawBody = string(buf)

		return next(ctx)
	}
}
