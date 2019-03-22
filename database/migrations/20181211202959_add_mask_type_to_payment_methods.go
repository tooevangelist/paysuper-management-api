package migrations

import (
	"github.com/globalsign/mgo"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/xakep666/mongo-migrate"
	"gopkg.in/mgo.v2/bson"
)

type PaymentMethodMaskType struct {
	Mask       string
	Type       string
	GroupAlias string
}

var paymentMethodMaskTypes = []*PaymentMethodMaskType{
	{
		GroupAlias: "bank_card",
		Mask:       "^(?:4[0-9]{12}(?:[0-9]{3})?|[25][1-7][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11})$",
		Type:       "bank_card",
	},
	{
		GroupAlias: "qiwi",
		Mask:       "^\\d{1,15}",
		Type:       "ewallet",
	},
	{
		GroupAlias: "webmoney",
		Mask:       "^[ZERB][0-9]{12}$",
		Type:       "ewallet",
	},
	{
		GroupAlias: "neteller",
		Mask:       "^([a-zA-Z0-9_.+-])+\\@(([a-zA-Z0-9-])+\\.)+([a-zA-Z0-9]{2,4})+$",
		Type:       "ewallet",
	},
	{
		GroupAlias: "alipay",
		Mask:       "^.*$",
		Type:       "ewallet",
	},
	{
		GroupAlias: "bitcoin",
		Mask:       "^[13][a-km-zA-HJ-NP-Z1-9]{25,34}$",
		Type:       "crypto",
	},
}

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			c := db.C(manager.TablePaymentMethod)

			for _, v := range paymentMethodMaskTypes {
				var pms []*model.PaymentMethod

				if err := c.Find(bson.M{"group_alias": v.GroupAlias}).All(&pms); err != nil {
					return err
				}

				for _, pm := range pms {
					pm.Type = v.Type
					pm.AccountRegexp = v.Mask

					if err := c.UpdateId(pm.Id, pm); err != nil {
						return err
					}
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
