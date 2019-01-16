package migrations

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/xakep666/mongo-migrate"
)

var pmCurrencies = []string{"USD", "RUB", "EUR", "GBP"}

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			var intCurrencies []int32
			var cur []*model.Currency
			var pms []*model.PaymentMethod

			err := db.C(manager.TableCurrency).Find(bson.M{"code_a3": bson.M{"$in": pmCurrencies}}).All(&cur)

			if err != nil {
				return err
			}

			for _, c := range cur {
				intCurrencies = append(intCurrencies, int32(c.CodeInt))
			}

			pmCol := db.C(manager.TablePaymentMethod)
			err = pmCol.Find(bson.M{}).All(&pms)

			if err != nil {
				return err
			}

			for _, pm := range pms {
				pm.Currencies = intCurrencies

				if err := pmCol.UpdateId(pm.Id, pm); err != nil {
					return err
				}
			}

			return nil
		},
		func(db *mgo.Database) error {
			return nil
		},
	)

	if err != nil {
		return
	}
}
