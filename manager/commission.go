package manager

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
	"go.uber.org/zap"
)

const (
	commissionNotFound = "commission not found for specified project and payment method"
)

type CommissionManager Manager

func InitCommissionManager(database dao.Database, logger *zap.SugaredLogger) *CommissionManager {
	cm := &CommissionManager{
		Database: database,
		Logger:   logger,
	}

	return cm
}

func (cm *CommissionManager) CalculateCommission(projectId bson.ObjectId, pmId bson.ObjectId, amount float64) (*model.CommissionOrder, error) {
	commission, err := cm.Database.Repository(TableCommission).FindCommissionByProjectIdAndPaymentMethodId(projectId, pmId)

	if err != nil {
		cm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableCommission, err)
	}

	if commission == nil {
		return nil, errors.New(commissionNotFound)
	}

	cOrder := &model.CommissionOrder{
		PMCommission:     amount * (commission.PaymentMethodCommission / 100),
		PspCommission:    amount * (commission.PspCommission / 100),
		ToUserCommission: amount * (commission.TotalCommissionToUser / 100),
	}

	return cOrder, nil
}
