package migrations

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/xakep666/mongo-migrate"
	"time"
)

type VatItem struct {
	CountryCode int
	Subdivision string
	Value       float64
}

var items = []*VatItem{
	{CountryCode: 36, Subdivision: "", Value: 10},
	{CountryCode: 40, Subdivision: "", Value: 20},
	{CountryCode: 31, Subdivision: "", Value: 18},
	{CountryCode: 8, Subdivision: "", Value: 30},
	{CountryCode: 12, Subdivision: "", Value: 30},
	{CountryCode: 24, Subdivision: "", Value: 30},
	{CountryCode: 20, Subdivision: "", Value: 30},
	{CountryCode: 10, Subdivision: "", Value: 30},
	{CountryCode: 28, Subdivision: "", Value: 30},
	{CountryCode: 32, Subdivision: "", Value: 21},
	{CountryCode: 51, Subdivision: "", Value: 20},
	{CountryCode: 533, Subdivision: "", Value: 30},
	{CountryCode: 4, Subdivision: "", Value: 30},
	{CountryCode: 44, Subdivision: "", Value: 30},
	{CountryCode: 50, Subdivision: "", Value: 30},
	{CountryCode: 52, Subdivision: "", Value: 30},
	{CountryCode: 48, Subdivision: "", Value: 30},
	{CountryCode: 112, Subdivision: "", Value: 20},
	{CountryCode: 84, Subdivision: "", Value: 30},
	{CountryCode: 56, Subdivision: "", Value: 21},
	{CountryCode: 204, Subdivision: "", Value: 30},
	{CountryCode: 60, Subdivision: "", Value: 30},
	{CountryCode: 100, Subdivision: "", Value: 20},
	{CountryCode: 68, Subdivision: "", Value: 30},
	{CountryCode: 70, Subdivision: "", Value: 17},
	{CountryCode: 72, Subdivision: "", Value: 30},
	{CountryCode: 76, Subdivision: "", Value: 30},
	{CountryCode: 96, Subdivision: "", Value: 30},
	{CountryCode: 854, Subdivision: "", Value: 30},
	{CountryCode: 108, Subdivision: "", Value: 30},
	{CountryCode: 64, Subdivision: "", Value: 30},
	{CountryCode: 548, Subdivision: "", Value: 30},
	{CountryCode: 336, Subdivision: "", Value: 30},
	{CountryCode: 826, Subdivision: "", Value: 20},
	{CountryCode: 348, Subdivision: "", Value: 27},
	{CountryCode: 862, Subdivision: "", Value: 11},
	{CountryCode: 626, Subdivision: "", Value: 30},
	{CountryCode: 704, Subdivision: "", Value: 10},
	{CountryCode: 266, Subdivision: "", Value: 30},
	{CountryCode: 332, Subdivision: "", Value: 30},
	{CountryCode: 328, Subdivision: "", Value: 16},
	{CountryCode: 270, Subdivision: "", Value: 30},
	{CountryCode: 288, Subdivision: "", Value: 30},
	{CountryCode: 312, Subdivision: "", Value: 30},
	{CountryCode: 320, Subdivision: "", Value: 30},
	{CountryCode: 324, Subdivision: "", Value: 30},
	{CountryCode: 624, Subdivision: "", Value: 30},
	{CountryCode: 276, Subdivision: "", Value: 19},
	{CountryCode: 292, Subdivision: "", Value: 30},
	{CountryCode: 340, Subdivision: "", Value: 30},
	{CountryCode: 344, Subdivision: "", Value: 13},
	{CountryCode: 308, Subdivision: "", Value: 30},
	{CountryCode: 304, Subdivision: "", Value: 30},
	{CountryCode: 300, Subdivision: "", Value: 24},
	{CountryCode: 268, Subdivision: "", Value: 18},
	{CountryCode: 316, Subdivision: "", Value: 30},
	{CountryCode: 208, Subdivision: "", Value: 25},
	{CountryCode: 180, Subdivision: "", Value: 30},
	{CountryCode: 262, Subdivision: "", Value: 30},
	{CountryCode: 212, Subdivision: "", Value: 12},
	{CountryCode: 214, Subdivision: "", Value: 12},
	{CountryCode: 818, Subdivision: "", Value: 30},
	{CountryCode: 894, Subdivision: "", Value: 30},
	{CountryCode: 732, Subdivision: "", Value: 30},
	{CountryCode: 716, Subdivision: "", Value: 30},
	{CountryCode: 376, Subdivision: "", Value: 17},
	{CountryCode: 356, Subdivision: "", Value: 12.5},
	{CountryCode: 360, Subdivision: "", Value: 30},
	{CountryCode: 400, Subdivision: "", Value: 30},
	{CountryCode: 368, Subdivision: "", Value: 30},
	{CountryCode: 364, Subdivision: "", Value: 30},
	{CountryCode: 372, Subdivision: "", Value: 23},
	{CountryCode: 352, Subdivision: "", Value: 24.5},
	{CountryCode: 724, Subdivision: "", Value: 21},
	{CountryCode: 380, Subdivision: "", Value: 22},
	{CountryCode: 887, Subdivision: "", Value: 30},
	{CountryCode: 132, Subdivision: "", Value: 30},
	{CountryCode: 398, Subdivision: "", Value: 12},
	{CountryCode: 116, Subdivision: "", Value: 30},
	{CountryCode: 120, Subdivision: "", Value: 30},
	{CountryCode: 124, Subdivision: "AB", Value: 5},
	{CountryCode: 124, Subdivision: "BC", Value: 12},
	{CountryCode: 124, Subdivision: "QC", Value: 14.975},
	{CountryCode: 124, Subdivision: "MB", Value: 13},
	{CountryCode: 124, Subdivision: "NS", Value: 15},
	{CountryCode: 124, Subdivision: "NU", Value: 5},
	{CountryCode: 124, Subdivision: "NB", Value: 13},
	{CountryCode: 124, Subdivision: "NL", Value: 13},
	{CountryCode: 124, Subdivision: "ON", Value: 13},
	{CountryCode: 124, Subdivision: "PE", Value: 14},
	{CountryCode: 124, Subdivision: "SK", Value: 10},
	{CountryCode: 124, Subdivision: "NT", Value: 5},
	{CountryCode: 124, Subdivision: "YT", Value: 5},
	{CountryCode: 634, Subdivision: "", Value: 30},
	{CountryCode: 404, Subdivision: "", Value: 30},
	{CountryCode: 196, Subdivision: "", Value: 19},
	{CountryCode: 417, Subdivision: "", Value: 30},
	{CountryCode: 296, Subdivision: "", Value: 30},
	{CountryCode: 156, Subdivision: "", Value: 17.6},
	{CountryCode: 170, Subdivision: "", Value: 30},
	{CountryCode: 178, Subdivision: "", Value: 30},
	{CountryCode: 188, Subdivision: "", Value: 30},
	{CountryCode: 384, Subdivision: "", Value: 20},
	{CountryCode: 192, Subdivision: "", Value: 30},
	{CountryCode: 414, Subdivision: "", Value: 30},
	{CountryCode: 428, Subdivision: "", Value: 21},
	{CountryCode: 426, Subdivision: "", Value: 30},
	{CountryCode: 430, Subdivision: "", Value: 30},
	{CountryCode: 422, Subdivision: "", Value: 10},
	{CountryCode: 434, Subdivision: "", Value: 30},
	{CountryCode: 440, Subdivision: "", Value: 21},
	{CountryCode: 438, Subdivision: "", Value: 30},
	{CountryCode: 442, Subdivision: "", Value: 17},
	{CountryCode: 480, Subdivision: "", Value: 30},
	{CountryCode: 478, Subdivision: "", Value: 30},
	{CountryCode: 450, Subdivision: "", Value: 30},
	{CountryCode: 446, Subdivision: "", Value: 13},
	{CountryCode: 807, Subdivision: "", Value: 18},
	{CountryCode: 454, Subdivision: "", Value: 30},
	{CountryCode: 458, Subdivision: "", Value: 5},
	{CountryCode: 466, Subdivision: "", Value: 30},
	{CountryCode: 462, Subdivision: "", Value: 30},
	{CountryCode: 470, Subdivision: "", Value: 18},
	{CountryCode: 504, Subdivision: "", Value: 30},
	{CountryCode: 474, Subdivision: "", Value: 30},
	{CountryCode: 484, Subdivision: "", Value: 16},
	{CountryCode: 508, Subdivision: "", Value: 30},
	{CountryCode: 498, Subdivision: "", Value: 30},
	{CountryCode: 492, Subdivision: "", Value: 30},
	{CountryCode: 496, Subdivision: "", Value: 30},
	{CountryCode: 104, Subdivision: "", Value: 30},
	{CountryCode: 516, Subdivision: "", Value: 30},
	{CountryCode: 520, Subdivision: "", Value: 30},
	{CountryCode: 524, Subdivision: "", Value: 30},
	{CountryCode: 562, Subdivision: "", Value: 30},
	{CountryCode: 566, Subdivision: "", Value: 30},
	{CountryCode: 528, Subdivision: "", Value: 21},
	{CountryCode: 558, Subdivision: "", Value: 30},
	{CountryCode: 554, Subdivision: "", Value: 15},
	{CountryCode: 578, Subdivision: "", Value: 25},
	{CountryCode: 784, Subdivision: "", Value: 30},
	{CountryCode: 512, Subdivision: "", Value: 30},
	{CountryCode: 586, Subdivision: "", Value: 30},
	{CountryCode: 591, Subdivision: "", Value: 30},
	{CountryCode: 598, Subdivision: "", Value: 30},
	{CountryCode: 600, Subdivision: "", Value: 10},
	{CountryCode: 604, Subdivision: "", Value: 18},
	{CountryCode: 616, Subdivision: "", Value: 23},
	{CountryCode: 620, Subdivision: "", Value: 23},
	{CountryCode: 643, Subdivision: "", Value: 20},
	{CountryCode: 646, Subdivision: "", Value: 30},
	{CountryCode: 642, Subdivision: "", Value: 19},
	{CountryCode: 222, Subdivision: "", Value: 13},
	{CountryCode: 882, Subdivision: "", Value: 30},
	{CountryCode: 682, Subdivision: "", Value: 30},
	{CountryCode: 748, Subdivision: "", Value: 30},
	{CountryCode: 686, Subdivision: "", Value: 30},
	{CountryCode: 702, Subdivision: "", Value: 5},
	{CountryCode: 760, Subdivision: "", Value: 30},
	{CountryCode: 703, Subdivision: "", Value: 20},
	{CountryCode: 705, Subdivision: "", Value: 22},
	{CountryCode: 840, Subdivision: "AL", Value: 13.5},
	{CountryCode: 840, Subdivision: "AK", Value: 7},
	{CountryCode: 840, Subdivision: "AZ", Value: 10.725},
	{CountryCode: 840, Subdivision: "AR", Value: 11.625},
	{CountryCode: 840, Subdivision: "CA", Value: 10.25},
	{CountryCode: 840, Subdivision: "CO", Value: 10},
	{CountryCode: 840, Subdivision: "CT", Value: 6.35},
	{CountryCode: 840, Subdivision: "DE", Value: 0},
	{CountryCode: 840, Subdivision: "FL", Value: 7.5},
	{CountryCode: 840, Subdivision: "GA", Value: 8},
	{CountryCode: 840, Subdivision: "HI", Value: 4.712},
	{CountryCode: 840, Subdivision: "ID", Value: 8.5},
	{CountryCode: 840, Subdivision: "IL", Value: 10.25},
	{CountryCode: 840, Subdivision: "IN", Value: 7},
	{CountryCode: 840, Subdivision: "IA", Value: 7},
	{CountryCode: 840, Subdivision: "KS", Value: 10.15},
	{CountryCode: 840, Subdivision: "KY", Value: 6},
	{CountryCode: 840, Subdivision: "LA", Value: 12},
	{CountryCode: 840, Subdivision: "ME", Value: 5.5},
	{CountryCode: 840, Subdivision: "MD", Value: 6},
	{CountryCode: 840, Subdivision: "MA", Value: 6.25},
	{CountryCode: 840, Subdivision: "MI", Value: 6},
	{CountryCode: 840, Subdivision: "MN", Value: 7.875},
	{CountryCode: 840, Subdivision: "MS", Value: 7.25},
	{CountryCode: 840, Subdivision: "MO", Value: 10.85},
	{CountryCode: 840, Subdivision: "MT", Value: 0},
	{CountryCode: 840, Subdivision: "NE", Value: 7.5},
	{CountryCode: 840, Subdivision: "NV", Value: 8.15},
	{CountryCode: 840, Subdivision: "NH", Value: 0},
	{CountryCode: 840, Subdivision: "NJ", Value: 12.875},
	{CountryCode: 840, Subdivision: "NM", Value: 8.688},
	{CountryCode: 840, Subdivision: "NY", Value: 8.875},
	{CountryCode: 840, Subdivision: "NC", Value: 7.50},
	{CountryCode: 840, Subdivision: "ND", Value: 8},
	{CountryCode: 840, Subdivision: "OH", Value: 8},
	{CountryCode: 840, Subdivision: "OK", Value: 11},
	{CountryCode: 840, Subdivision: "OR", Value: 0},
	{CountryCode: 840, Subdivision: "PA", Value: 8},
	{CountryCode: 840, Subdivision: "RI", Value: 7},
	{CountryCode: 840, Subdivision: "SC", Value: 9},
	{CountryCode: 840, Subdivision: "SD", Value: 6},
	{CountryCode: 840, Subdivision: "TN", Value: 9.75},
	{CountryCode: 840, Subdivision: "TX", Value: 8.25},
	{CountryCode: 840, Subdivision: "UT", Value: 8.35},
	{CountryCode: 840, Subdivision: "VT", Value: 7},
	{CountryCode: 840, Subdivision: "VA", Value: 6},
	{CountryCode: 840, Subdivision: "WA", Value: 10.4},
	{CountryCode: 840, Subdivision: "WV", Value: 7},
	{CountryCode: 840, Subdivision: "WI", Value: 6.75},
	{CountryCode: 840, Subdivision: "WY", Value: 6},
	{CountryCode: 840, Subdivision: "DC", Value: 5.75},
	{CountryCode: 840, Subdivision: "GU", Value: 4},
	{CountryCode: 840, Subdivision: "PR", Value: 11.5},
	{CountryCode: 706, Subdivision: "", Value: 30},
	{CountryCode: 736, Subdivision: "", Value: 30},
	{CountryCode: 740, Subdivision: "", Value: 30},
	{CountryCode: 762, Subdivision: "", Value: 30},
	{CountryCode: 764, Subdivision: "", Value: 7},
	{CountryCode: 158, Subdivision: "", Value: 30},
	{CountryCode: 834, Subdivision: "", Value: 30},
	{CountryCode: 768, Subdivision: "", Value: 30},
	{CountryCode: 776, Subdivision: "", Value: 30},
	{CountryCode: 780, Subdivision: "", Value: 15},
	{CountryCode: 798, Subdivision: "", Value: 30},
	{CountryCode: 788, Subdivision: "", Value: 30},
	{CountryCode: 795, Subdivision: "", Value: 30},
	{CountryCode: 792, Subdivision: "", Value: 18},
	{CountryCode: 800, Subdivision: "", Value: 30},
	{CountryCode: 860, Subdivision: "", Value: 20},
	{CountryCode: 804, Subdivision: "", Value: 20},
	{CountryCode: 858, Subdivision: "", Value: 23},
	{CountryCode: 242, Subdivision: "", Value: 30},
	{CountryCode: 608, Subdivision: "", Value: 12},
	{CountryCode: 246, Subdivision: "", Value: 24},
	{CountryCode: 250, Subdivision: "", Value: 20},
	{CountryCode: 191, Subdivision: "", Value: 22},
	{CountryCode: 148, Subdivision: "", Value: 30},
	{CountryCode: 203, Subdivision: "", Value: 21},
	{CountryCode: 152, Subdivision: "", Value: 19},
	{CountryCode: 756, Subdivision: "", Value: 8},
	{CountryCode: 752, Subdivision: "", Value: 25},
	{CountryCode: 144, Subdivision: "", Value: 15},
	{CountryCode: 218, Subdivision: "", Value: 12},
	{CountryCode: 232, Subdivision: "", Value: 30},
	{CountryCode: 233, Subdivision: "", Value: 20},
	{CountryCode: 231, Subdivision: "", Value: 30},
	{CountryCode: 891, Subdivision: "", Value: 30},
	{CountryCode: 710, Subdivision: "", Value: 30},
	{CountryCode: 410, Subdivision: "", Value: 10},
	{CountryCode: 388, Subdivision: "", Value: 20},
	{CountryCode: 392, Subdivision: "", Value: 8},
}

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			var vats []interface{}

			collection := db.C(manager.TableVat)

			for _, v := range items {
				err := collection.EnsureIndex(
					mgo.Index{
						Name:   "vat_country_code_int_subdivision_uniq",
						Key:    []string{"country.code_int", "subdivision_code"},
						Unique: true,
					},
				)

				if err != nil {
					return err
				}

				err = collection.EnsureIndex(
					mgo.Index{
						Name:   "vat_country_code_a2_subdivision_uniq",
						Key:    []string{"country.code_a2", "subdivision_code"},
						Unique: true,
					},
				)

				if err != nil {
					return err
				}

				err = collection.EnsureIndex(
					mgo.Index{
						Name:   "vat_country_code_a3_subdivision_uniq",
						Key:    []string{"country.code_a3", "subdivision_code"},
						Unique: true,
					},
				)

				if err != nil {
					return err
				}

				var c *model.Country

				err = db.C(manager.TableCountry).Find(bson.M{"code_int": v.CountryCode}).One(&c)

				if err != nil {
					return err
				}

				vat := &model.Vat{
					Id: bson.NewObjectId(),
					Country: &model.SimpleCountry{
						CodeA2: c.IsoCodeA2,
					},
					SubdivisionCode: v.Subdivision,
					Vat:             v.Value,
					CreatedAt:       time.Now(),
				}

				vats = append(vats, vat)
			}

			return collection.Insert(vats...)
		},
		func(db *mgo.Database) error {
			return db.C(manager.TableVat).DropCollection()
		},
	)

	if err != nil {
		return
	}
}
