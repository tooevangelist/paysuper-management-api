package manager

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/payment_system"
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

	orderSignatureElementsGlue = "|"

	orderDefaultDescription = "Payment by order # %s"
)

type OrderManager struct {
	*Manager

	geoDbReader          *geoip2.Reader
	projectManager       *ProjectManager
	paymentSystemManager *PaymentSystemManager
	paymentMethodManager *PaymentMethodManager
	currencyRateManager  *CurrencyRateManager
	currencyManager      *CurrencyManager
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

func InitOrderManager(database dao.Database, logger *zap.SugaredLogger, geoDbReader *geoip2.Reader) *OrderManager {
	om := &OrderManager{
		Manager:              &Manager{Database: database, Logger: logger},
		geoDbReader:          geoDbReader,
		projectManager:       InitProjectManager(database, logger),
		paymentSystemManager: InitPaymentSystemManager(database, logger),
		paymentMethodManager: InitPaymentMethodManager(database, logger),
		currencyRateManager:  InitCurrencyRateManager(database, logger),
		currencyManager:      InitCurrencyManager(database, logger),
	}

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

	check := &check{order: order, project: p, oCurrency: oCurrency}

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

	if pmOutcomeData, err = om.checkPaymentMethodLimits(check); err != nil {
		return nil, err
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
	id := bson.NewObjectId()

	nOrder := &model.Order{
		Id:                    id,
		ProjectId:             p.Id,
		Description:           fmt.Sprintf(orderDefaultDescription, id.Hex()),
		ProjectOrderId:        order.OrderId,
		ProjectAccount:        order.Account,
		ProjectIncomeAmount:   FormatAmount(order.Amount),
		ProjectIncomeCurrency: oCurrency,
		ProjectParams:         order.Other,
		PayerData: &model.PayerData{
			Ip:            order.CreateOrderIp,
			CountryCodeA2: gRecord.Country.IsoCode,
			CountryName:   &model.Name{EN: gRecord.Country.Names["en"], RU: gRecord.Country.Names["ru"]},
			City:          &model.Name{EN: gRecord.City.Names["en"], RU: gRecord.City.Names["ru"]},
			Timezone:      gRecord.Location.TimeZone,
			Phone:         order.PayerPhone,
			Email:         order.PayerEmail,
		},
		PaymentMethod: &model.OrderPaymentMethod{
			Id:            pm.Id,
			Name:          pm.Name,
			Params:        pm.Params,
			PaymentSystem: pm.PaymentSystem,
		},
		PaymentMethodOutcomeAmount:         FormatAmount(pmOutcomeData.amount),
		PaymentMethodOutcomeCurrency:       pmOutcomeData.currency,
		Status:                             model.OrderStatusCreated,
		CreatedAt:                          time.Now(),
		IsJsonRequest:                      order.IsJsonRequest,
		FixedPackage:                       ofp,
		AmountInMerchantAccountingCurrency: FormatAmount(mACAmount),
	}

	if order.Description != nil {
		nOrder.Description = *order.Description
	}

	if err = om.Database.Repository(TableOrder).InsertOrder(nOrder); err != nil {
		om.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableOrder, err)

		return nil, errors.New(orderErrorCanNotCreate)
	}

	return nOrder, nil
}

func (om *OrderManager) FindById(id string) *model.Order {
	o, err := om.Database.Repository(TableOrder).FindOrderById(bson.ObjectIdHex(id))

	if err != nil {
		om.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableOrder, err)
	}

	if o != nil {
		if p := om.projectManager.FindProjectById(o.ProjectId.Hex()); p != nil {
			o.ProjectData = p
		}

		o.ProjectOutcomeAmountPrintable = fmt.Sprintf("%.2f", o.PaymentMethodOutcomeAmount)
		o.OrderIdPrintable = o.Id.Hex()
	}

	return o
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
	var filter = make(bson.M)
	var f bson.M
	var pFilter []bson.ObjectId

	for k := range params.Projects {
		pFilter = append(pFilter, k)
	}

	filter["project_id"] = bson.M{"$in": pFilter}

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
				Id:   oValue.ProjectId,
				Name: params.Projects[oValue.ProjectId],
			},
			Account:        oValue.ProjectAccount,
			ProjectOrderId: oValue.ProjectOrderId,
			PayerData:      oValue.PayerData,
			PaymentMethod: &model.SimpleItem{
				Id:   oValue.PaymentMethod.Id,
				Name: oValue.PaymentMethod.Name,
			},
			ProjectTechnicalIncome: &model.OrderSimpleAmountObject{
				Amount: oValue.ProjectIncomeAmount,
				Currency: &model.SimpleCurrency{
					CodeInt: oValue.ProjectIncomeCurrency.CodeInt,
					CodeA3:  oValue.ProjectIncomeCurrency.CodeA3,
					Name:    oValue.ProjectIncomeCurrency.Name,
				},
			},
			ProjectAccountingIncome: &model.OrderSimpleAmountObject{
				Amount: oValue.AmountInMerchantAccountingCurrency,
				Currency: &model.SimpleCurrency{
					CodeInt: params.Merchant.Currency.CodeInt,
					CodeA3:  params.Merchant.Currency.CodeA3,
					Name:    params.Merchant.Currency.Name,
				},
			},
			FixedPackage: oValue.FixedPackage,
			Status: &model.Status{
				Status:      oValue.Status,
				Description: model.OrderStatusesDescription[oValue.Status],
			},
			CreatedAt:   oValue.CreatedAt,
			ConfirmedAt: oValue.PaymentMethodOrderClosedAt,
			ClosedAt:    oValue.ProjectLastRequestedAt,
		}

		if oValue.AmountInPaymentSystemAccountingCurrency > 0 {
			tOrder.PaymentSystemTechnicalIncome = &model.OrderSimpleAmountObject{
				Amount: oValue.AmountInPaymentSystemAccountingCurrency,
				Currency: &model.SimpleCurrency{
					CodeInt: oValue.PaymentMethod.PaymentSystem.AccountingCurrency.CodeInt,
					CodeA3:  oValue.PaymentMethod.PaymentSystem.AccountingCurrency.CodeA3,
					Name:    oValue.PaymentMethod.PaymentSystem.AccountingCurrency.Name,
				},
			}
		}

		if oValue.AmountOutMerchantAccountingCurrency > 0 {
			tOrder.ProjectAccountingOutcome = &model.OrderSimpleAmountObject{
				Amount: oValue.AmountOutMerchantAccountingCurrency,
				Currency: &model.SimpleCurrency{
					CodeInt: params.Merchant.Currency.CodeInt,
					CodeA3:  params.Merchant.Currency.CodeA3,
					Name:    params.Merchant.Currency.Name,
				},
			}
		}

		if oValue.ProjectOutcomeAmount > 0 {
			tOrder.ProjectTechnicalOutcome = &model.OrderSimpleAmountObject{
				Amount: oValue.ProjectOutcomeAmount,
				Currency: &model.SimpleCurrency{
					CodeInt: oValue.ProjectOutcomeCurrency.CodeInt,
					CodeA3:  oValue.ProjectOutcomeCurrency.CodeA3,
					Name:    oValue.ProjectOutcomeCurrency.Name,
				},
			}
		}

		tOrders = append(tOrders, tOrder)
	}

	return tOrders, nil
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

		filter["pm_id"] = bson.M{"$in": fPms}
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

func (om *OrderManager) ProcessCreatePayment(o *model.Order, psSettings map[string]interface{}) error {
	err := om.Database.Repository(TableOrder).UpdateOrder(o)

	if err != nil {
		return err
	}

	handler, err := payment_system.GetPaymentHandler(o, psSettings)

	if err != nil {
		return err
	}

	err = handler.CreatePayment()

	if err != nil {
		return err
	}

	return nil
}
