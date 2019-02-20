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
			m := &model.Merchant{
				Id: bson.NewObjectId(),
				ExternalId: "5be2c3022b9bb6000765d132",
				Email: "dmitriy.sinichkin@protocol.one",
			}

			return db.C(manager.TableMerchant).Insert(m)
		},
		func(db *mgo.Database) error {
			return db.C(manager.TableMerchant).Remove(bson.M{"external_id": "5be2c3022b9bb6000765d132"})
		},
	)

	if err != nil {
		return
	}
}
