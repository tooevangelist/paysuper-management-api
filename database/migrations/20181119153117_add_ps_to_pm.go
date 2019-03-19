package migrations

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/xakep666/mongo-migrate"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			ps := &model.PaymentSystem{}
			if err := db.C(manager.TablePaymentSystem).Find(bson.M{"name": "CardPay"}).One(ps); err != nil {
				return err
			}

			pm := &model.PaymentMethod{}
			if err := db.C(manager.TablePaymentMethod).Find(bson.M{"group_alias": "bank_card"}).One(&pm); err != nil {
				return err
			}

			pm.PaymentSystem = ps

			return db.C(manager.TablePaymentMethod).UpdateId(pm.Id, pm)
		},
		func(db *mgo.Database) error {
			pm := &model.PaymentMethod{}
			if err := db.C(manager.TablePaymentMethod).Find(bson.M{"group_alias": "bank_card"}).One(&pm); err != nil {
				return err
			}

			pm.PaymentSystem = nil

			return db.C(manager.TablePaymentMethod).UpdateId(pm.Id, pm)
		},
	)

	if err != nil {
		return
	}
}
