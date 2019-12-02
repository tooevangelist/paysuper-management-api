package dispatcher

import (
	"bytes"
	"fmt"
	jwtverifier "github.com/ProtocolONE/authone-jwt-verifier-golang"
	jwtMiddleware "github.com/ProtocolONE/authone-jwt-verifier-golang/middleware/echo"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	casbinMiddleware "github.com/paysuper/echo-casbin-middleware"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"io/ioutil"
	"net/http"
	"strconv"
)

// RecoverMiddleware
func (d *Dispatcher) RecoverMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					err, ok := r.(error)
					if !ok {
						err = fmt.Errorf("%v", r)
					}
					d.L().Critical("[PANIC RECOVER] %s", logger.Args(err.Error()), logger.Stack("stacktrace"))
					c.Error(err)
				}
			}()
			return next(c)
		}
	}
}

// GetUserDetailsMiddleware
func (d *Dispatcher) GetUserDetailsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		auth := ctx.Request().Header.Get(echo.HeaderAuthorization)

		if auth == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, common.ErrorMessageAuthorizationHeaderNotFound.Message)
		}

		match := common.TokenRegex.FindStringSubmatch(auth)

		if len(match) < 1 {
			return echo.NewHTTPError(http.StatusUnauthorized, common.ErrorMessageAuthorizationTokenNotFound.Message)
		}

		u, err := d.appSet.JwtVerifier.GetUserInfo(ctx.Request().Context(), match[1])

		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, common.ErrorMessageAuthorizedUserNotFound.Message)
		}

		user := common.ExtractUserContext(ctx)
		user.Id = u.UserID
		user.Email = u.Email
		common.SetUserContext(ctx, user)

		return next(ctx)
	}
}

// LimitOffsetSortPreMiddleware
func (d *Dispatcher) LimitOffsetSortPreMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Limit
		if c.Request().Method != http.MethodGet {
			return next(c)
		}
		limit, err := strconv.Atoi(c.QueryParam(common.QueryParameterNameLimit))
		if err != nil {
			limit = int(d.globalCfg.LimitDefault)
		}
		// Offset
		offset, err := strconv.Atoi(c.QueryParam(common.QueryParameterNameOffset))
		if err != nil {
			offset = int(d.globalCfg.OffsetDefault)
		}
		// Sort
		qParams := c.QueryParams()
		sort := common.DefaultSort
		if s, ok := qParams[common.QueryParameterNameSort]; ok {
			sort = s
		}
		//
		common.SetCursorContext(c, &common.Cursor{
			Limit:  int32(limit),
			Offset: int32(offset),
			Sort:   sort,
		})
		return next(c)
	}
}

// RawBodyPreMiddleware
func (d *Dispatcher) RawBodyPreMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		buf, _ := ioutil.ReadAll(c.Request().Body)
		rdr := ioutil.NopCloser(bytes.NewBuffer(buf))
		c.Request().Body = rdr
		common.SetRawBodyContext(c, buf)
		return next(c)
	}
}

// RawBodyPreMiddleware
func (d *Dispatcher) CasbinMiddleware(fn func(c echo.Context) string) echo.MiddlewareFunc {
	cfg := casbinMiddleware.Config{
		Skipper:          middleware.DefaultSkipper,
		Mode:             casbinMiddleware.EnforceModeEnforcing,
		Logger:           d.L(),
		CtxUserExtractor: fn,
	}
	return casbinMiddleware.MiddlewareWithConfig(d.ms.Client(), cfg)
}

// BodyDumpMiddleware
func (d *Dispatcher) BodyDumpMiddleware() echo.MiddlewareFunc {
	return middleware.BodyDump(func(ctx echo.Context, reqBody, resBody []byte) {
		data := map[string]interface{}{
			"request_headers":  common.RequestResponseHeadersToString(ctx.Request().Header),
			"request_body":     string(reqBody),
			"response_headers": common.RequestResponseHeadersToString(ctx.Response().Header()),
			"response_body":    string(resBody),
		}
		d.L().Info(ctx.Path(), logger.WithFields(data))
	})
}

// MerchantBinderPreMiddleware
func (d *Dispatcher) MerchantBinderPreMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if paramValue := c.Param(common.RequestParameterMerchantId); paramValue != "" {
			user := common.ExtractUserContext(c)
			if paramValue != user.MerchantId {
				return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestDataInvalid)
			}
		}
		common.SetBinder(c, common.MerchantBinderDefault)
		return next(c)
	}
}

// SystemBinderPreMiddleware
func (d *Dispatcher) SystemBinderPreMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		common.SetBinder(c, common.SystemBinderDefault)
		return next(c)
	}
}

// AuthOneMerchantPreMiddleware
func (d *Dispatcher) AuthOneMerchantPreMiddleware() echo.MiddlewareFunc {
	return common.ContextWrapperCallback(func(c echo.Context, next echo.HandlerFunc) error {
		handleFn := jwtMiddleware.AuthOneJwtCallableWithConfig(
			d.appSet.JwtVerifier,
			func(ui *jwtverifier.UserInfo) {
				user := common.ExtractUserContext(c)
				user.Name = "Merchant User"

				res, err := d.appSet.Services.Billing.GetMerchantsForUser(
					c.Request().Context(),
					&grpc.GetMerchantsForUserRequest{UserId: user.Id},
				)

				if err != nil {
					d.L().Error(c.Path(), logger.Args(err.Error()), logger.Stack("stacktrace"))
					return
				}

				if len(res.Merchants) < 1 {
					d.L().Error(c.Path(), logger.Args("user_id", user.Id))
					return
				}

				user.Role = res.Merchants[0].Role
				user.MerchantId = res.Merchants[0].Id
				common.SetUserContext(c, user)
			},
		)(next)
		return handleFn(c)
	})
}
