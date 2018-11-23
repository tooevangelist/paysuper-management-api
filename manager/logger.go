package manager

import "github.com/ProtocolONE/p1pay.api/database/model"

type LoggerManager Manager

func (lm *LoggerManager) Insert(log *model.Log) {
	err := lm.Database.Repository(TableLog).InsertLog(log)

	if err != nil {
		return
	}

	return
}