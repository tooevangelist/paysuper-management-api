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
	requestParameterToken                    = "token"

	userProfileFieldNumberOfEmployees = "NumberOfEmployees"
	userProfileFieldAnnualIncome      = "AnnualIncome"
	userProfileFieldCompanyName       = "CompanyName"
	userProfileFieldPosition          = "Position"
	userProfileFieldFirstName         = "FirstName"
	userProfileFieldLastName          = "LastName"
	userProfileFieldWebsite           = "Website"
	userProfileFieldKindOfActivity    = "KindOfActivity"
	userProfileFieldReview            = "Review"
	userProfileFieldPageId            = "PageId"

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

	zipRegexp = map[string]string{
		"AF": "^\\d{4}$",
		"AX": "^\\d{5}$",
		"AL": "^\\d{4}$",
		"DZ": "^\\d{5}$",
		"AS": "^\\d{5}(-{1}\\d{4,6})$",
		"AD": "^[Aa][Dd]\\d{3}$",
		"AI": "^[Aa][I][-][2][6][4][0]$",
		"AR": "^\\d{4}|[A-Za-z]\\d{4}[a-zA-Z]{3}$",
		"AM": "^\\d{4}$",
		"AC": "^[Aa][Ss][Cc][Nn]\\s{0,1}[1][Zz][Zz]$",
		"AU": "^\\d{4}$",
		"AT": "^\\d{4}$",
		"AZ": "^[Aa][Zz]\\d{4}$",
		"BH": "^\\d{3,4}$",
		"BD": "^\\d{4}$",
		"BB": "^[Aa][Zz]\\d{5}$",
		"BY": "^\\d{6}$",
		"BE": "^\\d{4}$",
		"BM": "^[A-Za-z]{2}\\s([A-Za-z]{2}|\\d{2})$",
		"BT": "^\\d{5}$",
		"BO": "^\\d{4}$",
		"BA": "^\\d{5}$",
		"BR": "^\\d{5}-\\d{3}$",
		"IO": "^[Bb]{2}[Nn][Dd]\\s{0,1}[1][Zz]{2}$",
		"VG": "^[Vv][Gg]\\d{4}$",
		"BN": "^[A-Za-z]{2}\\d{4}$",
		"BG": "^\\d{4}$",
		"KH": "^\\d{5}$",
		"CA": "^(?=[^DdFfIiOoQqUu\\d\\s])[A-Za-z]\\d(?=[^DdFfIiOoQqUu\\d\\s])[A-Za-z]\\s{0,1}\\d(?=[^DdFfIiOoQqUu\\d\\s])[A-Za-z]\\d$",
		"CV": "^\\d{4}$",
		"KY": "^[Kk][Yy]\\d[-\\s]{0,1}\\d{4}$",
		"TD": "^\\d{5}$",
		"CL": "^\\d{7}\\s\\(\\d{3}-\\d{4}\\)$",
		"CN": "^\\d{6}$",
		"CX": "^\\d{4}$",
		"CC": "^\\d{4}$",
		"CO": "^\\d{6}$",
		"CD": "^[Cc][Dd]$",
		"CR": "^\\d{4,5}$",
		"HR": "^\\d{5}$",
		"CU": "^\\d{5}$",
		"CY": "^\\d{4}$",
		"CZ": "^\\d{5}\\s\\(\\d{3}\\s\\d{2}\\)$",
		"DK": "^\\d{4}$",
		"DO": "^\\d{5}$",
		"EC": "^\\d{6}$",
		"SV": "^1101$",
		"EG": "^\\d{5}$",
		"EE": "^\\d{5}$",
		"ET": "^\\d{4}$",
		"FK": "^[Ff][Ii][Qq]{2}\\s{0,1}[1][Zz]{2}$",
		"FO": "^\\d{3}$",
		"FI": "^\\d{5}$",
		"FR": "^\\d{5}$",
		"GF": "^973\\d{2}$",
		"PF": "^987\\d{2}$",
		"GA": "^\\d{2}\\s[a-zA-Z-_ ]\\s\\d{2}$",
		"GE": "^\\d{4}$",
		"DE": "^\\d{2,5}$",
		"GI": "^[Gg][Xx][1]{2}\\s{0,1}[1][Aa]{2}$",
		"GR": "^\\d{3}\\s{0,1}\\d{2}$",
		"GL": "^\\d{4}$",
		"GP": "^971\\d{2}$",
		"GU": "^\\d{5}$",
		"GT": "^\\d{5}$",
		"GG": "^[A-Za-z]{2}\\d\\s{0,1}\\d[A-Za-z]{2}$",
		"GW": "^\\d{4}$",
		"HT": "^\\d{4}$",
		"HM": "^\\d{4}$",
		"HN": "^\\d{5}$",
		"HU": "^\\d{4}$",
		"IS": "^\\d{3}$",
		"IN": "^\\d{6}$",
		"ID": "^\\d{5}$",
		"IR": "^\\d{5}-\\d{5}$",
		"IQ": "^\\d{5}$",
		"IM": "^[Ii[Mm]\\d{1,2}\\s\\d\\[A-Z]{2}$",
		"IL": "^\\b\\d{5}(\\d{2})?$",
		"IT": "^\\d{5}$",
		"JM": "^\\d{2}$",
		"JP": "^\\d{7}\\s\\(\\d{3}-\\d{4}\\)$",
		"JE": "^[Jj][Ee]\\d\\s{0,1}\\d[A-Za-z]{2}$",
		"JO": "^\\d{5}$",
		"KZ": "^\\d{6}$",
		"KE": "^\\d{5}$",
		"KR": "^\\d{6}\\s\\(\\d{3}-\\d{3}\\)$",
		"XK": "^\\d{5}$",
		"KW": "^\\d{5}$",
		"KG": "^\\d{6}$",
		"LV": "^[Ll][Vv][- ]{0,1}\\d{4}$",
		"LA": "^\\d{5}$",
		"LB": "^\\d{4}\\s{0,1}\\d{4}$",
		"LS": "^\\d{3}$",
		"LR": "^\\d{4}$",
		"LY": "^\\d{5}$",
		"LI": "^\\d{4}$",
		"LT": "^[Ll][Tt][- ]{0,1}\\d{5}$",
		"LU": "^\\d{4}$",
		"MK": "^\\d{4}$",
		"MG": "^\\d{3}$",
		"MV": "^\\d{4,5}$",
		"MY": "^\\d{5}$",
		"MT": "^[A-Za-z]{3}\\s{0,1}\\d{4}$",
		"MH": "^\\d{5}$",
		"MQ": "^972\\d{2}$",
		"YT": "^976\\d{2}$",
		"MX": "^\\d{5}$",
		"FM": "^\\d{5}$",
		"MD": "^[Mm][Dd][- ]{0,1}\\d{4}$",
		"MC": "^980\\d{2}$",
		"MN": "^\\d{5}$",
		"ME": "^\\d{5}$",
		"MS": "^[Mm][Ss][Rr]\\s{0,1}\\d{4}$",
		"MA": "^\\d{5}$",
		"MZ": "^\\d{4}$",
		"MM": "^\\d{5}$",
		"NA": "^\\d{5}$",
		"NP": "^\\d{5}$",
		"NL": "^\\d{4}\\s{0,1}[A-Za-z]{2}$",
		"NC": "^988\\d{2}$",
		"NZ": "^\\d{4}$",
		"NI": "^\\d{5}$",
		"NE": "^\\d{4}$",
		"NG": "^\\d{6}$",
		"NF": "^\\d{4}$",
		"MP": "^\\d{5}$",
		"NO": "^\\d{4}$",
		"OM": "^\\d{3}$",
		"PK": "^\\d{5}$",
		"PW": "^\\d{5}$",
		"PA": "^\\d{6}$",
		"PG": "^\\d{3}$",
		"PY": "^\\d{4}$",
		"PE": "^\\d{5}$",
		"PH": "^\\d{4}$",
		"PN": "^[Pp][Cc][Rr][Nn]\\s{0,1}[1][Zz]{2}$",
		"PL": "^\\d{2}[- ]{0,1}\\d{3}$",
		"PT": "^\\d{4}$",
		"PR": "^\\d{5}$",
		"RE": "^974\\d{2}$",
		"RO": "^\\d{6}$",
		"RU": "^\\d{6}$",
		"BL": "^97133$",
		"SH": "^[Ss][Tt][Hh][Ll]\\s{0,1}[1][Zz]{2}$|^[Tt][Dd][Cc][Uu]\\s{0,1}[1][Zz]{2}$",
		"MF": "^97150$",
		"PM": "^97500$",
		"VC": "^[Vv][Cc]\\d{4}$",
		"SM": "^4789\\d$",
		"SA": "^\\d{5}(-{1}\\d{4})?$",
		"SN": "^\\d{5}$",
		"RS": "^\\d{5}$",
		"SG": "^\\d{6}$",
		"SK": "^\\d{5}\\s\\(\\d{3}\\s\\d{2}\\)$",
		"SI": "^([Ss][Ii][- ]{0,1}){0,1}\\d{4}$",
		"ZA": "^\\d{4}$",
		"GS": "^[Ss][Ii][Qq]{2}\\s{0,1}[1][Zz]{2}$",
		"ES": "^\\d{5}$",
		"LK": "^\\d{5}$",
		"SD": "^\\d{5}$",
		"SZ": "^[A-Za-z]\\d{3}$",
		"SE": "^\\d{3}\\s*\\d{2}$",
		"CH": "^\\d{4}$",
		"SJ": "^\\d{4}$",
		"TW": "^\\d{5}$",
		"TJ": "^\\d{6}$",
		"TH": "^\\d{5}$",
		"TT": "^\\d{6}$",
		"TN": "^\\d{4}$",
		"TR": "^\\d{5}$",
		"TM": "^\\d{6}$",
		"TC": "^[Tt][Kk][Cc][Aa]\\s{0,1}[1][Zz]{2}$",
		"UA": "^\\d{5}$",
		"GB": "^[A-Z]{1,2}[0-9R][0-9A-Z]?\\s*[0-9][A-Z-[CIKMOV]]{2}",
		"US": "^\\b\\d{5}\\b(?:[- ]{1}\\d{4})?$",
		"UY": "^\\d{5}$",
		"VI": "^\\d{5}$",
		"UZ": "^\\d{3} \\d{3}$",
		"VA": "^120$",
		"VE": "^\\d{4}(\\s[a-zA-Z]{1})?$",
		"VN": "^\\d{6}$",
		"WF": "^986\\d{2}$",
		"ZM": "^\\d{5}$",
	}
)
