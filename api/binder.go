package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/labstack/echo/v4"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-payment-link/proto"
	"io/ioutil"
	"strconv"
	"time"
)

const (
	binderErrorQueryParamsIsEmpty   = "required params not found"
	binderErrorPeriodIsRequire      = "period is required field"
	binderErrorFromIsRequire        = "date from is required field"
	binderErrorToIsRequire          = "date to is required field"
	binderErrorUnknownRevenuePeriod = "unknown revenue period"
)

type OrderFormBinder struct{}
type OrderJsonBinder struct{}
type OrderRevenueDynamicRequestBinder struct{}
type OrderAccountingPaymentRequestBinder struct{}
type PaymentCreateProcessBinder struct{}
type OnboardingMerchantListingBinder struct{}
type OnboardingChangeMerchantStatusBinder struct{}
type OnboardingNotificationsListBinder struct{}
type OnboardingGetPaymentMethodBinder struct{}
type OnboardingListPaymentMethodsBinder struct{}
type OnboardingChangePaymentMethodBinder struct{}
type OnboardingCreateNotificationBinder struct{}
type PaylinksListBinder struct{}
type PaylinksUrlBinder struct{}
type PaylinksCreateBinder struct{}
type PaylinksUpdateBinder struct{}
type ProductsGetProductsListBinder struct{}
type ProductsCreateProductBinder struct{}
type ProductsUpdateProductBinder struct{}
type ChangeMerchantDataRequestBinder struct {
	*Api
}

func (cb *OrderFormBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	db := new(echo.DefaultBinder)

	if err = db.Bind(i, ctx); err != nil {
		return err
	}

	params, err := ctx.FormParams()
	addParams := make(map[string]string)
	rawParams := make(map[string]string)

	if err != nil {
		return err
	}

	o := i.(*billing.OrderCreateRequest)

	for key, value := range params {
		if _, ok := model.OrderReservedWords[key]; !ok {
			addParams[key] = value[0]
		}

		rawParams[key] = value[0]
	}

	o.Other = addParams
	o.RawParams = rawParams

	return
}

func (cb *OrderJsonBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	var buf []byte

	if ctx.Request().Body != nil {
		buf, err = ioutil.ReadAll(ctx.Request().Body)
		rdr := ioutil.NopCloser(bytes.NewBuffer(buf))

		if err != nil {
			return err
		}

		ctx.Request().Body = rdr
	}

	db := new(echo.DefaultBinder)

	if err = db.Bind(i, ctx); err != nil {
		return err
	}

	i.(*billing.OrderCreateRequest).RawBody = string(buf)

	return
}

func (cb *OrderRevenueDynamicRequestBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	period := ctx.Param(model.OrderRevenueDynamicRequestFieldPeriod)

	if period == "" {
		return errors.New(binderErrorPeriodIsRequire)
	}

	if _, ok := model.RevenuePeriods[period]; !ok {
		return errors.New(binderErrorUnknownRevenuePeriod)
	}

	params := ctx.QueryParams()

	s := i.(*model.RevenueDynamicRequest)
	s.Period = period
	s.Project = []bson.ObjectId{}

	if projects, ok := params[model.OrderRevenueDynamicRequestFieldProject]; ok {
		for _, project := range projects {
			if bson.IsObjectIdHex(project) == false {
				return errors.New(model.ResponseMessageProjectIdIsInvalid)
			}

			s.Project = append(s.Project, bson.ObjectIdHex(project))
		}
	}

	s, err = OrderPrepareAccountingPaymentRequest(s, ctx)

	if err != nil {
		return err
	}

	return
}

func (cb *OrderAccountingPaymentRequestBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	rdr := i.(*model.RevenueDynamicRequest)
	rdr, err = OrderPrepareAccountingPaymentRequest(rdr, ctx)

	if err != nil {
		return err
	}

	return
}

func OrderPrepareAccountingPaymentRequest(rdr *model.RevenueDynamicRequest, ctx echo.Context) (*model.RevenueDynamicRequest, error) {
	params := ctx.QueryParams()

	if len(params) <= 0 {
		return nil, errors.New(binderErrorQueryParamsIsEmpty)
	}

	if _, ok := params[model.OrderRevenueDynamicRequestFieldFrom]; !ok {
		return nil, errors.New(binderErrorFromIsRequire)
	}

	if _, ok := params[model.OrderRevenueDynamicRequestFieldTo]; !ok {
		return nil, errors.New(binderErrorToIsRequire)
	}

	t, err := strconv.ParseInt(params[model.OrderRevenueDynamicRequestFieldFrom][0], 10, 64)

	if err != nil {
		return nil, err
	}

	rdr.From = time.Unix(t, 0)

	t, err = strconv.ParseInt(params[model.OrderRevenueDynamicRequestFieldTo][0], 10, 64)

	if err != nil {
		return nil, err
	}

	rdr.To = time.Unix(t, 0)

	return rdr, nil
}

func (cb *PaymentCreateProcessBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	db := new(echo.DefaultBinder)
	untypedData := make(map[string]interface{})

	if err = db.Bind(&untypedData, ctx); err != nil {
		return
	}

	data := i.(map[string]string)

	for k, v := range untypedData {
		switch sv := v.(type) {
		case bool:
			data[k] = "0"

			if sv == true {
				data[k] = "1"
			}
			break
		default:
			data[k] = fmt.Sprintf("%v", sv)
		}
	}

	return
}

func (cb *OnboardingMerchantListingBinder) Bind(i interface{}, ctx echo.Context) (err error) {
	db := new(echo.DefaultBinder)
	err = db.Bind(i, ctx)

	if err != nil {
		return err
	}

	params := ctx.QueryParams()
	structure := i.(*grpc.MerchantListingRequest)

	if structure.Limit <= 0 {
		structure.Limit = LimitDefault
	}

	if v, ok := params[requestParameterIsSigned]; ok {
		if v[0] == "0" || v[0] == "false" {
			structure.IsSigned = 2
		} else {
			if v[0] == "1" || v[0] == "true" {
				structure.IsSigned = 2
			} else {
				return errors.New(errorQueryParamsIncorrect)
			}
		}
	}

	return
}

func (cb *OnboardingNotificationsListBinder) Bind(i interface{}, ctx echo.Context) error {
	params := ctx.QueryParams()
	structure := i.(*grpc.ListingNotificationRequest)

	structure.Limit = LimitDefault
	structure.Offset = OffsetDefault

	merchantId := ctx.Param(requestParameterMerchantId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return errors.New(errorIncorrectMerchantId)
	}

	structure.MerchantId = merchantId

	if v, ok := params[requestParameterUserId]; ok {
		if bson.IsObjectIdHex(v[0]) == false {
			return errors.New(errorIncorrectUserId)
		}

		structure.UserId = v[0]
	}

	if v, ok := params[requestParameterIsSystem]; ok {
		if v[0] == "0" || v[0] == "false" {
			structure.IsSystem = 1
		} else {
			structure.IsSystem = 2
		}
	}

	if v, ok := params[requestParameterLimit]; ok {
		if i, err := strconv.ParseInt(v[0], 10, 32); err == nil {
			structure.Limit = int32(i)
		}
	}

	if v, ok := params[requestParameterOffset]; ok {
		if i, err := strconv.ParseInt(v[0], 10, 32); err == nil {
			structure.Offset = int32(i)
		}
	}

	return nil
}

func (cb *OnboardingGetPaymentMethodBinder) Bind(i interface{}, ctx echo.Context) error {
	merchantId := ctx.Param(requestParameterMerchantId)
	paymentMethodId := ctx.Param(requestParameterPaymentMethodId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return errors.New(errorIncorrectMerchantId)
	}

	if paymentMethodId == "" || bson.IsObjectIdHex(paymentMethodId) == false {
		return errors.New(errorIncorrectPaymentMethodId)
	}

	structure := i.(*grpc.GetMerchantPaymentMethodRequest)
	structure.MerchantId = merchantId
	structure.PaymentMethodId = paymentMethodId

	return nil
}

func (cb *OnboardingListPaymentMethodsBinder) Bind(i interface{}, ctx echo.Context) error {
	merchantId := ctx.Param(requestParameterMerchantId)
	paymentMethodName := ctx.QueryParam(requestParameterPaymentMethodName)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return errors.New(errorIncorrectMerchantId)
	}

	structure := i.(*grpc.ListMerchantPaymentMethodsRequest)
	structure.MerchantId = merchantId
	structure.PaymentMethodName = paymentMethodName

	return nil
}

func (cb *OnboardingChangePaymentMethodBinder) Bind(i interface{}, ctx echo.Context) error {
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)

	if err != nil {
		return err
	}

	structure := i.(*grpc.MerchantPaymentMethodRequest)
	merchantId := ctx.Param(requestParameterMerchantId)
	methodId := ctx.Param(requestParameterPaymentMethodId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return errors.New(errorIncorrectMerchantId)
	}

	if methodId == "" || bson.IsObjectIdHex(methodId) == false ||
		structure.PaymentMethod.Id != methodId {
		return errors.New(errorIncorrectPaymentMethodId)
	}

	structure.MerchantId = merchantId

	return nil
}

func (b *OnboardingChangeMerchantStatusBinder) Bind(i interface{}, ctx echo.Context) error {
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)

	if err != nil {
		return err
	}

	merchantId := ctx.Param(requestParameterId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return errors.New(errorIncorrectMerchantId)
	}

	structure := i.(*grpc.MerchantChangeStatusRequest)
	structure.MerchantId = merchantId

	return nil
}

func (b *OnboardingCreateNotificationBinder) Bind(i interface{}, ctx echo.Context) error {
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)

	if err != nil {
		return err
	}

	merchantId := ctx.Param(requestParameterMerchantId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return errors.New(errorIncorrectMerchantId)
	}

	structure := i.(*grpc.NotificationRequest)
	structure.MerchantId = merchantId

	return nil
}

func (b *PaylinksListBinder) Bind(i interface{}, ctx echo.Context) error {
	params := ctx.QueryParams()
	structure := i.(*paylink.GetPaylinksRequest)

	structure.Limit = LimitDefault
	structure.Offset = OffsetDefault

	if v, ok := params[requestParameterLimit]; ok {
		i, err := strconv.ParseInt(v[0], 10, 32)
		if err != nil {
			return err
		}
		structure.Limit = uint32(i)
	}

	if v, ok := params[requestParameterOffset]; ok {
		i, err := strconv.ParseInt(v[0], 10, 32)
		if err != nil {
			return err
		}
		structure.Offset = uint32(i)
	}

	structure.ProjectId = ctx.Param(requestParameterProjectId)

	return nil
}

func (b *PaylinksUrlBinder) Bind(i interface{}, ctx echo.Context) error {
	params := ctx.QueryParams()
	structure := i.(*paylink.GetPaylinkURLRequest)

	structure.Id = ctx.Param(requestParameterId)

	if v, ok := params[requestParameterUtmSource]; ok {
		structure.UtmSource = v[0]
	}

	if v, ok := params[requestParameterUtmMedium]; ok {
		structure.UtmMedium = v[0]
	}

	if v, ok := params[requestParameterUtmCampaign]; ok {
		structure.UtmCampaign = v[0]
	}

	return nil
}

func (b *PaylinksCreateBinder) Bind(i interface{}, ctx echo.Context) error {
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)
	if err != nil {
		return err
	}

	structure := i.(*paylink.CreatePaylinkRequest)
	structure.Id = ""

	return nil
}

func (b *PaylinksUpdateBinder) Bind(i interface{}, ctx echo.Context) error {
	id := ctx.Param(requestParameterId)
	if id == "" {
		return errors.New(errorIncorrectPaylinkId)
	}
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)
	if err != nil {
		return err
	}

	structure := i.(*paylink.CreatePaylinkRequest)
	structure.Id = id

	return nil
}

func (b *ProductsGetProductsListBinder) Bind(i interface{}, ctx echo.Context) error {
	limit := int32(LimitDefault)
	offset := int32(OffsetDefault)

	params := ctx.QueryParams()

	if v, ok := params[requestParameterLimit]; ok {
		i, err := strconv.ParseInt(v[0], 10, 32)
		if err != nil {
			return err
		}
		limit = int32(i)
	}

	if v, ok := params[requestParameterOffset]; ok {
		i, err := strconv.ParseInt(v[0], 10, 32)
		if err != nil {
			return err
		}
		offset = int32(i)
	}

	structure := i.(*grpc.ListProductsRequest)
	structure.Limit = limit
	structure.Offset = offset

	if v, ok := params[requestParameterName]; ok {
		if v[0] != "" {
			structure.Name = v[0]
		}
	}

	if v, ok := params[requestParameterSku]; ok {
		if v[0] != "" {
			structure.Sku = v[0]
		}
	}

	if v, ok := params[requestParameterProjectId]; ok {
		if v[0] != "" {
			structure.ProjectId = v[0]
		}
	}

	return nil
}

func (b *ProductsCreateProductBinder) Bind(i interface{}, ctx echo.Context) error {
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)
	if err != nil {
		return err
	}

	structure := i.(*grpc.Product)
	structure.Id = ""

	return nil
}

func (b *ProductsUpdateProductBinder) Bind(i interface{}, ctx echo.Context) error {
	id := ctx.Param(requestParameterId)
	if id == "" || bson.IsObjectIdHex(id) == false {
		return errors.New(errorIncorrectProductId)
	}
	db := new(echo.DefaultBinder)
	err := db.Bind(i, ctx)
	if err != nil {
		return err
	}

	structure := i.(*grpc.Product)
	structure.Id = id

	return nil
}

func (b *ChangeMerchantDataRequestBinder) Bind(i interface{}, ctx echo.Context) error {
	req := make(map[string]interface{})

	db := new(echo.DefaultBinder)
	err := db.Bind(&req, ctx)

	if err != nil {
		return errors.New(errorQueryParamsIncorrect)
	}

	merchantId := ctx.Param(requestParameterId)

	if merchantId == "" || bson.IsObjectIdHex(merchantId) == false {
		return errors.New(errorIncorrectMerchantId)
	}

	mReq := &grpc.GetMerchantByRequest{MerchantId: merchantId}
	mRsp, err := b.billingService.GetMerchantBy(context.Background(), mReq)

	if err != nil {
		b.logError(`Call billing server method "GetMerchantBy" failed`, []interface{}{"error", err.Error(), "request", mReq})
		return errors.New(errorUnknown)
	}

	if mRsp.Status != pkg.ResponseStatusOk {
		return errors.New(mRsp.Message)
	}

	structure := i.(*grpc.ChangeMerchantDataRequest)
	structure.MerchantId = merchantId
	structure.AgreementType = mRsp.Item.AgreementType
	structure.HasMerchantSignature = mRsp.Item.HasMerchantSignature
	structure.HasPspSignature = mRsp.Item.HasPspSignature
	structure.AgreementSentViaMail = mRsp.Item.AgreementSentViaMail
	structure.MailTrackingLink = mRsp.Item.MailTrackingLink

	if v, ok := req[requestParameterAgreementType]; ok {
		if tv, ok := v.(float64); !ok {
			return errors.New(errorMessageAgreementTypeIncorrectType)
		} else {
			structure.AgreementType = int32(tv)
		}
	}

	if v, ok := req[requestParameterHasMerchantSignature]; ok {
		if tv, ok := v.(bool); !ok {
			return errors.New(errorMessageHasMerchantSignatureIncorrectType)
		} else {
			structure.HasMerchantSignature = tv
		}
	}

	if v, ok := req[requestParameterHasPspSignature]; ok {
		if tv, ok := v.(bool); !ok {
			return errors.New(errorMessageHasPspSignatureIncorrectType)
		} else {
			structure.HasPspSignature = tv
		}
	}

	if v, ok := req[requestParameterAgreementSentViaMail]; ok {
		if tv, ok := v.(bool); !ok {
			return errors.New(errorMessageAgreementSentViaMailIncorrectType)
		} else {
			structure.AgreementSentViaMail = tv
		}
	}

	if v, ok := req[requestParameterMailTrackingLink]; ok {
		if tv, ok := v.(string); !ok {
			return errors.New(errorMessageMailTrackingLinkIncorrectType)
		} else {
			structure.MailTrackingLink = tv
		}
	}

	return nil
}
