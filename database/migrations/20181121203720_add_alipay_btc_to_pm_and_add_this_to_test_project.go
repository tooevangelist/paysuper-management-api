package migrations

import (
	"errors"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/paysuper/paysuper-management-api/payment_system"
	"github.com/xakep666/mongo-migrate"
	"time"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			cr := &model.Currency{}
			if err := db.C(manager.TableCurrency).Find(bson.M{"code_a3": "USD"}).One(&cr); err != nil {
				return err
			}

			ps := &model.PaymentSystem{}
			if err := db.C(manager.TablePaymentSystem).Find(bson.M{"name": "CardPay"}).One(ps); err != nil {
				return err
			}

			pms := []interface{}{
				&model.PaymentMethod{
					Id:               bson.NewObjectId(),
					Name:             "Alipay",
					PaymentSystem:    ps,
					Currency:         cr,
					GroupAlias:       "alipay",
					MinPaymentAmount: 0.01,
					MaxPaymentAmount: 15000.00,
					IsActive:         true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
					Params: &model.PaymentMethodParams{
						Handler:    payment_system.PaymentSystemHandlerCardPay,
						Terminal:   "16001",
						ExternalId: "ALIPAY",
					},
					Icon: "/images/alipay_logo.png",
				},
				&model.PaymentMethod{
					Id:               bson.NewObjectId(),
					Name:             "Bitcoin",
					PaymentSystem:    ps,
					Currency:         cr,
					GroupAlias:       "bitcoin",
					MinPaymentAmount: 0.01,
					MaxPaymentAmount: 15000.00,
					IsActive:         true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
					Params: &model.PaymentMethodParams{
						Handler:    payment_system.PaymentSystemHandlerCardPay,
						Terminal:   "16007",
						ExternalId: "BITCOIN",
					},
					Icon: "/images/btc_logo.png",
				},
			}

			err := db.C(manager.TablePaymentMethod).Insert(pms...)

			if err != nil {
				return err
			}

			p := &model.Project{}

			if err := db.C(manager.TableProject).Find(bson.M{"name": "Universe of Futurama"}).One(&p); err != nil {
				return err
			}

			var selectedPms []*model.PaymentMethod

			err = db.C(manager.TablePaymentMethod).Find(bson.M{"name": bson.M{"$in": []string{"Bank card", "Qiwi", "WebMoney", "Alipay", "Bitcoin"}}}).All(&selectedPms)

			if err != nil {
				return err
			}

			projectPms := make(map[string][]*model.ProjectPaymentModes)

			for _, pm := range selectedPms {
				projectPms[pm.GroupAlias] = append(projectPms[pm.GroupAlias], &model.ProjectPaymentModes{Id: pm.Id, AddedAt: time.Now()})
			}

			p.PaymentMethods = projectPms

			return db.C(manager.TableProject).UpdateId(p.Id, p)
		},
		func(db *mgo.Database) error {
			var pms []*model.PaymentMethod

			err := db.C(manager.TablePaymentMethod).Find(bson.M{"name": bson.M{"$in": []string{"Alipay", "Bitcoin"}}}).All(&pms)

			if err != nil {
				return err
			}

			if len(pms) < 3 {
				return errors.New("payment methods not found")
			}

			err = db.C(manager.TablePaymentMethod).Remove(pms)

			if err != nil {
				return err
			}

			p := &model.Project{}

			if err := db.C(manager.TableProject).Find(bson.M{"name": "Universe of Futurama"}).One(&p); err != nil {
				return err
			}

			var selectedPms []*model.PaymentMethod

			err = db.C(manager.TablePaymentMethod).Find(bson.M{"name": bson.M{"$in": []string{"Bank card", "Qiwi", "WebMoney"}}}).All(&selectedPms)

			if err != nil {
				return err
			}

			projectPms := make(map[string][]*model.ProjectPaymentModes)

			for _, pm := range selectedPms {
				projectPms[pm.GroupAlias] = append(projectPms[pm.GroupAlias], &model.ProjectPaymentModes{Id: pm.Id, AddedAt: time.Now()})
			}

			p.PaymentMethods = projectPms

			return db.C(manager.TableProject).UpdateId(p.Id, p)
		},
	)

	if err != nil {
		return
	}
}
