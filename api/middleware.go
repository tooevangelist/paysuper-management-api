package api

import (
	"bytes"
	"github.com/labstack/echo"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-management-api/database/model"
	"io/ioutil"
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

func (api *Api) AuthUserMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		api.authUser = &AuthUser{
			Id:        pkg.SystemUserId,
			Name:      "System User",
			Merchants: make(map[string]bool),
			Roles:     make(map[string]bool),
		}

		return next(c)
	}
}
