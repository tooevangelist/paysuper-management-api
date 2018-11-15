package manager

import (
	"errors"
	"fmt"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
	"go.uber.org/zap"
)

const (
	paymentMethodErrorsPaymentSystemsNotFound = "payment systems not found"
	paymentMethodErrorsPaymentSystemNotFound = "payment system with id \"%s\" not found"
)

type PaymentMethodManager struct {
	*Manager
	paymentSystemManager *PaymentSystemManager
}

func InitPaymentMethodManager(database dao.Database, logger *zap.SugaredLogger) *PaymentMethodManager {
	pmm := &PaymentMethodManager{
		Manager: &Manager{Database: database, Logger: logger},
		paymentSystemManager: InitPaymentSystemManager(database, logger),
	}

	return pmm
}

func (pmm *PaymentMethodManager) FindById(id bson.ObjectId) *model.PaymentMethod {
	pm, err := pmm.Database.Repository(TablePaymentMethod).FindPaymentMethodById(id)

	if err != nil {
		pmm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TablePaymentMethod, err)
	}

	return pm
}

func (pmm *PaymentMethodManager) FindAllWithPaymentSystem() ([]*model.PaymentMethod, error) {
	pms, err := pmm.Database.Repository(TablePaymentMethod).FindAllPaymentMethods()

	if err != nil {
		pmm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TablePaymentMethod, err)
	}

	if pms == nil || len(pms) <= 0 {
		return pms, nil
	}

	pss := pmm.paymentSystemManager.FindAll()

	if pss == nil || len(pss) <= 0 {
		return pms, errors.New(paymentMethodErrorsPaymentSystemsNotFound)
	}

	psMap := make(map[bson.ObjectId]*model.PaymentSystem)

	for _, ps := range pss {
		psMap[ps.Id] = ps
	}

	for _, pm := range pms {
		ps, ok := psMap[pm.PaymentSystemId]

		if !ok {
			return pms, errors.New(fmt.Sprintf(paymentMethodErrorsPaymentSystemNotFound, pm.PaymentSystemId))
		}

		pm.PaymentSystem = ps
	}

	return pms, nil
}

func (pmm *PaymentMethodManager) FindAllWithPaymentSystemAsMap() (map[bson.ObjectId]*model.PaymentMethod, error) {
	pms, err := pmm.FindAllWithPaymentSystem()

	if err != nil {
		return nil, err
	}

	pmsMap := make(map[bson.ObjectId]*model.PaymentMethod)

	for _, pm := range pms {
		pmsMap[pm.Id] = pm
	}

	return pmsMap, nil
}
