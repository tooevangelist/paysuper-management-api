package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
)

func (rep *Repository) InsertProject(p *model.Project) error {
	return rep.Collection.Insert(p)
}

func (rep *Repository) UpdateProject(p *model.Project) error {
	return rep.Collection.UpdateId(p.Id, p)
}

func (rep *Repository) FindProjectsByMerchantId(mId string, limit int, offset int) ([]*model.Project, error) {
	var p []*model.Project
	err := rep.Collection.Find(bson.M{"merchant.external_id": mId}).Limit(limit).Skip(offset).All(&p)

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

func (rep *Repository) FindFixedPackageByFilters(filters *model.FixedPackageFilters) ([]map[string]interface{}, error) {
	var res []map[string]interface{}

	fpSection := "fixed_package." + filters.Region
	qCond := bson.M{
		"$and": []bson.M{
			{"$eq": []interface{}{"$$fixed_package.is_active", true}},
		},
	}

	orCond := bson.M{"$or": []bson.M{}}

	if len(filters.Ids) > 0 {
		orCond["$or"] = append(orCond["$or"].([]bson.M), map[string]interface{}{"$in": []interface{}{"$$fixed_package.id", filters.Ids}})
	}

	if len(filters.Names) > 0 {
		orCond["$or"] = append(orCond["$or"].([]bson.M), map[string]interface{}{"$in": []interface{}{"$$fixed_package.name", filters.Names}})
	}

	if len(orCond["$or"].([]bson.M)) > 0 {
		qCond["$and"] = append(qCond["$and"].([]bson.M), orCond)
	}

	q := []bson.M{
		{
			"$match": bson.M{
				"_id":       bson.ObjectIdHex(filters.ProjectId),
				fpSection:   bson.M{"$exists": true},
				"is_active": true,
			},
		},
		{
			"$project": bson.M{
				"fixed_package": bson.M{
					"$filter": bson.M{
						"input": "$" + fpSection,
						"as":    "fixed_package",
						"cond":  qCond,
					},
				},
				"_id": 0,
			},
		},
	}

	err := rep.Collection.Pipe(q).All(&res)

	return res, err
}
