package api

import (
	"errors"
	"fmt"
	"github.com/ProtocolONE/p1pay.api/config"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const (
	errorMessage = "Field validation for '%s' failed on the '%s' tag"
)

type Merchant struct {
	Identifier string
	Projects   []string
}

type GetParams struct {
	limit  int
	offset int
}

type Api struct {
	Http             *echo.Echo
	config           *config.Config
	database         dao.Database
	logger           *zap.SugaredLogger
	validate         *validator.Validate
	accessRouteGroup *echo.Group
	handlers         map[string]interface{}

	Merchant
	GetParams
}

func NewServer(config *config.Jwt, database dao.Database, logger *zap.SugaredLogger) (*Api, error) {
	api := &Api{
		Http:     echo.New(),
		database: database,
		logger:   logger,
		validate: validator.New(),
		handlers: make(map[string]interface{}),
	}

	api.accessRouteGroup = api.Http.Group("/api/v1/s")
	api.accessRouteGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    config.SignatureSecret,
		SigningMethod: config.Algorithm,
	}))
	api.accessRouteGroup.Use(api.SetMerchantIdentifierMiddleware)

	api.Http.Use(api.LimitOffsetMiddleware)
	api.Http.Use(middleware.Logger())
	api.Http.Use(middleware.Recover())
	api.Http.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{"authorization", "content-type"},
	}))

	api.
		InitCurrencyRoutes()

	/*api.InitMerchantRoutes().
		InitProjectRoutes().
		InitCurrencyRoutes()*/

	return api, nil
}

func (api *Api) Start() error {
	return api.Http.Start(":3001")
}

func (api *Api) getFirstValidationError(err error) string {
	vErr := err.(validator.ValidationErrors)[0]

	return fmt.Sprintf(errorMessage, vErr.Field(), vErr.Tag())
}

func (api *Api) SetMerchantIdentifierMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)

		id, ok := claims["id"]

		if !ok {
			c.Error(errors.New("merchant identifier not found"))
		}

		api.Merchant.Identifier = id.(string)

		return next(c)
	}
}
