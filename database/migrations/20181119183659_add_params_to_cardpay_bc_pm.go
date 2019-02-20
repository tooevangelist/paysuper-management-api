package migrations

import (
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/xakep666/mongo-migrate"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			pm := &model.PaymentMethod{}
			if err := db.C(manager.TablePaymentMethod).Find(bson.M{"group_alias": "bank_card"}).One(&pm); err != nil {
				return err
			}

			pm.Params = &model.PaymentMethodParams{
				Handler: "cardpay",
				Terminal: "15985",
				ExternalId: "BANKCARD",
			}

			return db.C(manager.TablePaymentMethod).UpdateId(pm.Id, pm)
		},
		func(db *mgo.Database) error {
			pm := &model.PaymentMethod{}
			if err := db.C(manager.TablePaymentMethod).Find(bson.M{"group_alias": "bank_card"}).One(&pm); err != nil {
				return err
			}

			pm.Params = nil

			return db.C(manager.TablePaymentMethod).UpdateId(pm.Id, pm)
		},
	)

	if err != nil {
		return
	}
}
