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
			p := &model.Project{}

			if err := db.C(manager.TableProject).Find(bson.M{"name": "Universe of Futurama"}).One(&p); err != nil {
				return err
			}

			var pms []*model.PaymentMethod

			err := db.C(manager.TablePaymentMethod).Find(bson.M{"name": bson.M{"$in": []string{"Bank card", "Qiwi", "WebMoney"}}}).All(&pms)

			if err != nil {
				return err
			}

			projectPms := make(map[string][]*model.ProjectPaymentModes)

			for _, pm := range pms {
				projectPms[pm.GroupAlias] = append(projectPms[pm.GroupAlias], &model.ProjectPaymentModes{Id: pm.Id, AddedAt: time.Now()})
			}

			p.PaymentMethods = projectPms

			return db.C(manager.TableProject).UpdateId(p.Id, p)
		},
		func(db *mgo.Database) error {
			p := &model.Project{}

			if err := db.C(manager.TableProject).Find(bson.M{"name": "Universe of Futurama"}).One(&p); err != nil {
				return err
			}

			pm := model.PaymentMethod{}

			if err := db.C(manager.TablePaymentMethod).Find(bson.M{"group_alias": "bank_card"}).One(&pm); err != nil {
				return err
			}

			pms := make(map[string][]*model.ProjectPaymentModes)
			pms["bank_card"] = append(pms["bank_card"], &model.ProjectPaymentModes{Id: pm.Id, AddedAt: time.Now()})

			p.PaymentMethods = pms

			return db.C(manager.TableProject).UpdateId(p.Id, p)
		},
	)

	if err != nil {
		return
	}
}
