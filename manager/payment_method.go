package manager

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
	"go.uber.org/zap"
)

type PaymentMethodManager struct {
	*Manager
}

func InitPaymentMethodManager(database dao.Database, logger *zap.SugaredLogger) *PaymentMethodManager {
	pmm := &PaymentMethodManager{
		Manager: &Manager{Database: database, Logger: logger},
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
