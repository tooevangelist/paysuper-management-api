package repository

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
)

func (rep *Repository) FindOrderByProjectOrderId(prjOrderId string) (*model.Order, error) {
	var o *model.Order
	err := rep.Collection.Find(bson.M{"project_order_id": prjOrderId}).One(&o)

	return o, err
}

func (rep *Repository) FindOrderById(id bson.ObjectId) (*model.Order, error) {
	var o *model.Order
	err := rep.Collection.FindId(id).One(&o)

	return o, err
}

func (rep *Repository) InsertOrder(order *model.Order) error {
	return rep.Collection.Insert(order)
}

func (rep *Repository) UpdateOrder(o *model.Order) error {
	return rep.Collection.UpdateId(o.Id, o)
}

func (rep *Repository) FindAllOrders(filters bson.M, sort []string, limit int32, offset int32) ([]*model.Order, error) {
	var o []*model.Order
	err := rep.Collection.Find(filters).Sort(sort...).Limit(int(limit)).Skip(int(offset)).All(&o)

	return o, err
}

func (rep *Repository) GetOrdersCountByConditions(filters bson.M) (int, error) {
	return rep.Collection.Find(filters).Count()
}

func (rep *Repository) GetRevenueDynamic(rdr *model.RevenueDynamicRequest) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	mGroup := bson.M{
		"merchant_id":                          "$project.merchant.id",
		model.RevenueDynamicRequestPeriodYear:  "$year",
		model.RevenueDynamicRequestPeriodMonth: "$month",
		model.RevenueDynamicRequestPeriodWeek:  "$week",
		model.RevenueDynamicRequestPeriodDay:   "$day",
		model.RevenueDynamicRequestPeriodHour:  "$hour",
	}

	switch rdr.Period {
	case model.RevenueDynamicRequestPeriodYear:
		delete(mGroup, model.RevenueDynamicRequestPeriodHour)
		delete(mGroup, model.RevenueDynamicRequestPeriodDay)
		delete(mGroup, model.RevenueDynamicRequestPeriodWeek)
		delete(mGroup, model.RevenueDynamicRequestPeriodMonth)
		break
	case model.RevenueDynamicRequestPeriodMonth:
		delete(mGroup, model.RevenueDynamicRequestPeriodHour)
		delete(mGroup, model.RevenueDynamicRequestPeriodDay)
		delete(mGroup, model.RevenueDynamicRequestPeriodWeek)
		break
	case model.RevenueDynamicRequestPeriodWeek:
		delete(mGroup, model.RevenueDynamicRequestPeriodHour)
		delete(mGroup, model.RevenueDynamicRequestPeriodDay)
		break
	case model.RevenueDynamicRequestPeriodDay:
		delete(mGroup, model.RevenueDynamicRequestPeriodHour)
		break
	}

	q := []bson.M{
		{
			"$project": bson.M{
				model.RevenueDynamicRequestPeriodHour:  bson.M{"$hour": "$created_at"},
				model.RevenueDynamicRequestPeriodDay:   bson.M{"$dayOfMonth": "$created_at"},
				model.RevenueDynamicRequestPeriodWeek:  bson.M{"$week": "$created_at"},
				model.RevenueDynamicRequestPeriodMonth: bson.M{"$month": "$created_at"},
				model.RevenueDynamicRequestPeriodYear:  bson.M{"$year": "$created_at"},
				"project":                              true,
				"status":                               true,
				"amount_out_merchant_ac":               true,
				"created_at":                           true,
			},
		},
		{
			"$facet": bson.M{
				model.RevenueDynamicFacetFieldRevenue: []bson.M{
					{
						"$match": bson.M{
							"status":     model.OrderStatusProjectComplete,
							"created_at": bson.M{"$gte": rdr.From, "$lte": rdr.To},
							"project.id": bson.M{"$in": rdr.Project},
						},
					},
					{
						"$group": bson.M{
							model.RevenueDynamicFacetFieldId:    "$project.merchant._id",
							model.RevenueDynamicFacetFieldCount: bson.M{"$sum": 1},
							model.RevenueDynamicFacetFieldTotal: bson.M{"$sum": "$amount_out_merchant_ac"},
							model.RevenueDynamicFacetFieldAvg:   bson.M{"$avg": "$amount_out_merchant_ac"},
						},
					},
				},
				model.RevenueDynamicFacetFieldRefund: []bson.M{
					{
						"$match": bson.M{
							"status":     bson.M{"$in": []int{model.OrderStatusRefund, model.OrderStatusChargeback}},
							"created_at": bson.M{"$gte": rdr.From, "$lte": rdr.To},
							"project.id": bson.M{"$in": rdr.Project},
						},
					},
					{
						"$group": bson.M{
							model.RevenueDynamicFacetFieldId:    "$project.merchant._id",
							model.RevenueDynamicFacetFieldCount: bson.M{"$sum": 1},
							model.RevenueDynamicFacetFieldTotal: bson.M{"$sum": "$amount_out_merchant_ac"},
							model.RevenueDynamicFacetFieldAvg:   bson.M{"$avg": "$amount_out_merchant_ac"},
						},
					},
				},
				model.RevenueDynamicFacetFieldPointsRevenue: []bson.M{
					{
						"$match": bson.M{
							"status":     model.OrderStatusProjectComplete,
							"created_at": bson.M{"$gte": rdr.From, "$lte": rdr.To},
							"project.id": bson.M{"$in": rdr.Project},
						},
					},
					{
						"$group": bson.M{
							model.RevenueDynamicFacetFieldId:    mGroup,
							model.RevenueDynamicFacetFieldTotal: bson.M{"$sum": "$amount_out_merchant_ac"},
						},
					},
					{"$sort": bson.M{model.RevenueDynamicFacetFieldId: 1}},
				},
				model.RevenueDynamicFacetFieldPointsRefund: []bson.M{
					{
						"$match": bson.M{
							"status":     bson.M{"$in": []int{model.OrderStatusRefund, model.OrderStatusChargeback}},
							"created_at": bson.M{"$gte": rdr.From, "$lte": rdr.To},
							"project.id": bson.M{"$in": rdr.Project},
						},
					},
					{
						"$group": bson.M{
							model.RevenueDynamicFacetFieldId:    mGroup,
							model.RevenueDynamicFacetFieldTotal: bson.M{"$sum": "$amount_out_merchant_ac"},
						},
					},
					{"$sort": bson.M{model.RevenueDynamicFacetFieldId: 1}},
				},
			},
		},
	}

	err := rep.Collection.Pipe(q).All(&result)

	return result, err
}

func (rep *Repository) GetAccountingPayment(rdr *model.RevenueDynamicRequest, mId string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	q := []bson.M{
		{
			"$project": bson.M{
				"project":               true,
				"status":                true,
				"pm_order_close_date":   true,
				"project_fee_amount":    true,
				"to_payer_fee_amount":   true,
				"amount_in_merchant_ac": true,
				"total": bson.M{
					"$subtract": []interface{}{
						"$amount_in_merchant_ac",
						bson.M{
							"$add": []string{
								"$project_fee_amount.amount_merchant_currency",
								"$to_payer_fee_amount.amount_merchant_currency",
							},
						},
					},
				},
			},
		},
		{
			"$facet": bson.M{
				"total_success": []bson.M{
					{
						"$match": bson.M{
							"status":                       bson.M{"$in": []int{model.OrderStatusProjectComplete, model.OrderStatusPaymentSystemComplete, model.OrderStatusProjectInProgress, model.OrderStatusProjectPending}},
							"pm_order_close_date":          bson.M{"$gte": rdr.From, "$lte": rdr.To},
							"project.merchant.external_id": bson.M{"$eq": mId},
						},
					},
					{
						"$group": bson.M{
							"_id":   "$project.merchant._id",
							"total": bson.M{"$sum": "$amount_in_merchant_ac"},
						},
					},
				},
				"success": []bson.M{
					{
						"$match": bson.M{
							"status":                       bson.M{"$in": []int{model.OrderStatusProjectComplete, model.OrderStatusPaymentSystemComplete, model.OrderStatusProjectInProgress, model.OrderStatusProjectPending}},
							"pm_order_close_date":          bson.M{"$gte": rdr.From, "$lte": rdr.To},
							"project.merchant.external_id": bson.M{"$eq": mId},
						},
					},
					{
						"$group": bson.M{
							"_id":   "$project.merchant._id",
							"total": bson.M{"$sum": "$total"},
						},
					},
				},
				"refund": []bson.M{
					{
						"$match": bson.M{
							"status":                       model.OrderStatusRefund,
							"pm_order_close_date":          bson.M{"$gte": rdr.From, "$lte": rdr.To},
							"project.merchant.external_id": bson.M{"$eq": mId},
						},
					},
					{
						"$group": bson.M{
							"_id":   "$project.merchant._id",
							"total": bson.M{"$sum": "$amount_in_merchant_ac"},
						},
					},
				},
				"chargeback": []bson.M{
					{
						"$match": bson.M{
							"status":                       model.OrderStatusChargeback,
							"pm_order_close_date":          bson.M{"$gte": rdr.From, "$lte": rdr.To},
							"project.merchant.external_id": bson.M{"$eq": mId},
						},
					},
					{
						"$group": bson.M{
							"_id":   "$project.merchant._id",
							"total": bson.M{"$sum": "$amount_in_merchant_ac"},
						},
					},
				},
				"commission": []bson.M{
					{
						"$match": bson.M{
							"status":                       bson.M{"$in": []int{model.OrderStatusProjectComplete, model.OrderStatusPaymentSystemComplete, model.OrderStatusProjectInProgress, model.OrderStatusProjectPending}},
							"pm_order_close_date":          bson.M{"$gte": rdr.From, "$lte": rdr.To},
							"project.merchant.external_id": bson.M{"$eq": mId},
						},
					},
					{
						"$group": bson.M{
							"_id": "$project.merchant._id",
							"total": bson.M{
								"$sum": bson.M{
									"$add": []string{
										"$project_fee_amount.amount_merchant_currency",
										"$to_payer_fee_amount.amount_merchant_currency",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	err := rep.Collection.Pipe(q).All(&result)

	return result, err
}
