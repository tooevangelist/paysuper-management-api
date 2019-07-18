package api

type Dictionary struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

const (
	apiWebHookGroupPath     = "/webhook"
	apiAuthProjectGroupPath = "/api/v1"
	apiAuthUserGroupPath    = "/admin/api/v1"

	LimitDefault  = 100
	OffsetDefault = 0

	requestParameterId                       = "id"
	requestParameterName                     = "name"
	requestParameterSku                      = "sku"
	requestParameterIsSigned                 = "is_signed"
	requestParameterMerchantId               = "merchant_id"
	requestParameterProject                  = "project[]"
	requestParameterPaymentMethod            = "payment_method[]"
	requestParameterCountry                  = "country"
	requestParameterCountries                = "country[]"
	requestParameterStatuses                 = "status[]"
	requestParameterProjectId                = "project_id"
	requestParameterPaymentMethodId          = "method_id"
	requestParameterOrderId                  = "order_id"
	requestParameterRefundId                 = "refund_id"
	requestParameterNotificationId           = "notification_id"
	requestParameterUserId                   = "user"
	requestParameterLimit                    = "limit"
	requestParameterOffset                   = "offset"
	requestParameterFile                     = "file"
	requestParameterUtmSource                = "utm_source"
	requestParameterUtmMedium                = "utm_medium"
	requestParameterUtmCampaign              = "utm_campaign"
	requestParameterIsSystem                 = "is_system"
	requestParameterAgreementType            = "agreement_type"
	requestParameterHasMerchantSignature     = "has_merchant_signature"
	requestParameterHasPspSignature          = "has_psp_signature"
	requestParameterAgreementSentViaMail     = "agreement_sent_via_mail"
	requestParameterMailTrackingLink         = "mail_tracking_link"
	requestParameterImage                    = "image"
	requestParameterCallbackCurrency         = "callback_currency"
	requestParameterCallbackProtocol         = "callback_protocol"
	requestParameterCreateOrderAllowedUrls   = "create_order_allowed_urls"
	requestParameterAllowDynamicNotifyUrls   = "allow_dynamic_notify_urls"
	requestParameterAllowDynamicRedirectUrls = "allow_dynamic_redirect_urls"
	requestParameterLimitsCurrency           = "limits_currency"
	requestParameterMinPaymentAmount         = "min_payment_amount"
	requestParameterMaxPaymentAmount         = "max_payment_amount"
	requestParameterNotifyEmails             = "notify_emails"
	requestParameterIsProductsCheckout       = "is_products_checkout"
	requestParameterSecretKey                = "secret_key"
	requestParameterSignatureRequired        = "signature_required"
	requestParameterSendNotifyEmail          = "send_notify_email"
	requestParameterUrlCheckAccount          = "url_check_account"
	requestParameterUrlProcessPayment        = "url_process_payment"
	requestParameterUrlRedirectFail          = "url_redirect_fail"
	requestParameterUrlRedirectSuccess       = "url_redirect_success"
	requestParameterUrlChargebackPayment     = "url_chargeback_payment"
	requestParameterUrlCancelPayment         = "url_cancel_payment"
	requestParameterUrlFraudPayment          = "url_fraud_payment"
	requestParameterUrlRefundPayment         = "url_refund_payment"
	requestParameterStatus                   = "status"
	requestAuthorizationTokenRegex           = "Bearer ([A-z0-9_.-]{10,})"
	requestParameterZipUsa                   = "zip_usa"

	userProfileFieldNumberOfEmployees = "NumberOfEmployees"
	userProfileFieldAnnualIncome      = "AnnualIncome"
	userProfileFieldCompanyName       = "CompanyName"
	userProfileFieldPosition          = "Position"
	userProfileFieldFirstName         = "FirstName"
	userProfileFieldLastName          = "LastName"
	userProfileFieldWebsite           = "Website"
	userProfileFieldKindOfActivity    = "KindOfActivity"

	orderFieldProjectId     = "PP_PROJECT_ID"
	orderFieldSignature     = "PP_SIGNATURE"
	orderFieldAmount        = "PP_AMOUNT"
	orderFieldCurrency      = "PP_CURRENCY"
	orderFieldAccount       = "PP_ACCOUNT"
	orderFieldOrderId       = "PP_ORDER_ID"
	orderFieldPaymentMethod = "PP_PAYMENT_METHOD"
	orderFieldUrlVerify     = "PP_URL_VERIFY"
	orderFieldUrlNotify     = "PP_URL_NOTIFY"
	orderFieldUrlSuccess    = "PP_URL_SUCCESS"
	orderFieldUrlFail       = "PP_URL_FAIL"
	orderFieldPayerEmail    = "PP_PAYER_EMAIL"
	orderFieldPayerPhone    = "PP_PAYER_PHONE"
	orderFieldDescription   = "PP_DESCRIPTION"
	orderFieldRegion        = "PP_REGION"

	QueryParameterNameLimit  = "limit"
	QueryParameterNameOffset = "offset"
	QueryParameterNameSort   = "sort[]"

	errorMessageMask = "field validation for '%s' failed on the '%s' tag"

	HeaderAcceptLanguage      = "Accept-Language"
	HeaderUserAgent           = "User-Agent"
	HeaderXApiSignatureHeader = "X-API-SIGNATURE"
	HeaderReferer             = "referer"

	EnvironmentProduction        = "prod"
	CustomerTokenCookiesName     = "_ps_ctkn"
	CustomerTokenCookiesLifetime = 2592000

	CardPayPaymentResponseHeaderSignature = "Signature"

	agreementPageTemplateName = "agreement.html"

	UserProfilePositionCEO               = "CEO"
	UserProfilePositionCTO               = "CTO"
	UserProfilePositionCMO               = "CMO"
	UserProfilePositionCFO               = "CFO"
	UserProfilePositionProjectManagement = "Project Management"
	UserProfilePositionGenericManagement = "Generic Management"
	UserProfilePositionSoftwareDeveloper = "Software Developer"
	UserProfilePositionMarketing         = "Marketing"
	UserProfilePositionSupport           = "Support"
)

var (
	DefaultSort = []string{"_id"}

	OrderReservedWords = map[string]bool{
		orderFieldProjectId:     true,
		orderFieldSignature:     true,
		orderFieldAmount:        true,
		orderFieldCurrency:      true,
		orderFieldAccount:       true,
		orderFieldOrderId:       true,
		orderFieldDescription:   true,
		orderFieldPaymentMethod: true,
		orderFieldUrlVerify:     true,
		orderFieldUrlNotify:     true,
		orderFieldUrlSuccess:    true,
		orderFieldUrlFail:       true,
		orderFieldPayerEmail:    true,
		orderFieldPayerPhone:    true,
		orderFieldRegion:        true,
	}
)
