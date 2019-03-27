package mock

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"github.com/micro/go-micro/client"
	"github.com/paysuper/paysuper-payment-link/proto"
)

var (
	pl = &paylink.Paylink{
		Id:         "21784001599a47e5a69ac28f7af2ec22",
		ProjectId:  "5c10ff51d5be4b0001bca600",
		MerchantId: "5c8f6a914dad6a0001839408",
		CreatedAt:  ptypes.TimestampNow(),
		UpdatedAt:  ptypes.TimestampNow(),
		ExpiresAt:  ptypes.TimestampNow(),
		Products: []string{
			"5c3c962781258d0001e65930",
			"5c9b68df68add437582ad84b",
		},
	}

	ps = &paylink.GetPaylinksResponse{
		ProjectId:  "5c10ff51d5be4b0001bca600",
		MerchantId: "5c8f6a914dad6a0001839408",
		Limit:      2,
		Offset:     0,
		Total:      100,
		Paylinks: []string{
			"21784001599a47e5a69ac28f7af2ec22",
			"10159921784a47e8f7af2ec225a69ac2",
		},
	}

	stat = &paylink.GetPaylinkStatResponse{
		Visits: 10,
	}

	url = &paylink.GetPaylinkUrlResponse{
		Url: "http://paysuprt.online/paylink/21784001599a47e5a69ac28f7af2ec22?utm_campagin=campagin1%26campagin2&utm_medium=123&utm_source=myUtm+Source",
	}
)

type PaymentLinkOkMock struct{}

func NewPaymentLinkOkMock() paylink.PaylinkService {
	return &PaymentLinkOkMock{}
}

func (PaymentLinkOkMock) GetPaylinks(ctx context.Context, in *paylink.GetPaylinksRequest, opts ...client.CallOption) (*paylink.GetPaylinksResponse, error) {
	return ps, nil
}

func (PaymentLinkOkMock) GetPaylink(ctx context.Context, in *paylink.PaylinkRequest, opts ...client.CallOption) (*paylink.Paylink, error) {
	return pl, nil
}

func (PaymentLinkOkMock) GetPaylinkStat(ctx context.Context, in *paylink.PaylinkRequest, opts ...client.CallOption) (*paylink.GetPaylinkStatResponse, error) {
	return stat, nil
}

func (PaymentLinkOkMock) IncrPaylinkVisits(ctx context.Context, in *paylink.PaylinkRequest, opts ...client.CallOption) (*paylink.EmptyResponse, error) {
	return &paylink.EmptyResponse{}, nil
}

func (PaymentLinkOkMock) GetPaylinkURL(ctx context.Context, in *paylink.GetPaylinkURLRequest, opts ...client.CallOption) (*paylink.GetPaylinkUrlResponse, error) {
	return url, nil
}

func (PaymentLinkOkMock) CreateOrUpdatePaylink(ctx context.Context, in *paylink.CreatePaylinkRequest, opts ...client.CallOption) (*paylink.Paylink, error) {
	return pl, nil
}

func (PaymentLinkOkMock) DeletePaylink(ctx context.Context, in *paylink.PaylinkRequest, opts ...client.CallOption) (*paylink.EmptyResponse, error) {
	return &paylink.EmptyResponse{}, nil
}
