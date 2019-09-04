package mock

import (
	"github.com/globalsign/mgo/bson"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
)

const (
	SomeAgreementName  = "some_name.pdf"
	SomeAgreementName1 = "some_name1.pdf"
	SomeAgreementName2 = "some_name2.pdf"
)

var (
	SomeError = &grpc.ResponseErrorMessage{Message: "some error"}

	SomeMerchantId  = bson.NewObjectId().Hex()
	SomeMerchantId1 = bson.NewObjectId().Hex()
	SomeMerchantId2 = bson.NewObjectId().Hex()
	SomeMerchantId3 = bson.NewObjectId().Hex()

	OnboardingMerchantMock = &billing.Merchant{
		Id: bson.NewObjectId().Hex(),
		Company: &billing.MerchantCompanyInfo{
			Name:    "merchant1",
			Country: "RU",
			Zip:     "190000",
			City:    "St.Petersburg",
		},
		Contacts: &billing.MerchantContact{
			Authorized: &billing.MerchantContactAuthorized{
				Name:     "Unit Test",
				Email:    "test@unit.test",
				Phone:    "123456789",
				Position: "Unit Test",
			},
			Technical: &billing.MerchantContactTechnical{
				Name:  "Unit Test",
				Email: "test@unit.test",
				Phone: "123456789",
			},
		},
		Banking: &billing.MerchantBanking{
			Currency: "RUB",
			Name:     "Bank name",
		},
		IsVatEnabled:              true,
		IsCommissionToUserEnabled: true,
		Status:                    pkg.MerchantStatusOnReview,
		LastPayout:                &billing.MerchantLastPayout{},
		IsSigned:                  true,
		PaymentMethods: map[string]*billing.MerchantPaymentMethod{
			bson.NewObjectId().Hex(): {
				PaymentMethod: &billing.MerchantPaymentMethodIdentification{
					Id:   bson.NewObjectId().Hex(),
					Name: "Bank card",
				},
				Commission: &billing.MerchantPaymentMethodCommissions{
					Fee: 2.5,
					PerTransaction: &billing.MerchantPaymentMethodPerTransactionCommission{
						Fee:      30,
						Currency: "RUB",
					},
				},
				Integration: &billing.MerchantPaymentMethodIntegration{
					TerminalId:       "1234567890",
					TerminalPassword: "0987654321",
					Integrated:       true,
				},
				IsActive: true,
			},
		},
	}

	ProductPrice = &grpc.ProductPrice{
		Currency: "USD",
		Amount:   1010.23,
	}

	Product = &grpc.Product{
		Id:              "5c99391568add439ccf0ffaf",
		Object:          "product",
		Type:            "simple_product",
		Sku:             "ru_double_yeti_rel",
		Name:            map[string]string{"en": "Double Yeti"},
		DefaultCurrency: "USD",
		Enabled:         true,
		Description:     map[string]string{"en": "Yet another cool game"},
		LongDescription: map[string]string{"en": "Super game steam keys"},
		Url:             "http://mygame.ru/duoble_yeti",
		Images:          []string{"/home/image.jpg"},
		MerchantId:      "5bdc35de5d1e1100019fb7db",
		Metadata: map[string]string{
			"SomeKey": "SomeValue",
		},
		Prices: []*grpc.ProductPrice{
			ProductPrice,
		},
	}
)
