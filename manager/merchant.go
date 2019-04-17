package manager

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-management-api/database/dao"
	"github.com/paysuper/paysuper-management-api/database/model"
	"go.uber.org/zap"
	"time"
)

type MerchantManager Manager

func InitMerchantManager(database dao.Database, logger *zap.SugaredLogger) *MerchantManager {
	return &MerchantManager{Database: database, Logger: logger}
}

func (mm *MerchantManager) FindById(id string) *billing.Merchant {
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

func (mm *MerchantManager) Update(m *billing.Merchant, mn *model.MerchantScalar) (*billing.Merchant, error) {
	if mn.Currency != nil && (m.Banking.Currency == nil || int(m.Banking.Currency.CodeInt) != *mn.Currency) {
		c, err := mm.Database.Repository(TableCurrency).FindCurrencyById(*mn.Currency)

		if err == nil {
			m.Banking.Currency = &billing.Currency{
				CodeInt:  int32(c.CodeInt),
				CodeA3:   c.CodeA3,
				Name:     &billing.Name{En: c.Name.EN, Ru: c.Name.RU},
				IsActive: c.IsActive,
			}
		}
	}

	if mn.Country != nil && (m.Country == nil || int(m.Country.CodeInt) != *mn.Country) {
		ctr, err := mm.Database.Repository(TableCountry).FindCountryById(*mn.Country)

		if err == nil {
			m.Country = &billing.Country{
				CodeInt:  int32(ctr.CodeInt),
				CodeA2:   ctr.CodeA2,
				CodeA3:   ctr.CodeA3,
				Name:     &billing.Name{En: ctr.Name.EN, Ru: ctr.Name.RU},
				IsActive: ctr.IsActive,
			}
		}
	}

	if mn.Email != nil && m.User.Email != *mn.Email {
		m.User.Email = *mn.Email
	}

	if mn.Name != nil && m.Name != *mn.Name {
		m.Name = *mn.Name
	}

	return m, nil
}

func (mm *MerchantManager) Delete(m *billing.Merchant) error {
	m.Status = model.MerchantStatusDeleted

	return mm.Database.Repository(TableMerchant).UpdateMerchant(m)
}

func (mm *MerchantManager) IsComplete(m *model.Merchant) bool {
	return m.ExternalId != "" && m.Name != nil && m.Country != nil && m.Currency != nil && m.AccountingPeriod != nil
}
