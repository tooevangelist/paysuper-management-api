package config

import (
	"github.com/kelseyhightower/envconfig"
)

const (
	DefaultJwtSignAlgorithm = "RS256"
)

type Auth1 struct {
	Issuer       string `envconfig:"AUTH1_ISSUER" required:"true" default:"https://dev-auth1.tst.protocol.one"`
	ClientId     string `envconfig:"AUTH1_CLIENTID" required:"true"`
	ClientSecret string `envconfig:"AUTH1_CLIENTSECRET" required:"true"`
	RedirectUrl  string `envconfig:"AUTH1_REDIRECTURL" required:"true"`
}

type S3 struct {
	AwsAccessKeyId     string `envconfig:"AWS_ACCESS_KEY_ID" required:"true"`
	AwsSecretAccessKey string `envconfig:"AWS_SECRET_ACCESS_KEY" required:"true"`
	AwsRegion          string `envconfig:"AWS_REGION" default:"eu-west-1"`
	AwsBucket          string `envconfig:"AWS_BUCKET" required:"true"`
}

type Config struct {
	Auth1
	S3

	HttpScheme  string `envconfig:"HTTP_SCHEME" default:"https"`
	AmqpAddress string `envconfig:"AMQP_ADDRESS" required:"true" default:"amqp://127.0.0.1:5672"`
	Environment string `envconfig:"ENVIRONMENT" default:"test"`
}

func NewConfig() (error, *Config) {
	var err error

	config := Config{}

	if err = envconfig.Process("", &config); err != nil {
		return err, nil
	}

	return nil, &config
}
