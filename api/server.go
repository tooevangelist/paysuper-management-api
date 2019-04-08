package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	jwtMiddleware "github.com/ProtocolONE/authone-jwt-verifier-golang/middleware/echo"
	"github.com/ProtocolONE/geoip-service/pkg"
	"github.com/ProtocolONE/geoip-service/pkg/proto"
	"github.com/ProtocolONE/rabbitmq/pkg"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/micro/go-micro"
	k8s "github.com/micro/kubernetes/go/micro"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/config"
	"github.com/paysuper/paysuper-management-api/database/dao"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/utils"
	paylinkServiceConst "github.com/paysuper/paysuper-payment-link/pkg"
	"github.com/paysuper/paysuper-payment-link/proto"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/paysuper/paysuper-recurring-repository/pkg/proto/repository"
	taxServiceConst "github.com/paysuper/paysuper-tax-service/pkg"
	"github.com/paysuper/paysuper-tax-service/proto"
	"github.com/sidmal/slug"
	"github.com/ttacon/libphonenumber"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"
)

const (
	apiWebHookGroupPath  = "/webhook"
	apiAuthUserGroupPath = "/admin/api/v1"

	LimitDefault  = 100
	OffsetDefault = 0

	requestParameterId                   = "id"
	requestParameterName                 = "name"
	requestParameterSku                  = "sku"
	requestParameterIsSigned             = "is_signed"
	requestParameterLastPayoutDateFrom   = "last_payout_date_from"
	requestParameterLastPayoutDateTo     = "last_payout_date_to"
	requestParameterLastPayoutAmount     = "last_payout_amount"
	requestParameterMerchantId           = "merchant_id"
	requestParameterProjectId            = "project_id"
	requestParameterPaymentMethodId      = "method_id"
	requestParameterOrderId              = "order_id"
	requestParameterRefundId             = "refund_id"
	requestParameterNotificationId       = "notification_id"
	requestParameterPaymentMethodName    = "method_name"
	requestParameterUserId               = "user"
	requestParameterSort                 = "sort[]"
	requestParameterLimit                = "limit"
	requestParameterOffset               = "offset"
	requestParameterQuickSearch          = "quick_search"
	requestParameterFile                 = "file"
	requestParameterUtmSource            = "utm_source"
	requestParameterUtmMedium            = "utm_medium"
	requestParameterUtmCampagin          = "utm_campagin"
	requestParameterIsSystem             = "is_system"
	requestParameterAgreementType        = "agreement_type"
	requestParameterHasMerchantSignature = "has_merchant_signature"
	requestParameterHasPspSignature      = "has_psp_signature"
	requestParameterAgreementSentViaMail = "agreement_sent_via_mail"
	requestParameterMailTrackingLink     = "mail_tracking_link"
	requestAuthorizationTokenRegex       = "Bearer ([A-z0-9_.-]{10,})"

	errorIdIsEmpty                                = "identifier can't be empty"
	errorIncorrectMerchantId                      = "incorrect merchant identifier"
	errorIncorrectNotificationId                  = "incorrect notification identifier"
	errorIncorrectOrderId                         = "incorrect order identifier"
	errorIncorrectPaymentMethodId                 = "incorrect payment method identifier"
	errorIncorrectProductId                       = "incorrect product identifier"
	errorIncorrectRefundId                        = "incorrect refund identifier"
	errorIncorrectPaylinkId                       = "incorrect paylink identifier"
	errorIncorrectUserId                          = "incorrect user identifier"
	errorMessageAccessDenied                      = "access denied"
	errorMessageAuthorizationHeaderNotFound       = "authorization header not found"
	errorMessageAuthorizationTokenNotFound        = "authorization token not found"
	errorMessageAuthorizedUserNotFound            = "information about authorized user not found"
	errorMessageMask                              = "field validation for '%s' failed on the '%s' tag"
	errorQueryParamsIncorrect                     = "incorrect query parameters"
	errorUnknown                                  = "unknown error. try request later"
	errorMessageAgreementNotGenerated             = "agreement for merchant not generated early"
	errorMessageAgreementNotFound                 = "agreement for merchant not found"
	errorMessageAgreementUploadMaxSize            = "agreement document max upload size can't be greater than %d"
	errorMessageAgreementContentType              = "agreement document type must be a pdf"
	errorMessageAgreementCanNotBeGenerate         = "agreement can't be generated for not checked merchant data"
	errorMessageAgreementTypeIncorrectType        = "agreement type parameter have incorrect type"
	errorMessageHasMerchantSignatureIncorrectType = "merchant signature parameter has incorrect type"
	errorMessageHasPspSignatureIncorrectType      = "paysuper signature parameter has incorrect type"
	errorMessageAgreementSentViaMailIncorrectType = "agreement sent via email parameter has incorrect type"
	errorMessageMailTrackingLinkIncorrectType     = "mail tracking link parameter has incorrect type"

	HeaderAcceptLanguage = "Accept-Language"

	agreementPageTemplateName = "agreement.html"
)

var funcMap = template.FuncMap{
	"Marshal": func(v interface{}) template.JS {
		a, _ := json.Marshal(v)
		return template.JS(a)
	},
}

type Template struct {
	templates *template.Template
}

type ServerInitParams struct {
	Config      *config.Config
	Database    dao.Database
	Logger      *zap.SugaredLogger
	HttpScheme  string
	K8sHost     string
	AmqpAddress string
	Auth1       *config.Auth1
}

type Merchant struct {
	Identifier string
}

type GetParams struct {
	limit  int32
	offset int32
	sort   []string
}

type Order struct {
	PayerPhone *libphonenumber.PhoneNumber
}

type AuthUser struct {
	Id        string
	Name      string
	Email     string
	Roles     map[string]bool
	Merchants map[string]bool
}

type Api struct {
	Http     *echo.Echo
	config   *config.Config
	database dao.Database
	logger   *zap.SugaredLogger
	validate *validator.Validate

	accessRouteGroup  *echo.Group
	webhookRouteGroup *echo.Group
	jwtVerifier       *jwtverifier.JwtVerifier

	authUserRouteGroup *echo.Group
	authUser           *AuthUser

	httpScheme string

	service        micro.Service
	serviceContext context.Context
	serviceCancel  context.CancelFunc

	repository     repository.RepositoryService
	geoService     proto.GeoIpService
	billingService grpc.BillingService
	taxService     tax_service.TaxService
	paylinkService paylink.PaylinkService

	AmqpAddress string
	notifierPub *rabbitmq.Broker

	k8sHost string
	rawBody string

	Merchant
	GetParams
	Order
}

func NewServer(p *ServerInitParams) (*Api, error) {
	api := &Api{
		Http:        echo.New(),
		database:    p.Database,
		logger:      p.Logger,
		validate:    validator.New(),
		httpScheme:  p.HttpScheme,
		k8sHost:     p.K8sHost,
		AmqpAddress: p.AmqpAddress,
		config:      p.Config,
	}
	api.InitService()

	jwtVerifierSettings := jwtverifier.Config{
		ClientID:     p.Auth1.ClientId,
		ClientSecret: p.Auth1.ClientSecret,
		Scopes:       []string{"openid", "offline"},
		RedirectURL:  p.Auth1.RedirectUrl,
		Issuer:       p.Auth1.Issuer,
	}
	api.jwtVerifier = jwtverifier.NewJwtVerifier(jwtVerifierSettings)

	renderer := &Template{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("web/template/*.html")),
	}
	api.Http.Renderer = renderer

	api.Http.Static("/", "web/static")
	api.Http.Static("/spec", "spec")

	api.validate.RegisterStructValidation(ProjectStructValidator, model.ProjectScalar{})
	api.validate.RegisterStructValidation(api.OrderStructValidator, model.OrderScalar{})
	err := api.validate.RegisterValidation("phone", api.PhoneValidator)

	if err != nil {
		return nil, err
	}

	err = api.validate.RegisterValidation("uuid", api.UuidValidator)

	if err != nil {
		return nil, err
	}

	api.accessRouteGroup = api.Http.Group("/api/v1/s")

	api.accessRouteGroup.Use(
		jwtMiddleware.AuthOneJwtCallableWithConfig(
			api.jwtVerifier,
			func(ui *jwtverifier.UserInfo) {
				api.Merchant.Identifier = string(ui.UserID)
				// TODO: Remove this line after merchant registration is completed.
				api.Merchant.Identifier = "5be2c3022b9bb6000765d132"
			},
		),
	)

	api.authUserRouteGroup = api.Http.Group(apiAuthUserGroupPath)
	api.authUserRouteGroup.Use(
		jwtMiddleware.AuthOneJwtCallableWithConfig(
			api.jwtVerifier,
			func(ui *jwtverifier.UserInfo) {
				api.authUser = &AuthUser{
					Id:        ui.UserID,
					Name:      "System User",
					Merchants: make(map[string]bool),
					Roles:     make(map[string]bool),
				}
			},
		),
	)
	api.authUserRouteGroup.Use(api.getUserDetailsMiddleware)

	api.webhookRouteGroup = api.Http.Group(apiWebHookGroupPath)
	api.webhookRouteGroup.Use(middleware.BodyDump(func(ctx echo.Context, reqBody, resBody []byte) {
		data := []interface{}{
			"request_headers", utils.RequestResponseHeadersToString(ctx.Request().Header),
			"request_body", string(reqBody),
			"response_headers", utils.RequestResponseHeadersToString(ctx.Response().Header()),
			"response_body", string(resBody),
		}

		api.logger.Infow(ctx.Path(), data...)
	}))
	api.webhookRouteGroup.Use(api.RawBodyMiddleware)

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
		InitProductRoutes().
		InitProjectRoutes().
		InitOrderV1Routes().
		InitPaylinkRoutes().
		InitPaymentMethodRoutes().
		InitCardPayWebHookRoutes().
		initTaxesRoutes()

	_, err = api.initOnboardingRoutes()

	if err != nil {
		return nil, err
	}

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

	options := []micro.Option{
		micro.Name("p1payapi"),
		micro.Version(constant.PayOneMicroserviceVersion),
	}

	if api.k8sHost == "" {
		api.service = micro.NewService(options...)
		log.Println("Initialize micro service")
	} else {
		api.service = k8s.NewService(options...)
		log.Println("Initialize k8s service")
	}

	api.service.Init()

	api.repository = repository.NewRepositoryService(constant.PayOneRepositoryServiceName, api.service.Client())
	api.geoService = proto.NewGeoIpService(geoip.ServiceName, api.service.Client())
	api.billingService = grpc.NewBillingService(pkg.ServiceName, api.service.Client())
	api.taxService = tax_service.NewTaxService(taxServiceConst.ServiceName, api.service.Client())
	api.paylinkService = paylink.NewPaylinkService(paylinkServiceConst.ServiceName, api.service.Client())
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

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func (api *Api) SetMerchantIdentifierMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		//user := c.Get("user").(*jwt.Token)
		//claims := user.Claims.(jwt.MapClaims)

		//id, ok := claims["id"]

		//if !ok {
		//	c.Error(errors.New("merchant identifier not found"))
		//}

		api.Merchant.Identifier = "5be2c3022b9bb6000765d132"

		return next(c)
	}
}

func (api *Api) getUserDetailsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ctx echo.Context) error {
		auth := ctx.Request().Header.Get(echo.HeaderAuthorization)

		if auth == "" {
			return errors.New(errorMessageAuthorizationHeaderNotFound)
		}

		r := regexp.MustCompile(requestAuthorizationTokenRegex)
		match := r.FindStringSubmatch(auth)

		if len(match) < 1 {
			return errors.New(errorMessageAuthorizationTokenNotFound)
		}

		u, err := api.jwtVerifier.GetUserInfo(ctx.Request().Context(), match[1])

		if err != nil {
			return errors.New(errorMessageAuthorizedUserNotFound)
		}

		api.authUser.Email = u.Email

		return next(ctx)
	}
}

func (api *Api) getValidationError(err error) string {
	vErr := err.(validator.ValidationErrors)[0]

	return fmt.Sprintf(errorMessageMask, vErr.Field(), vErr.Tag())
}

func (api *Api) onboardingBeforeHandler(st interface{}, ctx echo.Context) *echo.HTTPError {
	err := ctx.Bind(st)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorQueryParamsIncorrect)
	}

	err = api.validate.Struct(st)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, api.getValidationError(err))
	}

	return nil
}

func (api *Api) logError(msg string, data []interface{}) {
	zap.S().Errorw(fmt.Sprintf("[PAYSUPER_MANAGEMENT_API] %s", msg), data...)
}
