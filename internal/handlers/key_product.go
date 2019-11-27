package handlers

import (
	"github.com/ProtocolONE/geoip-service/pkg/proto"
	"github.com/ProtocolONE/go-core/v2/pkg/logger"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	"github.com/labstack/echo/v4"
	"github.com/micro/go-micro/client"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	keyProductsPath               = "/key-products"
	keyProductsIdPath             = "/key-products/:key_product_id"
	keyProductsPublishPath        = "/key-products/:key_product_id/publish"
	keyProductsUnPublishPath      = "/key-products/:key_product_id/unpublish"
	platformsPath                 = "/platforms"
	keyProductsPlatformsFilePath  = "/key-products/:key_product_id/platforms/:platform_id/file"
	keyProductsPlatformsCountPath = "/key-products/:key_product_id/platforms/:platform_id/count"
)

type KeyProductRoute struct {
	dispatch common.HandlerSet
	cfg      common.Config
	provider.LMT
}

func NewKeyProductRoute(set common.HandlerSet, cfg *common.Config) *KeyProductRoute {
	set.AwareSet.Logger = set.AwareSet.Logger.WithFields(logger.Fields{"router": "KeyProductRoute"})
	return &KeyProductRoute{
		dispatch: set,
		cfg:      *cfg,
		LMT:      &set.AwareSet,
	}
}

func (h *KeyProductRoute) Route(groups *common.Groups) {
	groups.AuthUser.GET(keyProductsPath, h.getKeyProductList)
	groups.AuthUser.POST(keyProductsPath, h.createKeyProduct)
	groups.AuthUser.GET(keyProductsIdPath, h.getKeyProductById)
	groups.AuthUser.PUT(keyProductsIdPath, h.changeKeyProduct)
	groups.AuthUser.POST(keyProductsPublishPath, h.publishKeyProduct)
	groups.AuthUser.POST(keyProductsUnPublishPath, h.unpublishKeyProduct)
	groups.AuthUser.DELETE(keyProductsIdPath, h.deleteKeyProductById)
	groups.AuthUser.GET(platformsPath, h.getPlatformsList)

	groups.AuthUser.POST(keyProductsPlatformsFilePath, h.uploadKeys)
	groups.AuthUser.GET(keyProductsPlatformsCountPath, h.getCountOfKeys)

	groups.AuthProject.GET(keyProductsIdPath, h.getKeyProduct)
}

func (h *KeyProductRoute) unpublishKeyProduct(ctx echo.Context) error {
	req := &grpc.UnPublishKeyProductRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.KeyProductId = ctx.Param("key_product_id")

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.UnPublishKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Product)
}

func (h *KeyProductRoute) publishKeyProduct(ctx echo.Context) error {
	req := &grpc.PublishKeyProductRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.KeyProductId = ctx.Param("key_product_id")

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.PublishKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Product)
}

func (h *KeyProductRoute) getPlatformsList(ctx echo.Context) error {
	req := &grpc.ListPlatformsRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if req.Limit == 0 {
		req.Limit = h.cfg.LimitDefault
	}

	if req.Limit > h.cfg.LimitMax {
		req.Limit = h.cfg.LimitMax
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetPlatforms(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *KeyProductRoute) deleteKeyProductById(ctx echo.Context) error {
	req := &grpc.RequestKeyProductMerchant{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.Id = ctx.Param("key_product_id")

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.DeleteKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.NoContent(http.StatusOK)
}

func (h *KeyProductRoute) changeKeyProduct(ctx echo.Context) error {
	req := &grpc.CreateOrUpdateKeyProductRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.Id = ctx.Param("key_product_id")

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.CreateOrUpdateKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Product)
}

func (h *KeyProductRoute) getKeyProductById(ctx echo.Context) error {
	req := &grpc.RequestKeyProductMerchant{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.Id = ctx.Param("key_product_id")

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.Product)
}

func (h *KeyProductRoute) createKeyProduct(ctx echo.Context) error {
	req := &grpc.CreateOrUpdateKeyProductRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	h.L().Info("createKeyProduct", logger.PairArgs("req", req))

	res, err := h.dispatch.Services.Billing.CreateOrUpdateKeyProduct(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusCreated, res.Product)
}

func (h *KeyProductRoute) getKeyProductList(ctx echo.Context) error {
	authUser := common.ExtractUserContext(ctx)
	req := &grpc.ListKeyProductsRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	if req.Limit > int64(h.cfg.LimitMax) {
		req.Limit = int64(h.cfg.LimitMax)
	}

	if req.Limit <= 0 {
		req.Limit = int64(h.cfg.LimitDefault)
	}

	req.MerchantId = authUser.MerchantId

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetKeyProducts(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Message != nil {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *KeyProductRoute) getKeyProduct(ctx echo.Context) error {
	req := &grpc.GetKeyProductInfoRequest{}

	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.KeyProductId = ctx.Param("key_product_id")

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	if req.Currency == "" && req.Country == "" {
		res, err := h.dispatch.Services.Geo.GetIpData(ctx.Request().Context(), &proto.GeoIpDataRequest{IP: ctx.RealIP()})
		if err != nil {
			h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		} else {
			req.Country = res.Country.IsoCode
		}
	}

	if req.Language == "" {
		req.Language, _ = h.getCountryFromAcceptLanguage(ctx.Request().Header.Get(common.HeaderAcceptLanguage))
	}

	res, err := h.dispatch.Services.Billing.GetKeyProductInfo(ctx.Request().Context(), req)
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res.KeyProduct)
}

func (h *KeyProductRoute) uploadKeys(ctx echo.Context) error {
	req := &grpc.PlatformKeysFileRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		h.L().Error(common.ErrorMessageFileNotFound.String(), logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorMessageFileNotFound)
	}

	src, err := file.Open()
	if err != nil {
		h.L().Error(common.ErrorMessageCantReadFile.String(), logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorMessageCantReadFile)
	}
	defer src.Close()

	req.File, err = ioutil.ReadAll(src)

	if err != nil {
		h.L().Error(common.ErrorMessageCantReadFile.String(), logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorMessageCantReadFile)
	}

	req.KeyProductId = ctx.Param("key_product_id")
	req.PlatformId = ctx.Param("platform_id")

	keyProductRes, err := h.dispatch.Services.Billing.GetKeyProduct(ctx.Request().Context(), &grpc.RequestKeyProductMerchant{Id: req.KeyProductId, MerchantId: req.MerchantId})
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if keyProductRes.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(keyProductRes.Status), keyProductRes.Message)
	}

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.UploadKeysFile(ctx.Request().Context(), req, client.WithRequestTimeout(time.Minute*10))
	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}


func (h *KeyProductRoute) getCountOfKeys(ctx echo.Context) error {
	req := &grpc.GetPlatformKeyCountRequest{}
	if err := ctx.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.ErrorRequestParamsIncorrect)
	}

	req.KeyProductId = ctx.Param("key_product_id")
	req.PlatformId = ctx.Param("platform_id")

	if err := h.dispatch.Validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, common.GetValidationError(err))
	}

	res, err := h.dispatch.Services.Billing.GetAvailableKeysCount(ctx.Request().Context(), req)

	if err != nil {
		h.L().Error(common.InternalErrorTemplate, logger.PairArgs("err", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, common.ErrorInternal)
	}

	if res.Status != pkg.ResponseStatusOk {
		return echo.NewHTTPError(int(res.Status), res.Message)
	}

	return ctx.JSON(http.StatusOK, res)
}

func (h *KeyProductRoute) getCountryFromAcceptLanguage(acceptLanguage string) (string, string) {
	it := strings.Split(acceptLanguage, ",")

	if len(it) == 0 {
		return "", ""
	}

	if strings.Index(it[0], "-") == -1 {
		return "", ""
	}

	it = strings.Split(it[0], "-")

	return strings.ToLower(it[0]), strings.ToUpper(it[1])
}
