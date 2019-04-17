package mock

import (
	"context"
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/micro/go-micro/client"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
)

const (
	SomeError          = "some error"
	SomeError1         = "some error1"
	SomeAgreementName  = "some_name.pdf"
	SomeAgreementName1 = "some_name1.pdf"
	SomeAgreementName2 = "some_name2.pdf"
)

var (
	SomeMerchantId  = bson.NewObjectId().Hex()
	SomeMerchantId1 = bson.NewObjectId().Hex()
	SomeMerchantId2 = bson.NewObjectId().Hex()
	SomeMerchantId3 = bson.NewObjectId().Hex()

	OnboardingMerchantMock = &billing.Merchant{
		Id:   bson.NewObjectId().Hex(),
		Name: "Unit test",
		Country: &billing.Country{
			CodeInt:  643,
			CodeA2:   "RU",
			CodeA3:   "RUS",
			Name:     &billing.Name{Ru: "Россия", En: "Russia (Russian Federation)"},
			IsActive: true,
		},
		Zip:  "190000",
		City: "St.Petersburg",
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
			Currency: &billing.Currency{
				CodeInt:  643,
				CodeA3:   "RUB",
				Name:     &billing.Name{Ru: "Российский рубль", En: "Russian ruble"},
				IsActive: true,
			},
			Name: "Bank name",
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

type BillingServerOkMock struct{}
type BillingServerOkTemporaryMock struct{}
type BillingServerErrorMock struct{}
type BillingServerSystemErrorMock struct{}

func NewBillingServerOkMock() grpc.BillingService {
	return &BillingServerOkMock{}
}

func NewBillingServerErrorMock() grpc.BillingService {
	return &BillingServerErrorMock{}
}

func NewBillingServerSystemErrorMock() grpc.BillingService {
	return &BillingServerSystemErrorMock{}
}

func NewBillingServerOkTemporaryMock() grpc.BillingService {
	return &BillingServerOkTemporaryMock{}
}

func (s *BillingServerOkMock) GetProductsForOrder(
	ctx context.Context,
	in *grpc.GetProductsForOrderRequest,
	opts ...client.CallOption,
) (*grpc.ListProductsResponse, error) {
	return &grpc.ListProductsResponse{}, nil
}

func (s *BillingServerOkMock) OrderCreateProcess(
	ctx context.Context,
	in *billing.OrderCreateRequest,
	opts ...client.CallOption,
) (*billing.Order, error) {
	if in.ProjectId == SomeMerchantId {
		return nil, errors.New(SomeError)
	}

	if in.ProjectId == SomeMerchantId1 {
		return &billing.Order{Id: SomeMerchantId2, Uuid: SomeMerchantId2}, nil
	}

	return &billing.Order{}, nil
}

func (s *BillingServerOkMock) PaymentFormJsonDataProcess(
	ctx context.Context,
	in *grpc.PaymentFormJsonDataRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormJsonDataResponse, error) {
	if in.OrderId == SomeMerchantId2 {
		return nil, errors.New(SomeError1)
	}

	rsp := &grpc.PaymentFormJsonDataResponse{
		Id:          bson.NewObjectId().Hex(),
		Account:     bson.NewObjectId().Hex(),
		HasVat:      false,
		Vat:         0,
		Amount:      10,
		TotalAmount: 10,
		Currency:    "RUB",
		Project: &grpc.PaymentFormJsonDataProject{
			Name:       bson.NewObjectId().Hex(),
			UrlSuccess: "",
			UrlFail:    "",
		},
		PaymentMethods: []*billing.PaymentFormPaymentMethod{
			{
				Id:            bson.NewObjectId().Hex(),
				Name:          "Bank card",
				Icon:          "",
				Type:          "bank_card",
				Group:         "BANKCARD",
				AccountRegexp: "^(?:4[0-9]{12}(?:[0-9]{3})?|[25][1-7][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11})$",
				HasSavedCards: false,
			},
		},
		Token:                 bson.NewObjectId().Hex(),
		InlineFormRedirectUrl: "",
		HasCustomerToken:      false,
		CustomerToken:         bson.NewObjectId().Hex(),
	}

	if in.OrderId == SomeMerchantId1 {
		rsp.HasCustomerToken = true
	} else {
		rsp.HasCustomerToken = false
	}

	return rsp, nil
}

func (s *BillingServerOkMock) PaymentCreateProcess(
	ctx context.Context,
	in *grpc.PaymentCreateRequest,
	opts ...client.CallOption,
) (*grpc.PaymentCreateResponse, error) {
	return &grpc.PaymentCreateResponse{}, nil
}

func (s *BillingServerOkMock) PaymentCallbackProcess(
	ctx context.Context,
	in *grpc.PaymentNotifyRequest,
	opts ...client.CallOption,
) (*grpc.PaymentNotifyResponse, error) {
	return &grpc.PaymentNotifyResponse{}, nil
}

func (s *BillingServerOkMock) RebuildCache(
	ctx context.Context,
	in *grpc.EmptyRequest,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerOkMock) UpdateOrder(
	ctx context.Context,
	in *billing.Order,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerOkMock) UpdateMerchant(
	ctx context.Context,
	in *billing.Merchant,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerOkMock) GetConvertRate(
	ctx context.Context,
	in *grpc.ConvertRateRequest,
	opts ...client.CallOption,
) (*grpc.ConvertRateResponse, error) {
	return &grpc.ConvertRateResponse{}, nil
}

func (s *BillingServerOkMock) GetMerchantBy(
	ctx context.Context,
	in *grpc.GetMerchantByRequest,
	opts ...client.CallOption,
) (*grpc.MerchantGetMerchantResponse, error) {
	if in.MerchantId == OnboardingMerchantMock.Id {
		OnboardingMerchantMock.S3AgreementName = SomeAgreementName
	} else {
		if in.MerchantId == SomeMerchantId1 {
			OnboardingMerchantMock.S3AgreementName = SomeAgreementName1
		} else {
			if in.MerchantId == SomeMerchantId2 {
				OnboardingMerchantMock.S3AgreementName = SomeAgreementName2
			} else {
				OnboardingMerchantMock.S3AgreementName = ""
			}
		}
	}

	if in.MerchantId == SomeMerchantId3 {
		OnboardingMerchantMock.Status = pkg.MerchantStatusDraft
	} else {
		OnboardingMerchantMock.Status = pkg.MerchantStatusOnReview
	}

	rsp := &grpc.MerchantGetMerchantResponse{
		Status:  pkg.ResponseStatusOk,
		Message: "",
		Item:    OnboardingMerchantMock,
	}

	return rsp, nil
}

func (s *BillingServerOkMock) ListMerchants(
	ctx context.Context,
	in *grpc.MerchantListingRequest,
	opts ...client.CallOption,
) (*grpc.MerchantListingResponse, error) {
	return &grpc.MerchantListingResponse{
		Count: 3,
		Items: []*billing.Merchant{OnboardingMerchantMock, OnboardingMerchantMock, OnboardingMerchantMock},
	}, nil
}

func (s *BillingServerOkMock) ChangeMerchant(
	ctx context.Context,
	in *grpc.OnboardingRequest,
	opts ...client.CallOption,
) (*billing.Merchant, error) {
	m := &billing.Merchant{
		User: &billing.MerchantUser{
			Id:    bson.NewObjectId().Hex(),
			Email: "test@unit.test",
		},
		Name:            in.Name,
		AlternativeName: in.AlternativeName,
		Website:         in.Website,
		Country: &billing.Country{
			CodeInt:  643,
			CodeA3:   "RUS",
			CodeA2:   in.Country,
			IsActive: true,
		},
		State:              in.State,
		Zip:                in.Zip,
		City:               in.City,
		Address:            in.Address,
		AddressAdditional:  in.AddressAdditional,
		RegistrationNumber: in.RegistrationNumber,
		TaxId:              in.TaxId,
		Contacts:           in.Contacts,
		Banking: &billing.MerchantBanking{
			Currency: &billing.Currency{
				CodeInt:  643,
				CodeA3:   in.Banking.Currency,
				Name:     &billing.Name{Ru: "Российский рубль", En: "Russian ruble"},
				IsActive: true,
			},
			Name:          in.Banking.Name,
			Address:       in.Banking.Address,
			AccountNumber: in.Banking.AccountNumber,
			Swift:         in.Banking.Swift,
			Details:       in.Banking.Details,
		},
		Status: pkg.MerchantStatusDraft,
	}

	if in.Id != "" {
		m.Id = in.Id
	} else {
		m.Id = bson.NewObjectId().Hex()
	}

	return m, nil
}

func (s *BillingServerOkMock) ChangeMerchantStatus(
	ctx context.Context,
	in *grpc.MerchantChangeStatusRequest,
	opts ...client.CallOption,
) (*billing.Merchant, error) {
	return &billing.Merchant{Id: in.MerchantId, Status: in.Status}, nil
}

func (s *BillingServerOkMock) CreateNotification(
	ctx context.Context,
	in *grpc.NotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return &billing.Notification{}, nil
}

func (s *BillingServerOkMock) GetNotification(
	ctx context.Context,
	in *grpc.GetNotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return &billing.Notification{}, nil
}

func (s *BillingServerOkMock) ListNotifications(
	ctx context.Context,
	in *grpc.ListingNotificationRequest,
	opts ...client.CallOption,
) (*grpc.Notifications, error) {
	return &grpc.Notifications{}, nil
}

func (s *BillingServerOkMock) MarkNotificationAsRead(
	ctx context.Context,
	in *grpc.GetNotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return &billing.Notification{}, nil
}

func (s *BillingServerOkMock) ListMerchantPaymentMethods(
	ctx context.Context,
	in *grpc.ListMerchantPaymentMethodsRequest,
	opts ...client.CallOption,
) (*grpc.ListingMerchantPaymentMethod, error) {
	return &grpc.ListingMerchantPaymentMethod{}, nil
}

func (s *BillingServerOkMock) GetMerchantPaymentMethod(
	ctx context.Context,
	in *grpc.GetMerchantPaymentMethodRequest,
	opts ...client.CallOption,
) (*billing.MerchantPaymentMethod, error) {
	return &billing.MerchantPaymentMethod{}, nil
}

func (s *BillingServerOkMock) ChangeMerchantPaymentMethod(
	ctx context.Context,
	in *grpc.MerchantPaymentMethodRequest,
	opts ...client.CallOption,
) (*grpc.MerchantPaymentMethodResponse, error) {
	return &grpc.MerchantPaymentMethodResponse{
		Status: pkg.ResponseStatusOk,
		Item: &billing.MerchantPaymentMethod{
			PaymentMethod: &billing.MerchantPaymentMethodIdentification{
				Id:   in.PaymentMethod.Id,
				Name: in.PaymentMethod.Name,
			},
			Commission:  in.Commission,
			Integration: in.Integration,
			IsActive:    in.IsActive,
		},
	}, nil
}

func (s *BillingServerOkMock) CreateRefund(
	ctx context.Context,
	in *grpc.CreateRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return &grpc.CreateRefundResponse{
		Status: pkg.ResponseStatusOk,
		Item:   &billing.Refund{},
	}, nil
}

func (s *BillingServerOkMock) ListRefunds(
	ctx context.Context,
	in *grpc.ListRefundsRequest,
	opts ...client.CallOption,
) (*grpc.ListRefundsResponse, error) {
	return &grpc.ListRefundsResponse{}, nil
}

func (s *BillingServerOkMock) GetRefund(
	ctx context.Context,
	in *grpc.GetRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return &grpc.CreateRefundResponse{
		Status: pkg.ResponseStatusOk,
		Item:   &billing.Refund{},
	}, nil
}

func (s *BillingServerOkMock) ProcessRefundCallback(
	ctx context.Context,
	in *grpc.CallbackRequest,
	opts ...client.CallOption,
) (*grpc.PaymentNotifyResponse, error) {
	return &grpc.PaymentNotifyResponse{
		Status: pkg.ResponseStatusOk,
	}, nil
}

func (s *BillingServerOkMock) PaymentFormLanguageChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangeLangRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return &grpc.PaymentFormDataChangeResponse{
		Status: pkg.ResponseStatusOk,
		Item: &grpc.PaymentFormDataChangeResponseItem{
			UserAddressDataRequired: true,
			UserIpData: &grpc.UserIpData{
				Country: "RU",
				City:    "St.Petersburg",
				Zip:     "190000",
			},
		},
	}, nil
}

func (s *BillingServerOkMock) PaymentFormPaymentAccountChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangePaymentAccountRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return &grpc.PaymentFormDataChangeResponse{
		Status: pkg.ResponseStatusOk,
		Item: &grpc.PaymentFormDataChangeResponseItem{
			UserAddressDataRequired: true,
			UserIpData: &grpc.UserIpData{
				Country: "RU",
				City:    "St.Petersburg",
				Zip:     "190000",
			},
		},
	}, nil
}

func (s *BillingServerOkMock) ProcessBillingAddress(
	ctx context.Context,
	in *grpc.ProcessBillingAddressRequest,
	opts ...client.CallOption,
) (*grpc.ProcessBillingAddressResponse, error) {
	return &grpc.ProcessBillingAddressResponse{
		Status: pkg.ResponseStatusOk,
		Item: &grpc.ProcessBillingAddressResponseItem{
			HasVat:      true,
			Vat:         10,
			Amount:      10,
			TotalAmount: 20,
		},
	}, nil
}

func (s *BillingServerOkMock) ChangeMerchantData(
	ctx context.Context,
	in *grpc.ChangeMerchantDataRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	rsp := &grpc.ChangeMerchantDataResponse{
		Status: pkg.ResponseStatusOk,
		Item:   OnboardingMerchantMock,
	}

	if in.MerchantId == SomeMerchantId {
		return nil, errors.New(SomeError)
	}

	return rsp, nil
}

func (s *BillingServerOkMock) SetMerchantS3Agreement(
	ctx context.Context,
	in *grpc.SetMerchantS3AgreementRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	rsp := &grpc.ChangeMerchantDataResponse{
		Status: pkg.ResponseStatusOk,
		Item:   OnboardingMerchantMock,
	}

	if in.MerchantId == SomeMerchantId {
		return nil, errors.New(SomeError)
	}

	return rsp, nil
}

func (s *BillingServerOkMock) ChangeCustomer(
	ctx context.Context,
	in *billing.Customer,
	opts ...client.CallOption,
) (*grpc.ChangeCustomerResponse, error) {
	return &grpc.ChangeCustomerResponse{
		Status: pkg.ResponseStatusOk,
		Item: &billing.Customer{
			Token: bson.NewObjectId().Hex(),
		},
	}, nil
}

func (s *BillingServerOkMock) CheckProjectRequestSignature(
	ctx context.Context,
	in *grpc.CheckProjectRequestSignatureRequest,
	opts ...client.CallOption,
) (*grpc.CheckProjectRequestSignatureResponse, error) {
	return &grpc.CheckProjectRequestSignatureResponse{Status: pkg.ResponseStatusOk}, nil
}

func (s *BillingServerErrorMock) GetProductsForOrder(
	ctx context.Context,
	in *grpc.GetProductsForOrderRequest,
	opts ...client.CallOption,
) (*grpc.ListProductsResponse, error) {
	return &grpc.ListProductsResponse{}, nil
}

func (s *BillingServerErrorMock) OrderCreateProcess(
	ctx context.Context,
	in *billing.OrderCreateRequest,
	opts ...client.CallOption,
) (*billing.Order, error) {
	return &billing.Order{}, nil
}

func (s *BillingServerErrorMock) PaymentFormJsonDataProcess(
	ctx context.Context,
	in *grpc.PaymentFormJsonDataRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormJsonDataResponse, error) {
	return &grpc.PaymentFormJsonDataResponse{}, nil
}

func (s *BillingServerErrorMock) PaymentCreateProcess(
	ctx context.Context,
	in *grpc.PaymentCreateRequest,
	opts ...client.CallOption,
) (*grpc.PaymentCreateResponse, error) {
	return &grpc.PaymentCreateResponse{}, nil
}

func (s *BillingServerErrorMock) PaymentCallbackProcess(
	ctx context.Context,
	in *grpc.PaymentNotifyRequest,
	opts ...client.CallOption,
) (*grpc.PaymentNotifyResponse, error) {
	return &grpc.PaymentNotifyResponse{}, nil
}

func (s *BillingServerErrorMock) RebuildCache(
	ctx context.Context,
	in *grpc.EmptyRequest,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerErrorMock) UpdateOrder(
	ctx context.Context,
	in *billing.Order,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerErrorMock) UpdateMerchant(
	ctx context.Context,
	in *billing.Merchant,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerErrorMock) GetConvertRate(
	ctx context.Context,
	in *grpc.ConvertRateRequest,
	opts ...client.CallOption,
) (*grpc.ConvertRateResponse, error) {
	return &grpc.ConvertRateResponse{}, nil
}

func (s *BillingServerErrorMock) GetMerchantBy(
	ctx context.Context,
	in *grpc.GetMerchantByRequest,
	opts ...client.CallOption,
) (*grpc.MerchantGetMerchantResponse, error) {
	return &grpc.MerchantGetMerchantResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) ListMerchants(
	ctx context.Context,
	in *grpc.MerchantListingRequest,
	opts ...client.CallOption,
) (*grpc.MerchantListingResponse, error) {
	return &grpc.MerchantListingResponse{}, nil
}

func (s *BillingServerErrorMock) ChangeMerchant(
	ctx context.Context,
	in *grpc.OnboardingRequest,
	opts ...client.CallOption,
) (*billing.Merchant, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) ChangeMerchantStatus(
	ctx context.Context,
	in *grpc.MerchantChangeStatusRequest,
	opts ...client.CallOption,
) (*billing.Merchant, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) CreateNotification(
	ctx context.Context,
	in *grpc.NotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) GetNotification(
	ctx context.Context,
	in *grpc.GetNotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) ListNotifications(
	ctx context.Context,
	in *grpc.ListingNotificationRequest,
	opts ...client.CallOption,
) (*grpc.Notifications, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) MarkNotificationAsRead(
	ctx context.Context,
	in *grpc.GetNotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) ListMerchantPaymentMethods(
	ctx context.Context,
	in *grpc.ListMerchantPaymentMethodsRequest,
	opts ...client.CallOption,
) (*grpc.ListingMerchantPaymentMethod, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) GetMerchantPaymentMethod(
	ctx context.Context,
	in *grpc.GetMerchantPaymentMethodRequest,
	opts ...client.CallOption,
) (*billing.MerchantPaymentMethod, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) ChangeMerchantPaymentMethod(
	ctx context.Context,
	in *grpc.MerchantPaymentMethodRequest,
	opts ...client.CallOption,
) (*grpc.MerchantPaymentMethodResponse, error) {
	return &grpc.MerchantPaymentMethodResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) CreateRefund(
	ctx context.Context,
	in *grpc.CreateRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return &grpc.CreateRefundResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) ListRefunds(
	ctx context.Context,
	in *grpc.ListRefundsRequest,
	opts ...client.CallOption,
) (*grpc.ListRefundsResponse, error) {
	return &grpc.ListRefundsResponse{}, nil
}

func (s *BillingServerErrorMock) GetRefund(
	ctx context.Context,
	in *grpc.GetRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return &grpc.CreateRefundResponse{
		Status:  pkg.ResponseStatusNotFound,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) ProcessRefundCallback(
	ctx context.Context,
	in *grpc.CallbackRequest,
	opts ...client.CallOption,
) (*grpc.PaymentNotifyResponse, error) {
	return &grpc.PaymentNotifyResponse{
		Status: pkg.ResponseStatusNotFound,
		Error:  SomeError,
	}, nil
}

func (s *BillingServerErrorMock) PaymentFormLanguageChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangeLangRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return &grpc.PaymentFormDataChangeResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) PaymentFormPaymentAccountChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangePaymentAccountRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return &grpc.PaymentFormDataChangeResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) ProcessBillingAddress(
	ctx context.Context,
	in *grpc.ProcessBillingAddressRequest,
	opts ...client.CallOption,
) (*grpc.ProcessBillingAddressResponse, error) {
	return &grpc.ProcessBillingAddressResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) ChangeMerchantData(
	ctx context.Context,
	in *grpc.ChangeMerchantDataRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	return &grpc.ChangeMerchantDataResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) SetMerchantS3Agreement(
	ctx context.Context,
	in *grpc.SetMerchantS3AgreementRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	return &grpc.ChangeMerchantDataResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) ChangeCustomer(
	ctx context.Context,
	in *billing.Customer,
	opts ...client.CallOption,
) (*grpc.ChangeCustomerResponse, error) {
	return &grpc.ChangeCustomerResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerErrorMock) CheckProjectRequestSignature(
	ctx context.Context,
	in *grpc.CheckProjectRequestSignatureRequest,
	opts ...client.CallOption,
) (*grpc.CheckProjectRequestSignatureResponse, error) {
	if in.ProjectId == SomeMerchantId {
		return &grpc.CheckProjectRequestSignatureResponse{Status: pkg.ResponseStatusOk}, nil
	}

	return &grpc.CheckProjectRequestSignatureResponse{
		Status:  pkg.ResponseStatusBadData,
		Message: SomeError,
	}, nil
}

func (s *BillingServerSystemErrorMock) GetProductsForOrder(
	ctx context.Context,
	in *grpc.GetProductsForOrderRequest,
	opts ...client.CallOption,
) (*grpc.ListProductsResponse, error) {
	return &grpc.ListProductsResponse{}, nil
}

func (s *BillingServerSystemErrorMock) OrderCreateProcess(
	ctx context.Context,
	in *billing.OrderCreateRequest,
	opts ...client.CallOption,
) (*billing.Order, error) {
	return &billing.Order{}, nil
}

func (s *BillingServerSystemErrorMock) PaymentFormJsonDataProcess(
	ctx context.Context,
	in *grpc.PaymentFormJsonDataRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormJsonDataResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) PaymentCreateProcess(
	ctx context.Context,
	in *grpc.PaymentCreateRequest,
	opts ...client.CallOption,
) (*grpc.PaymentCreateResponse, error) {
	return &grpc.PaymentCreateResponse{}, nil
}

func (s *BillingServerSystemErrorMock) PaymentCallbackProcess(
	ctx context.Context,
	in *grpc.PaymentNotifyRequest,
	opts ...client.CallOption,
) (*grpc.PaymentNotifyResponse, error) {
	return &grpc.PaymentNotifyResponse{}, nil
}

func (s *BillingServerSystemErrorMock) RebuildCache(
	ctx context.Context,
	in *grpc.EmptyRequest,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerSystemErrorMock) UpdateOrder(
	ctx context.Context,
	in *billing.Order,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerSystemErrorMock) UpdateMerchant(
	ctx context.Context,
	in *billing.Merchant,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerSystemErrorMock) GetConvertRate(
	ctx context.Context,
	in *grpc.ConvertRateRequest,
	opts ...client.CallOption,
) (*grpc.ConvertRateResponse, error) {
	return &grpc.ConvertRateResponse{}, nil
}

func (s *BillingServerSystemErrorMock) GetMerchantBy(
	ctx context.Context,
	in *grpc.GetMerchantByRequest,
	opts ...client.CallOption,
) (*grpc.MerchantGetMerchantResponse, error) {
	return nil, errors.New("some error")
}

func (s *BillingServerSystemErrorMock) ListMerchants(
	ctx context.Context,
	in *grpc.MerchantListingRequest,
	opts ...client.CallOption,
) (*grpc.MerchantListingResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) ChangeMerchant(
	ctx context.Context,
	in *grpc.OnboardingRequest,
	opts ...client.CallOption,
) (*billing.Merchant, error) {
	return &billing.Merchant{}, nil
}

func (s *BillingServerSystemErrorMock) ChangeMerchantStatus(
	ctx context.Context,
	in *grpc.MerchantChangeStatusRequest,
	opts ...client.CallOption,
) (*billing.Merchant, error) {
	return &billing.Merchant{}, nil
}

func (s *BillingServerSystemErrorMock) CreateNotification(
	ctx context.Context,
	in *grpc.NotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return &billing.Notification{}, nil
}

func (s *BillingServerSystemErrorMock) GetNotification(
	ctx context.Context,
	in *grpc.GetNotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return &billing.Notification{}, nil
}

func (s *BillingServerSystemErrorMock) ListNotifications(
	ctx context.Context,
	in *grpc.ListingNotificationRequest,
	opts ...client.CallOption,
) (*grpc.Notifications, error) {
	return &grpc.Notifications{}, nil
}

func (s *BillingServerSystemErrorMock) MarkNotificationAsRead(
	ctx context.Context,
	in *grpc.GetNotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return &billing.Notification{}, nil
}

func (s *BillingServerSystemErrorMock) ListMerchantPaymentMethods(
	ctx context.Context,
	in *grpc.ListMerchantPaymentMethodsRequest,
	opts ...client.CallOption,
) (*grpc.ListingMerchantPaymentMethod, error) {
	return &grpc.ListingMerchantPaymentMethod{}, nil
}

func (s *BillingServerSystemErrorMock) GetMerchantPaymentMethod(
	ctx context.Context,
	in *grpc.GetMerchantPaymentMethodRequest,
	opts ...client.CallOption,
) (*billing.MerchantPaymentMethod, error) {
	return &billing.MerchantPaymentMethod{}, nil
}

func (s *BillingServerSystemErrorMock) ChangeMerchantPaymentMethod(
	ctx context.Context,
	in *grpc.MerchantPaymentMethodRequest,
	opts ...client.CallOption,
) (*grpc.MerchantPaymentMethodResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) CreateRefund(
	ctx context.Context,
	in *grpc.CreateRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) ListRefunds(
	ctx context.Context,
	in *grpc.ListRefundsRequest,
	opts ...client.CallOption,
) (*grpc.ListRefundsResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) GetRefund(
	ctx context.Context,
	in *grpc.GetRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) ProcessRefundCallback(
	ctx context.Context,
	in *grpc.CallbackRequest,
	opts ...client.CallOption,
) (*grpc.PaymentNotifyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) PaymentFormLanguageChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangeLangRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) PaymentFormPaymentAccountChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangePaymentAccountRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) ProcessBillingAddress(
	ctx context.Context,
	in *grpc.ProcessBillingAddressRequest,
	opts ...client.CallOption,
) (*grpc.ProcessBillingAddressResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) ChangeMerchantData(
	ctx context.Context,
	in *grpc.ChangeMerchantDataRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) SetMerchantS3Agreement(
	ctx context.Context,
	in *grpc.SetMerchantS3AgreementRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) ChangeCustomer(
	ctx context.Context,
	in *billing.Customer,
	opts ...client.CallOption,
) (*grpc.ChangeCustomerResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) CheckProjectRequestSignature(
	ctx context.Context,
	in *grpc.CheckProjectRequestSignatureRequest,
	opts ...client.CallOption,
) (*grpc.CheckProjectRequestSignatureResponse, error) {
	if in.ProjectId == SomeMerchantId {
		return &grpc.CheckProjectRequestSignatureResponse{Status: pkg.ResponseStatusOk}, nil
	}

	return nil, errors.New(SomeError)
}

func (s *BillingServerOkTemporaryMock) GetProductsForOrder(
	ctx context.Context,
	in *grpc.GetProductsForOrderRequest,
	opts ...client.CallOption,
) (*grpc.ListProductsResponse, error) {
	return &grpc.ListProductsResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) OrderCreateProcess(
	ctx context.Context,
	in *billing.OrderCreateRequest,
	opts ...client.CallOption,
) (*billing.Order, error) {
	return &billing.Order{}, nil
}

func (s *BillingServerOkTemporaryMock) PaymentFormJsonDataProcess(
	ctx context.Context,
	in *grpc.PaymentFormJsonDataRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormJsonDataResponse, error) {
	return &grpc.PaymentFormJsonDataResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) PaymentCreateProcess(
	ctx context.Context,
	in *grpc.PaymentCreateRequest,
	opts ...client.CallOption,
) (*grpc.PaymentCreateResponse, error) {
	return &grpc.PaymentCreateResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) PaymentCallbackProcess(
	ctx context.Context,
	in *grpc.PaymentNotifyRequest,
	opts ...client.CallOption,
) (*grpc.PaymentNotifyResponse, error) {
	return &grpc.PaymentNotifyResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) RebuildCache(
	ctx context.Context,
	in *grpc.EmptyRequest,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) UpdateOrder(
	ctx context.Context,
	in *billing.Order,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) UpdateMerchant(
	ctx context.Context,
	in *billing.Merchant,
	opts ...client.CallOption,
) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) GetConvertRate(
	ctx context.Context,
	in *grpc.ConvertRateRequest,
	opts ...client.CallOption,
) (*grpc.ConvertRateResponse, error) {
	return &grpc.ConvertRateResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) GetMerchantBy(
	ctx context.Context,
	in *grpc.GetMerchantByRequest,
	opts ...client.CallOption,
) (*grpc.MerchantGetMerchantResponse, error) {
	rsp := &grpc.MerchantGetMerchantResponse{
		Status:  pkg.ResponseStatusOk,
		Message: "",
		Item:    OnboardingMerchantMock,
	}

	return rsp, nil
}

func (s *BillingServerOkTemporaryMock) ListMerchants(
	ctx context.Context,
	in *grpc.MerchantListingRequest,
	opts ...client.CallOption,
) (*grpc.MerchantListingResponse, error) {
	return &grpc.MerchantListingResponse{
		Count: 3,
		Items: []*billing.Merchant{OnboardingMerchantMock, OnboardingMerchantMock, OnboardingMerchantMock},
	}, nil
}

func (s *BillingServerOkTemporaryMock) ChangeMerchant(
	ctx context.Context,
	in *grpc.OnboardingRequest,
	opts ...client.CallOption,
) (*billing.Merchant, error) {
	m := &billing.Merchant{
		Name:            in.Name,
		AlternativeName: in.AlternativeName,
		Website:         in.Website,
		Country: &billing.Country{
			CodeInt:  643,
			CodeA3:   "RUS",
			CodeA2:   in.Country,
			IsActive: true,
		},
		State:              in.State,
		Zip:                in.Zip,
		City:               in.City,
		Address:            in.Address,
		AddressAdditional:  in.AddressAdditional,
		RegistrationNumber: in.RegistrationNumber,
		TaxId:              in.TaxId,
		Contacts:           in.Contacts,
		Banking: &billing.MerchantBanking{
			Currency: &billing.Currency{
				CodeInt:  643,
				CodeA3:   in.Banking.Currency,
				Name:     &billing.Name{Ru: "Российский рубль", En: "Russian ruble"},
				IsActive: true,
			},
			Name:          in.Banking.Name,
			Address:       in.Banking.Address,
			AccountNumber: in.Banking.AccountNumber,
			Swift:         in.Banking.Swift,
			Details:       in.Banking.Details,
		},
		Status: pkg.MerchantStatusDraft,
	}

	if in.Id != "" {
		m.Id = in.Id
	} else {
		m.Id = bson.NewObjectId().Hex()
	}

	return m, nil
}

func (s *BillingServerOkTemporaryMock) ChangeMerchantStatus(
	ctx context.Context,
	in *grpc.MerchantChangeStatusRequest,
	opts ...client.CallOption,
) (*billing.Merchant, error) {
	return &billing.Merchant{Id: in.MerchantId, Status: in.Status}, nil
}

func (s *BillingServerOkTemporaryMock) CreateNotification(
	ctx context.Context,
	in *grpc.NotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return &billing.Notification{}, nil
}

func (s *BillingServerOkTemporaryMock) GetNotification(
	ctx context.Context,
	in *grpc.GetNotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return &billing.Notification{}, nil
}

func (s *BillingServerOkTemporaryMock) ListNotifications(
	ctx context.Context,
	in *grpc.ListingNotificationRequest,
	opts ...client.CallOption,
) (*grpc.Notifications, error) {
	return &grpc.Notifications{}, nil
}

func (s *BillingServerOkTemporaryMock) MarkNotificationAsRead(
	ctx context.Context,
	in *grpc.GetNotificationRequest,
	opts ...client.CallOption,
) (*billing.Notification, error) {
	return &billing.Notification{}, nil
}

func (s *BillingServerOkTemporaryMock) ListMerchantPaymentMethods(
	ctx context.Context,
	in *grpc.ListMerchantPaymentMethodsRequest,
	opts ...client.CallOption,
) (*grpc.ListingMerchantPaymentMethod, error) {
	return &grpc.ListingMerchantPaymentMethod{}, nil
}

func (s *BillingServerOkTemporaryMock) GetMerchantPaymentMethod(
	ctx context.Context,
	in *grpc.GetMerchantPaymentMethodRequest,
	opts ...client.CallOption,
) (*billing.MerchantPaymentMethod, error) {
	return &billing.MerchantPaymentMethod{}, nil
}

func (s *BillingServerOkTemporaryMock) ChangeMerchantPaymentMethod(
	ctx context.Context,
	in *grpc.MerchantPaymentMethodRequest,
	opts ...client.CallOption,
) (*grpc.MerchantPaymentMethodResponse, error) {
	return &grpc.MerchantPaymentMethodResponse{
		Status: pkg.ResponseStatusOk,
		Item: &billing.MerchantPaymentMethod{
			PaymentMethod: &billing.MerchantPaymentMethodIdentification{
				Id:   in.PaymentMethod.Id,
				Name: in.PaymentMethod.Name,
			},
			Commission:  in.Commission,
			Integration: in.Integration,
			IsActive:    in.IsActive,
		},
	}, nil
}

func (s *BillingServerOkTemporaryMock) CreateRefund(
	ctx context.Context,
	in *grpc.CreateRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return &grpc.CreateRefundResponse{
		Status: pkg.ResponseStatusOk,
		Item:   &billing.Refund{},
	}, nil
}

func (s *BillingServerOkTemporaryMock) ListRefunds(
	ctx context.Context,
	in *grpc.ListRefundsRequest,
	opts ...client.CallOption,
) (*grpc.ListRefundsResponse, error) {
	return &grpc.ListRefundsResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) GetRefund(
	ctx context.Context,
	in *grpc.GetRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return &grpc.CreateRefundResponse{
		Status: pkg.ResponseStatusOk,
		Item:   &billing.Refund{},
	}, nil
}

func (s *BillingServerOkTemporaryMock) ProcessRefundCallback(
	ctx context.Context,
	in *grpc.CallbackRequest,
	opts ...client.CallOption,
) (*grpc.PaymentNotifyResponse, error) {
	return &grpc.PaymentNotifyResponse{
		Status: pkg.ResponseStatusOk,
		Error:  SomeError,
	}, nil
}

func (s *BillingServerOkTemporaryMock) ChangeCustomer(
	ctx context.Context,
	in *billing.Customer,
	opts ...client.CallOption,
) (*grpc.ChangeCustomerResponse, error) {
	return &grpc.ChangeCustomerResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) CheckProjectRequestSignature(
	ctx context.Context,
	in *grpc.CheckProjectRequestSignatureRequest,
	opts ...client.CallOption,
) (*grpc.CheckProjectRequestSignatureResponse, error) {
	return &grpc.CheckProjectRequestSignatureResponse{}, nil
}

func (s *BillingServerOkMock) CreateOrUpdateProduct(ctx context.Context, in *grpc.Product, opts ...client.CallOption) (*grpc.Product, error) {
	return Product, nil
}

func (s *BillingServerOkMock) ListProducts(ctx context.Context, in *grpc.ListProductsRequest, opts ...client.CallOption) (*grpc.ListProductsResponse, error) {
	return &grpc.ListProductsResponse{
		Limit:  1,
		Offset: 0,
		Total:  200,
		Products: []*grpc.Product{
			Product,
		},
	}, nil
}

func (s *BillingServerOkMock) GetProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.Product, error) {
	return Product, nil
}

func (s *BillingServerOkMock) DeleteProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) CreateOrUpdateProduct(ctx context.Context, in *grpc.Product, opts ...client.CallOption) (*grpc.Product, error) {
	return Product, nil
}

func (s *BillingServerOkTemporaryMock) ListProducts(ctx context.Context, in *grpc.ListProductsRequest, opts ...client.CallOption) (*grpc.ListProductsResponse, error) {
	return &grpc.ListProductsResponse{
		Limit:  1,
		Offset: 0,
		Total:  200,
		Products: []*grpc.Product{
			Product,
		},
	}, nil
}

func (s *BillingServerOkTemporaryMock) GetProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.Product, error) {
	return Product, nil
}

func (s *BillingServerOkTemporaryMock) DeleteProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	return &grpc.EmptyResponse{}, nil
}

func (s *BillingServerErrorMock) CreateOrUpdateProduct(ctx context.Context, in *grpc.Product, opts ...client.CallOption) (*grpc.Product, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) ListProducts(ctx context.Context, in *grpc.ListProductsRequest, opts ...client.CallOption) (*grpc.ListProductsResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) GetProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.Product, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerErrorMock) DeleteProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) CreateOrUpdateProduct(ctx context.Context, in *grpc.Product, opts ...client.CallOption) (*grpc.Product, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) ListProducts(ctx context.Context, in *grpc.ListProductsRequest, opts ...client.CallOption) (*grpc.ListProductsResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) GetProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.Product, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerSystemErrorMock) DeleteProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerOkTemporaryMock) PaymentFormLanguageChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangeLangRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerOkTemporaryMock) PaymentFormPaymentAccountChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangePaymentAccountRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerOkTemporaryMock) ProcessBillingAddress(
	ctx context.Context,
	in *grpc.ProcessBillingAddressRequest,
	opts ...client.CallOption,
) (*grpc.ProcessBillingAddressResponse, error) {
	return nil, errors.New(SomeError)
}

func (s *BillingServerOkTemporaryMock) ChangeMerchantData(
	ctx context.Context,
	in *grpc.ChangeMerchantDataRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	return &grpc.ChangeMerchantDataResponse{}, nil
}

func (s *BillingServerOkTemporaryMock) SetMerchantS3Agreement(
	ctx context.Context,
	in *grpc.SetMerchantS3AgreementRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	return &grpc.ChangeMerchantDataResponse{}, nil
}
