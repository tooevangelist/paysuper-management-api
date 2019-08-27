package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Auth1 struct {
	Issuer       string `envconfig:"AUTH1_ISSUER" required:"true" default:"https://dev-auth1.tst.protocol.one"`
	ClientId     string `envconfig:"AUTH1_CLIENTID" required:"true"`
	ClientSecret string `envconfig:"AUTH1_CLIENTSECRET" required:"true"`
	RedirectUrl  string `envconfig:"AUTH1_REDIRECTURL" required:"true"`
}

type Config struct {
	Auth1

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
