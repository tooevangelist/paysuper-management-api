package migrations

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/xakep666/mongo-migrate"
	"time"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			p := &model.Project{}

			if err := db.C(manager.TableProject).Find(bson.M{"name": "Герои Меча и Магии", "merchant.external_id": "5be2c3022b9bb6000765d132"}).One(&p); err != nil {
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
		func(db *mgo.Database) error {
			return nil
		},
	)

	if err != nil {
		return
	}
}
