package repository

import (
	"github.com/globalsign/mgo"
)

type Repository struct {
	Collection *mgo.Collection
}
