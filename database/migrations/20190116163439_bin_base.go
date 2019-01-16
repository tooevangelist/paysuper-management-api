package migrations

import (
	"github.com/globalsign/mgo"
	"github.com/xakep666/mongo-migrate"
)

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			c := db.C("bank_bin")
			return c.EnsureIndex(mgo.Index{Name: "bank_bin_card_bin_unq", Key: []string{"card_bin"}, Unique: true})
		},
		func(db *mgo.Database) error {
			return db.C("bank_bin").DropCollection()
		},
	)

	if err != nil {
		return
	}
}
