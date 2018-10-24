package repository_mongo

import (
	"fmt"
	"github.com/ProtocolONE/p1pay.api/database/dao"
	"github.com/ProtocolONE/p1pay.api/database/dao/mongo"
)

type CurrencyRepository struct {
	database *dao.Database
}

func (mgo *mongo.Source) CurrencyRepository() {
	fmt.Println("Ja tut")
}
