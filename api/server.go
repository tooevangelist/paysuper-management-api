package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ProtocolONE/p1pay.api/api/webhook"
	"github.com/ProtocolONE/p1pay.api/config"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/payment_system"
	"github.com/ProtocolONE/p1pay.api/utils"
	"github.com/ProtocolONE/payone-repository/pkg/constant"
	"github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/micro/go-micro"
	"github.com/oschwald/geoip2-golang"
	"github.com/sidmal/slug"
	"github.com/ttacon/libphonenumber"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"html/template"
	"io"
	"log"
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
	"Marshal": func(v interface{}) template.JS {
		a, _ := json.Marshal(v)
		return template.JS(a)
	},
}

type ServerInitParams struct {
	Config                  *config.Jwt
	Database                dao.Database
	Logger                  *zap.SugaredLogger
	GeoDbReader             *geoip2.Reader
	PaymentSystemConfig     map[string]interface{}
	PSPAccountingCurrencyA3 string
	HttpScheme              string
	CentrifugoSecret        string
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
	sort   []string
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
	paymentSystemsSettings  *payment_system.PaymentSystemSetting
	httpScheme              string
	centrifugoSecret        string

	service        micro.Service
	serviceContext context.Context
	serviceCancel  context.CancelFunc

	publisher micro.Publisher

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
		httpScheme:              p.HttpScheme,
		paymentSystemsSettings:  &payment_system.PaymentSystemSetting{Logger: p.Logger},
		centrifugoSecret:        p.CentrifugoSecret,
	}

	api.InitService()

	renderer := &Template{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("web/template/*.html")),
	}
	api.Http.Renderer = renderer
	api.Http.Static("/", "web/static")
	api.Http.Static("/spec", "spec")

	api.validate.RegisterStructValidation(ProjectStructValidator, model.ProjectScalar{})
	api.validate.RegisterStructValidation(api.OrderStructValidator, model.OrderScalar{})

	api.accessRouteGroup = api.Http.Group("/api/v1/s")
	api.accessRouteGroup.Use(middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:    p.Config.SignatureSecret,
		SigningMethod: p.Config.Algorithm,
	}))
	api.accessRouteGroup.Use(api.SetMerchantIdentifierMiddleware)

	api.Http.Use(api.LimitOffsetSortMiddleware)
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
		InitOrderV1Routes().
		InitPaymentMethodRoutes()

	// init webhook endpoints section
	api.InitWebHooks()

	api.Http.GET("/docs", func(ctx echo.Context) error {
		return ctx.Render(http.StatusOK, "docs.html", map[string]interface{}{})
	})
	api.Http.GET("/slug", func(ctx echo.Context) error {
		text := ctx.QueryParam("text")

		if text == "" {
			return ctx.NoContent(http.StatusBadRequest)
		}

		got := slug.MakeLang(text, slug.DefaultLang, model.FixedPackageSlugSeparator)

		return ctx.JSON(http.StatusOK, map[string]string{"slug": got})
	})

	return api, nil
}

func (api *Api) Start() error {
	go func() {
		if err := api.service.Run(); err != nil {
			return
		}
	}()

	go func() {
		if err := api.Http.Start(":3001"); err != nil {
			api.Http.Logger.Info("shutting down the server")
		}
	}()

	return nil
}

func (api *Api) InitService() {
	api.serviceContext, api.serviceCancel = context.WithCancel(context.Background())

	api.service = micro.NewService(
		micro.Name("go.p1.payone.api"),
		micro.Version(constant.PayOneMicroserviceVersion),
	)
	api.service.Init()
	api.publisher = micro.NewPublisher(constant.PayOneTopicNotifyPaymentName, api.service.Client())
}

func (api *Api) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := api.Http.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}

	log.Println("Http server exiting")

	api.serviceCancel()
	log.Println("Micro server exiting")
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
	whGroup := api.Http.Group(apiWebHookGroupPath)
	whGroup.Use(middleware.BodyDump(func(ctx echo.Context, reqBody, resBody []byte) {
		data := []interface{}{
			"request_headers", utils.RequestResponseHeadersToString(ctx.Request().Header),
			"request_body", string(reqBody),
			"response_headers", utils.RequestResponseHeadersToString(ctx.Response().Header()),
			"response_body", string(resBody),
		}

		api.logger.Infow(ctx.Path(), data...)
	}))

	wh := webhook.InitWebHook(
		api.database,
		api.logger,
		api.validate,
		api.geoDbReader,
		api.pspAccountingCurrencyA3,
		whGroup,
		api.PaymentSystemConfig,
		api.paymentSystemsSettings,
		api.publisher,
		api.centrifugoSecret,
	)

	whGroup.Use(wh.RawBodyMiddleware)
	wh.InitCardPayWebHookRoutes()
}
