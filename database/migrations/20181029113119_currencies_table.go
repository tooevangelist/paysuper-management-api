package migrations

import (
	"github.com/globalsign/mgo"
	"github.com/paysuper/paysuper-management-api/database/model"
	"github.com/paysuper/paysuper-management-api/manager"
	"github.com/xakep666/mongo-migrate"
	"time"
)

var currencies = []*model.Currency{
	{
		CodeInt:   36,
		CodeA3:    "AUD",
		Name:      &model.Name{RU: "Австралийский доллар", EN: "Australian Dollar"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   944,
		CodeA3:    "AZN",
		Name:      &model.Name{RU: "Азербайджанский манат", EN: "Azerbaijan Manat"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   51,
		CodeA3:    "AMD",
		Name:      &model.Name{RU: "Армянский драм", EN: "Armenia Dram"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   933,
		CodeA3:    "BYN",
		Name:      &model.Name{RU: "Белорусский рубль", EN: "Belarussian Ruble"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   975,
		CodeA3:    "BGN",
		Name:      &model.Name{RU: "Болгарский лев", EN: "Bulgarian Lev"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   986,
		CodeA3:    "BRL",
		Name:      &model.Name{RU: "Бразильский реал", EN: "Brazil Real"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   348,
		CodeA3:    "HUF",
		Name:      &model.Name{RU: "Венгерский форинт", EN: "Hungarian Forint"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   410,
		CodeA3:    "KRW",
		Name:      &model.Name{RU: "Вон Республики Корея", EN: "South Korean Won"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   344,
		CodeA3:    "HKD",
		Name:      &model.Name{RU: "Гонконгский доллар", EN: "Hong Kong Dollar"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   208,
		CodeA3:    "DKK",
		Name:      &model.Name{RU: "Датская крона", EN: "Danish Krone"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   840,
		CodeA3:    "USD",
		Name:      &model.Name{RU: "Доллар США", EN: "US Dollar"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   978,
		CodeA3:    "EUR",
		Name:      &model.Name{RU: "Евро", EN: "Euro"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   356,
		CodeA3:    "INR",
		Name:      &model.Name{RU: "Индийская рупия", EN: "Indian Rupee"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   398,
		CodeA3:    "KZT",
		Name:      &model.Name{RU: "Казахстанский тенге", EN: "Kazakhstan Tenge"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   124,
		CodeA3:    "CAD",
		Name:      &model.Name{RU: "Канадский доллар", EN: "Canadian Dollar"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   417,
		CodeA3:    "KGS",
		Name:      &model.Name{RU: "Киргизский сом", EN: "Kyrgyzstan Som"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   156,
		CodeA3:    "CNY",
		Name:      &model.Name{RU: "Китайский юань", EN: "China Yuan"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   498,
		CodeA3:    "MDL",
		Name:      &model.Name{RU: "Молдавская лея", EN: "Moldova Lei"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   934,
		CodeA3:    "TMT",
		Name:      &model.Name{RU: "Туркменский манат", EN: "Turkmenistan Manat"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   578,
		CodeA3:    "NOK",
		Name:      &model.Name{RU: "Норвежская крона", EN: "Norwegian Krone"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   985,
		CodeA3:    "PLN",
		Name:      &model.Name{RU: "Польский злотый", EN: "Polish Zloty"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   946,
		CodeA3:    "RON",
		Name:      &model.Name{RU: "Румынский лей", EN: "Romanian Leu"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   702,
		CodeA3:    "SGD",
		Name:      &model.Name{RU: "Сингапурский доллар", EN: "Singapore Dollar"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{CodeInt: 972,
		CodeA3: "TJS",
		Name: &model.Name{RU: "Таджикский сомони",
			EN: "Tajikistan Ruble",
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   949,
		CodeA3:    "TRY",
		Name:      &model.Name{RU: "Турецкая лира", EN: "Turkish Lira"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   860,
		CodeA3:    "UZS",
		Name:      &model.Name{RU: "Узбекский сум", EN: "Uzbekistan Sum"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   980,
		CodeA3:    "UAH",
		Name:      &model.Name{RU: "Украинская гривна", EN: "Ukrainian Hryvnia"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   826,
		CodeA3:    "GBP",
		Name:      &model.Name{RU: "Фунт стерлингов Соединенного королевства", EN: "British Pound Sterling"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   203,
		CodeA3:    "CZK",
		Name:      &model.Name{RU: "Чешская крона", EN: "Czech Koruna"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   752,
		CodeA3:    "SEK",
		Name:      &model.Name{RU: "Шведская крона", EN: "Swedish Krona"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   756,
		CodeA3:    "CHF",
		Name:      &model.Name{RU: "Швейцарский франк", EN: "Swiss Franc"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   710,
		CodeA3:    "ZAR",
		Name:      &model.Name{RU: "Южноафриканский рэнд", EN: "South African Rand"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   392,
		CodeA3:    "JPY",
		Name:      &model.Name{RU: "Японская иена", EN: "Japanese Yen"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		CodeInt:   643,
		CodeA3:    "RUB",
		Name:      &model.Name{RU: "Российский рубль", EN: "Russian ruble"},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
}

func init() {
	err := migrate.Register(
		func(db *mgo.Database) error {
			var err error

			c := db.C(manager.TableCurrency)

			err = c.EnsureIndex(mgo.Index{Name: "currency_code_int_uniq", Key: []string{"code_int"}, Unique: true})

			if err != nil {
				return err
			}

			err = c.EnsureIndex(mgo.Index{Name: "currency_code_a3_uniq", Key: []string{"code_a3"}, Unique: true})

			if err != nil {
				return err
			}

			err = c.EnsureIndex(mgo.Index{Name: "currency_is_active_idx", Key: []string{"is_active"}})

			if err != nil {
				return err
			}

			var iCurrencies []interface{}

			for _, t := range currencies {
				iCurrencies = append(iCurrencies, t)
			}

			return c.Insert(iCurrencies...)
		},
		func(db *mgo.Database) error {
			return db.C(manager.TableCurrency).DropCollection()
		},
	)

	if err != nil {
		return
	}
}
