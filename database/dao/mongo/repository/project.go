package repository

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"gopkg.in/mgo.v2/bson"
)

func (rep *Repository) InsertProject(p *model.Project) error {
	return rep.Collection.Insert(p)
}

func (rep *Repository) UpdateProject(p *model.Project) error {
	return rep.Collection.UpdateId(p.Id, p)
}

func (rep *Repository) FindProjectsByMerchantId(mId bson.ObjectId, limit int, offset int) ([]*model.Project, error) {
	var p []*model.Project
	err := rep.Collection.Find(bson.M{"merchant._id": mId}).Limit(limit).Skip(offset).All(&p)

	return p, err
}

func (rep *Repository) FindProjectByMerchantIdAndName(mId bson.ObjectId, pName string) (*model.Project, error) {
	var p *model.Project
	err := rep.Collection.Find(bson.M{"merchant._id": mId, "name": pName}).One(&p)

	return p, err
}

func (rep *Repository) FindProjectById(id bson.ObjectId) (*model.Project, error) {
	var p *model.Project
	err := rep.Collection.Find(bson.M{"_id": id}).One(&p)

	return p, err
}
