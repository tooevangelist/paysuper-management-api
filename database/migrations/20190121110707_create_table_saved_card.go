package migrations

import (
	"github.com/globalsign/mgo"
	"github.com/xakep666/mongo-migrate"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			c := db.C("saved_card")

			err := c.EnsureIndex(
				mgo.Index{
					Name: "saved_card_account_project_id_pan_unq",
					Key: []string{"account", "project_id", "pan"},
					Unique: true,
				},
			)

			if err != nil {
				return err
			}

			return c.EnsureIndex(
				mgo.Index{
					Name: "saved_card_account_project_id_is_active_idx",
					Key: []string{"account", "project_id", "is_active"},
					Unique: false,
				},
			)
		},
		func(db *mgo.Database) error {
			return db.C("saved_card").DropCollection()
		},
	)

	if err != nil {
		return
	}
}

