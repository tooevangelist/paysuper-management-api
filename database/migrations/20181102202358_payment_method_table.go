package migrations

import (
	"github.com/globalsign/mgo"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/xakep666/mongo-migrate"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			var err error

			c := db.C(manager.TablePaymentMethod)

			err = c.EnsureIndex(mgo.Index{Name: "county_group_id_is_active_idx", Key: []string{"group_id", "is_active"}})

			if err != nil {
				return err
			}

			err = c.EnsureIndex(mgo.Index{Name: "county_payment_system_id_idx", Key: []string{"payment_system_id"}})

			if err != nil {
				return err
			}

			return c.EnsureIndex(mgo.Index{Name: "county_payment_system_id_is_active_idx", Key: []string{"payment_system_id", "is_active"}})
		},
		func(db *mgo.Database) error {
			return db.C(manager.TablePaymentMethod).DropCollection()
		},
	)

	if err != nil {
		return
	}
}
