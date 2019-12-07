package helpers

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func SetResponseCookie(ctx echo.Context, name, value, domain string, expires time.Time) {
	if name == "" || value == "" {
		return
	}

	now := time.Now()
	zap.S().Infow("SetResponseCookie", "name", name, "expires", expires, "domain", domain, "time.Now()", now, "utc", now.UTC())

	cookie := new(http.Cookie)
	cookie.Name = name
	cookie.Value = value
	cookie.Domain = domain
	cookie.Expires = expires
	cookie.HttpOnly = true
	ctx.SetCookie(cookie)
}

func GetRequestCookie(ctx echo.Context, name string) string {
	if name == "" {
		return ""
	}
	cookie, err := ctx.Cookie(name)
	if err == nil && cookie != nil && cookie.Value != "" {
		return cookie.Value
	}
	return ""
}
