package manager

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/ProtocolONE/authone-jwt-verifier-golang"
	"github.com/labstack/echo"
	"go.uber.org/zap"
)

type OauthManager struct {
	JwtVerifier *jwtverifier.JwtVerifier
	Logger      *zap.SugaredLogger
}

func InitOauthManager(jwtVerifier *jwtverifier.JwtVerifier, logger *zap.SugaredLogger) *OauthManager {
	return &OauthManager{JwtVerifier: jwtVerifier, Logger: logger}
}

func (m *OauthManager) GetAuthUrl(state map[string]interface{}) string {
	data, _ := json.Marshal(&state)
	s := base64.StdEncoding.EncodeToString([]byte(data))

	return m.JwtVerifier.CreateAuthUrl(s)
}

func (m *OauthManager) ExchangeCodeToTokens(ctx echo.Context, code string) (*jwtverifier.Token, error) {
	tokens, err := m.JwtVerifier.Exchange(ctx.Request().Context(), code)
	if err != nil {
		m.Logger.Errorf("Unable to change code to tokens with error: %s", err)
		return nil, errors.New("unable to change code to tokens")
	}

	return tokens, nil
}
