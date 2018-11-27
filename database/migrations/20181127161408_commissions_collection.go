package migrations

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/ProtocolONE/p1pay.api/manager"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/xakep666/mongo-migrate"
	"time"
)

const (
	pspCommission = 1.5
)

var pmCommissions = map[string]float32{
	"bank_card": 3,
	"qiwi":      5,
	"webmoney":  2,
	"neteller":  7,
	"alipay":    4.5,
	"bitcoin":   6.7,
}

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			var pms []*model.PaymentMethod
			var err error

			collection := db.C(manager.TableCommission)

			err = collection.EnsureIndex(
				mgo.Index{
					Name: "commission_project_id_pm_id_start_date_idx",
					Key:  []string{"project_id", "pm_id", "start_date"},
				},
			)

			if err != nil {
				return err
			}

			err = db.C(manager.TablePaymentMethod).Find(bson.M{"is_active": true}).All(&pms)

			if err != nil {
				return err
			}

			var p *model.Project

			err = db.C(manager.TableProject).Find(bson.M{"name": "Universe of Futurama"}).One(&p)

			var commissions []interface{}

			for _, pm := range pms {
				commission := &model.Commission{
					Id:                      bson.NewObjectId(),
					PaymentMethodId:         pm.Id,
					ProjectId:               p.Id,
					PaymentMethodCommission: pmCommissions[pm.GroupAlias],
					PspCommission:           pspCommission,
					TotalCommissionToUser:   0,
					StartDate:               time.Now(),
					CreatedAt:               time.Now(),
				}

				commissions = append(commissions, commission)
			}

			return collection.Insert(commissions...)
		},
		func(db *mgo.Database) error {
			return db.C(manager.TableCommission).DropCollection()
		},
	)

	if err != nil {
		return
	}
}
