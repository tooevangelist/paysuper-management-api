package handlers

import (
	"github.com/ProtocolONE/go-core/v2/pkg/config"
	"github.com/ProtocolONE/go-core/v2/pkg/provider"
	awsWrapper "github.com/paysuper/paysuper-aws-manager"
	"github.com/paysuper/paysuper-management-api/internal/dispatcher/common"
	"gopkg.in/go-playground/validator.v9"
)

func ProviderHandlers(initial config.Initial, srv common.Services, validator *validator.Validate, set provider.AwareSet, cfg *common.Config) (common.Handlers, func(), error) {
	hSet := common.HandlerSet{
		Services: srv,
		Validate: validator,
		AwareSet: set,
	}
	copyCfg := *cfg

	// Agreement S3 AWS Client
	awsOptions := []awsWrapper.Option{
		awsWrapper.AccessKeyId(cfg.AwsAccessKeyIdAgreement),
		awsWrapper.SecretAccessKey(cfg.AwsSecretAccessKeyAgreement),
		awsWrapper.Region(cfg.AwsRegionAgreement),
		awsWrapper.Bucket(cfg.AwsBucketAgreement),
	}
	awsManagerAgreement, err := awsWrapper.New(awsOptions...)
	if err != nil {
		return nil, func() {}, err
	}

	// Reporter S3 AWS Client
	awsOptions = []awsWrapper.Option{
		awsWrapper.AccessKeyId(cfg.AwsAccessKeyIdReporter),
		awsWrapper.SecretAccessKey(cfg.AwsSecretAccessKeyReporter),
		awsWrapper.Region(cfg.AwsRegionReporter),
		awsWrapper.Bucket(cfg.AwsBucketReporter),
	}
	awsManagerReporter, err := awsWrapper.New(awsOptions...)
	if err != nil {
		return nil, func() {}, err
	}

	return []common.Handler{
		NewCardPayWebHook(hSet, &copyCfg),
		NewCountryApiV1(hSet, &copyCfg),
		NewDashboardRoute(hSet, &copyCfg),
		NewKeyRoute(hSet, &copyCfg),
		NewKeyProductRoute(hSet, &copyCfg),
		NewOnboardingRoute(hSet, initial, awsManagerAgreement, &copyCfg),
		NewOrderRoute(hSet, &copyCfg),
		NewPayLinkRoute(hSet, &copyCfg),
		NewPaymentCostRoute(hSet, &copyCfg),
		NewPaymentMethodApiV1(hSet, &copyCfg),
		NewPriceGroupRoute(hSet, &copyCfg),
		NewProductRoute(hSet, &copyCfg),
		NewProjectRoute(hSet, &copyCfg),
		NewReportFileRoute(hSet, awsManagerReporter, &copyCfg),
		NewRoyaltyReportsRoute(hSet, &copyCfg),
		NewTaxesRoute(hSet, &copyCfg),
		NewTokenRoute(hSet, &copyCfg),
		NewUserProfileRoute(hSet, &copyCfg),
		NewVatReportsRoute(hSet, &copyCfg),
		NewZipCodeRoute(hSet, &copyCfg),
		NewBalanceRoute(hSet, &copyCfg),
		NewPayoutDocumentsRoute(hSet, &copyCfg),
		NewPricingRoute(hSet, &copyCfg),
		NewRecurringRoute(hSet, &copyCfg),
		NewOperatingCompanyRoute(hSet, &copyCfg),
		NewPaymentMinLimitSystemRoute(hSet, &copyCfg),
		NewAdminUsersRoute(hSet, &copyCfg),
		NewMerchantUsersRoute(hSet, &copyCfg),
		NewUserRoute(hSet, &copyCfg),
	}, func() {}, nil
}
