package repository

import "github.com/ProtocolONE/p1pay.api/database/model"

func (rep *Repository) InsertLog(log *model.Log) error {
	return rep.Collection.Insert(log)
}
