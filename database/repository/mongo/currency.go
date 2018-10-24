package mongo

import (
	"fmt"
	"github.com/ProtocolONE/p1pay.api/database/dao"
)

type CurrencyRepository struct {
	database *dao.Database
}

func InitCyrrencyRepository() {
	fmt.Println("Ja tut")
}


