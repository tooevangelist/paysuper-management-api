package api

import (
	"errors"
	"fmt"
	"github.com/ProtocolONE/p1pay.api/config"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/oschwald/geoip2-golang"
	"github.com/ttacon/libphonenumber"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"html/template"
	"io"
	"net/http"
	"time"
)

const (
	errorMessage                      = "Field validation for '%s' failed on the '%s' tag"
	ResponseMessageInvalidRequestData = "Invalid request data"
	ResponseMessageAccessDenied       = "Access denied"
	ResponseMessageNotFound           = "Not found"

	apiWebHookGroupPath = "/webhook"
)

var funcMap = template.FuncMap{
	"For": func(start, end int) (stream chan int) {
		stream = make(chan int)

		go func() {
			for i := start; i <= end; i++ {
				stream <- i
			}
			close(stream)
		}()

		return
	},
	"Now": time.Now,
	"Increment": func(i int, add int) int {
		return i + add
	},
	"BsonObjectIdToString": func(objectId bson.ObjectId) string {
		return objectId.Hex()
	},
}

type ServerInitParams struct {
	Config              *config.Jwt
	Database            dao.Database
	Logger              *zap.SugaredLogger
	GeoDbReader         *geoip2.Reader
	PaymentSystemConfig map[string]interface{}
}

type Template struct {
	templates *template.Template
}

type Merchant struct {
	Identifier string
}

type GetParams struct {
	limit  int
	offset int
}

type Order struct {
	PayerPhone *libphonenumber.PhoneNumber
}

type Api struct {
	Http                *echo.Echo
	config              *config.Config
	Database            dao.Database
	Logger              *zap.SugaredLogger
	Validate            *validator.Validate
	accessRouteGroup    *echo.Group
	geoDbReader         *geoip2.Reader
	PaymentSystemConfig map[string]interface{}
	WebHookGroup        *echo.Group

	Merchant
	GetParams
	Order
}

func NewServer(p *ServerInitParams) (*Api, error) {
	api := &Api{
		Http:                echo.New(),
		Database:            p.Database,
		Logger:              p.Logger,
		Validate:            validator.New(),
		geoDbReader:         p.GeoDbReader,
		PaymentSystemConfig: p.PaymentSystemConfig,
	}

	renderer := &Template{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("web/template/*.html")),
	}
	api.Http.Renderer = renderer
	api.Http.Static("/", "web/static")

	api.Validate.RegisterStructValidation(ProjectStructValidator, model.ProjectScalar{})
	api.Validate.RegisterStructValidation(api.OrderStructValidator, model.OrderScalar{})

	api.accessRouteGroup = api.Http.Group("/api/v1/s")
	api.accessRouteGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    p.Config.SignatureSecret,
		SigningMethod: p.Config.Algorithm,
	}))
	api.accessRouteGroup.Use(api.SetMerchantIdentifierMiddleware)

	api.WebHookGroup = api.Http.Group(apiWebHookGroupPath)

	api.Http.Use(api.LimitOffsetMiddleware)
	api.Http.Use(middleware.Logger())
	api.Http.Use(middleware.Recover())
	api.Http.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowHeaders: []string{"authorization", "content-type"},
	}))

	api.
		InitCurrencyRoutes().
		InitCountryRoutes().
		InitMerchantRoutes().
		InitProjectRoutes().
		InitOrderV1Routes()

	api.Http.GET("/docs", func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, "docs.html", map[string]interface{}{})
	})

	return api, nil
}

func (api *Api) Start() error {
	return api.Http.Start(":3001")
}

func (api *Api) GetFirstValidationError(err error) string {
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

func (t *Template) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
