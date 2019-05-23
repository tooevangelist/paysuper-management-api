package manager

import (
	"context"
	"github.com/ProtocolONE/geoip-service/pkg/proto"
	"github.com/ProtocolONE/rabbitmq/pkg"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/dao"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/payment_system"
	"github.com/paysuper/paysuper-management-api/utils"
	"github.com/paysuper/paysuper-recurring-repository/pkg/proto/repository"
	"go.uber.org/zap"
	"net/url"
)

const (
	orderErrorProjectNotFound                          = "project with specified identifier not found"
	orderErrorProjectInactive                          = "project with specified identifier is inactive"
	orderErrorPaymentMethodNotAllowed                  = "payment method not specified for project"
	orderErrorPaymentMethodNotFound                    = "payment method with specified id not found"
	orderErrorPaymentMethodInactive                    = "payment method with specified is inactive"
	orderErrorPaymentSystemNotFound                    = "payment system for specified payment method not found"
	orderErrorPaymentSystemInactive                    = "payment system for specified payment method is inactive"
	orderErrorPayerRegionUnknown                       = "payer region can't be found"
	orderErrorFixedPackageForRegionNotFound            = "project not have fixed packages for payer region"
	orderErrorFixedPackageNotFound                     = "project not have fixed package with specified amount or currency"
	orderErrorProjectOrderIdIsDuplicate                = "request with specified project order identifier processed early"
	orderErrorDynamicNotifyUrlsNotAllowed              = "dynamic verify url or notify url not allowed for project"
	orderErrorDynamicRedirectUrlsNotAllowed            = "dynamic payer redirect urls not allowed for project"
	orderErrorCurrencyNotFound                         = "currency received from request not found"
	orderErrorAmountLowerThanMinAllowed                = "order amount is lower than min allowed payment amount for project"
	orderErrorAmountGreaterThanMaxAllowed              = "order amount is greater than max allowed payment amount for project"
	orderErrorAmountLowerThanMinAllowedPaymentMethod   = "order amount is lower than min allowed payment amount for payment method"
	orderErrorAmountGreaterThanMaxAllowedPaymentMethod = "order amount is greater than max allowed payment amount for payment method"
	orderErrorCanNotCreate                             = "order can't create. try request later"
	orderErrorSignatureInvalid                         = "order request signature is invalid"
	orderErrorNotFound                                 = "order with specified identifier not found"
	orderErrorOrderAlreadyComplete                     = "order with specified identifier payed early"
	orderErrorOrderPaymentMethodIncomeCurrencyNotFound = "unknown currency received from payment system"
	orderErrorOrderPSPAccountingCurrencyNotFound       = "unknown PSP accounting currency"
	orderErrorOrderDeclined                            = "payment system decline order with specified identifier early"
	orderErrorOrderCanceled                            = "payment system cancel order with specified identifier early"

	orderErrorCreatePaymentRequiredFieldIdNotFound            = "required field with order identifier not found"
	orderErrorCreatePaymentRequiredFieldPaymentMethodNotFound = "required field with payment method identifier not found"
	orderErrorCreatePaymentRequiredFieldEmailNotFound         = "required field \"email\" not found"

	orderSignatureElementsGlue = "|"

	orderDefaultDescription = "Payment by order # %s"
)

type OrderManager struct {
	*Manager

	projectManager         *ProjectManager
	paymentSystemManager   *PaymentSystemManager
	paymentMethodManager   *PaymentMethodManager
	currencyRateManager    *CurrencyRateManager
	currencyManager        *CurrencyManager
	pspAccountingCurrency  *model.Currency
	paymentSystemsSettings *payment_system.PaymentSystemSetting
	vatManager             *VatManager
	commissionManager      *CommissionManager
	centrifugoSecret       string

	rep repository.RepositoryService
	geo proto.GeoIpService
	ctx context.Context
	pub *rabbitmq.Broker
}

type check struct {
	order         *model.OrderScalar
	project       *model.Project
	oCurrency     *model.Currency
	paymentMethod *model.PaymentMethod
}

type pmOutcomeData struct {
	amount   float64
	currency *model.Currency
}

type FindAll struct {
	Values   url.Values
	Projects map[bson.ObjectId]string
	Merchant *model.Merchant
	Limit    int32
	Offset   int32
	SortBy   []string
}

type OrderHttp struct {
	Host   string
	Scheme string
}

func InitOrderManager(
	database dao.Database,
	logger *zap.SugaredLogger,
	publisher *rabbitmq.Broker,
	repository repository.RepositoryService,
	geoService proto.GeoIpService,
) *OrderManager {
	om := &OrderManager{
		Manager:              &Manager{Database: database, Logger: logger},
		projectManager:       InitProjectManager(database, logger),
		paymentSystemManager: InitPaymentSystemManager(database, logger),
		paymentMethodManager: InitPaymentMethodManager(database, logger),
		currencyRateManager:  InitCurrencyRateManager(database, logger),
		currencyManager:      InitCurrencyManager(database, logger),
		vatManager:           InitVatManager(database, logger),
		commissionManager:    InitCommissionManager(database, logger),

		rep: repository,
		geo: geoService,
		ctx: context.TODO(),
		pub: publisher,
	}

	return om
}

/*func (om *OrderManager) FindAll(params *FindAll) (*model.OrderPaginate, error) {
	var f bson.M
	var pFilter []bson.ObjectId

	for k := range params.Projects {
		pFilter = append(pFilter, k)
	}

	filter := bson.M{"project.id": bson.M{"$in": pFilter}}

	if quickFilter, ok := params.Values[model.OrderFilterFieldQuickFilter]; ok {
		r := bson.RegEx{Pattern: ".*" + quickFilter[0] + ".*", Options: "i"}

		filter["$or"] = []bson.M{
			{"project.name": bson.M{"$regex": r}},
			{"project_account": bson.M{"$regex": r}},
			{"project_order_id": bson.M{"$regex": r, "$exists": true}},
			{"fixed_package.name": bson.M{"$regex": r, "$exists": true}},
			{"payment_method.name": bson.M{"$regex": r, "$exists": true}},
			{"id_string": bson.M{"$regex": r, "$exists": true}},
		}

		f = filter
	} else {
		f = om.processFilters(params.Values, filter)
	}

	co, err := om.Database.Repository(TableOrder).GetOrdersCountByConditions(f)

	if err != nil {
		om.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableOrder, err)
	}

	o, err := om.Database.Repository(TableOrder).FindAllOrders(f, params.SortBy, params.Limit, params.Offset)

	if err != nil {
		om.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableOrder, err)
	}

	var ot []*model.OrderSimple

	if o != nil && len(o) > 0 {
		ot, err = om.transformOrders(o, params)

		if err != nil {
			return nil, err
		}
	}

	return &model.OrderPaginate{Count: co, Items: ot}, nil
}*/

/*func (om *OrderManager) transformOrders(orders []*model.Order, params *FindAll) ([]*model.OrderSimple, error) {
	var tOrders []*model.OrderSimple

	for _, oValue := range orders {
		tOrder := &model.OrderSimple{
			Id: oValue.Id,
			Project: &model.SimpleItem{
				Id:   oValue.Project.Id,
				Name: oValue.Project.Name,
			},
			Account:        oValue.ProjectAccount,
			ProjectOrderId: oValue.ProjectOrderId,
			PayerData:      oValue.PayerData,
			ProjectAmountIncome: &model.OrderSimpleAmountObject{
				Amount: oValue.ProjectIncomeAmount,
				Currency: &model.SimpleCurrency{
					CodeInt: oValue.ProjectIncomeCurrency.CodeInt,
					CodeA3:  oValue.ProjectIncomeCurrency.CodeA3,
					Name:    oValue.ProjectIncomeCurrency.Name,
				},
			},
			FixedPackage: oValue.FixedPackage,
			Status: &model.Status{
				Status:      oValue.Status,
				Name:        model.OrderStatusesNames[oValue.Status],
				Description: model.OrderStatusesDescription[oValue.Status],
			},
			VatAmount: oValue.VatAmount,
			CreatedAt: oValue.CreatedAt.Unix(),
		}

		if oValue.PaymentMethodIncomeAmount > 0 {
			tOrder.PaymentMethodAmountIncome = &model.OrderSimpleAmountObject{
				Amount: oValue.PaymentMethodIncomeAmount,
				Currency: &model.SimpleCurrency{
					CodeInt: oValue.PaymentMethodIncomeCurrency.CodeInt,
					CodeA3:  oValue.PaymentMethodIncomeCurrency.CodeA3,
					Name:    oValue.PaymentMethodIncomeCurrency.Name,
				},
			}
		}

		if oValue.AmountOutMerchantAccountingCurrency > 0 {
			tOrder.ProjectAmountOutcome = &model.OrderSimpleAmountObject{
				Amount: oValue.AmountOutMerchantAccountingCurrency,
				Currency: &model.SimpleCurrency{
					CodeInt: params.Merchant.Currency.CodeInt,
					CodeA3:  params.Merchant.Currency.CodeA3,
					Name:    params.Merchant.Currency.Name,
				},
			}
		}

		if oValue.PaymentMethod != nil {
			tOrder.PaymentMethod = &model.SimpleItem{
				Id:   oValue.PaymentMethod.Id,
				Name: oValue.PaymentMethod.Name,
			}

			tOrder.PspFeeAmount = oValue.PspFeeAmount
			tOrder.PaymentSystemFeeAmount = oValue.PaymentSystemFeeAmount
			tOrder.ProjectFeeAmount = oValue.ProjectFeeAmount
			tOrder.ToPayerFeeAmount = oValue.ToPayerFeeAmount

			tOrder.PaymentRequisites = om.preparePaymentRequisites(oValue)
		}

		if oValue.PaymentMethodOrderClosedAt != nil {
			tOrder.ConfirmedAt = oValue.PaymentMethodOrderClosedAt.Unix()
		}

		if oValue.ProjectLastRequestedAt != nil {
			tOrder.ClosedAt = oValue.ProjectLastRequestedAt.Unix()
		}

		tOrders = append(tOrders, tOrder)
	}

	return tOrders, nil
}*/

/*// Prepare payer payment requisites for frontend
func (om *OrderManager) preparePaymentRequisites(o *model.Order) map[string]string {
	requisites := make(map[string]string)

	for k, v := range o.PaymentRequisites {
		n, ok := entity.PaymentRequisitesNames[k]

		if !ok {
			continue
		}

		requisites[n] = v
	}

	if len(o.PaymentMethodTxnParams) <= 0 {
		return requisites
	}

	for k, v := range o.PaymentMethodTxnParams {
		n, ok := entity.PaymentRequisitesNames[k]

		if !ok {
			continue
		}

		requisites[n] = v.(string)
	}

	return requisites
}*/

/*func (om *OrderManager) processFilters(values url.Values, filter bson.M) bson.M {
	if id, ok := values[model.OrderFilterFieldId]; ok {
		if bson.IsObjectIdHex(id[0]) {
			filter["_id"] = bson.ObjectIdHex(id[0])
		} else {
			filter["_id"] = id[0]
		}
	}

	if pms, ok := values[model.OrderFilterFieldPaymentMethods]; ok {
		var fPms []bson.ObjectId

		for _, pm := range pms {
			fPms = append(fPms, bson.ObjectIdHex(pm))
		}

		filter["payment_method.id"] = bson.M{"$in": fPms}
	}

	if cs, ok := values[model.OrderFilterFieldCountries]; ok {
		filter["payer_data.country_code_a2"] = bson.M{"$in": cs}
	}

	if ss, ok := values[model.OrderFilterFieldStatuses]; ok {
		var ssi []int

		for _, s := range ss {
			si, err := strconv.Atoi(s)

			if err != nil {
				continue
			}

			ssi = append(ssi, si)
		}

		if len(ssi) > 0 {
			filter["status"] = bson.M{"$in": ssi}
		}
	}

	if a, ok := values[model.OrderFilterFieldAccount]; ok {
		ar := bson.RegEx{Pattern: ".*" + a[0] + ".*", Options: "i"}
		filter["$or"] = bson.M{"project_account": ar, "pm_account": ar, "payer_data.phone": ar, "payer_data.email": ar}
	}

	pmDates := make(bson.M)

	if pmDateFrom, ok := values[model.OrderFilterFieldPMDateFrom]; ok {
		if ts, err := strconv.ParseInt(pmDateFrom[0], 10, 64); err == nil {
			pmDates["$gte"] = time.Unix(ts, 0)
		}
	}

	if pmDateTo, ok := values[model.OrderFilterFieldPMDateTo]; ok {
		if ts, err := strconv.ParseInt(pmDateTo[0], 10, 64); err == nil {
			pmDates["$lte"] = time.Unix(ts, 0)
		}
	}

	if len(pmDates) > 0 {
		filter["pm_order_close_date"] = pmDates
	}

	prjDates := make(bson.M)

	if prjDateFrom, ok := values[model.OrderFilterFieldProjectDateFrom]; ok {
		if ts, err := strconv.ParseInt(prjDateFrom[0], 10, 64); err == nil {
			prjDates["$gte"] = time.Unix(ts, 0)
		}
	}

	if prjDateTo, ok := values[model.OrderFilterFieldProjectDateTo]; ok {
		if ts, err := strconv.ParseInt(prjDateTo[0], 10, 64); err == nil {
			prjDates["$lte"] = time.Unix(ts, 0)
		}
	}

	if len(prjDates) > 0 {
		filter["created_at"] = prjDates
	}

	return filter
}*/

func (om *OrderManager) GetRevenueDynamic(rdr *model.RevenueDynamicRequest) (*model.RevenueDynamicResult, error) {
	rdr.From = utils.GetTimeRangeFrom(rdr.From)
	rdr.To = utils.GetTimeRangeFrom(rdr.To)

	res, err := om.Database.Repository(TableOrder).GetRevenueDynamic(rdr)

	if err != nil {
		return nil, err
	}

	refPoints := make(map[string]float64)

	var revPoints []*model.RevenueDynamicPoint

	pRefund := res[0][model.RevenueDynamicFacetFieldPointsRefund].([]interface{})
	pRevenue := res[0][model.RevenueDynamicFacetFieldPointsRevenue].([]interface{})

	for _, v := range pRefund {
		vm := v.(map[string]interface{})

		vmId := vm[model.RevenueDynamicFacetFieldId].(map[string]interface{})
		vmTotal := vm[model.RevenueDynamicFacetFieldTotal].(float64)

		refPoints[om.getRevenueDynamicPointsKey(vmId).String()] = FormatAmount(vmTotal)
	}

	for _, v := range pRevenue {
		vm := v.(map[string]interface{})

		vmId := vm[model.RevenueDynamicFacetFieldId].(map[string]interface{})
		vmTotal := vm[model.RevenueDynamicFacetFieldTotal].(float64)

		revPointDate := om.getRevenueDynamicPointsKey(vmId)
		refVal, ok := refPoints[revPointDate.String()]

		revPoint := &model.RevenueDynamicPoint{
			Date: revPointDate,
		}

		if ok {
			revPoint.Amount = FormatAmount(vmTotal - refVal)
		} else {
			revPoint.Amount = FormatAmount(vmTotal)
		}

		revPoints = append(revPoints, revPoint)
	}

	rd := &model.RevenueDynamicResult{
		Points:  revPoints,
		Revenue: &model.RevenueDynamicMainData{Count: 0, Total: 0, Avg: 0},
		Refund:  &model.RevenueDynamicMainData{Count: 0, Total: 0, Avg: 0},
	}

	rev := res[0][model.RevenueDynamicFacetFieldRevenue].([]interface{})

	if len(rev) > 0 {
		mRev := rev[0].(map[string]interface{})

		if v, ok := mRev[model.RevenueDynamicFacetFieldCount]; ok {
			rd.Revenue.Count = v.(int)
		}

		if v, ok := mRev[model.RevenueDynamicFacetFieldTotal]; ok {
			rd.Revenue.Total = FormatAmount(v.(float64))
		}

		if v, ok := mRev[model.RevenueDynamicFacetFieldAvg]; ok {
			rd.Revenue.Avg = FormatAmount(v.(float64))
		}
	}

	ref := res[0][model.RevenueDynamicFacetFieldRefund].([]interface{})

	if len(ref) > 0 {
		mRef := ref[0].(map[string]interface{})

		if v, ok := mRef[model.RevenueDynamicFacetFieldCount]; ok {
			rd.Refund.Count = v.(int)
		}

		if v, ok := mRef[model.RevenueDynamicFacetFieldTotal]; ok {
			rd.Refund.Total = FormatAmount(v.(float64))
		}

		if v, ok := mRef[model.RevenueDynamicFacetFieldAvg]; ok {
			rd.Refund.Avg = FormatAmount(v.(float64))
		}
	}

	return rd, err
}

func (om *OrderManager) getRevenueDynamicPointsKey(pointId map[string]interface{}) *model.RevenueDynamicPointDate {
	revPointDate := &model.RevenueDynamicPointDate{
		Year: pointId[model.RevenueDynamicRequestPeriodYear].(int),
	}

	if val, ok := pointId[model.RevenueDynamicRequestPeriodMonth]; ok {
		revPointDate.Month = val.(int)
	}

	if val, ok := pointId[model.RevenueDynamicRequestPeriodWeek]; ok {
		revPointDate.Week = val.(int)
	}

	if val, ok := pointId[model.RevenueDynamicRequestPeriodDay]; ok {
		revPointDate.Day = val.(int)
	}

	if val, ok := pointId[model.RevenueDynamicRequestPeriodHour]; ok {
		revPointDate.Hour = val.(int)
	}

	return revPointDate
}

// Get data about accounting payment by accounting period of merchant
func (om *OrderManager) GetAccountingPayment(rdr *model.RevenueDynamicRequest, mId string) (*model.AccountingPayment, error) {
	rdr.From = utils.GetTimeRangeFrom(rdr.From)
	rdr.To = utils.GetTimeRangeTo(rdr.To)

	res, err := om.Database.Repository(TableOrder).GetAccountingPayment(rdr, mId)

	if err != nil {
		return nil, err
	}

	apData := res[0]
	ap := &model.AccountingPayment{}

	if v, ok := apData["total_success"]; ok {
		sV := v.([]interface{})

		if len(sV) > 0 {
			mV := sV[0].(map[string]interface{})

			if vv, ok := mV["total"]; ok {
				ap.SuccessWithCommissions = vv.(float64)
			}
		}
	}

	if v, ok := apData["success"]; ok {
		ap.SuccessWithoutCommissions = om.getValueFromAccountingPaymentReport(v)
	}

	if v, ok := apData["refund"]; ok {
		ap.TotalRefund = om.getValueFromAccountingPaymentReport(v)
	}

	if v, ok := apData["chargeback"]; ok {
		ap.TotalChargeback = om.getValueFromAccountingPaymentReport(v)
	}

	if v, ok := apData["commission"]; ok {
		ap.TotalCommission = om.getValueFromAccountingPaymentReport(v)
	}

	return ap, nil
}

func (om *OrderManager) getValueFromAccountingPaymentReport(v interface{}) float64 {
	var fV float64

	sV := v.([]interface{})

	if len(sV) > 0 {
		mV := sV[0].(map[string]interface{})

		if vv, ok := mV["total"]; ok {
			fV = FormatAmount(vv.(float64))
		}
	}

	return fV
}
