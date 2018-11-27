package manager

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/globalsign/mgo/bson"
	"go.uber.org/zap"
)

type CommissionManager Manager

func InitCommissionManager(database dao.Database, logger *zap.SugaredLogger) *CommissionManager {
	cm := &CommissionManager{
		Database: database,
		Logger:   logger,
	}

	return cm
}

func (cm *CommissionManager) CalculateCommission(pmId bson.ObjectId, projectId bson.ObjectId, amount float64) {

}
