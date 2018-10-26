package repository

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"gopkg.in/mgo.v2/bson"
)

func (rep *Repository) InsertProject(p *model.Project) error {

}

func (rep *Repository) UpdateProject(p *model.Project) error {

}

func (rep *Repository) FindProjectsByMerchantId(string) ([]*model.Project, error) {

}

func (rep *Repository) FindProjectsByMerchantIdAndName(mId bson.ObjectId, pName string) *model.Project {

}

func (rep *Repository) FindProjectById() (*model.Project, error) {

}
