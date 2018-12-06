package manager

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/payment_system"
	"github.com/ProtocolONE/p1pay.api/payment_system/entity"
	"github.com/globalsign/mgo/bson"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	orderErrorProjectNotFound                          = "project with specified identifier not found"
	orderErrorProjectInactive                          = "project with specified identifier is inactive"
	orderErrorPaymentMethodNotAllowed                  = "payment method not specified for project"
	orderErrorPaymentMethodNotFound                    = "payment method with specified not found"
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

	geoDbReader            *geoip2.Reader
	projectManager         *ProjectManager
	paymentSystemManager   *PaymentSystemManager
	paymentMethodManager   *PaymentMethodManager
	currencyRateManager    *CurrencyRateManager
	currencyManager        *CurrencyManager
	pspAccountingCurrency  *model.Currency
	paymentSystemsSettings *payment_system.PaymentSystemSetting
	vatManager             *VatManager
	commissionManager      *CommissionManager
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
	Limit    int
	Offset   int
}

func InitOrderManager(
	database dao.Database,
	logger *zap.SugaredLogger,
	geoDbReader *geoip2.Reader,
	pspAccountingCurrencyA3 string,
	paymentSystemsSettings *payment_system.PaymentSystemSetting,
) *OrderManager {
	om := &OrderManager{
		Manager:                &Manager{Database: database, Logger: logger},
		geoDbReader:            geoDbReader,
		projectManager:         InitProjectManager(database, logger),
		paymentSystemManager:   InitPaymentSystemManager(database, logger),
		paymentMethodManager:   InitPaymentMethodManager(database, logger),
		currencyRateManager:    InitCurrencyRateManager(database, logger),
		currencyManager:        InitCurrencyManager(database, logger),
		paymentSystemsSettings: paymentSystemsSettings,
		vatManager:             InitVatManager(database, logger),
		commissionManager:      InitCommissionManager(database, logger),
	}

	om.pspAccountingCurrency = om.currencyManager.FindByCodeA3(pspAccountingCurrencyA3)

	return om
}

func (om *OrderManager) Process(order *model.OrderScalar) (*model.Order, error) {
	var pm *model.PaymentMethod
	var pmOutcomeData *pmOutcomeData
	var gRecord *geoip2.City
	var ofp *model.OrderFixedPackage
	var err error

	p := om.projectManager.FindProjectById(order.ProjectId)

	if p == nil {
		return nil, errors.New(orderErrorProjectNotFound)
	}

	if p.IsActive == false {
		return nil, errors.New(orderErrorProjectInactive)
	}

	var oCurrency *model.Currency

	if order.Currency != nil {
		oCurrency = om.currencyManager.FindByCodeA3(*order.Currency)

		if oCurrency == nil {
			return nil, errors.New(orderErrorCurrencyNotFound)
		}
	}

	check := &check{
		order: &model.OrderScalar{
			Amount:           order.Amount,
			Currency:         order.Currency,
			Region:           order.Region,
			CreateOrderIp:    order.CreateOrderIp,
			RawRequestParams: order.RawRequestParams,
			Signature:        order.Signature,
			PaymentMethod:    order.PaymentMethod,
		},
		project:   p,
		oCurrency: oCurrency,
	}

	if order.Signature != nil {
		err = om.checkSignature(check)

		if err != nil {
			return nil, err
		}
	}

	if order.PaymentMethod != nil {
		pm, err = om.checkPaymentMethod(check)

		if err != nil {
			return nil, err
		}
	}

	check.paymentMethod = pm

	if p.OnlyFixedAmounts == true {
		gRecord, ofp, err = om.getOrderFixedPackage(check)

		if err != nil {
			return nil, err
		}
	}

	if err = om.checkProjectLimits(check); err != nil {
		return nil, err
	}

	if order.PaymentMethod != nil {
		if pmOutcomeData, err = om.checkPaymentMethodLimits(check); err != nil {
			return nil, err
		}
	}

	if order.OrderId != nil {
		if err = om.checkProjectOrderIdUnique(order); err != nil {
			return nil, err
		}
	}

	if (order.UrlVerify != nil || order.UrlNotify != nil) && p.IsAllowDynamicNotifyUrls == false {
		return nil, errors.New(orderErrorDynamicNotifyUrlsNotAllowed)
	}

	if (order.UrlSuccess != nil || order.UrlFail != nil) && p.IsAllowDynamicRedirectUrls == false {
		return nil, errors.New(orderErrorDynamicRedirectUrlsNotAllowed)
	}

	mACAmount, _ := om.currencyRateManager.convert(oCurrency.CodeInt, p.Merchant.Currency.CodeInt, order.Amount)
	pOutAmount, _ := om.currencyRateManager.convert(oCurrency.CodeInt, p.CallbackCurrency.CodeInt, order.Amount)

	id := bson.NewObjectId()

	nOrder := &model.Order{
		Id: id,
		Project: &model.ProjectOrder{
			Id:       p.Id,
			Name:     p.Name,
			Merchant: p.Merchant,
		},
		Description:            fmt.Sprintf(orderDefaultDescription, id.Hex()),
		ProjectOrderId:         order.OrderId,
		ProjectAccount:         order.Account,
		ProjectIncomeAmount:    FormatAmount(order.Amount),
		ProjectIncomeCurrency:  oCurrency,
		ProjectOutcomeAmount:   FormatAmount(pOutAmount),
		ProjectOutcomeCurrency: p.CallbackCurrency,
		ProjectParams:          order.Other,
		PayerData: &model.PayerData{
			Ip:            order.CreateOrderIp,
			CountryCodeA2: gRecord.Country.IsoCode,
			CountryName:   &model.Name{EN: gRecord.Country.Names["en"], RU: gRecord.Country.Names["ru"]},
			City:          &model.Name{EN: gRecord.City.Names["en"], RU: gRecord.City.Names["ru"]},
			Subdivision:   gRecord.Subdivisions[0].IsoCode,
			Timezone:      gRecord.Location.TimeZone,
			Phone:         order.PayerPhone,
			Email:         order.PayerEmail,
		},
		Status:                             model.OrderStatusNew,
		CreatedAt:                          time.Now(),
		IsJsonRequest:                      order.IsJsonRequest,
		FixedPackage:                       ofp,
		AmountInMerchantAccountingCurrency: FormatAmount(mACAmount),
	}

	if order.Description != nil {
		nOrder.Description = *order.Description
	}

	if order.PaymentMethod != nil {
		nOrder.PaymentMethod = &model.OrderPaymentMethod{
			Id:            pm.Id,
			Name:          pm.Name,
			Params:        pm.Params,
			PaymentSystem: pm.PaymentSystem,
			GroupAlias:    pm.GroupAlias,
		}

		nOrder.PaymentMethodOutcomeAmount = FormatAmount(pmOutcomeData.amount)
		nOrder.PaymentMethodOutcomeCurrency = pmOutcomeData.currency

		if nOrder, err = om.ProcessOrderCommissions(nOrder); err != nil {
			return nil, err
		}
	}

	if err = om.Database.Repository(TableOrder).InsertOrder(nOrder); err != nil {
		om.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableOrder, err)

		return nil, errors.New(orderErrorCanNotCreate)
	}

	return nOrder, nil
}

// Calculate all possible commissions for order, i.e. payment system fee amount, PSP (P1) fee amount,
// commission shifted from project to user and VAT
func (om *OrderManager) ProcessOrderCommissions(o *model.Order) (*model.Order, error) {
	pmOutAmount := o.PaymentMethodOutcomeAmount

	// if merchant enable VAT calculation then we're need to calculate VAT for payer
	if o.Project.Merchant.IsVatEnabled == true {
		vatAmount, err := om.vatManager.CalculateVat(o.PayerData.CountryCodeA2, o.PayerData.Subdivision, o.PaymentMethodOutcomeAmount)

		if err != nil {
			return nil, err
		}

		o.VatAmount = FormatAmount(vatAmount)

		// add VAT amount to payment amount
		pmOutAmount += vatAmount
	}

	// calculate commissions to selected payment method
	commissions, err := om.commissionManager.CalculateCommission(o.Project.Id, o.PaymentMethod.Id, o.PaymentMethodOutcomeAmount)

	if err != nil {
		return nil, err
	}

	var cCom float64

	mAccCur := o.Project.Merchant.Currency.CodeInt
	pmOutCur := o.PaymentMethodOutcomeCurrency.CodeInt

	totalCommission := commissions.PMCommission + commissions.PspCommission

	// if merchant enable to shift commissions form project to payer then we're need to calculate commissions shifting
	if o.Project.Merchant.IsCommissionToUserEnabled == true {
		// subtract commission to user from project's commission
		totalCommission -= commissions.ToUserCommission

		// add commission to user to payment amount
		pmOutAmount += commissions.ToUserCommission

		o.ToPayerFeeAmount = &model.OrderFee{
			AmountPaymentMethodCurrency: FormatAmount(commissions.ToUserCommission),
		}

		// convert amount of fee shifted to user to accounting currency of merchant
		if cCom, err = om.currencyRateManager.convert(pmOutCur, mAccCur, commissions.ToUserCommission); err != nil {
			return nil, err
		}

		o.ToPayerFeeAmount.AmountMerchantCurrency = FormatAmount(cCom)
	}

	o.ProjectFeeAmount = &model.OrderFee{
		AmountPaymentMethodCurrency: FormatAmount(totalCommission),
	}

	// convert amount of fee to project to accounting currency of merchant
	if cCom, err = om.currencyRateManager.convert(pmOutCur, mAccCur, totalCommission); err != nil {
		return nil, err
	}

	o.ProjectFeeAmount.AmountMerchantCurrency = FormatAmount(cCom)

	o.PspFeeAmount = &model.OrderFeePsp{
		AmountPaymentMethodCurrency: commissions.PspCommission,
	}

	// convert PSP amount of fee to accounting currency of merchant
	if cCom, err = om.currencyRateManager.convert(pmOutCur, mAccCur, commissions.PspCommission); err != nil {
		return nil, err
	}

	o.PspFeeAmount.AmountMerchantCurrency = FormatAmount(cCom)

	// convert PSP amount of fee to accounting currency of PSP
	if cCom, err = om.currencyRateManager.convert(pmOutCur, om.pspAccountingCurrency.CodeInt, commissions.PspCommission); err != nil {
		return nil, err
	}

	o.PspFeeAmount.AmountPspCurrency = FormatAmount(cCom)

	// save information about payment system commission
	o.PaymentSystemFeeAmount = &model.OrderFeePaymentSystem{
		AmountPaymentMethodCurrency: FormatAmount(commissions.PMCommission),
	}

	// convert payment system amount of fee to accounting currency of payment system
	cCom, err = om.currencyRateManager.convert(pmOutCur, o.PaymentMethod.PaymentSystem.AccountingCurrency.CodeInt, commissions.PMCommission)

	if err != nil {
		return nil, err
	}

	o.PaymentSystemFeeAmount.AmountPaymentSystemCurrency = FormatAmount(cCom)

	// convert payment system amount of fee to accounting currency of merchant
	if cCom, err = om.currencyRateManager.convert(pmOutCur, mAccCur, commissions.PMCommission); err != nil {
		return nil, err
	}

	o.PaymentSystemFeeAmount.AmountMerchantCurrency = FormatAmount(cCom)

	return o, nil
}

func (om *OrderManager) FindById(id string) *model.Order {
	o, err := om.Database.Repository(TableOrder).FindOrderById(bson.ObjectIdHex(id))

	if err != nil {
		om.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableOrder, err)
	}

	return o
}

func (om *OrderManager) GetOrderByIdWithPaymentMethods(id string) (*model.Order, []*model.PaymentMethod, error) {
	order := om.FindById(id)

	if order == nil {
		return nil, nil, errors.New(orderErrorNotFound)
	}

	projectPms, err := om.projectManager.GetProjectPaymentMethods(order.Project.Id)

	if err != nil {
		return nil, nil, err
	}

	pdMap := make(map[string]*model.PaymentMethodsPreparedFormData)

	for _, pm := range projectPms {
		amount, err := om.currencyRateManager.convert(order.ProjectIncomeCurrency.CodeInt, pm.Currency.CodeInt, order.ProjectIncomeAmount)

		if err != nil {
			return nil, nil, err
		}

		pmPreparedData := &model.PaymentMethodsPreparedFormData{
			Amount:   amount,
			Currency: pm.Currency,
		}

		// if commission to user enabled for merchant then calculate commissions to user
		// for every allowed payment methods
		if order.Project.Merchant.IsCommissionToUserEnabled == true {
			commissions, err := om.commissionManager.CalculateCommission(order.Project.Id, pm.Id, pmPreparedData.Amount)

			if err != nil {
				return nil, nil, err
			}

			amount += commissions.ToUserCommission
			pmPreparedData.ToUserCommissionAmount = FormatAmount(commissions.ToUserCommission)
		}

		// if merchant enable VAT calculation then we're calculate VAT for payer
		if order.Project.Merchant.IsVatEnabled == true {
			vat, err := om.vatManager.CalculateVat(order.PayerData.CountryCodeA2, order.PayerData.Subdivision, pmPreparedData.Amount)

			if err != nil {
				return nil, nil, err
			}

			amount += vat
			pmPreparedData.Vat = FormatAmount(vat)
		}

		pmPreparedData.Amount = FormatAmount(amount)

		pdMap[pm.GroupAlias] = pmPreparedData
	}

	order.PaymentMethodsPreparedFormData = pdMap

	return order, projectPms, nil
}

func (om *OrderManager) getPaymentMethod(order *model.OrderScalar, pms map[string][]*model.ProjectPaymentModes) (*model.ProjectPaymentModes, error) {
	cpms, ok := pms[*order.PaymentMethod]

	if !ok || len(cpms) <= 0 {
		return nil, errors.New(orderErrorPaymentMethodNotAllowed)
	}

	var opm *model.ProjectPaymentModes

	for _, ppm := range cpms {
		if opm == nil || opm.AddedAt.Before(ppm.AddedAt) == true {
			opm = ppm
		}
	}

	return opm, nil
}

func (om *OrderManager) checkProjectOrderIdUnique(order *model.OrderScalar) error {
	if order.OrderId == nil {
		return nil
	}

	o, err := om.Database.Repository(TableOrder).FindOrderByProjectOrderId(*order.OrderId)

	if err != nil {
		om.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableOrder, err)
	}

	if o == nil {
		return nil
	}

	return errors.New(orderErrorProjectOrderIdIsDuplicate)
}

func (om *OrderManager) checkProjectLimits(c *check) error {
	var err error
	cAmount := c.order.Amount

	if c.oCurrency != nil && c.oCurrency.CodeA3 != c.project.LimitsCurrency.CodeA3 {
		cAmount, err = om.currencyRateManager.convert(c.oCurrency.CodeInt, c.project.LimitsCurrency.CodeInt, c.order.Amount)

		if err != nil {
			return err
		}
	}

	if cAmount < c.project.MinPaymentAmount {
		return errors.New(orderErrorAmountLowerThanMinAllowed)
	}

	if cAmount > c.project.MaxPaymentAmount {
		return errors.New(orderErrorAmountGreaterThanMaxAllowed)
	}

	return nil
}

func (om *OrderManager) checkPaymentMethodLimits(c *check) (*pmOutcomeData, error) {
	var err error
	cAmount := c.order.Amount

	if c.oCurrency != nil && c.oCurrency.CodeA3 != c.paymentMethod.Currency.CodeA3 {
		cAmount, err = om.currencyRateManager.convert(c.oCurrency.CodeInt, c.paymentMethod.Currency.CodeInt, c.order.Amount)

		if err != nil {
			return nil, err
		}
	}

	if cAmount < c.paymentMethod.MinPaymentAmount {
		return nil, errors.New(orderErrorAmountLowerThanMinAllowedPaymentMethod)
	}

	if cAmount > c.paymentMethod.MaxPaymentAmount {
		return nil, errors.New(orderErrorAmountGreaterThanMaxAllowedPaymentMethod)
	}

	pmOutcomeData := &pmOutcomeData{
		amount:   cAmount,
		currency: c.paymentMethod.Currency,
	}

	return pmOutcomeData, nil
}

func (om *OrderManager) checkPaymentMethod(c *check) (*model.PaymentMethod, error) {
	opm, err := om.getPaymentMethod(c.order, c.project.PaymentMethods)

	if err != nil {
		return nil, err
	}

	pm := om.paymentMethodManager.FindById(opm.Id)

	if pm == nil {
		return nil, errors.New(orderErrorPaymentMethodNotFound)
	}

	if pm.IsActive == false {
		return nil, errors.New(orderErrorPaymentMethodInactive)
	}

	ps := pm.PaymentSystem

	if ps == nil {
		return nil, errors.New(orderErrorPaymentSystemNotFound)
	}

	if ps.IsActive == false {
		return nil, errors.New(orderErrorPaymentSystemInactive)
	}

	return pm, nil
}

func (om *OrderManager) getOrderFixedPackage(c *check) (*geoip2.City, *model.OrderFixedPackage, error) {
	var region string

	if c.order.Region != nil {
		region = *c.order.Region
	}

	ip := net.ParseIP(c.order.CreateOrderIp)
	gRecord, err := om.geoDbReader.City(ip)

	if err != nil {
		return nil, nil, errors.New(orderErrorPayerRegionUnknown)
	}

	if region == "" {
		region = gRecord.Country.IsoCode
	}

	fps, ok := c.project.FixedPackage[region]

	if !ok || len(fps) <= 0 {
		return nil, nil, errors.New(orderErrorFixedPackageForRegionNotFound)
	}

	var ofp *model.FixedPackage
	var ofpId int

	for i, fp := range fps {
		if fp.Price != c.order.Amount || fp.Currency.CodeA3 != *c.order.Currency {
			continue
		}

		ofp = fp
		ofpId = i
	}

	if ofp == nil {
		return nil, nil, errors.New(orderErrorFixedPackageNotFound)
	}

	orderFp := &model.OrderFixedPackage{
		Id:          ofpId,
		Region:      region,
		Name:        ofp.Name,
		CurrencyInt: ofp.CurrencyInt,
		Price:       ofp.Price,
	}

	return gRecord, orderFp, nil
}

func (om *OrderManager) checkSignature(check *check) error {
	keys := make([]string, 0, len(check.order.RawRequestParams))
	gs := make([]string, 0, len(check.order.RawRequestParams))

	for k := range check.order.RawRequestParams {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		value := k + "=" + check.order.RawRequestParams[k]
		gs = append(gs, value)
	}

	h := sha256.New()
	h.Write([]byte(strings.Join(gs, orderSignatureElementsGlue) + orderSignatureElementsGlue + check.project.SecretKey))

	if string(h.Sum(nil)) != *check.order.Signature {
		return errors.New(orderErrorSignatureInvalid)
	}

	return nil
}

func (om *OrderManager) FindAll(params *FindAll) (*model.OrderPaginate, error) {
	var f bson.M
	var pFilter []bson.ObjectId

	for k := range params.Projects {
		pFilter = append(pFilter, k)
	}

	filter := bson.M{"project.id": bson.M{"$in": pFilter}}

	if len(params.Values) > 0 {
		f = om.ProcessFilters(params.Values, filter)
	}

	co, err := om.Database.Repository(TableOrder).GetOrdersCountByConditions(f)

	if err != nil {
		om.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableOrder, err)
	}

	o, err := om.Database.Repository(TableOrder).FindAllOrders(f, params.Limit, params.Offset)

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
}

func (om *OrderManager) transformOrders(orders []*model.Order, params *FindAll) ([]*model.OrderSimple, error) {
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
}

// Prepare payer payment requisites for frontend
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
}

func (om *OrderManager) ProcessFilters(values url.Values, filter bson.M) bson.M {
	if id, ok := values[model.OrderFilterFieldId]; ok {
		filter["_id"] = bson.ObjectIdHex(id[0])
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
		filter["pm_order_close_date"] = prjDates
	}

	return filter
}

func (om *OrderManager) ProcessCreatePayment(data map[string]string, psSettings map[string]interface{}) *payment_system.PaymentResponse {
	var err error

	if err = om.validateCreatePaymentData(data); err != nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorValidation, err.Error())
	}

	o := om.FindById(data[model.OrderPaymentCreateRequestFieldOrderId])

	if o == nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorValidation, orderErrorNotFound)
	}

	if o.IsComplete() == true {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorValidation, orderErrorOrderAlreadyComplete)
	}

	pm := om.paymentMethodManager.FindById(bson.ObjectIdHex(data[model.OrderPaymentCreateRequestFieldOPaymentMethodId]))

	if pm == nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorValidation, orderErrorPaymentMethodNotFound)
	}

	if o, err = om.modifyOrderAfterOrderFormSubmit(o, pm); err != nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorValidation, err.Error())
	}

	email, _ := data[model.OrderPaymentCreateRequestFieldEmail]
	o.PayerData.Email = &email

	delete(data, model.OrderPaymentCreateRequestFieldOrderId)
	delete(data, model.OrderPaymentCreateRequestFieldOPaymentMethodId)
	delete(data, model.OrderPaymentCreateRequestFieldEmail)

	o.PaymentRequisites = data
	o.PaymentMethodTerminalId = pm.Params.Terminal

	if o, err = om.updateOrderAccountingAmounts(o); err != nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorSystem, err.Error())
	}

	handler, err := om.paymentSystemsSettings.GetPaymentHandler(o, psSettings)

	if err != nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorSystem, err.Error())
	}

	res := handler.CreatePayment()

	if res.Status == payment_system.PaymentStatusOK {
		o.Status = model.OrderStatusPaymentSystemCreate
	}

	if res.Status == payment_system.PaymentStatusErrorSystem {
		o.Status = model.OrderStatusPaymentSystemRejectOnCreate
	}

	if err = om.Database.Repository(TableOrder).UpdateOrder(o); err != nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorSystem, err.Error())
	}

	return res
}

func (om *OrderManager) ProcessNotifyPayment(opn *model.OrderPaymentNotification, psSettings map[string]interface{}) *payment_system.PaymentResponse {
	o := om.FindById(opn.Id)

	if o == nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorValidation, orderErrorNotFound)
	}

	if !o.CanProcessNotify() {
		if o.Status == model.OrderStatusPaymentSystemDeclined {
			return payment_system.NewPaymentResponse(payment_system.PaymentStatusOK, orderErrorOrderDeclined)
		}

		if o.Status == model.OrderStatusPaymentSystemCanceled {
			return payment_system.NewPaymentResponse(payment_system.PaymentStatusOK, orderErrorOrderCanceled)
		}

		return payment_system.NewPaymentResponse(payment_system.PaymentStatusOK, orderErrorOrderAlreadyComplete)
	}

	handler, err := om.paymentSystemsSettings.GetPaymentHandler(o, psSettings)

	if err != nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorSystem, err.Error())
	}

	o.UpdatedAt = time.Now()

	res := handler.ProcessPayment(o, opn)

	if res.Status == payment_system.PaymentStatusTemporary {
		return res
	}

	if res.Status == payment_system.PaymentStatusOK {
		o = res.Order

		if o, err = om.processNotifyPaymentAmounts(o); err != nil {
			return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorSystem, err.Error())
		}
	}

	if _, err = om.UpdateOrder(o); err != nil {
		return payment_system.NewPaymentResponse(payment_system.PaymentStatusErrorSystem, model.ResponseMessageUnknownDbError)
	}

	return res
}

func (om *OrderManager) processNotifyPaymentAmounts(o *model.Order) (*model.Order, error) {
	var err error

	o.PaymentMethodIncomeCurrency = om.currencyManager.FindByCodeA3(o.PaymentMethodIncomeCurrencyA3)

	if o.PaymentMethodIncomeCurrency == nil {
		return nil, errors.New(orderErrorOrderPaymentMethodIncomeCurrencyNotFound)
	}

	o.ProjectOutcomeAmount, err = om.currencyRateManager.convert(
		o.PaymentMethodIncomeCurrency.CodeInt,
		o.ProjectOutcomeCurrency.CodeInt,
		o.PaymentMethodIncomeAmount,
	)

	if err != nil {
		return nil, err
	}

	if om.pspAccountingCurrency == nil {
		return nil, errors.New(orderErrorOrderPSPAccountingCurrencyNotFound)
	}

	o.AmountInPSPAccountingCurrency, err = om.currencyRateManager.convert(
		o.PaymentMethodIncomeCurrency.CodeInt,
		om.pspAccountingCurrency.CodeInt,
		o.PaymentMethodIncomeAmount,
	)

	if err != nil {
		return nil, err
	}

	o.AmountOutMerchantAccountingCurrency, err = om.currencyRateManager.convert(
		o.PaymentMethodIncomeCurrency.CodeInt,
		o.Project.Merchant.Currency.CodeInt,
		o.PaymentMethodIncomeAmount,
	)

	if err != nil {
		return nil, err
	}

	o.AmountInPaymentSystemAccountingCurrency, err = om.currencyRateManager.convert(
		o.PaymentMethodIncomeCurrency.CodeInt,
		o.PaymentMethod.PaymentSystem.AccountingCurrency.CodeInt,
		o.PaymentMethodIncomeAmount,
	)

	if err != nil {
		return nil, err
	}

	return o, nil
}

func (om *OrderManager) validateCreatePaymentData(data map[string]string) error {
	if _, ok := data[model.OrderPaymentCreateRequestFieldOrderId]; !ok {
		return errors.New(orderErrorCreatePaymentRequiredFieldIdNotFound)
	}

	if _, ok := data[model.OrderPaymentCreateRequestFieldOPaymentMethodId]; !ok {
		return errors.New(orderErrorCreatePaymentRequiredFieldPaymentMethodNotFound)
	}

	if _, ok := data[model.OrderPaymentCreateRequestFieldEmail]; !ok {
		return errors.New(orderErrorCreatePaymentRequiredFieldEmailNotFound)
	}

	return nil
}

func (om *OrderManager) modifyOrderAfterOrderFormSubmit(o *model.Order, pm *model.PaymentMethod) (*model.Order, error) {
	if o.PaymentMethod != nil && o.PaymentMethod.Id == pm.Id {
		return o, nil
	}

	p := om.projectManager.FindProjectById(o.Project.Id.Hex())

	if p == nil {
		return nil, errors.New(orderErrorProjectNotFound)
	}

	check := &check{
		order: &model.OrderScalar{
			Amount: o.ProjectIncomeAmount,
		},
		project:       p,
		oCurrency:     o.ProjectIncomeCurrency,
		paymentMethod: pm,
	}

	pmOutData, err := om.checkPaymentMethodLimits(check)

	if err != nil {
		return nil, err
	}

	o.PaymentMethod = &model.OrderPaymentMethod{
		Id:            pm.Id,
		Name:          pm.Name,
		Params:        pm.Params,
		PaymentSystem: pm.PaymentSystem,
		GroupAlias:    pm.GroupAlias,
	}

	o.PaymentMethodOutcomeAmount = FormatAmount(pmOutData.amount)
	o.PaymentMethodOutcomeCurrency = pmOutData.currency

	if o, err = om.ProcessOrderCommissions(o); err != nil {
		return nil, err
	}

	return o, nil
}

func (om *OrderManager) GetRevenueDynamic(rdr *model.RevenueDynamicRequest) (*model.RevenueDynamicResult, error) {
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

	rev := res[0][model.RevenueDynamicFacetFieldRevenue].([]interface{})[0].(map[string]interface{})
	ref := res[0][model.RevenueDynamicFacetFieldRefund].([]interface{})[0].(map[string]interface{})

	rd := &model.RevenueDynamicResult{
		Points: revPoints,
		Revenue: &model.RevenueDynamicMainData{
			Count: rev[model.RevenueDynamicFacetFieldCount].(int),
			Total: FormatAmount(rev[model.RevenueDynamicFacetFieldTotal].(float64)),
			Avg:   FormatAmount(rev[model.RevenueDynamicFacetFieldAvg].(float64)),
		},
		Refund: &model.RevenueDynamicMainData{
			Count: ref[model.RevenueDynamicFacetFieldCount].(int),
			Total: FormatAmount(ref[model.RevenueDynamicFacetFieldTotal].(float64)),
			Avg:   FormatAmount(ref[model.RevenueDynamicFacetFieldAvg].(float64)),
		},
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

// Calculate order accounting amounts and update this amounts in order struct
func (om *OrderManager) updateOrderAccountingAmounts(o *model.Order) (*model.Order, error) {
	var cAmount float64
	var err error

	pmCodeInt := o.PaymentMethodOutcomeCurrency.CodeInt

	// calculate and save order amount in PSP accounting currency
	cAmount, err = om.currencyRateManager.convert(pmCodeInt, om.pspAccountingCurrency.CodeInt, o.PaymentMethodOutcomeAmount)

	if err != nil {
		return nil, err
	}

	o.AmountInPSPAccountingCurrency = FormatAmount(cAmount)

	// calculate and save order amount in merchant accounting currency
	cAmount, err = om.currencyRateManager.convert(o.ProjectIncomeCurrency.CodeInt, o.Project.Merchant.Currency.CodeInt, o.ProjectIncomeAmount)

	if err != nil {
		return nil, err
	}

	o.AmountOutMerchantAccountingCurrency = FormatAmount(cAmount)

	// calculate and save order amount in payment system accounting currency
	cAmount, err = om.currencyRateManager.convert(pmCodeInt, o.PaymentMethod.PaymentSystem.AccountingCurrency.CodeInt, o.PaymentMethodOutcomeAmount)

	if err != nil {
		return nil, err
	}

	o.AmountInPaymentSystemAccountingCurrency = FormatAmount(cAmount)

	return o, nil
}

func (om *OrderManager) UpdateOrder(o *model.Order) (*model.Order, error) {
	err := om.Database.Repository(TableOrder).UpdateOrder(o)

	if err != nil {
		om.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableOrder, err)
		return nil, err
	}

	return o, nil
}
