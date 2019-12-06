package common

import (
	"encoding/json"
	geoService "github.com/ProtocolONE/geoip-service/pkg/proto"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	billingService "github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	recurringService "github.com/paysuper/paysuper-recurring-repository/pkg/proto/repository"
	reporterPkg "github.com/paysuper/paysuper-reporter/pkg"
	reporterProto "github.com/paysuper/paysuper-reporter/pkg/proto"
	reporterService "github.com/paysuper/paysuper-reporter/pkg/proto"
	taxService "github.com/paysuper/paysuper-tax-service/proto"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

const (
	Prefix                   = "internal.dispatcher"
	UnmarshalKey             = "dispatcher"
	UnmarshalGlobalConfigKey = "dispatcher.global"
	AuthProjectGroupPath     = "/auth/api/v1"
	AuthUserGroupPath        = "/admin/api/v1"
	SystemUserGroupPath      = "/system/api/v1"
	NoAuthGroupPath          = "/api/v1"
	WebHookGroupPath         = "/webhook"
)

// Cursor
type Cursor struct {
	Limit, Offset int32
	Sort          []string
}

// ExtractUserContext
func ExtractUserContext(ctx echo.Context) *AuthUser {
	if user, ok := ctx.Get("user").(*AuthUser); ok {
		return user
	}
	return &AuthUser{}
}

// ExtractRawBodyContext
func ExtractRawBodyContext(ctx echo.Context) []byte {
	if rawBody, ok := ctx.Get("rawBody").([]byte); ok {
		return rawBody
	}
	return nil
}

// ExtractCursorContext
func ExtractCursorContext(ctx echo.Context) *Cursor {
	if cursor, ok := ctx.Get("cursor").(*Cursor); ok {
		return cursor
	}
	return &Cursor{}
}

// ExtractBinderContext
func ExtractBinderContext(ctx echo.Context) echo.Binder {
	if binder, ok := ctx.Get("binder").(echo.Binder); ok {
		return binder
	}
	return nil
}

// SetUserContext
func SetUserContext(ctx echo.Context, user *AuthUser) {
	ctx.Set("user", user)
}

// SetRawBodyContext
func SetRawBodyContext(ctx echo.Context, rawBody []byte) {
	ctx.Set("rawBody", rawBody)
}

// SetCursorContext
func SetCursorContext(ctx echo.Context, cursor *Cursor) {
	ctx.Set("cursor", cursor)
}

// SetBinder
func SetBinder(ctx echo.Context, binder echo.Binder) {
	ctx.Set("binder", binder)
}

// Groups
type Groups struct {
	AuthProject *echo.Group
	Access      *echo.Group
	AuthUser    *echo.Group
	WebHooks    *echo.Group
	Common      *echo.Group
	SystemUser  *echo.Group
}

// Handler
type Handler interface {
	Route(groups *Groups)
}

// Validate
type Validator interface {
	Use(validator *validator.Validate)
}

// Services
type Services struct {
	Repository recurringService.RepositoryService
	Geo        geoService.GeoIpService
	Billing    billingService.BillingService
	Tax        taxService.TaxService
	Reporter   reporterService.ReporterService
}

// Handlers
type Handlers []Handler

// HandlerSet
type HandlerSet struct {
	Services Services
	Validate *validator.Validate
	AwareSet provider.AwareSet
}

// BindAndValidate
func (h HandlerSet) BindAndValidate(req interface{}, ctx echo.Context) *echo.HTTPError {
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, ErrorRequestParamsIncorrect)
	}
	if err := h.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, GetValidationError(err))
	}
	return nil
}

// SrvCallHandler returns error if present, otherwise response as JSON with 200 OK
func (h HandlerSet) SrvCallHandler(req interface{}, err error, name, method string) *echo.HTTPError {
	h.AwareSet.L().Error(pkg.ErrorGrpcServiceCallFailed,
		logger.PairArgs(
			ErrorFieldService, name,
			ErrorFieldMethod, method,
		),
		logger.WithPrettyFields(logger.Fields{"err": err, ErrorFieldRequest: req}),
	)
	return echo.NewHTTPError(http.StatusInternalServerError, ErrorInternal)
}

// AuthUser
type AuthUser struct {
	Id         string
	Name       string
	Email      string
	Role       string
	MerchantId string
}

type ReportFileRequest struct {
	MerchantId string                 `json:"merchant_id" validate:"required"`
	FileType   string                 `json:"file_type" validate:"required"`
	ReportType string                 `json:"report_type"`
	Template   string                 `json:"template"`
	Params     map[string]interface{} `json:"params"`
}

func (h *HandlerSet) RequestReportFile(ctx echo.Context, data *ReportFileRequest) error {
	params, err := json.Marshal(data.Params)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, ErrorRequestDataInvalid)
	}

	req := &reporterProto.ReportFile{
		UserId:           ExtractUserContext(ctx).Id,
		MerchantId:       data.MerchantId,
		ReportType:       data.ReportType,
		FileType:         data.FileType,
		Template:         data.Template,
		Params:           params,
		SendNotification: true,
	}

	if err = h.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, GetValidationError(err))
	}

	res, err := h.Services.Reporter.CreateFile(ctx.Request().Context(), req)

	if err != nil {
		return h.SrvCallHandler(req, err, reporterPkg.ServiceName, "CreateFile")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}
