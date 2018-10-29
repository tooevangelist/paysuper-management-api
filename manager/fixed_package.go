package manager

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"go.uber.org/zap"
)

type FixedPackageManager Manager

func InitFixedPackageManager(database dao.Database, logger *zap.SugaredLogger) *FixedPackageManager {
	fpm := &FixedPackageManager{
		Database: database,
		Logger: logger,
	}

	return fpm
}
