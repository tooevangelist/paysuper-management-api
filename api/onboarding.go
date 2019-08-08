package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

const (
	agreementFileMask      = "agreement_%s.pdf"
	agreementContentType   = "application/pdf"
	agreementExtension     = "pdf"
	agreementUrlMask       = "%s://%s/admin/api/v1/merchants/%s/agreement/document"
	agreementUploadMaxSize = 3145728
)

type onboardingRoute struct {
	*Api
	mClt *minio.Client
}

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

func (api *Api) initOnboardingRoutes() (*Api, error) {
	route := &onboardingRoute{Api: api}

	mClt, err := minio.New(
		api.config.S3.Endpoint,
		api.config.S3.AccessKeyId,
		api.config.S3.SecretKey,
		api.config.S3.Secure,
	)

	if err != nil {
		return nil, err
	}

	err = mClt.MakeBucket(api.config.S3.BucketName, api.config.S3.Region)

	if err != nil {
		return nil, err
	}

	route.mClt = mClt

	api.authUserRouteGroup.GET("/merchants", route.listMerchants)
	api.authUserRouteGroup.GET("/merchants/:id", route.getMerchant)
	api.authUserRouteGroup.GET("/merchants/user", route.getMerchantByUser)
	api.authUserRouteGroup.POST("/merchants", route.changeMerchant)
	api.authUserRouteGroup.PUT("/merchants", route.changeMerchant)

	api.authUserRouteGroup.PUT("/merchants/company", route.setMerchantCompany)
	api.authUserRouteGroup.PUT("/merchants/contacts", route.setMerchantContacts)
	api.authUserRouteGroup.PUT("/merchants/banking", route.setMerchantBanking)
	api.authUserRouteGroup.PUT("/merchants/tariff", route.setMerchantTariff)
	api.authUserRouteGroup.PUT("/merchants/:id/company", route.setMerchantCompany)
	api.authUserRouteGroup.PUT("/merchants/:id/contacts", route.setMerchantContacts)
	api.authUserRouteGroup.PUT("/merchants/:id/banking", route.setMerchantBanking)
	api.authUserRouteGroup.PUT("/merchants/:id/tariff", route.setMerchantTariff)
	api.authUserRouteGroup.GET("/merchants/:id/status", route.getMerchantStatus)

	api.authUserRouteGroup.PUT("/merchants/:id/change-status", route.changeMerchantStatus)
	api.authUserRouteGroup.PATCH("/merchants/:id", route.changeAgreement)

	api.authUserRouteGroup.GET("/merchants/:id/agreement", route.generateAgreement)
	api.authUserRouteGroup.GET("/merchants/:id/agreement/document", route.getAgreementDocument)
	api.authUserRouteGroup.POST("/merchants/:id/agreement/document", route.uploadAgreementDocument)
	api.authUserRouteGroup.PUT("/merchants/:id/agreement/signature", route.createAgreementSignature)

	api.authUserRouteGroup.POST("/merchants/:merchant_id/notifications", route.createNotification)
	api.authUserRouteGroup.GET("/merchants/:merchant_id/notifications/:notification_id", route.getNotification)
	api.authUserRouteGroup.GET("/merchants/:merchant_id/notifications", route.listNotifications)
	api.authUserRouteGroup.PUT("/merchants/:merchant_id/notifications/:notification_id/mark-as-read", route.markAsReadNotification)

	api.authUserRouteGroup.GET("/merchants/:merchant_id/methods/:method_id", route.getPaymentMethod)
	api.authUserRouteGroup.GET("/merchants/:merchant_id/methods", route.listPaymentMethods)
	api.authUserRouteGroup.PUT("/merchants/:merchant_id/methods/:method_id", route.changePaymentMethod)

	return api, nil
}

func (r *onboardingRoute) getMerchant(ctx echo.Context) error {
	id := ctx.Param(requestParameterId)

	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorIdIsEmpty)
	}

	req := &grpc.GetMerchantByRequest{MerchantId: id}
	rsp, err := r.billingService.GetMerchantBy(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("Call billing-server method GetMerchantBy failed", "error", err.Error(), "request", req)
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) getMerchantByUser(ctx echo.Context) error {
	if r.authUser.Id == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, errorMessageAccessDenied)
	}

	rsp, err := r.billingService.GetMerchantBy(ctx.Request().Context(), &grpc.GetMerchantByRequest{UserId: r.authUser.Id})

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) listMerchants(ctx echo.Context) error {
	req := &grpc.MerchantListingRequest{}
	err := (&OnboardingMerchantListingBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ListMerchants(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) changeMerchant(ctx echo.Context) error {
	req := &grpc.OnboardingRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.User = &billing.MerchantUser{
		Id:    r.authUser.Id,
		Email: r.authUser.Email,
	}
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ChangeMerchant(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	if rsp.Status != http.StatusOK {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) changeMerchantStatus(ctx echo.Context) error {
	req := &grpc.MerchantChangeStatusRequest{}
	err := (&OnboardingChangeMerchantStatusBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	req.UserId = r.authUser.Id
	rsp, err := r.billingService.ChangeMerchantStatus(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	if rsp.Status != http.StatusOK {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) createNotification(ctx echo.Context) error {
	req := &grpc.NotificationRequest{}
	err := (&OnboardingCreateNotificationBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	req.UserId = r.authUser.Id
	rsp, err := r.billingService.CreateNotification(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	if rsp.Status != http.StatusOK {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusCreated, rsp.Item)
}

func (r *onboardingRoute) getNotification(ctx echo.Context) error {
	merchantId := ctx.Param(requestParameterMerchantId)
	notificationId := ctx.Param(requestParameterNotificationId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectMerchantId)
	}

	if notificationId == "" || bson.IsObjectIdHex(notificationId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectNotificationId)
	}

	req := &grpc.GetNotificationRequest{
		MerchantId:     merchantId,
		NotificationId: notificationId,
	}
	rsp, err := r.billingService.GetNotification(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, errorNotificationNotFound)
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) listNotifications(ctx echo.Context) error {
	req := &grpc.ListingNotificationRequest{}
	err := (&OnboardingNotificationsListBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ListNotifications(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) markAsReadNotification(ctx echo.Context) error {
	merchantId := ctx.Param(requestParameterMerchantId)
	notificationId := ctx.Param(requestParameterNotificationId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectMerchantId)
	}

	if notificationId == "" || bson.IsObjectIdHex(notificationId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorIncorrectNotificationId)
	}

	req := &grpc.GetNotificationRequest{
		MerchantId:     merchantId,
		NotificationId: notificationId,
	}
	rsp, err := r.billingService.MarkNotificationAsRead(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp)
}

func (r *onboardingRoute) getPaymentMethod(ctx echo.Context) error {
	req := &grpc.GetMerchantPaymentMethodRequest{}
	err := (&OnboardingGetPaymentMethodBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	rsp, err := r.billingService.GetMerchantPaymentMethod(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) listPaymentMethods(ctx echo.Context) error {
	req := &grpc.ListMerchantPaymentMethodsRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.MerchantId = ctx.Param(requestParameterMerchantId)
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ListMerchantPaymentMethods(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	return ctx.JSON(http.StatusOK, rsp.PaymentMethods)
}

func (r *onboardingRoute) changePaymentMethod(ctx echo.Context) error {
	req := &grpc.MerchantPaymentMethodRequest{}
	err := (&OnboardingChangePaymentMethodBinder{}).Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.UserId = r.authUser.Id

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ChangeMerchantPaymentMethod(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) changeAgreement(ctx echo.Context) error {
	req := &grpc.ChangeMerchantDataRequest{}
	binder := &ChangeMerchantDataRequestBinder{Api: r.Api}
	err := binder.Bind(req, ctx)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ChangeMerchantData(ctx.Request().Context(), req)

	if err != nil {
		zap.S().Errorf(`Call billing server method "ChangeMerchantData" failed`,
			"error", err.Error(), "request", req)

		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

func (r *onboardingRoute) generateAgreement(ctx echo.Context) error {
	merchantId := ctx.Param(requestParameterId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req := &grpc.GetMerchantByRequest{MerchantId: merchantId}
	rsp, err := r.billingService.GetMerchantBy(context.Background(), req)

	if err != nil {
		zap.S().Errorf(`Call billing server method "GetMerchantBy" failed`,
			"error", err.Error(), "request", req)
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	if rsp.Item.S3AgreementName != "" {
		filePath := os.TempDir() + string(os.PathSeparator) + rsp.Item.S3AgreementName
		err = r.mClt.FGetObject(r.config.S3.BucketName, rsp.Item.S3AgreementName, filePath, minio.GetObjectOptions{})

		if err != nil {
			zap.S().Errorf("internal error", "err", err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
		}

		fData, err := r.getAgreementStructure(ctx, merchantId, agreementExtension, agreementContentType, filePath)

		if err != nil {
			zap.S().Errorf("internal error", "err", err.Error())
			return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
		}

		return ctx.JSON(http.StatusOK, fData)
	}

	if rsp.Item.CanGenerateAgreement() == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorMessageAgreementCanNotBeGenerate)
	}

	buf := new(bytes.Buffer)
	data := map[string]interface{}{"Merchant": rsp.Item}

	err = ctx.Echo().Renderer.Render(buf, agreementPageTemplateName, data, ctx)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	pdf, err := wkhtmltopdf.NewPDFGenerator()

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	pdf.AddPage(wkhtmltopdf.NewPageReader(strings.NewReader(buf.String())))
	err = pdf.Create()

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	agrName := fmt.Sprintf(agreementFileMask, rsp.Item.Id)
	filePath := os.TempDir() + string(os.PathSeparator) + agrName

	err = pdf.WriteFile(filePath)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	_, err = r.mClt.FPutObject(r.config.S3.BucketName, agrName, filePath, minio.PutObjectOptions{ContentType: agreementContentType})

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	req1 := &grpc.SetMerchantS3AgreementRequest{MerchantId: merchantId, S3AgreementName: agrName}
	_, err = r.billingService.SetMerchantS3Agreement(context.Background(), req1)

	if err != nil {
		zap.S().Errorf(`Call billing server method "SetMerchantS3Agreement" failed`,
			"error", err.Error(), "request", req1)
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	fData, err := r.getAgreementStructure(ctx, merchantId, agreementExtension, agreementContentType, filePath)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	return ctx.JSON(http.StatusOK, fData)
}

func (r *onboardingRoute) getAgreementDocument(ctx echo.Context) error {
	merchantId := ctx.Param(requestParameterId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req := &grpc.GetMerchantByRequest{MerchantId: merchantId}
	rsp, err := r.billingService.GetMerchantBy(context.Background(), req)

	if err != nil {
		zap.S().Errorf(`Call billing server method "GetMerchantBy" failed`,
			"error", err.Error(), "request", req)
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	if rsp.Item.S3AgreementName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errorMessageAgreementNotGenerated)
	}

	filePath := os.TempDir() + string(os.PathSeparator) + rsp.Item.S3AgreementName
	err = r.mClt.FGetObject(r.config.S3.BucketName, rsp.Item.S3AgreementName, filePath, minio.GetObjectOptions{})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorAgreementFileNotExist)
	}

	return ctx.File(filePath)
}

func (r *onboardingRoute) uploadAgreementDocument(ctx echo.Context) error {
	merchantId := ctx.Param(requestParameterId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req := &grpc.GetMerchantByRequest{MerchantId: merchantId}
	rsp, err := r.billingService.GetMerchantBy(context.Background(), req)

	if err != nil {
		zap.S().Errorf(`Call billing server method "GetMerchantBy" failed`,
			"error", err.Error(), "request", req)
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	file, err := ctx.FormFile(requestParameterFile)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorNotMultipartForm)
	}

	src, err := r.validateUpload(file)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	defer func() {
		if err := src.Close(); err != nil {
			return
		}
	}()

	agrName := fmt.Sprintf(agreementFileMask, rsp.Item.Id)
	filePath := os.TempDir() + string(os.PathSeparator) + agrName
	dst, err := os.Create(filePath)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	defer func() {
		if err := dst.Close(); err != nil {
			return
		}
	}()

	_, err = io.Copy(dst, src)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	_, err = r.mClt.FPutObject(r.config.S3.BucketName, agrName, filePath, minio.PutObjectOptions{ContentType: agreementContentType})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUploadFailed)
	}

	req1 := &grpc.SetMerchantS3AgreementRequest{MerchantId: merchantId, S3AgreementName: agrName}
	_, err = r.billingService.SetMerchantS3Agreement(context.Background(), req1)

	if err != nil {
		zap.S().Errorf(`Call billing server method "SetMerchantS3Agreement" failed`,
			"error", err.Error(), "request", req1)
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	fData, err := r.getAgreementStructure(ctx, merchantId, agreementExtension, agreementContentType, filePath)

	if err != nil {
		zap.S().Errorf("internal error", "err", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, errorInternal)
	}

	return ctx.JSON(http.StatusOK, fData)
}

func (r *onboardingRoute) getAgreementStructure(
	ctx echo.Context,
	merchantId, ext, ct, fPath string,
) (interface{}, error) {
	file, err := os.Open(fPath)

	if err != nil {
		return nil, errorMessageAgreementNotFound
	}

	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	fi, err := file.Stat()

	if err != nil {
		return nil, errorMessageAgreementNotFound
	}

	data := &OnboardingFileData{
		Url: fmt.Sprintf(agreementUrlMask, r.config.HttpScheme, ctx.Request().Host, merchantId),
		Metadata: &OnboardingFileMetadata{
			Name:        fi.Name(),
			Extension:   ext,
			ContentType: ct,
			Size:        fi.Size(),
		},
	}

	return data, nil
}

func (r *onboardingRoute) validateUpload(file *multipart.FileHeader) (multipart.File, error) {
	if file.Size > agreementUploadMaxSize {
		return nil, errorMessageAgreementUploadMaxSize
	}

	src, err := file.Open()

	if err != nil {
		zap.S().Errorf("validate upload error", "err", err.Error())
		return nil, errorUnknown
	}

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)

	if err != nil {
		zap.S().Errorf("validate upload error", "err", err.Error())
		return nil, errorUnknown
	}

	_, err = src.Seek(0, 0)

	if err != nil {
		zap.S().Errorf("validate upload error", "err", err.Error())
		return nil, errorUnknown
	}

	ct := http.DetectContentType(buffer)

	if ct != agreementContentType {
		return nil, errorMessageAgreementContentType
	}

	return src, nil
}

// @Description Set company information in merchant onboarding process
// @Example curl -X PATCH -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  -d '{"name": "Roga and Copita LLC", "alternative_name": "Apple Inc", "website": "http://localhost", "country": "RU",
//    	"state": "St.Petersburg", "zip": "190000", "city": "St.Petersburg", "address": "Nevskiy st. 1"}' \
//  https://api.paysuper.online/admin/api/v1/merchants/company
//
// @Example curl -X PATCH -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  -d '{"name": "Roga and Copita LLC", "alternative_name": "Apple Inc", "website": "http://localhost", "country": "RU",
//    	"state": "St.Petersburg", "zip": "190000", "city": "St.Petersburg", "address": "Nevskiy st. 2"}' \
//  https://api.paysuper.online/admin/api/v1/merchants/5d4847f61986ee46ec581e26/company
func (r *onboardingRoute) setMerchantCompany(ctx echo.Context) error {
	in := &billing.MerchantCompanyInfo{}
	err := ctx.Bind(in)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(in)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	req := &grpc.OnboardingRequest{Company: in}
	merchantId := ctx.Param(requestParameterId)

	if merchantId != "" {
		req.Id = merchantId
	} else {
		req.User = &billing.MerchantUser{Id: r.authUser.Id}
	}

	rsp, err := r.billingService.ChangeMerchant(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

// @Description Set company contact information in merchant onboarding process
// @Example curl -X PATCH -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  -d '{"authorized": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890", "position": "CEO"},
//    	"technical": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"}}' \
//  https://api.paysuper.online/admin/api/v1/merchants/contacts
//
// @Example curl -X PATCH -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  -d '{"authorized": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567891", "position": "CEO"},
//    	"technical": {"name": "Unit Test", "email": "test@unit.test", "phone": "1234567890"}}' \
//  https://api.paysuper.online/admin/api/v1/merchants/5d4847f61986ee46ec581e26/contacts
func (r *onboardingRoute) setMerchantContacts(ctx echo.Context) error {
	in := &billing.MerchantContact{}
	err := ctx.Bind(in)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(in)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	req := &grpc.OnboardingRequest{Contacts: in}
	merchantId := ctx.Param(requestParameterId)

	if merchantId != "" {
		req.Id = merchantId
	} else {
		req.User = &billing.MerchantUser{Id: r.authUser.Id}
	}

	rsp, err := r.billingService.ChangeMerchant(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

// @Description Set company banking information in merchant onboarding process
// @Example curl -X PATCH -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  -d '{"currency": "RUB", "name": "Bank Name-Spb.", "address": "St.Petersburg, Nevskiy st. 1",
//  	"account_number": "408000000001", "swift": "ALFARUMM", "correspondent_account": "408000000001"}' \
//  https://api.paysuper.online/admin/api/v1/merchants/banking
//
// @Example curl -X PATCH -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  -d '{"currency": "RUB", "name": "Bank Name-Spb.", "address": "St.Petersburg, Nevskiy st. 1",
//  	"account_number": "408000000001", "swift": "ALFARUMM", "correspondent_account": "408000000002"}' \
//  https://api.paysuper.online/admin/api/v1/merchants/5d4847f61986ee46ec581e26/banking
func (r *onboardingRoute) setMerchantBanking(ctx echo.Context) error {
	in := &billing.MerchantBanking{}
	err := ctx.Bind(in)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	err = r.validate.Struct(in)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	req := &grpc.OnboardingRequest{Banking: in}
	merchantId := ctx.Param(requestParameterId)

	if merchantId != "" {
		req.Id = merchantId
	} else {
		req.User = &billing.MerchantUser{Id: r.authUser.Id}
	}

	rsp, err := r.billingService.ChangeMerchant(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

// Set company tariff in merchant onboarding process
// @Example curl -X PATCH -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  -d '{"tariff": "ffffffffffffffffffffffff"}'
//  https://api.paysuper.online/admin/api/v1/merchants/tariff
//
// @Example curl -X PATCH -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
//  -d '{"tariff": "000000000000000000000000"}'
//  https://api.paysuper.online/admin/api/v1/merchants/5d4847f61986ee46ec581e26/tariff
func (r *onboardingRoute) setMerchantTariff(ctx echo.Context) error {
	req := &grpc.OnboardingRequest{}
	err := ctx.Bind(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errorRequestParamsIncorrect)
	}

	req.Id = ctx.Param(requestParameterId)
	req.User = &billing.MerchantUser{
		Id:    r.authUser.Id,
		Email: r.authUser.Email,
	}
	err = r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.ChangeMerchant(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

// @Description Get merchant completion information
// curl -X GET -H 'Authorization: Bearer %access_token_here%' -H 'Content-Type: application/json' \
// https://api.paysuper.online/admin/api/v1/merchants/5d4847f61986ee46ec581e26/status
func (r *onboardingRoute) getMerchantStatus(ctx echo.Context) error {
	req := &grpc.SetMerchantS3AgreementRequest{
		MerchantId: ctx.Param(requestParameterId),
	}

	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.GetMerchantOnboardingCompleteData(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.Item)
}

// @Description get hellosign (https://www.hellosign.com) signature to sign license agreement
// @Example PUT /admin/api/v1/merchants/ffffffffffffffffffffffff/agreement/signature
func (r *onboardingRoute) createAgreementSignature(ctx echo.Context) error {
	req := &grpc.SetMerchantS3AgreementRequest{
		MerchantId: ctx.Param(requestParameterId),
	}
	err := r.validate.Struct(req)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, r.getValidationError(err))
	}

	rsp, err := r.billingService.AgreementSign(ctx.Request().Context(), req)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errorUnknown)
	}

	if rsp.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(rsp.Status), rsp.Message)
	}

	return ctx.JSON(http.StatusOK, rsp.SignatureRequest)
}
