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
	Environment string `envconfig:"ENVIRONMENT" default:"test"`
	PaymentFormJsLibraryUrl string `envconfig:"PAYMENT_FORM_JS_LIBRARY_URL" required:"true"`
	WebsocketUrl            string `envconfig:"WEBSOCKET_URL" default:"wss://cf.tst.protocol.one/connection/websocket"`
}

func NewConfig() (error, *Config) {
	var err error

	config := Config{}

	if err = envconfig.Process("", &config); err != nil {
		return err, nil
	}

	return nil, &config
}
