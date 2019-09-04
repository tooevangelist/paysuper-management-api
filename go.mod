module github.com/paysuper/paysuper-management-api

require (
	github.com/ProtocolONE/authone-jwt-verifier-golang v0.0.0-20190327070329-4dd563b01681
	github.com/ProtocolONE/geoip-service v0.0.0-20190130072841-bf3b3b79a742
	github.com/SebastiaanKlippert/go-wkhtmltopdf v1.4.1
	github.com/aws/aws-sdk-go v1.23.8
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/golang/protobuf v1.3.2 // indirects
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo/v4 v4.1.6
	github.com/micro/go-micro v1.8.0
	github.com/micro/go-plugins v1.2.0
	github.com/paysuper/paysuper-aws-manager v0.0.0-20190827071211-4aff35ed4d82
	github.com/paysuper/paysuper-billing-server v0.0.0-20190826080453-5c3b4dafd15e
	github.com/paysuper/paysuper-payment-link v0.0.0-20190410180823-800306b3fd7c
	github.com/paysuper/paysuper-recurring-repository v1.0.123
	github.com/paysuper/paysuper-reporter v0.0.0-20190904051107-e889f74e5e0b
	github.com/paysuper/paysuper-tax-service v0.0.0-20190722140034-a37f835eaad7
	github.com/stretchr/testify v1.4.0
	github.com/ttacon/libphonenumber v1.0.1
	go.uber.org/zap v1.10.0
	gopkg.in/go-playground/validator.v9 v9.29.1
)

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
