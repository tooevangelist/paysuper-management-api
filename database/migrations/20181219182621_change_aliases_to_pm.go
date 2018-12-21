package migrations

import (
	"errors"
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/xakep666/mongo-migrate"
)

var aliases = map[string]string{

}

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			/*var pms []*model.PaymentMethod

			err := db.C(manager.TablePaymentMethod).Find(bson.M{}).All(&pms)

			if err != nil {
				return err
			}

			for _, v := range pms {

			}

			err := db.C(manager.TablePaymentMethod).Insert(pms...)

			if err != nil {
				return err
			}

			var pm *model.PaymentMethod

			if err := db.C(manager.TablePaymentMethod).Find(bson.M{"name": "Bank card"}).One(&pm); err != nil {
				return err
			}

			pm.Params = &model.PaymentMethodParams{
				Handler:    payment_system.PaymentSystemHandlerCardPay,
				Terminal:   "15985",
				ExternalId: "BANKCARD",
			}
			pm.Icon = "/images/bank_card_logo.png"

			return db.C(manager.TablePaymentMethod).UpdateId(pm.Id, pm)*/
			return nil
		},
		func(db *mgo.Database) error {
			var pms []*model.PaymentMethod

			err := db.C(manager.TablePaymentMethod).Find(bson.M{"name": bson.M{"$in": []string{"Qiwi", "WebMoney", "Neteller"}}}).All(&pms)

			if err != nil {
				return err
			}

			if len(pms) < 3 {
				return errors.New("payment methods not found")
			}

			return db.C(manager.TablePaymentMethod).Remove(pms)
		},
	)

	if err != nil {
		return
	}
}
