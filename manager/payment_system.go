package manager

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/dao"
	"github.com/paysuper/paysuper-management-api/database/model"
	"go.uber.org/zap"
)

type PaymentSystemManager struct {
	*Manager
}

func InitPaymentSystemManager(database dao.Database, logger *zap.SugaredLogger) *PaymentSystemManager {
	psm := &PaymentSystemManager{
		Manager: &Manager{Database: database, Logger: logger},
	}

	return psm
}

func (psm *PaymentSystemManager) FindById(id bson.ObjectId) *model.PaymentSystem {
	ps, err := psm.Database.Repository(TablePaymentSystem).FindPaymentSystemById(id)

	if err != nil {
		psm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TablePaymentSystem, err)
	}

	return ps
}

func (psm *PaymentSystemManager) FindAll() []*model.PaymentSystem {
	pss, err := psm.Database.Repository(TablePaymentSystem).FindAllPaymentSystem()

	if err != nil {
		psm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TablePaymentSystem, err)
	}

	return pss
}
