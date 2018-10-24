package manager

import (
	"github.com/ProtocolONE/p1pay.api/database/dao"
)

type CurrencyManager struct {
	database dao.Database
}

func NewCurrencyManager(database dao.Database) *CurrencyManager {
	return &CurrencyManager{ database: database }
}
