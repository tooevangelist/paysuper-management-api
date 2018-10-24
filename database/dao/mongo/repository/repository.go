package repository

import (
	"gopkg.in/mgo.v2"
)

type Repository struct {
	Collection *mgo.Collection
}
