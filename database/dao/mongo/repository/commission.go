package repository

import (
	"github.com/ProtocolONE/p1pay.api/database/model"
	"github.com/globalsign/mgo/bson"
	"time"
)

func (rep *Repository) FindCommissionByProjectIdAndPaymentMethodId(projectId bson.ObjectId, pmId bson.ObjectId) (*model.Commission, error) {
	var commission *model.Commission

	err := rep.Collection.Find(
		bson.M{
			"project_id": projectId,
			"pm_id":      pmId,
			"start_date": bson.M{"$lte": time.Now()},
		},
	).Sort("-start_date").Limit(1).One(&commission)

	return commission, err
}
