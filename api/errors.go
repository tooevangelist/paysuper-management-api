package api

import "github.com/paysuper/paysuper-billing-server/pkg/proto/grpc"

func newManagementApiResponseError(code, msg string, details ...string) *grpc.ResponseErrorMessage {
	var det string
	if len(details) > 0 && details[0] != "" {
		det = details[0]
	} else {
		det = ""
	}
	return &grpc.ResponseErrorMessage{Code: code, Message: msg, Details: det}
}

func newValidationError(details string) *grpc.ResponseErrorMessage {
	return newManagementApiResponseError(errorValidationFailed.Code, errorValidationFailed.Message, details)
}

const (
	errorNamespaceMerchantCompanyInfoName                 = "OnboardingRequest.Company.Name"
	errorNamespaceMerchantCompanyInfoAlternativeName      = "OnboardingRequest.Company.AlternativeName"
	errorNamespaceMerchantCompanyInfoWebsite              = "OnboardingRequest.Company.Website"
	errorNamespaceMerchantCompanyInfoCountry              = "OnboardingRequest.Company.Country"
	errorNamespaceMerchantCompanyInfoState                = "OnboardingRequest.Company.State"
	errorNamespaceMerchantCompanyInfoZip                  = "OnboardingRequest.Company.Zip"
	errorNamespaceMerchantCompanyInfoCity                 = "OnboardingRequest.Company.City"
	errorNamespaceMerchantCompanyInfoAddress              = "OnboardingRequest.Company.Address"
	errorNamespaceMerchantContactAuthorized               = "OnboardingRequest.Contacts.Authorized"
	errorNamespaceMerchantContactTechnical                = "OnboardingRequest.Contacts.Technical"
	errorNamespaceMerchantContactAuthorizedName           = "OnboardingRequest.Contacts.Authorized.Name"
	errorNamespaceMerchantContactAuthorizedEmail          = "OnboardingRequest.Contacts.Authorized.Email"
	errorNamespaceMerchantContactAuthorizedPhone          = "OnboardingRequest.Contacts.Authorized.Phone"
	errorNamespaceMerchantContactAuthorizedPosition       = "OnboardingRequest.Contacts.Authorized.Position"
	errorNamespaceMerchantContactTechnicalName            = "OnboardingRequest.Contacts.Technical.Name"
	errorNamespaceMerchantContactTechnicalEmail           = "OnboardingRequest.Contacts.Technical.Email"
	errorNamespaceMerchantContactTechnicalPhone           = "OnboardingRequest.Contacts.Technical.Phone"
	errorNamespaceMerchantBankingCurrency                 = "OnboardingRequest.Banking.Currency"
	errorNamespaceMerchantBankingName                     = "OnboardingRequest.Banking.Name"
	errorNamespaceMerchantBankingAddress                  = "OnboardingRequest.Banking.Address"
	errorNamespaceMerchantBankingAccountNumber            = "OnboardingRequest.Banking.AccountNumber"
	errorNamespaceMerchantBankingSwift                    = "OnboardingRequest.Banking.Swift"
	errorNamespaceMerchantBankingCorrespondentAccount     = "OnboardingRequest.Banking.CorrespondentAccount"
	errorNamespaceGetDashboardMainRequestPeriod           = "GetDashboardMainRequest.Period"
	errorNamespaceGetDashboardMainRequestMerchantId       = "GetDashboardMainRequest.MerchantId"
	errorNamespaceGetDashboardBaseReportRequestPeriod     = "GetDashboardBaseReportRequest.Period"
	errorNamespaceGetDashboardBaseReportRequestMerchantId = "GetDashboardBaseReportRequest.MerchantId"
)

var (
	errorUnknown                                      = newManagementApiResponseError("ma000001", "unknown error. try request later")
	errorValidationFailed                             = newManagementApiResponseError("ma000002", "validation failed")
	errorInternal                                     = newManagementApiResponseError("ma000003", "internal error")
	errorMessageAccessDenied                          = newManagementApiResponseError("ma000004", "access denied")
	errorIdIsEmpty                                    = newManagementApiResponseError("ma000005", "identifier can't be empty")
	errorIncorrectMerchantId                          = newManagementApiResponseError("ma000006", "incorrect merchant identifier")
	errorIncorrectNotificationId                      = newManagementApiResponseError("ma000007", "incorrect notification identifier")
	errorIncorrectOrderId                             = newManagementApiResponseError("ma000008", "incorrect order identifier")
	errorIncorrectProductId                           = newManagementApiResponseError("ma000009", "incorrect product identifier")
	errorIncorrectCountryIdentifier                   = newManagementApiResponseError("ma000010", "incorrect country identifier")
	errorIncorrectCurrencyIdentifier                  = newManagementApiResponseError("ma000011", "incorrect currency identifier")
	errorMessageOrdersNotFound                        = newManagementApiResponseError("ma000012", "orders not found")
	errorCountryNotFound                              = newManagementApiResponseError("ma000013", "country not found")
	errorCurrencyNotFound                             = newManagementApiResponseError("ma000014", "currency not found")
	errorNotificationNotFound                         = newManagementApiResponseError("ma000015", "notification not found")
	errorMessageAgreementCanNotBeGenerate             = newManagementApiResponseError("ma000020", "agreement can't be generated for not checked merchant data")
	errorMessageAgreementNotGenerated                 = newManagementApiResponseError("ma000021", "agreement for merchant not generated early")
	errorMessageSignatureHeaderIsEmpty                = newManagementApiResponseError("ma000022", "header with request signature can't be empty")
	errorRequestParamsIncorrect                       = newManagementApiResponseError("ma000023", "incorrect request parameters")
	errorEmailFieldIncorrect                          = newManagementApiResponseError("ma000024", "incorrect email")
	errorRequestDataInvalid                           = newManagementApiResponseError("ma000026", "request data invalid")
	errorCountriesListError                           = newManagementApiResponseError("ma000027", "countries list error")
	errorAgreementFileNotExist                        = newManagementApiResponseError("ma000028", "file for the specified key does not exist")
	errorNotMultipartForm                             = newManagementApiResponseError("ma000029", "no multipart boundary param in Content-Type")
	errorUploadFailed                                 = newManagementApiResponseError("ma000030", "upload failed")
	errorIncorrectProjectId                           = newManagementApiResponseError("ma000031", "incorrect project identifier")
	errorIncorrectPaymentMethodId                     = newManagementApiResponseError("ma000032", "incorrect payment method identifier")
	errorIncorrectPaylinkId                           = newManagementApiResponseError("ma000033", "incorrect paylink identifier")
	errorMessageAuthorizationHeaderNotFound           = newManagementApiResponseError("ma000034", "authorization header not found")
	errorMessageAuthorizationTokenNotFound            = newManagementApiResponseError("ma000035", "authorization token not found")
	errorMessageAuthorizedUserNotFound                = newManagementApiResponseError("ma000036", "information about authorized user not found")
	errorMessageStatusIncorrectType                   = newManagementApiResponseError("ma000037", "status parameter has incorrect type")
	errorMessageAgreementNotFound                     = newManagementApiResponseError("ma000038", "agreement for merchant not found")
	errorMessageAgreementUploadMaxSize                = newManagementApiResponseError("ma000039", "agreement document max upload size exceeded")
	errorMessageAgreementContentType                  = newManagementApiResponseError("ma000040", "agreement document type must be a pdf")
	errorMessageAgreementTypeIncorrectType            = newManagementApiResponseError("ma000041", "agreement type parameter have incorrect type")
	errorMessageHasMerchantSignatureIncorrectType     = newManagementApiResponseError("ma000042", "merchant signature parameter has incorrect type")
	errorMessageHasPspSignatureIncorrectType          = newManagementApiResponseError("ma000043", "paysuper signature parameter has incorrect type")
	errorMessageAgreementSentViaMailIncorrectType     = newManagementApiResponseError("ma000044", "agreement sent via email parameter has incorrect type")
	errorMessageMailTrackingLinkIncorrectType         = newManagementApiResponseError("ma000045", "mail tracking link parameter has incorrect type")
	errorMessageNameIncorrectType                     = newManagementApiResponseError("ma000046", "name parameter has incorrect type")
	errorMessageImageIncorrectType                    = newManagementApiResponseError("ma000047", "image parameter has incorrect type")
	errorMessageCallbackCurrencyIncorrectType         = newManagementApiResponseError("ma000048", "callback currency parameter has incorrect type")
	errorMessageCallbackProtocolIncorrectType         = newManagementApiResponseError("ma000049", "callback protocol parameter has incorrect type")
	errorMessageCreateOrderAllowedUrlsIncorrectType   = newManagementApiResponseError("ma000050", "create order allowed urls parameter has incorrect type")
	errorMessageAllowDynamicNotifyUrlsIncorrectType   = newManagementApiResponseError("ma000051", "allow dynamic notify urls parameter has incorrect type")
	errorMessageAllowDynamicRedirectUrlsIncorrectType = newManagementApiResponseError("ma000052", "allow dynamic redirect urls parameter has incorrect type")
	errorMessageLimitsCurrencyIncorrectType           = newManagementApiResponseError("ma000053", "limits currency parameter has incorrect type")
	errorMessageMinPaymentAmountIncorrectType         = newManagementApiResponseError("ma000054", "min payment amount parameter has incorrect type")
	errorMessageMaxPaymentAmountIncorrectType         = newManagementApiResponseError("ma000055", "max payment amount parameter has incorrect type")
	errorMessageNotifyEmailsIncorrectType             = newManagementApiResponseError("ma000056", "notify emails parameter has incorrect type")
	errorMessageIsProductsCheckoutIncorrectType       = newManagementApiResponseError("ma000057", "is products checkout parameter has incorrect type")
	errorMessageSecretKeyIncorrectType                = newManagementApiResponseError("ma000058", "secret key parameter has incorrect type")
	errorMessageSignatureRequiredIncorrectType        = newManagementApiResponseError("ma000059", "signature required parameter has incorrect type")
	errorMessageSendNotifyEmailIncorrectType          = newManagementApiResponseError("ma000060", "send notify email parameter has incorrect type")
	errorMessageUrlCheckAccountIncorrectType          = newManagementApiResponseError("ma000061", "url check account parameter has incorrect type")
	errorMessageUrlProcessPaymentIncorrectType        = newManagementApiResponseError("ma000062", "url process payment parameter has incorrect type")
	errorMessageUrlRedirectFailIncorrectType          = newManagementApiResponseError("ma000063", "url redirect fail parameter has incorrect type")
	errorMessageUrlRedirectSuccessIncorrectType       = newManagementApiResponseError("ma000064", "url redirect success parameter has incorrect type")
	errorMessageUrlChargebackPayment                  = newManagementApiResponseError("ma000065", "url chargeback payment parameter has incorrect type")
	errorMessageUrlCancelPayment                      = newManagementApiResponseError("ma000066", "url cancel payment parameter has incorrect type")
	errorMessageUrlFraudPayment                       = newManagementApiResponseError("ma000067", "url fraud payment parameter has incorrect type")
	errorMessageUrlRefundPayment                      = newManagementApiResponseError("ma000068", "url refund payment parameter has incorrect type")
	errorMessagePriceGroupByCountry                   = newManagementApiResponseError("ma000069", "unable to get price group by country")
	errorMessagePriceGroupCurrencyList                = newManagementApiResponseError("ma000070", "unable to get price group currencies")
	errorMessagePriceGroupCurrencyByRegion            = newManagementApiResponseError("ma000071", "unable to get price group currency by region")
	errorMessagePriceGroupRecommendedList             = newManagementApiResponseError("ma000072", "unable to get price group recommended prices")
	errorMessageGetProductPrice                       = newManagementApiResponseError("ma000072", "unable to get price of product")
	errorMessageUpdateProductPrice                    = newManagementApiResponseError("ma000072", "unable to update price of product")
	errorMessageIncorrectZip                          = newManagementApiResponseError("ma000073", "incorrect zip code")
	errorMessageIncorrectNumberOfEmployees            = newManagementApiResponseError("ma000074", "incorrect number of employees value")
	errorMessageIncorrectAnnualIncome                 = newManagementApiResponseError("ma000075", "incorrect annual income value")
	errorMessageIncorrectCompanyName                  = newManagementApiResponseError("ma000076", "incorrect company name")
	errorMessageIncorrectPosition                     = newManagementApiResponseError("ma000077", "incorrect position")
	errorMessageIncorrectFirstName                    = newManagementApiResponseError("ma000078", "incorrect first name")
	errorMessageIncorrectLastName                     = newManagementApiResponseError("ma000079", "incorrect last name")
	errorMessageIncorrectWebsite                      = newManagementApiResponseError("ma000080", "incorrect website")
	errorMessageIncorrectKindOfActivity               = newManagementApiResponseError("ma000081", "incorrect kind of activity")
	errorMessageIncorrectReview                       = newManagementApiResponseError("ma000082", "review must be text with length lower than or equal 500 characters")
	errorMessageIncorrectPageId                       = newManagementApiResponseError("ma000083", "review page identifier must be one of next values: primary_onboarding, merchant_onboarding")
	ErrorMessageKeyProductIdInvalid                   = newManagementApiResponseError("ma000082", "key product id is invalid")
	ErrorMessagePlatformIdInvalid                     = newManagementApiResponseError("ma000083", "platform id is invalid")

	errorMessageIncorrectAlternativeName          = newManagementApiResponseError("ma000084", "incorrect brand")
	errorMessageIncorrectState                    = newManagementApiResponseError("ma000085", "incorrect state")
	errorMessageIncorrectCity                     = newManagementApiResponseError("ma000086", "incorrect city")
	errorMessageIncorrectAddress                  = newManagementApiResponseError("ma000087", "incorrect address")
	errorMessageRequiredContactAuthorized         = newManagementApiResponseError("ma000088", "company authorized contact information is required")
	errorMessageRequiredContactTechnical          = newManagementApiResponseError("ma000089", "company technical contact information is required")
	errorMessageIncorrectName                     = newManagementApiResponseError("ma000090", "incorrect name")
	errorMessageIncorrectPhone                    = newManagementApiResponseError("ma000091", "incorrect phone")
	errorMessageIncorrectBankName                 = newManagementApiResponseError("ma000092", "incorrect bank name")
	errorMessageIncorrectBankAddress              = newManagementApiResponseError("ma000093", "incorrect bank address")
	errorMessageIncorrectBankAccountNumber        = newManagementApiResponseError("ma000094", "incorrect bank accounting number")
	errorMessageIncorrectBankSwift                = newManagementApiResponseError("ma000095", "incorrect bank swift code")
	errorMessageIncorrectBankCorrespondentAccount = newManagementApiResponseError("ma000096", "incorrect bank correspondent account")
	errorMessageFileNotFound                      = newManagementApiResponseError("ma000097", "file with key was not specified")
	errorMessageCantReadFile                      = newManagementApiResponseError("ma000098", "file can not be read")
	errorIncorrectPeriod                          = newManagementApiResponseError("ma000099", "incorrect period")
	errorMessageMerchantNotFound                  = newManagementApiResponseError("ma000100", "merchant not found")
	errorMessageCreateReportFile                  = newManagementApiResponseError("ma000101", "unable to create report file")
	errorMessageDownloadReportFile                = newManagementApiResponseError("ma000102", "unable to download report file")

	validationErrors = map[string]*grpc.ResponseErrorMessage{
		userProfileFieldNumberOfEmployees: errorMessageIncorrectNumberOfEmployees,
		userProfileFieldAnnualIncome:      errorMessageIncorrectAnnualIncome,
		userProfileFieldCompanyName:       errorMessageIncorrectCompanyName,
		userProfileFieldPosition:          errorMessageIncorrectPosition,
		userProfileFieldFirstName:         errorMessageIncorrectFirstName,
		userProfileFieldLastName:          errorMessageIncorrectLastName,
		userProfileFieldWebsite:           errorMessageIncorrectWebsite,
		userProfileFieldKindOfActivity:    errorMessageIncorrectKindOfActivity,
		userProfileFieldReview:            errorMessageIncorrectReview,
		userProfileFieldPageId:            errorMessageIncorrectPageId,
	}

	validationNamespaceErrors = map[string]*grpc.ResponseErrorMessage{
		errorNamespaceMerchantCompanyInfoName:                 errorMessageIncorrectCompanyName,
		errorNamespaceMerchantCompanyInfoAlternativeName:      errorMessageIncorrectAlternativeName,
		errorNamespaceMerchantCompanyInfoWebsite:              errorMessageIncorrectWebsite,
		errorNamespaceMerchantCompanyInfoCountry:              errorIncorrectCountryIdentifier,
		errorNamespaceMerchantCompanyInfoState:                errorMessageIncorrectState,
		errorNamespaceMerchantCompanyInfoZip:                  errorMessageIncorrectZip,
		errorNamespaceMerchantCompanyInfoCity:                 errorMessageIncorrectCity,
		errorNamespaceMerchantCompanyInfoAddress:              errorMessageIncorrectAddress,
		errorNamespaceMerchantContactAuthorized:               errorMessageRequiredContactAuthorized,
		errorNamespaceMerchantContactTechnical:                errorMessageRequiredContactTechnical,
		errorNamespaceMerchantContactAuthorizedName:           errorMessageIncorrectName,
		errorNamespaceMerchantContactAuthorizedEmail:          errorEmailFieldIncorrect,
		errorNamespaceMerchantContactAuthorizedPhone:          errorMessageIncorrectPhone,
		errorNamespaceMerchantContactAuthorizedPosition:       errorMessageIncorrectPosition,
		errorNamespaceMerchantContactTechnicalName:            errorMessageIncorrectName,
		errorNamespaceMerchantContactTechnicalEmail:           errorEmailFieldIncorrect,
		errorNamespaceMerchantContactTechnicalPhone:           errorMessageIncorrectPhone,
		errorNamespaceMerchantBankingCurrency:                 errorIncorrectCurrencyIdentifier,
		errorNamespaceMerchantBankingName:                     errorMessageIncorrectBankName,
		errorNamespaceMerchantBankingAddress:                  errorMessageIncorrectBankAddress,
		errorNamespaceMerchantBankingAccountNumber:            errorMessageIncorrectBankAccountNumber,
		errorNamespaceMerchantBankingSwift:                    errorMessageIncorrectBankSwift,
		errorNamespaceMerchantBankingCorrespondentAccount:     errorMessageIncorrectBankCorrespondentAccount,
		errorNamespaceGetDashboardMainRequestMerchantId:       errorIncorrectMerchantId,
		errorNamespaceGetDashboardMainRequestPeriod:           errorIncorrectPeriod,
		errorNamespaceGetDashboardBaseReportRequestPeriod:     errorIncorrectPeriod,
		errorNamespaceGetDashboardBaseReportRequestMerchantId: errorIncorrectMerchantId,
	}
)
