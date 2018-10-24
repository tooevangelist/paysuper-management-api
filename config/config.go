package config

import (
	"crypto/rsa"
	"encoding/base64"
	"github.com/dgrijalva/jwt-go"
	"github.com/kelseyhightower/envconfig"
)

const (
	DefaultJwtSignAlgorithm = "RS256"
)

type Database struct {
	Host     string `envconfig:"MONGO_HOST"`
	Database string `envconfig:"MONGO_DB"`
	User     string `envconfig:"MONGO_USER"`
	Password string `envconfig:"MONGO_PASSWORD"`
}

type Jwt struct {
	SignatureSecret       *rsa.PublicKey
	SignatureSecretBase64 string `envconfig:"JWT_SIGNATURE_SECRET"`
	Algorithm             string `envconfig:"JWT_ALGORITHM"`
}

type Config struct {
	Jwt
	Database
}

func NewConfig() (error, *Config) {
	var err error

	config := Config{}

	if err = envconfig.Process("", &config); err != nil {
		return err, nil
	}

	if config.Jwt.Algorithm == "" {
		config.Jwt.Algorithm = DefaultJwtSignAlgorithm
	}

	pemKey, err := base64.StdEncoding.DecodeString(config.Jwt.SignatureSecretBase64)

	if err != nil {
		return err, nil
	}

	config.Jwt.SignatureSecret, err = jwt.ParseRSAPublicKeyFromPEM(pemKey)

	if err != nil {
		return err, nil
	}

	return nil, &config
}
