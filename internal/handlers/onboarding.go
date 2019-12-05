package handlers

import (
	"fmt"
	"github.com/ProtocolONE/go-core/v2/pkg/config"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/micro/go-micro/client"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

const (
	merchantsPath                      = "/merchants"
	merchantsIdPath                    = "/merchants/:merchant_id"
	merchantsCompanyPath               = "/merchants/company"
	merchantsContactsPath              = "/merchants/contacts"
	merchantsBankingPath               = "/merchants/banking"
	merchantsIdCompanyPath             = "/merchants/:merchant_id/company"
	merchantsIdContactsPath            = "/merchants/:merchant_id/contacts"
	merchantsIdBankingPath             = "/merchants/:merchant_id/banking"
	merchantsStatusCompanyPath         = "/merchants/status"
	merchantsIdChangeStatusCompanyPath = "/merchants/:merchant_id/change-status"
	merchantsNotificationsPath         = "/merchants/notifications"
	merchantsIdNotificationsPath       = "/merchants/:merchant_id/notifications"
	merchantsAgreementPath             = "/merchants/agreement"
	merchantsIdAgreementPath           = "/merchants/:merchant_id/agreement"
	merchantsAgreementDocumentPath     = "/merchants/agreement/document"
	merchantsIdAgreementDocumentPath   = "/merchants/:merchant_id/agreement/document"
	merchantsNotificationsIdPath       = "/merchants/notifications/:notification_id"
	merchantsNotificationsMarkReadPath = "/merchants/notifications/:notification_id/mark-as-read"
	merchantsTariffsPath               = "/merchants/tariffs"
	merchantsIdTariffsPath             = "/merchants/:merchant_id/tariffs"
	merchantsIdManualPayoutEnablePath  = "/merchants/manual_payout/enable"
	merchantsIdManualPayoutDisablePath = "/merchants/manual_payout/disable"
	merchantsIdSetOperatingCompanyPath = "/merchants/:merchant_id/set_operating_company"
)

const (
	agreementFileMask        = "agreement_%s.pdf"
	agreementContentType     = "application/pdf"
	agreementExtension       = "pdf"
	merchantAgreementUrlMask = "%s://%s/admin/api/v1/merchants/agreement/document"
	systemAgreementUrlMask   = "%s://%s/system/api/v1/merchants/%s/agreement/document"
	agreementUploadMaxSize   = 3145728
)

type OnboardingFileMetadata struct {
	Name        string `json:"name"`
	Extension   string `json:"extension"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type OnboardingFileData struct {
	Url      string                  `json:"url"`
	Metadata *OnboardingFileMetadata `json:"metadata"`
}

type OnboardingRoute struct {
	dispatch   common.HandlerSet
	awsManager awsWrapper.AwsManagerInterface
	cfg        common.Config
	provider.LMT
}

func NewOnboardingRoute(set common.HandlerSet, initial config.Initial, awsManager awsWrapper.AwsManagerInterface, globalCfg *common.Config) *OnboardingRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "OnboardingRoute"})
	return &OnboardingRoute{
		dispatch:   set,
		LMT:        &set.AwareSet,
		cfg:        *globalCfg,
		awsManager: awsManager,
	}
}

func (h *OnboardingRoute) Route(groups *common.Groups) {
	groups.SystemUser.GET(merchantsPath, h.listMerchants)
	groups.SystemUser.GET(merchantsIdPath, h.getMerchant)

	groups.AuthUser.PUT(merchantsCompanyPath, h.setMerchantCompany)
	groups.AuthUser.PUT(merchantsContactsPath, h.setMerchantContacts)
	groups.AuthUser.PUT(merchantsBankingPath, h.setMerchantBanking)
	groups.SystemUser.PUT(merchantsIdCompanyPath, h.setMerchantCompany)
	groups.SystemUser.PUT(merchantsIdContactsPath, h.setMerchantContacts)
	groups.SystemUser.PUT(merchantsIdBankingPath, h.setMerchantBanking)
	groups.AuthUser.GET(merchantsStatusCompanyPath, h.getMerchantStatus)

	groups.SystemUser.PUT(merchantsIdChangeStatusCompanyPath, h.changeMerchantStatus)
	groups.AuthUser.PATCH(merchantsPath, h.changeAgreement)

	groups.AuthUser.GET(merchantsAgreementPath, h.getMerchantAgreementData)
	groups.SystemUser.GET(merchantsIdAgreementPath, h.getSystemAgreementData)
	groups.AuthUser.GET(merchantsAgreementDocumentPath, h.getAgreementDocument)
	groups.SystemUser.GET(merchantsIdAgreementDocumentPath, h.getAgreementDocument)

	groups.SystemUser.POST(merchantsIdNotificationsPath, h.createNotification)
	groups.SystemUser.GET(merchantsIdNotificationsPath, h.listNotifications)
	groups.AuthUser.GET(merchantsNotificationsIdPath, h.getNotification)
	groups.AuthUser.GET(merchantsNotificationsPath, h.listNotifications)
	groups.AuthUser.PUT(merchantsNotificationsMarkReadPath, h.markAsReadNotification)

	groups.AuthUser.GET(merchantsTariffsPath, h.getTariffRates)
	groups.AuthUser.POST(merchantsTariffsPath, h.setTariffRates)
	groups.SystemUser.POST(merchantsIdTariffsPath, h.setTariffRates)

	groups.AuthUser.PUT(merchantsIdManualPayoutEnablePath, h.enableMerchantManualPayout)
	groups.AuthUser.PUT(merchantsIdManualPayoutDisablePath, h.disableMerchantManualPayout)

	groups.SystemUser.POST(merchantsIdSetOperatingCompanyPath, h.setOperatingCompany)
}

func (h *OnboardingRoute) getMerchant(ctx echo.Context) error {
	req := &grpc.GetMerchantByRequest{}
	err := ctx.Bind(req)

	res, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OnboardingRoute) getMerchantByUser(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	if authUser.Id == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, common.ErrorMessageAccessDenied)
	}

	req := &grpc.GetMerchantByRequest{UserId: authUser.Id}
	res, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantBy", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OnboardingRoute) listMerchants(ctx echo.Context) error {
	req := &grpc.MerchantListingRequest{}
	err := (&common.OnboardingMerchantListingBinder{
		LimitDefault:  int64(h.cfg.LimitDefault),
		OffsetDefault: int64(h.cfg.OffsetDefault),
	}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ListMerchants(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ListMerchants", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *OnboardingRoute) changeMerchantStatus(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	req := &grpc.MerchantChangeStatusRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	req.UserId = authUser.Id
	res, err := h.dispatch.Services.Billing.ChangeMerchantStatus(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ChangeMerchantStatus", req)
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OnboardingRoute) createNotification(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	req := &grpc.NotificationRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	req.UserId = authUser.Id
	res, err := h.dispatch.Services.Billing.CreateNotification(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "CreateNotification", req)
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusCreated, res.Item)
}

func (h *OnboardingRoute) getNotification(ctx echo.Context) error {
	req := &grpc.GetNotificationRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetNotification(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetNotification")
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *OnboardingRoute) listNotifications(ctx echo.Context) error {
	req := &grpc.ListingNotificationRequest{}
	err := (&common.OnboardingNotificationsListBinder{
		LimitDefault:  int64(h.cfg.LimitDefault),
		OffsetDefault: int64(h.cfg.OffsetDefault),
	}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ListNotifications(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ListNotifications", req)
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorUnknown)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *OnboardingRoute) markAsReadNotification(ctx echo.Context) error {
	req := &grpc.GetNotificationRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.MarkNotificationAsRead(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "MarkNotificationAsRead")
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *OnboardingRoute) changeAgreement(ctx echo.Context) error {
	req := &grpc.ChangeMerchantDataRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.ChangeMerchantData(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ChangeMerchantData")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OnboardingRoute) getAgreementDocument(ctx echo.Context) error {
	req := &grpc.GetMerchantByRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	res, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetMerchantBy")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	if res.Item.S3AgreementName == "" {
		return echo.NewHTTPError(http.StatusNotFound, common.ErrorMessageAgreementNotGenerated)
	}

	filePath := os.TempDir() + string(os.PathSeparator) + res.Item.S3AgreementName
	_, err = h.awsManager.Download(ctx.Request().Context(), filePath, &awsWrapper.DownloadInput{FileName: res.Item.S3AgreementName})

	if err != nil {
		h.L().Error("AWS api call to download file failed", logger.PairArgs("err", err.Error(), "file_name", res.Item.S3AgreementName))

		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorAgreementFileNotExist)
	}

	return ctx.File(filePath)
}

func (h *OnboardingRoute) getAgreementStructure(
	ctx echo.Context,
	merchantId, ext, ct, fPath string,
	signerType int32,
) (interface{}, error) {
	file, err := os.Open(fPath)

	if err != nil {
		return nil, common.ErrorMessageAgreementNotFound
	}

	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	fi, err := file.Stat()

	if err != nil {
		return nil, common.ErrorMessageAgreementNotFound
	}

	url := fmt.Sprintf(systemAgreementUrlMask, h.cfg.HttpScheme, ctx.Request().Host, merchantId)

	if signerType == pkg.SignerTypeMerchant {
		url = fmt.Sprintf(merchantAgreementUrlMask, h.cfg.HttpScheme, ctx.Request().Host)
	}

	data := &OnboardingFileData{
		Url: url,
		Metadata: &OnboardingFileMetadata{
			Name:        fi.Name(),
			Extension:   ext,
			ContentType: ct,
			Size:        fi.Size(),
		},
	}

	return data, nil
}

func (h *OnboardingRoute) validateUpload(file *multipart.FileHeader) (multipart.File, error) {
	if file.Size > agreementUploadMaxSize {
		return nil, common.ErrorMessageAgreementUploadMaxSize
	}

	src, err := file.Open()

	if err != nil {
		h.L().Error("validate upload error", logger.PairArgs("err", err.Error()))
		return nil, common.ErrorUnknown
	}

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)

	if err != nil {
		h.L().Error("validate upload error", logger.PairArgs("err", err.Error()))
		return nil, common.ErrorUnknown
	}

	_, err = src.Seek(0, 0)

	if err != nil {
		h.L().Error("validate upload error", logger.PairArgs("err", err.Error()))
		return nil, common.ErrorUnknown
	}

	ct := http.DetectContentType(buffer)

	if ct != agreementContentType {
		return nil, common.ErrorMessageAgreementContentType
	}

	return src, nil
}

func (h *OnboardingRoute) setMerchantCompany(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	in := &billing.MerchantCompanyInfo{}
	err := ctx.Bind(in)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req := &grpc.OnboardingRequest{
		Company: in,
		Id:      ctx.Param(common.RequestParameterMerchantId),
		User: &billing.MerchantUser{
			Id:    authUser.Id,
			Email: authUser.Email,
		},
	}
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ChangeMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ChangeMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OnboardingRoute) setMerchantContacts(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	in := &billing.MerchantContact{}
	err := ctx.Bind(in)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req := &grpc.OnboardingRequest{
		Contacts: in,
		Id:       ctx.Param(common.RequestParameterMerchantId),
		User: &billing.MerchantUser{
			Id:    authUser.Id,
			Email: authUser.Email,
		},
	}
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ChangeMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ChangeMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OnboardingRoute) setMerchantBanking(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	in := &billing.MerchantBanking{}
	err := ctx.Bind(in)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req := &grpc.OnboardingRequest{
		Banking: in,
		Id:      ctx.Param(common.RequestParameterMerchantId),
		User: &billing.MerchantUser{
			Id:    authUser.Id,
			Email: authUser.Email,
		},
	}
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.ChangeMerchant(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "ChangeMerchant", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OnboardingRoute) getMerchantStatus(ctx echo.Context) error {
	req := &grpc.SetMerchantS3AgreementRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetMerchantOnboardingCompleteData(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetMerchantOnboardingCompleteData")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OnboardingRoute) getTariffRates(ctx echo.Context) error {
	req := &grpc.GetMerchantTariffRatesRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetMerchantTariffRates(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "GetMerchantTariffRates", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Items)
}

func (h *OnboardingRoute) setTariffRates(ctx echo.Context) error {
	req := &grpc.SetMerchantTariffRatesRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.SetMerchantTariffRates(
		ctx.Request().Context(),
		req,
		client.WithRequestTimeout(time.Minute*10),
	)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "SetMerchantTariffRates", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
}

func (h *OnboardingRoute) getMerchantAgreementData(ctx echo.Context) error {
	return h.getAgreementData(ctx, pkg.SignerTypeMerchant)
}

func (h *OnboardingRoute) getSystemAgreementData(ctx echo.Context) error {
	return h.getAgreementData(ctx, pkg.SignerTypePs)
}

func (h *OnboardingRoute) getAgreementData(ctx echo.Context, signerType int32) error {
	req := &grpc.GetMerchantByRequest{}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.GetMerchantBy(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "GetMerchantBy")
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	if res.Item.S3AgreementName == "" {
		return echo.NewHTTPError(http.StatusNotFound, common.ErrorMessageAgreementNotFound)
	}

	filePath := os.TempDir() + string(os.PathSeparator) + res.Item.S3AgreementName
	_, err = h.awsManager.Download(ctx.Request().Context(), filePath, &awsWrapper.DownloadInput{FileName: res.Item.S3AgreementName})

	if err != nil {
		h.L().Error("AWS api call to download file failed", logger.PairArgs("err", err.Error(), "file_name", res.Item.S3AgreementName))

		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	fData, err := h.getAgreementStructure(ctx, req.MerchantId, agreementExtension, agreementContentType, filePath, signerType)

	if err != nil {
		h.L().Error("Get agreement structure failed", logger.PairArgs("err", err.Error(), "merchant_id", req.MerchantId))

		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	return ctx.JSON(http.StatusOK, fData)
}

func (h *OnboardingRoute) enableMerchantManualPayout(ctx echo.Context) error {
	return h.changeMerchantManualPayout(ctx, true)
}

func (h *OnboardingRoute) disableMerchantManualPayout(ctx echo.Context) error {
	return h.changeMerchantManualPayout(ctx, false)
}

func (h *OnboardingRoute) changeMerchantManualPayout(ctx echo.Context, enableManualPayout bool) error {
	req := &grpc.ChangeMerchantManualPayoutsRequest{ManualPayoutsEnabled: enableManualPayout}

	if err := h.dispatch.BindAndValidate(req, ctx); err != nil {
		return err
	}

	res, err := h.dispatch.Services.Billing.ChangeMerchantManualPayouts(ctx.Request().Context(), req)

	if err != nil {
		return h.dispatch.SrvCallHandler(req, err, pkg.ServiceName, "ChangeMerchantManualPayouts")
	}

	if res.Status != http.StatusOK {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}

func (h *OnboardingRoute) setOperatingCompany(ctx echo.Context) error {
	req := &grpc.SetMerchantOperatingCompanyRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.MerchantId = ctx.Param(common.RequestParameterMerchantId)
	err = h.dispatch.Validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.SetMerchantOperatingCompany(ctx.Request().Context(), req)

	if err != nil {
		common.LogSrvCallFailedGRPC(h.L(), err, pkg.ServiceName, "SetMerchantOperatingCompany", req)
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorUnknown)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Item)
}
