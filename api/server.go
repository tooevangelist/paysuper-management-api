package api

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/api/webhook"
	"github.com/ProtocolONE/p1pay.api/config"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
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
	Config                  *config.Jwt
	Database                dao.Database
	Logger                  *zap.SugaredLogger
	GeoDbReader             *geoip2.Reader
	PaymentSystemConfig     map[string]interface{}
	PSPAccountingCurrencyA3 string
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
	Http                    *echo.Echo
	config                  *config.Config
	database                dao.Database
	logger                  *zap.SugaredLogger
	validate                *validator.Validate
	accessRouteGroup        *echo.Group
	geoDbReader             *geoip2.Reader
	PaymentSystemConfig     map[string]interface{}
	pspAccountingCurrencyA3 string
	webHookRawBody          string

	Merchant
	GetParams
	Order
}

func NewServer(p *ServerInitParams) (*Api, error) {
	api := &Api{
		Http:                    echo.New(),
		database:                p.Database,
		logger:                  p.Logger,
		validate:                validator.New(),
		geoDbReader:             p.GeoDbReader,
		PaymentSystemConfig:     p.PaymentSystemConfig,
		pspAccountingCurrencyA3: p.PSPAccountingCurrencyA3,
	}

	renderer := &Template{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("web/template/*.html")),
	}
	api.Http.Renderer = renderer
	api.Http.Static("/", "web/static")

	api.validate.RegisterStructValidation(ProjectStructValidator, model.ProjectScalar{})
	api.validate.RegisterStructValidation(api.OrderStructValidator, model.OrderScalar{})

	api.accessRouteGroup = api.Http.Group("/api/v1/s")
	api.accessRouteGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    p.Config.SignatureSecret,
		SigningMethod: p.Config.Algorithm,
	}))
	api.accessRouteGroup.Use(api.SetMerchantIdentifierMiddleware)

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

	// init webhook endpoints section
	api.InitWebHooks()

	api.Http.GET("/docs", func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, "docs.html", map[string]interface{}{})
	})

	return api, nil
}

func (api *Api) Start() error {
	return api.Http.Start(":3001")
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

func (api *Api) InitWebHooks() {
	var headerToString = func(headers map[string][]string) string {
		var out string

		for k, v := range headers {
			out += k + ":" + v[0] + "\n "
		}

		return out
	}

	whGroup := api.Http.Group(apiWebHookGroupPath)
	whGroup.Use(api.WebHookRequestLoggerMiddleware)
	whGroup.Use(middleware.BodyDump(func(ctx echo.Context, reqBody, resBody []byte) {
		data := []interface{}{
			"request_headers", headerToString(ctx.Request().Header),
			"request_body", string(reqBody),
			"response_headers", headerToString(ctx.Response().Header()),
			"response_body", string(resBody),
		}

		api.logger.Infow(ctx.Path(), data...)

		log := &model.Log{
			RequestHeaders: headerToString(ctx.Request().Header),
			RequestBody: string(reqBody),
			ResponseHeaders: headerToString(ctx.Response().Header()),
			ResponseBody: string(resBody),
		}

		(&manager.LoggerManager{Database: api.database, Logger: api.logger}).Insert(log)
	}))

	wh := webhook.InitWebHook(
		api.database,
		api.logger,
		api.validate,
		api.geoDbReader,
		api.pspAccountingCurrencyA3,
		whGroup,
		api.webHookRawBody,
		api.PaymentSystemConfig,
	)
	wh.InitCardPayWebHookRoutes()
}