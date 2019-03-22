package manager

import "github.com/paysuper/paysuper-management-api/database/model"

type LoggerManager Manager

func (lm *LoggerManager) Insert(log *model.Log) {
	err := lm.Database.Repository(TableLog).InsertLog(log)

	if err != nil {
		return
	}

	return
}
