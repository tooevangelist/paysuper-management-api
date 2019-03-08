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
	SomeError = "some error"
)

var (
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
		Status:                    pkg.MerchantStatusDraft,
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
)

type BillingServerOkMock struct{}
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

func (s *BillingServerOkMock) OrderCreateProcess(
	ctx context.Context,
	in *billing.OrderCreateRequest,
	opts ...client.CallOption,
) (*billing.Order, error) {
	return &billing.Order{}, nil
}

func (s *BillingServerOkMock) PaymentFormJsonDataProcess(
	ctx context.Context,
	in *grpc.PaymentFormJsonDataRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormJsonDataResponse, error) {
	return &grpc.PaymentFormJsonDataResponse{}, nil
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

func (s *BillingServerOkMock) GetMerchantById(
	ctx context.Context,
	in *grpc.FindByIdRequest,
	opts ...client.CallOption,
) (*grpc.MerchantGetMerchantResponse, error) {
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
) (*grpc.Merchants, error) {
	return &grpc.Merchants{
		Merchants: []*billing.Merchant{OnboardingMerchantMock, OnboardingMerchantMock, OnboardingMerchantMock},
	}, nil
}

func (s *BillingServerOkMock) ChangeMerchant(
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
	in *grpc.FindByIdRequest,
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
	in *grpc.FindByIdRequest,
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

func (s *BillingServerErrorMock) GetMerchantById(
	ctx context.Context,
	in *grpc.FindByIdRequest,
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
) (*grpc.Merchants, error) {
	return &grpc.Merchants{}, nil
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
	in *grpc.FindByIdRequest,
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
	in *grpc.FindByIdRequest,
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
	return &grpc.PaymentFormJsonDataResponse{}, nil
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

func (s *BillingServerSystemErrorMock) GetMerchantById(
	ctx context.Context,
	in *grpc.FindByIdRequest,
	opts ...client.CallOption,
) (*grpc.MerchantGetMerchantResponse, error) {
	return nil, errors.New("some error")
}

func (s *BillingServerSystemErrorMock) ListMerchants(
	ctx context.Context,
	in *grpc.MerchantListingRequest,
	opts ...client.CallOption,
) (*grpc.Merchants, error) {
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
	in *grpc.FindByIdRequest,
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
	in *grpc.FindByIdRequest,
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
