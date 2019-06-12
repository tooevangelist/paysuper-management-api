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
	CountryCode string
	Subdivision string
	Value       float64
}

var items = []*VatItem{
	{CountryCode: "AU", Subdivision: "", Value: 10},
	{CountryCode: "AT", Subdivision: "", Value: 20},
	{CountryCode: "AZ", Subdivision: "", Value: 18},
	{CountryCode: "AL", Subdivision: "", Value: 30},
	{CountryCode: "DZ", Subdivision: "", Value: 30},
	{CountryCode: "AO", Subdivision: "", Value: 30},
	{CountryCode: "AD", Subdivision: "", Value: 30},
	{CountryCode: "AQ", Subdivision: "", Value: 30},
	{CountryCode: "AG", Subdivision: "", Value: 30},
	{CountryCode: "AR", Subdivision: "", Value: 21},
	{CountryCode: "AM", Subdivision: "", Value: 20},
	{CountryCode: "AW", Subdivision: "", Value: 30},
	{CountryCode: "AF", Subdivision: "", Value: 30},
	{CountryCode: "BS", Subdivision: "", Value: 30},
	{CountryCode: "BD", Subdivision: "", Value: 30},
	{CountryCode: "BB", Subdivision: "", Value: 30},
	{CountryCode: "BH", Subdivision: "", Value: 30},
	{CountryCode: "BY", Subdivision: "", Value: 20},
	{CountryCode: "BZ", Subdivision: "", Value: 30},
	{CountryCode: "BE", Subdivision: "", Value: 21},
	{CountryCode: "BJ", Subdivision: "", Value: 30},
	{CountryCode: "BM", Subdivision: "", Value: 30},
	{CountryCode: "BG", Subdivision: "", Value: 20},
	{CountryCode: "BO", Subdivision: "", Value: 30},
	{CountryCode: "BA", Subdivision: "", Value: 17},
	{CountryCode: "BW", Subdivision: "", Value: 30},
	{CountryCode: "BR", Subdivision: "", Value: 30},
	{CountryCode: "BN", Subdivision: "", Value: 30},
	{CountryCode: "BF", Subdivision: "", Value: 30},
	{CountryCode: "BI", Subdivision: "", Value: 30},
	{CountryCode: "BT", Subdivision: "", Value: 30},
	{CountryCode: "VU", Subdivision: "", Value: 30},
	{CountryCode: "VA", Subdivision: "", Value: 30},
	{CountryCode: "GB", Subdivision: "", Value: 20},
	{CountryCode: "HU", Subdivision: "", Value: 27},
	{CountryCode: "VE", Subdivision: "", Value: 11},
	{CountryCode: "TL", Subdivision: "", Value: 30},
	{CountryCode: "VN", Subdivision: "", Value: 10},
	{CountryCode: "GA", Subdivision: "", Value: 30},
	{CountryCode: "HT", Subdivision: "", Value: 30},
	{CountryCode: "GY", Subdivision: "", Value: 16},
	{CountryCode: "GM", Subdivision: "", Value: 30},
	{CountryCode: "GH", Subdivision: "", Value: 30},
	{CountryCode: "GP", Subdivision: "", Value: 30},
	{CountryCode: "GT", Subdivision: "", Value: 30},
	{CountryCode: "GN", Subdivision: "", Value: 30},
	{CountryCode: "GW", Subdivision: "", Value: 30},
	{CountryCode: "DE", Subdivision: "", Value: 19},
	{CountryCode: "GI", Subdivision: "", Value: 30},
	{CountryCode: "HN", Subdivision: "", Value: 30},
	{CountryCode: "HK", Subdivision: "", Value: 13},
	{CountryCode: "GD", Subdivision: "", Value: 30},
	{CountryCode: "GL", Subdivision: "", Value: 30},
	{CountryCode: "GR", Subdivision: "", Value: 24},
	{CountryCode: "GE", Subdivision: "", Value: 18},
	{CountryCode: "GU", Subdivision: "", Value: 30},
	{CountryCode: "DK", Subdivision: "", Value: 25},
	{CountryCode: "CD", Subdivision: "", Value: 30},
	{CountryCode: "DJ", Subdivision: "", Value: 30},
	{CountryCode: "DM", Subdivision: "", Value: 12},
	{CountryCode: "DO", Subdivision: "", Value: 12},
	{CountryCode: "EG", Subdivision: "", Value: 30},
	{CountryCode: "ZM", Subdivision: "", Value: 30},
	{CountryCode: "EH", Subdivision: "", Value: 30},
	{CountryCode: "ZW", Subdivision: "", Value: 30},
	{CountryCode: "IL", Subdivision: "", Value: 17},
	{CountryCode: "IN", Subdivision: "", Value: 12.5},
	{CountryCode: "ID", Subdivision: "", Value: 30},
	{CountryCode: "JO", Subdivision: "", Value: 30},
	{CountryCode: "IQ", Subdivision: "", Value: 30},
	{CountryCode: "IR", Subdivision: "", Value: 30},
	{CountryCode: "IE", Subdivision: "", Value: 23},
	{CountryCode: "IS", Subdivision: "", Value: 24.5},
	{CountryCode: "ES", Subdivision: "", Value: 21},
	{CountryCode: "IT", Subdivision: "", Value: 22},
	{CountryCode: "YE", Subdivision: "", Value: 30},
	{CountryCode: "CV", Subdivision: "", Value: 30},
	{CountryCode: "KZ", Subdivision: "", Value: 12},
	{CountryCode: "KH", Subdivision: "", Value: 30},
	{CountryCode: "CM", Subdivision: "", Value: 30},
	{CountryCode: "CA", Subdivision: "AB", Value: 5},
	{CountryCode: "CA", Subdivision: "BC", Value: 12},
	{CountryCode: "CA", Subdivision: "QC", Value: 14.975},
	{CountryCode: "CA", Subdivision: "MB", Value: 13},
	{CountryCode: "CA", Subdivision: "NS", Value: 15},
	{CountryCode: "CA", Subdivision: "NU", Value: 5},
	{CountryCode: "CA", Subdivision: "NB", Value: 13},
	{CountryCode: "CA", Subdivision: "NL", Value: 13},
	{CountryCode: "CA", Subdivision: "ON", Value: 13},
	{CountryCode: "CA", Subdivision: "PE", Value: 14},
	{CountryCode: "CA", Subdivision: "SK", Value: 10},
	{CountryCode: "CA", Subdivision: "NT", Value: 5},
	{CountryCode: "CA", Subdivision: "YT", Value: 5},
	{CountryCode: "QA", Subdivision: "", Value: 30},
	{CountryCode: "KE", Subdivision: "", Value: 30},
	{CountryCode: "CY", Subdivision: "", Value: 19},
	{CountryCode: "KG", Subdivision: "", Value: 30},
	{CountryCode: "KI", Subdivision: "", Value: 30},
	{CountryCode: "CN", Subdivision: "", Value: 17.6},
	{CountryCode: "CO", Subdivision: "", Value: 30},
	{CountryCode: "CG", Subdivision: "", Value: 30},
	{CountryCode: "CR", Subdivision: "", Value: 30},
	{CountryCode: "CI", Subdivision: "", Value: 20},
	{CountryCode: "CU", Subdivision: "", Value: 30},
	{CountryCode: "KW", Subdivision: "", Value: 30},
	{CountryCode: "LV", Subdivision: "", Value: 21},
	{CountryCode: "LS", Subdivision: "", Value: 30},
	{CountryCode: "LR", Subdivision: "", Value: 30},
	{CountryCode: "LB", Subdivision: "", Value: 10},
	{CountryCode: "LY", Subdivision: "", Value: 30},
	{CountryCode: "LT", Subdivision: "", Value: 21},
	{CountryCode: "LI", Subdivision: "", Value: 30},
	{CountryCode: "LU", Subdivision: "", Value: 17},
	{CountryCode: "MU", Subdivision: "", Value: 30},
	{CountryCode: "MR", Subdivision: "", Value: 30},
	{CountryCode: "MG", Subdivision: "", Value: 30},
	{CountryCode: "MO", Subdivision: "", Value: 13},
	{CountryCode: "MK", Subdivision: "", Value: 18},
	{CountryCode: "MW", Subdivision: "", Value: 30},
	{CountryCode: "MY", Subdivision: "", Value: 5},
	{CountryCode: "ML", Subdivision: "", Value: 30},
	{CountryCode: "MV", Subdivision: "", Value: 30},
	{CountryCode: "MT", Subdivision: "", Value: 18},
	{CountryCode: "MA", Subdivision: "", Value: 30},
	{CountryCode: "MQ", Subdivision: "", Value: 30},
	{CountryCode: "MX", Subdivision: "", Value: 16},
	{CountryCode: "MZ", Subdivision: "", Value: 30},
	{CountryCode: "MD", Subdivision: "", Value: 30},
	{CountryCode: "MC", Subdivision: "", Value: 30},
	{CountryCode: "MN", Subdivision: "", Value: 30},
	{CountryCode: "MM", Subdivision: "", Value: 30},
	{CountryCode: "NA", Subdivision: "", Value: 30},
	{CountryCode: "NR", Subdivision: "", Value: 30},
	{CountryCode: "NP", Subdivision: "", Value: 30},
	{CountryCode: "NE", Subdivision: "", Value: 30},
	{CountryCode: "NG", Subdivision: "", Value: 30},
	{CountryCode: "NL", Subdivision: "", Value: 21},
	{CountryCode: "NI", Subdivision: "", Value: 30},
	{CountryCode: "NZ", Subdivision: "", Value: 15},
	{CountryCode: "NO", Subdivision: "", Value: 25},
	{CountryCode: "AE", Subdivision: "", Value: 30},
	{CountryCode: "OM", Subdivision: "", Value: 30},
	{CountryCode: "PK", Subdivision: "", Value: 30},
	{CountryCode: "PA", Subdivision: "", Value: 30},
	{CountryCode: "PG", Subdivision: "", Value: 30},
	{CountryCode: "PY", Subdivision: "", Value: 10},
	{CountryCode: "PE", Subdivision: "", Value: 18},
	{CountryCode: "PL", Subdivision: "", Value: 23},
	{CountryCode: "PT", Subdivision: "", Value: 23},
	{CountryCode: "RU", Subdivision: "", Value: 20},
	{CountryCode: "RW", Subdivision: "", Value: 30},
	{CountryCode: "RO", Subdivision: "", Value: 19},
	{CountryCode: "SV", Subdivision: "", Value: 13},
	{CountryCode: "WS", Subdivision: "", Value: 30},
	{CountryCode: "SA", Subdivision: "", Value: 30},
	{CountryCode: "SZ", Subdivision: "", Value: 30},
	{CountryCode: "SN", Subdivision: "", Value: 30},
	{CountryCode: "SG", Subdivision: "", Value: 5},
	{CountryCode: "SY", Subdivision: "", Value: 30},
	{CountryCode: "SK", Subdivision: "", Value: 20},
	{CountryCode: "SI", Subdivision: "", Value: 22},
	{CountryCode: "US", Subdivision: "AL", Value: 13.5},
	{CountryCode: "US", Subdivision: "AK", Value: 7},
	{CountryCode: "US", Subdivision: "AZ", Value: 10.725},
	{CountryCode: "US", Subdivision: "AR", Value: 11.625},
	{CountryCode: "US", Subdivision: "CA", Value: 10.25},
	{CountryCode: "US", Subdivision: "CO", Value: 10},
	{CountryCode: "US", Subdivision: "CT", Value: 6.35},
	{CountryCode: "US", Subdivision: "DE", Value: 0},
	{CountryCode: "US", Subdivision: "FL", Value: 7.5},
	{CountryCode: "US", Subdivision: "GA", Value: 8},
	{CountryCode: "US", Subdivision: "HI", Value: 4.712},
	{CountryCode: "US", Subdivision: "ID", Value: 8.5},
	{CountryCode: "US", Subdivision: "IL", Value: 10.25},
	{CountryCode: "US", Subdivision: "IN", Value: 7},
	{CountryCode: "US", Subdivision: "IA", Value: 7},
	{CountryCode: "US", Subdivision: "KS", Value: 10.15},
	{CountryCode: "US", Subdivision: "KY", Value: 6},
	{CountryCode: "US", Subdivision: "LA", Value: 12},
	{CountryCode: "US", Subdivision: "ME", Value: 5.5},
	{CountryCode: "US", Subdivision: "MD", Value: 6},
	{CountryCode: "US", Subdivision: "MA", Value: 6.25},
	{CountryCode: "US", Subdivision: "MI", Value: 6},
	{CountryCode: "US", Subdivision: "MN", Value: 7.875},
	{CountryCode: "US", Subdivision: "MS", Value: 7.25},
	{CountryCode: "US", Subdivision: "MO", Value: 10.85},
	{CountryCode: "US", Subdivision: "MT", Value: 0},
	{CountryCode: "US", Subdivision: "NE", Value: 7.5},
	{CountryCode: "US", Subdivision: "NV", Value: 8.15},
	{CountryCode: "US", Subdivision: "NH", Value: 0},
	{CountryCode: "US", Subdivision: "NJ", Value: 12.875},
	{CountryCode: "US", Subdivision: "NM", Value: 8.688},
	{CountryCode: "US", Subdivision: "NY", Value: 8.875},
	{CountryCode: "US", Subdivision: "NC", Value: 7.50},
	{CountryCode: "US", Subdivision: "ND", Value: 8},
	{CountryCode: "US", Subdivision: "OH", Value: 8},
	{CountryCode: "US", Subdivision: "OK", Value: 11},
	{CountryCode: "US", Subdivision: "OR", Value: 0},
	{CountryCode: "US", Subdivision: "PA", Value: 8},
	{CountryCode: "US", Subdivision: "RI", Value: 7},
	{CountryCode: "US", Subdivision: "SC", Value: 9},
	{CountryCode: "US", Subdivision: "SD", Value: 6},
	{CountryCode: "US", Subdivision: "TN", Value: 9.75},
	{CountryCode: "US", Subdivision: "TX", Value: 8.25},
	{CountryCode: "US", Subdivision: "UT", Value: 8.35},
	{CountryCode: "US", Subdivision: "VT", Value: 7},
	{CountryCode: "US", Subdivision: "VA", Value: 6},
	{CountryCode: "US", Subdivision: "WA", Value: 10.4},
	{CountryCode: "US", Subdivision: "WV", Value: 7},
	{CountryCode: "US", Subdivision: "WI", Value: 6.75},
	{CountryCode: "US", Subdivision: "WY", Value: 6},
	{CountryCode: "US", Subdivision: "DC", Value: 5.75},
	{CountryCode: "US", Subdivision: "GU", Value: 4},
	{CountryCode: "US", Subdivision: "PR", Value: 11.5},
	{CountryCode: "SO", Subdivision: "", Value: 30},
	{CountryCode: "SD", Subdivision: "", Value: 30},
	{CountryCode: "SR", Subdivision: "", Value: 30},
	{CountryCode: "TJ", Subdivision: "", Value: 30},
	{CountryCode: "TH", Subdivision: "", Value: 7},
	{CountryCode: "TW", Subdivision: "", Value: 30},
	{CountryCode: "TZ", Subdivision: "", Value: 30},
	{CountryCode: "TG", Subdivision: "", Value: 30},
	{CountryCode: "TO", Subdivision: "", Value: 30},
	{CountryCode: "TT", Subdivision: "", Value: 15},
	{CountryCode: "TV", Subdivision: "", Value: 30},
	{CountryCode: "TN", Subdivision: "", Value: 30},
	{CountryCode: "TM", Subdivision: "", Value: 30},
	{CountryCode: "TR", Subdivision: "", Value: 18},
	{CountryCode: "UG", Subdivision: "", Value: 30},
	{CountryCode: "UZ", Subdivision: "", Value: 20},
	{CountryCode: "UA", Subdivision: "", Value: 20},
	{CountryCode: "UY", Subdivision: "", Value: 23},
	{CountryCode: "FJ", Subdivision: "", Value: 30},
	{CountryCode: "PH", Subdivision: "", Value: 12},
	{CountryCode: "FI", Subdivision: "", Value: 24},
	{CountryCode: "FR", Subdivision: "", Value: 20},
	{CountryCode: "HR", Subdivision: "", Value: 22},
	{CountryCode: "TD", Subdivision: "", Value: 30},
	{CountryCode: "CZ", Subdivision: "", Value: 21},
	{CountryCode: "CL", Subdivision: "", Value: 19},
	{CountryCode: "CH", Subdivision: "", Value: 8},
	{CountryCode: "SE", Subdivision: "", Value: 25},
	{CountryCode: "LK", Subdivision: "", Value: 15},
	{CountryCode: "EC", Subdivision: "", Value: 12},
	{CountryCode: "ER", Subdivision: "", Value: 30},
	{CountryCode: "EE", Subdivision: "", Value: 20},
	{CountryCode: "ET", Subdivision: "", Value: 30},
	{CountryCode: "ZA", Subdivision: "", Value: 30},
	{CountryCode: "KR", Subdivision: "", Value: 10},
	{CountryCode: "JM", Subdivision: "", Value: 20},
	{CountryCode: "JP", Subdivision: "", Value: 8},
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

				vat := &model.Vat{
					Id:              bson.NewObjectId(),
					Country:         v.CountryCode,
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
