package migrations

import (
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/xakep666/mongo-migrate"
	"time"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			c := &model.Country{}
			if err := db.C(manager.TableCountry).Find(bson.M{"code_a3": "CYP"}).One(&c); err != nil {
				return err
			}

			cr := &model.Currency{}
			if err := db.C(manager.TableCurrency).Find(bson.M{"code_a3": "EUR"}).One(&cr); err != nil {
				return err
			}

			ps := model.PaymentSystem{
				Id:                 bson.NewObjectId(),
				Name:               "CardPay",
				Country:            c,
				AccountingCurrency: cr,
				AccountingPeriod:   "2week",
				IsActive:           true,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}

			return db.C(manager.TablePaymentSystem).Insert(ps)
		},
		func(db *mgo.Database) error {
			return db.C(manager.TablePaymentSystem).Remove(bson.M{"name": "CardPay"})
		},
	)

	if err != nil {
		return
	}
}
