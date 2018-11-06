package manager

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
	"net"
)

const (
	orderErrorProjectNotFound               = "project with specified identifier not found"
	orderErrorProjectInactive               = "project with specified identifier is inactive"
	orderErrorPaymentMethodNotAllowed       = "payment method not specified for project"
	orderErrorPaymentMethodNotFound         = "payment method with specified not found"
	orderErrorPaymentMethodInactive         = "payment method with specified is inactive"
	orderErrorPaymentSystemNotFound         = "payment system for specified payment method not found"
	orderErrorPaymentSystemInactive         = "payment system for specified payment method is inactive"
	orderErrorPayerRegionUnknown            = "payer region can't be found"
	orderErrorFixedPackageForRegionNotFound = "project not have fixed packages for payer region"
	orderErrorFixedPackageNotFound          = "project not have fixed package with specified amount or currency"
	orderErrorProjectOrderIdIsDuplicate     = "request with specified project order identifier processed early"
	orderErrorDynamicNotifyUrlsNotAllowed   = "dynamic verify url or notify url not allowed for project"
	orderErrorDynamicRedirectUrlsNotAllowed = "dynamic payer redirect urls not allowed for project"
)

type OrderManager struct {
	*Manager

	geoDbReader          *geoip2.Reader
	projectManager       *ProjectManager
	paymentSystemManager *PaymentSystemManager
	paymentMethodManager *PaymentMethodManager
}

func InitOrderManager(database dao.Database, logger *zap.SugaredLogger, geoDbReader *geoip2.Reader) *OrderManager {
	om := &OrderManager{
		Manager:              &Manager{Database: database, Logger: logger},
		geoDbReader:          geoDbReader,
		projectManager:       InitProjectManager(database, logger),
		paymentSystemManager: InitPaymentSystemManager(database, logger),
		paymentMethodManager: InitPaymentMethodManager(database, logger),
	}

	return om
}

func (om *OrderManager) Validate(order *model.OrderScalar) error {
	p := om.projectManager.FindProjectById(order.ProjectId)

	if p == nil {
		return errors.New(orderErrorProjectNotFound)
	}

	if p.IsActive == true {
		return errors.New(orderErrorProjectInactive)
	}

	if order.PaymentMethod != nil {
		opm, err := om.getPaymentMethod(order, p.PaymentMethods)

		if err != nil {
			return err
		}

		pm := om.paymentMethodManager.FindById(opm.Id)

		if pm == nil {
			return errors.New(orderErrorPaymentMethodNotFound)
		}

		if pm.IsActive == false {
			return errors.New(orderErrorPaymentMethodInactive)
		}

		ps := om.paymentSystemManager.FindById(pm.PaymentSystemId)

		if ps == nil {
			return errors.New(orderErrorPaymentSystemNotFound)
		}

		if ps.IsActive == false {
			return errors.New(orderErrorPaymentSystemInactive)
		}
	}

	if p.OnlyFixedAmounts == true {
		region := *order.Region

		if region == "" {
			ip := net.ParseIP(order.CreateOrderIp)
			gRecord, err := om.geoDbReader.City(ip)

			if err != nil {
				return errors.New(orderErrorPayerRegionUnknown)
			}

			region = gRecord.Country.IsoCode
		}

		fps, ok := p.FixedPackage[region]

		if !ok || len(fps) <= 0 {
			return errors.New(orderErrorFixedPackageForRegionNotFound)
		}

		var ofp *model.FixedPackage

		for _, fp := range fps {
			if fp.Price != order.Amount || fp.Currency.CodeA3 != *order.Currency {
				continue
			}

			ofp = fp
		}

		if ofp == nil {
			return errors.New(orderErrorFixedPackageNotFound)
		}
	}

	if order.OrderId != nil {
		if err := om.checkProjectOrderIdUnique(order); err != nil {
			return err
		}
	}

	if (order.UrlVerify != nil || order.UrlNotify != nil) && p.IsAllowDynamicNotifyUrls == false {
		return errors.New(orderErrorDynamicNotifyUrlsNotAllowed)
	}

	if (order.UrlSuccess != nil || order.UrlFail != nil) && p.IsAllowDynamicRedirectUrls == false {
		return errors.New(orderErrorDynamicRedirectUrlsNotAllowed)
	}

	return nil
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
