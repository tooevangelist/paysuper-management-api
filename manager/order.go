package manager

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
	"net"
	"sort"
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
		gRecord, err = om.getOrderFixedPackage(check)

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

	nOrder := &model.Order{
		Id:                    bson.NewObjectId(),
		ProjectId:             p.Id,
		ProjectOrderId:        order.OrderId,
		ProjectAccount:        order.Account,
		ProjectIncomeAmount:   order.Amount,
		ProjectIncomeCurrency: oCurrency,
		ProjectParams:         order.Other,
		PayerData: &model.PayerData{
			Ip:            order.CreateOrderIp,
			CountryCodeA3: gRecord.Country.IsoCode,
			City:          gRecord.City.Names["en"],
			Timezone:      gRecord.Location.TimeZone,
			Phone:         order.PayerPhone,
			Email:         order.PayerEmail,
		},
		PaymentMethodId:              pm.Id,
		PaymentMethodOutcomeAmount:   pmOutcomeData.amount,
		PaymentMethodOutcomeCurrency: pmOutcomeData.currency,
		Status:                       model.OrderStatusCreated,
		CreatedAt:                    time.Now(),
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

func (om *OrderManager) GetCardYears() []int {
	var years []int

	start := time.Now().Year()
	end := start + 10

	for i := start; i < end; i++ {
		years = append(years, i)
	}

	return years
}

func (om *OrderManager) GetCardMonths() map[string]string {
	return map[string]string{
		"01": "January",
		"02": "February",
		"03": "March",
		"04": "April",
		"05": "May",
		"06": "June",
		"07": "July",
		"08": "August",
		"09": "September",
		"10": "October",
		"11": "November",
		"12": "December",
	}
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

	ps := om.paymentSystemManager.FindById(pm.PaymentSystemId)

	if ps == nil {
		return nil, errors.New(orderErrorPaymentSystemNotFound)
	}

	if ps.IsActive == false {
		return nil, errors.New(orderErrorPaymentSystemInactive)
	}

	return pm, nil
}

func (om *OrderManager) getOrderFixedPackage(c *check) (*geoip2.City, error) {
	var region string

	if c.order.Region != nil {
		region = *c.order.Region
	}

	ip := net.ParseIP(c.order.CreateOrderIp)
	gRecord, err := om.geoDbReader.City(ip)

	if err != nil {
		return nil, errors.New(orderErrorPayerRegionUnknown)
	}

	if region == "" {
		region = gRecord.Country.IsoCode
	}

	fps, ok := c.project.FixedPackage[region]

	if !ok || len(fps) <= 0 {
		return nil, errors.New(orderErrorFixedPackageForRegionNotFound)
	}

	var ofp *model.FixedPackage

	for _, fp := range fps {
		if fp.Price != c.order.Amount || fp.Currency.CodeA3 != *c.order.Currency {
			continue
		}

		ofp = fp
	}

	if ofp == nil {
		return nil, errors.New(orderErrorFixedPackageNotFound)
	}

	return gRecord, nil
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
