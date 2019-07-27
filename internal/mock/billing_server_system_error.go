package mock

import (
	"context"
	"errors"
	"github.com/globalsign/mgo/bson"
	"github.com/micro/go-micro/client"
	"github.com/paysuper/paysuper-billing-server/pkg"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/billing"
	"github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"
	"net/http"
)

type BillingServerSystemErrorMock struct{}

func NewBillingServerSystemErrorMock() grpc.BillingService {
	return &BillingServerSystemErrorMock{}
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
) (*grpc.OrderCreateProcessResponse, error) {
	return &grpc.OrderCreateProcessResponse{
		Status:  http.StatusBadRequest,
		Message: SomeError,
	}, nil
}

func (s *BillingServerSystemErrorMock) PaymentFormJsonDataProcess(
	ctx context.Context,
	in *grpc.PaymentFormJsonDataRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormJsonDataResponse, error) {
	return nil, SomeError
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

func (s *BillingServerSystemErrorMock) GetMerchantBy(
	ctx context.Context,
	in *grpc.GetMerchantByRequest,
	opts ...client.CallOption,
) (*grpc.GetMerchantResponse, error) {
	return nil, errors.New("some error")
}

func (s *BillingServerSystemErrorMock) ListMerchants(
	ctx context.Context,
	in *grpc.MerchantListingRequest,
	opts ...client.CallOption,
) (*grpc.MerchantListingResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) ChangeMerchant(
	ctx context.Context,
	in *grpc.OnboardingRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantResponse, error) {
	return &grpc.ChangeMerchantResponse{
		Status: pkg.ResponseStatusOk,
		Item:   &billing.Merchant{},
	}, nil
}

func (s *BillingServerSystemErrorMock) ChangeMerchantStatus(
	ctx context.Context,
	in *grpc.MerchantChangeStatusRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantStatusResponse, error) {
	return &grpc.ChangeMerchantStatusResponse{
		Status: pkg.ResponseStatusOk,
		Item:   &billing.Merchant{},
	}, nil
}

func (s *BillingServerSystemErrorMock) CreateNotification(
	ctx context.Context,
	in *grpc.NotificationRequest,
	opts ...client.CallOption,
) (*grpc.CreateNotificationResponse, error) {
	return &grpc.CreateNotificationResponse{
		Status:  http.StatusBadRequest,
		Message: SomeError,
	}, nil
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
) (*grpc.GetMerchantPaymentMethodResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) ChangeMerchantPaymentMethod(
	ctx context.Context,
	in *grpc.MerchantPaymentMethodRequest,
	opts ...client.CallOption,
) (*grpc.MerchantPaymentMethodResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) CreateRefund(
	ctx context.Context,
	in *grpc.CreateRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) ListRefunds(
	ctx context.Context,
	in *grpc.ListRefundsRequest,
	opts ...client.CallOption,
) (*grpc.ListRefundsResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetRefund(
	ctx context.Context,
	in *grpc.GetRefundRequest,
	opts ...client.CallOption,
) (*grpc.CreateRefundResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) ProcessRefundCallback(
	ctx context.Context,
	in *grpc.CallbackRequest,
	opts ...client.CallOption,
) (*grpc.PaymentNotifyResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) PaymentFormLanguageChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangeLangRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) PaymentFormPaymentAccountChanged(
	ctx context.Context,
	in *grpc.PaymentFormUserChangePaymentAccountRequest,
	opts ...client.CallOption,
) (*grpc.PaymentFormDataChangeResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) ProcessBillingAddress(
	ctx context.Context,
	in *grpc.ProcessBillingAddressRequest,
	opts ...client.CallOption,
) (*grpc.ProcessBillingAddressResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) ChangeMerchantData(
	ctx context.Context,
	in *grpc.ChangeMerchantDataRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) SetMerchantS3Agreement(
	ctx context.Context,
	in *grpc.SetMerchantS3AgreementRequest,
	opts ...client.CallOption,
) (*grpc.ChangeMerchantDataResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) ChangeProject(
	ctx context.Context,
	in *billing.Project,
	opts ...client.CallOption,
) (*grpc.ChangeProjectResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetProject(
	ctx context.Context,
	in *grpc.GetProjectRequest,
	opts ...client.CallOption,
) (*grpc.ChangeProjectResponse, error) {
	if in.ProjectId == SomeMerchantId {
		return &grpc.ChangeProjectResponse{
			Status: pkg.ResponseStatusOk,
			Item: &billing.Project{
				MerchantId:         bson.NewObjectId().Hex(),
				Name:               map[string]string{"en": "A", "ru": "А"},
				CallbackCurrency:   "RUB",
				CallbackProtocol:   pkg.ProjectCallbackProtocolEmpty,
				LimitsCurrency:     "RUB",
				MinPaymentAmount:   0,
				MaxPaymentAmount:   15000,
				IsProductsCheckout: false,
			},
		}, nil
	}

	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) DeleteProject(
	ctx context.Context,
	in *grpc.GetProjectRequest,
	opts ...client.CallOption,
) (*grpc.ChangeProjectResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) CreateToken(
	ctx context.Context,
	in *grpc.TokenRequest,
	opts ...client.CallOption,
) (*grpc.TokenResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) CheckProjectRequestSignature(
	ctx context.Context,
	in *grpc.CheckProjectRequestSignatureRequest,
	opts ...client.CallOption,
) (*grpc.CheckProjectRequestSignatureResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) CreateOrUpdateProduct(ctx context.Context, in *grpc.Product, opts ...client.CallOption) (*grpc.Product, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) ListProducts(ctx context.Context, in *grpc.ListProductsRequest, opts ...client.CallOption) (*grpc.ListProductsResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.Product, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) DeleteProduct(ctx context.Context, in *grpc.RequestProduct, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) FindAllOrders(ctx context.Context, in *grpc.ListOrdersRequest, opts ...client.CallOption) (*billing.OrderPaginate, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) IsOrderCanBePaying(
	ctx context.Context,
	in *grpc.IsOrderCanBePayingRequest,
	opts ...client.CallOption,
) (*grpc.IsOrderCanBePayingResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetCountry(ctx context.Context, in *billing.GetCountryRequest, opts ...client.CallOption) (*billing.Country, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) UpdateCountry(ctx context.Context, in *billing.Country, opts ...client.CallOption) (*billing.Country, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetPriceGroup(ctx context.Context, in *billing.GetPriceGroupRequest, opts ...client.CallOption) (*billing.PriceGroup, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) UpdatePriceGroup(ctx context.Context, in *billing.PriceGroup, opts ...client.CallOption) (*billing.PriceGroup, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) SetUserNotifySales(ctx context.Context, in *grpc.SetUserNotifyRequest, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) SetUserNotifyNewRegion(ctx context.Context, in *grpc.SetUserNotifyRequest, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetCountriesList(ctx context.Context, in *grpc.EmptyRequest, opts ...client.CallOption) (*billing.CountriesList, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetPaymentChannelCostSystem(ctx context.Context, in *billing.PaymentChannelCostSystemRequest, opts ...client.CallOption) (*grpc.PaymentChannelCostSystemResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) SetPaymentChannelCostSystem(ctx context.Context, in *billing.PaymentChannelCostSystem, opts ...client.CallOption) (*grpc.PaymentChannelCostSystemResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) DeletePaymentChannelCostSystem(ctx context.Context, in *billing.PaymentCostDeleteRequest, opts ...client.CallOption) (*grpc.ResponseError, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetPaymentChannelCostMerchant(ctx context.Context, in *billing.PaymentChannelCostMerchantRequest, opts ...client.CallOption) (*grpc.PaymentChannelCostMerchantResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) SetPaymentChannelCostMerchant(ctx context.Context, in *billing.PaymentChannelCostMerchant, opts ...client.CallOption) (*grpc.PaymentChannelCostMerchantResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) DeletePaymentChannelCostMerchant(ctx context.Context, in *billing.PaymentCostDeleteRequest, opts ...client.CallOption) (*grpc.ResponseError, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetMoneyBackCostSystem(ctx context.Context, in *billing.MoneyBackCostSystemRequest, opts ...client.CallOption) (*grpc.MoneyBackCostSystemResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) SetMoneyBackCostSystem(ctx context.Context, in *billing.MoneyBackCostSystem, opts ...client.CallOption) (*grpc.MoneyBackCostSystemResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) DeleteMoneyBackCostSystem(ctx context.Context, in *billing.PaymentCostDeleteRequest, opts ...client.CallOption) (*grpc.ResponseError, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetMoneyBackCostMerchant(ctx context.Context, in *billing.MoneyBackCostMerchantRequest, opts ...client.CallOption) (*grpc.MoneyBackCostMerchantResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) SetMoneyBackCostMerchant(ctx context.Context, in *billing.MoneyBackCostMerchant, opts ...client.CallOption) (*grpc.MoneyBackCostMerchantResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) DeleteMoneyBackCostMerchant(ctx context.Context, in *billing.PaymentCostDeleteRequest, opts ...client.CallOption) (*grpc.ResponseError, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetAllPaymentChannelCostSystem(ctx context.Context, in *grpc.EmptyRequest, opts ...client.CallOption) (*grpc.PaymentChannelCostSystemListResponse, error) {
	return nil, errors.New("Some error")
}

func (s *BillingServerSystemErrorMock) GetAllPaymentChannelCostMerchant(ctx context.Context, in *billing.PaymentChannelCostMerchantListRequest, opts ...client.CallOption) (*grpc.PaymentChannelCostMerchantListResponse, error) {
	return nil, errors.New("Some error")
}

func (s *BillingServerSystemErrorMock) GetAllMoneyBackCostSystem(ctx context.Context, in *grpc.EmptyRequest, opts ...client.CallOption) (*grpc.MoneyBackCostSystemListResponse, error) {
	return nil, errors.New("Some error")
}

func (s *BillingServerSystemErrorMock) GetAllMoneyBackCostMerchant(ctx context.Context, in *billing.MoneyBackCostMerchantListRequest, opts ...client.CallOption) (*grpc.MoneyBackCostMerchantListResponse, error) {
	return nil, errors.New("Some error")
}

func (s *BillingServerSystemErrorMock) CreateOrUpdatePaymentMethodTestSettings(ctx context.Context, in *grpc.ChangePaymentMethodParamsRequest, opts ...client.CallOption) (*grpc.ChangePaymentMethodParamsResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) DeletePaymentMethodTestSettings(ctx context.Context, in *grpc.GetPaymentMethodSettingsRequest, opts ...client.CallOption) (*grpc.ChangePaymentMethodParamsResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) FindByZipCode(
	ctx context.Context,
	in *grpc.FindByZipCodeRequest,
	opts ...client.CallOption,
) (*grpc.FindByZipCodeResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) CreateOrUpdatePaymentMethod(
	ctx context.Context,
	in *billing.PaymentMethod,
	opts ...client.CallOption,
) (*grpc.ChangePaymentMethodResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) CreateOrUpdatePaymentMethodProductionSettings(
	ctx context.Context,
	in *grpc.ChangePaymentMethodParamsRequest,
	opts ...client.CallOption,
) (*grpc.ChangePaymentMethodParamsResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) DeletePaymentMethodProductionSettings(
	ctx context.Context,
	in *grpc.GetPaymentMethodSettingsRequest,
	opts ...client.CallOption,
) (*grpc.ChangePaymentMethodParamsResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) CreateAccountingEntry(ctx context.Context, in *grpc.CreateAccountingEntryRequest, opts ...client.CallOption) (*grpc.CreateAccountingEntryResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) CreateRoyaltyReport(ctx context.Context, in *grpc.CreateRoyaltyReportRequest, opts ...client.CallOption) (*grpc.CreateRoyaltyReportRequest, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) ListRoyaltyReports(ctx context.Context, in *grpc.ListRoyaltyReportsRequest, opts ...client.CallOption) (*grpc.ListRoyaltyReportsResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) ChangeRoyaltyReportStatus(ctx context.Context, in *grpc.CreateRoyaltyReportRequest, opts ...client.CallOption) (*grpc.CreateRoyaltyReportRequest, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) ListRoyaltyReportOrders(ctx context.Context, in *grpc.ListRoyaltyReportOrdersRequest, opts ...client.CallOption) (*grpc.TransactionsResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetVatReportsDashboard(ctx context.Context, in *grpc.EmptyRequest, opts ...client.CallOption) (*grpc.VatReportsResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetVatReportsForCountry(ctx context.Context, in *grpc.VatReportsRequest, opts ...client.CallOption) (*grpc.VatReportsResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetVatReportTransactions(ctx context.Context, in *grpc.VatTransactionsRequest, opts ...client.CallOption) (*grpc.TransactionsResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) CalcAnnualTurnovers(ctx context.Context, in *grpc.EmptyRequest, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) ListProjects(ctx context.Context, in *grpc.ListProjectsRequest, opts ...client.CallOption) (*grpc.ListProjectsResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetOrder(ctx context.Context, in *grpc.GetOrderRequest, opts ...client.CallOption) (*billing.Order, error) {
	return nil, SomeError
}
func (s *BillingServerSystemErrorMock) ProcessVatReports(ctx context.Context, in *grpc.ProcessVatReportsRequest, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) UpdateVatReportStatus(ctx context.Context, in *grpc.UpdateVatReportStatusRequest, opts ...client.CallOption) (*grpc.ResponseError, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) UpdateProductPrices(
	ctx context.Context,
	in *grpc.UpdateProductPricesRequest,
	opts ...client.CallOption,
) (*grpc.ResponseError, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetProductPrices(
	ctx context.Context,
	in *grpc.RequestProduct,
	opts ...client.CallOption,
) (*grpc.ProductPricesResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetPriceGroupRecommendedPrice(
	ctx context.Context,
	in *grpc.PriceGroupRecommendedPriceRequest,
	opts ...client.CallOption,
) (*grpc.PriceGroupRecommendedPriceResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetPriceGroupCurrencyByRegion(
	ctx context.Context,
	in *grpc.PriceGroupByRegionRequest,
	opts ...client.CallOption,
) (*grpc.PriceGroupCurrenciesResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetPriceGroupCurrencies(
	ctx context.Context,
	in *grpc.EmptyRequest,
	opts ...client.CallOption,
) (*grpc.PriceGroupCurrenciesResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetPriceGroupByCountry(
	ctx context.Context,
	in *grpc.PriceGroupByCountryRequest,
	opts ...client.CallOption,
) (*billing.PriceGroup, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetPaymentMethodProductionSettings(ctx context.Context, in *grpc.GetPaymentMethodSettingsRequest, opts ...client.CallOption) (*grpc.GetPaymentMethodSettingsResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) GetPaymentMethodTestSettings(ctx context.Context, in *grpc.GetPaymentMethodSettingsRequest, opts ...client.CallOption) (*grpc.GetPaymentMethodSettingsResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) ChangeRoyaltyReport(ctx context.Context, in *grpc.ChangeRoyaltyReportRequest, opts ...client.CallOption) (*grpc.ResponseError, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) AutoAcceptRoyaltyReports(ctx context.Context, in *grpc.EmptyRequest, opts ...client.CallOption) (*grpc.EmptyResponse, error) {
	panic("implement me")
}

func (s *BillingServerSystemErrorMock) GetUserProfile(
	ctx context.Context,
	in *grpc.GetUserProfileRequest,
	opts ...client.CallOption,
) (*grpc.GetUserProfileResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) CreateOrUpdateUserProfile(
	ctx context.Context,
	in *grpc.UserProfile,
	opts ...client.CallOption,
) (*grpc.GetUserProfileResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) ConfirmUserEmail(
	ctx context.Context,
	in *grpc.ConfirmUserEmailRequest,
	opts ...client.CallOption,
) (*grpc.CheckProjectRequestSignatureResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) CreatePageReview(
	ctx context.Context,
	in *grpc.CreatePageReviewRequest,
	opts ...client.CallOption,
) (*grpc.CheckProjectRequestSignatureResponse, error) {
	return nil, SomeError
}

func (s *BillingServerSystemErrorMock) MerchantAcceptRoyaltyReport(ctx context.Context, in *grpc.MerchantAcceptRoyaltyReportRequest, opts ...client.CallOption) (*grpc.ResponseError, error) {
	panic("implement me")
}
