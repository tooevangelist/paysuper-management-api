module github.com/paysuper/paysuper-management-api

require (
	github.com/ProtocolONE/authone-jwt-verifier-golang v0.0.0-20190327070329-4dd563b01681
	github.com/ProtocolONE/geoip-service v0.0.0-20190903084234-1d5ae6b96679
	github.com/ProtocolONE/rabbitmq v0.0.0-20190129162844-9f24367e139c
	github.com/SebastiaanKlippert/go-wkhtmltopdf v1.4.1
	github.com/amalfra/etag v0.0.0-20180217025506-c1ee3b8b3121
	github.com/apex/log v1.1.0
	github.com/aws/aws-sdk-go v1.23.8
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/labstack/echo/v4 v4.1.6
	github.com/micro/go-grpc v0.6.0 // indirect
	github.com/micro/go-micro v1.8.0
	github.com/micro/go-plugins v1.2.0
	github.com/micro/grpc-go v0.0.0-20180913204047-2c703400301b // indirect
	github.com/micro/kubernetes v0.2.0 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/paysuper/paysuper-aws-manager v0.0.0-20190827071211-4aff35ed4d82
	github.com/paysuper/paysuper-billing-server v0.0.0-20190904133438-ca187438857e
	github.com/paysuper/paysuper-payment-link v0.0.0-20190903143854-b799a77c03ce
	github.com/paysuper/paysuper-recurring-repository v1.0.123
	github.com/paysuper/paysuper-tax-service v0.0.0-20190903084038-7849f394f122
	github.com/spf13/viper v1.3.1 // indirect
	github.com/streadway/amqp v0.0.0-20190404075320-75d898a42a94
	github.com/stretchr/testify v1.4.0
	github.com/ttacon/builder v0.0.0-20170518171403-c099f663e1c2 // indirect
	github.com/ttacon/libphonenumber v1.0.1
	go.uber.org/zap v1.10.0
	gopkg.in/go-playground/validator.v9 v9.29.1
	gopkg.in/mgo.v2 v2.0.0-20180705113604-9856a29383ce
)

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
