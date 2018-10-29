package database

import (
	"github.com/ProtocolONE/p1pay.api/config"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/dao/mongo"
	_ "github.com/ProtocolONE/p1pay.api/database/migrations"
	"github.com/globalsign/mgo"
	"github.com/xakep666/mongo-migrate"
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

func Migrate(db *mgo.Database, direction string) error {
	var err error

	migrate.SetDatabase(db)

	if direction == "up" {
		err = migrate.Up(migrate.AllAvailable)
	}

	if direction == "down" {
		err = migrate.Down(migrate.AllAvailable)
	}

	return err
}
