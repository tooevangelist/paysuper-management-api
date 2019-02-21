package migrations

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-recurring-repository/pkg/constant"
	"github.com/xakep666/mongo-migrate"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			var pms []*model.PaymentMethod

			c := db.C("payment_method")
			err := c.Find(bson.M{}).All(&pms)

			if err != nil {
				return err
			}

			for _, pm := range pms {
				switch pm.GroupAlias {
				case "bank_card":
					pm.GroupAlias = constant.PaymentSystemGroupAliasBankCard
					break
				case "qiwi":
					pm.GroupAlias = constant.PaymentSystemGroupAliasQiwi
					break
				case "webmoney":
					pm.GroupAlias = constant.PaymentSystemGroupAliasWebMoney
					break
				case "neteller":
					pm.GroupAlias = constant.PaymentSystemGroupAliasNeteller
					break
				case "alipay":
					pm.GroupAlias = constant.PaymentSystemGroupAliasAlipay
					break
				case "bitcoin":
					pm.GroupAlias = constant.PaymentSystemGroupAliasBitcoin
					break
				}

				if err := c.UpdateId(pm.Id, pm); err != nil {
					return err
				}
			}

			return nil
		},
		func(db *mgo.Database) error {
			var pms []*model.PaymentMethod

			c := db.C("payment_method")
			err := c.Find(bson.M{}).All(&pms)

			if err != nil {
				return err
			}

			for _, pm := range pms {
				switch pm.GroupAlias {
				case constant.PaymentSystemGroupAliasBankCard:
					pm.GroupAlias = "bank_card"
					break
				case constant.PaymentSystemGroupAliasQiwi:
					pm.GroupAlias = "qiwi"
					break
				case constant.PaymentSystemGroupAliasWebMoney:
					pm.GroupAlias = "webmoney"
					break
				case constant.PaymentSystemGroupAliasNeteller:
					pm.GroupAlias = "neteller"
					break
				case constant.PaymentSystemGroupAliasAlipay:
					pm.GroupAlias = "alipay"
					break
				case constant.PaymentSystemGroupAliasBitcoin:
					pm.GroupAlias = "bitcoin"
					break
				}

				if err := c.UpdateId(pm.Id, pm); err != nil {
					return err
				}
			}

			return nil
		},
	)

	if err != nil {
		return
	}
}
