package database

import (
	"github.com/ProtocolONE/p1payments.api/config"
	"github.com/ProtocolONE/p1payments.api/database/dao"
	"github.com/ProtocolONE/p1payments.api/database/dao/mongo"
)

func NewConnection(config *config.Database) (dao.Database, error) {
	settings := mongo.Connection{
		Host:     config.Host,
		Database: config.Database,
		User:     config.User,
		Password: config.Password,
	}

	return mongo.Open(settings)
}
