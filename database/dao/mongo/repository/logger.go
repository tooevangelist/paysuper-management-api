package repository

import "github.com/paysuper/paysuper-management-api/database/model"

func (rep *Repository) InsertLog(log *model.Log) error {
	return rep.Collection.Insert(log)
}
