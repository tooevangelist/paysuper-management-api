package manager

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/dao"
	"github.com/paysuper/paysuper-management-api/database/model"
	"go.uber.org/zap"
	"time"
)

type MerchantManager Manager

func InitMerchantManager(database dao.Database, logger *zap.SugaredLogger) *MerchantManager {
	return &MerchantManager{Database: database, Logger: logger}
}

func (mm *MerchantManager) FindById(id string) *model.Merchant {
	m, err := mm.Database.Repository(TableMerchant).FindMerchantById(id)

	if err != nil {
		mm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableMerchant, err)
	}

	return m
}

func (mm *MerchantManager) Create(ms *model.MerchantScalar) (*model.Merchant, error) {
	m := &model.Merchant{
		Id:         bson.NewObjectId(),
		Email:      *ms.Email,
		ExternalId: ms.Id,
		Status:     model.MerchantStatusCreated,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if ms.Currency != nil {
		cur, err := mm.Database.Repository(TableCurrency).FindCurrencyById(*ms.Currency)

		if err == nil {
			m.Currency = cur
		}
	}

	if ms.Country != nil {
		ctr, err := mm.Database.Repository(TableCountry).FindCountryById(*ms.Country)

		if err == nil {
			m.Country = ctr
		}
	}

	if ms.Name != nil {
		m.Name = ms.Name
	}

	if ms.AccountingPeriod != nil {
		m.AccountingPeriod = ms.AccountingPeriod
	}

	err := mm.Database.Repository(TableMerchant).InsertMerchant(m)

	if err != nil {
		mm.Logger.Errorf("Query from table \"%s\" ended with error: %s", TableMerchant, err)
	}

	return m, err
}

func (mm *MerchantManager) Update(m *model.Merchant, mn *model.MerchantScalar) (*model.Merchant, error) {
	if mn.Currency != nil && (m.Currency == nil || m.Currency.CodeInt != *mn.Currency) {
		cur, err := mm.Database.Repository(TableCurrency).FindCurrencyById(*mn.Currency)

		if err == nil {
			m.Currency = cur
		}
	}

	if mn.Country != nil && (m.Country == nil || m.Country.CodeInt != *mn.Country) {
		ctr, err := mm.Database.Repository(TableCountry).FindCountryById(*mn.Country)

		if err == nil {
			m.Country = ctr
		}
	}

	if mn.Email != nil && m.Email != *mn.Email {
		m.Email = *mn.Email
	}

	if mn.Name != nil && m.Name != mn.Name {
		m.Name = mn.Name
	}

	if mn.AccountingPeriod != nil && m.AccountingPeriod != mn.AccountingPeriod {
		m.AccountingPeriod = mn.AccountingPeriod
	}

	if mm.IsComplete(m) {
		m.Status = model.MerchantStatusCompleted
	}

	m.UpdatedAt = time.Now()

	err := mm.Database.Repository(TableMerchant).UpdateMerchant(m)

	if err != nil {
		return nil, err
	}

	return m, nil
}

func (mm *MerchantManager) Delete(m *model.Merchant) error {
	m.Status = model.MerchantStatusDeleted

	return mm.Database.Repository(TableMerchant).UpdateMerchant(m)
}

func (mm *MerchantManager) IsComplete(m *model.Merchant) bool {
	return m.ExternalId != "" && m.Name != nil && m.Country != nil && m.Currency != nil && m.AccountingPeriod != nil
}
