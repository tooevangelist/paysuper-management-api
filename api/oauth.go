package api

import (
	"github.com/labstack/echo"
	"github.com/paysuper/paysuper-management-api/manager"
	"net/http"
)

type OauthApiV1 struct {
	*Api
	oauthManager *manager.OauthManager
}

func (api *Api) InitOauthRoutes() *Api {
	apiV1 := OauthApiV1{
		Api:          api,
		oauthManager: manager.InitOauthManager(api.jwtVerifier, api.logger),
	}

	api.Http.GET("/oauth/auth", apiV1.auth)
	api.Http.GET("/oauth/callback", apiV1.callback)

	return api
}

// @Summary Authenticate and registration endpoint
// @Description Generates an authorization link to the authorization server and redirects the user there
// @Tags Auth
// @Accept text/html
// @Produce html
// @Success 302 {string} html "Redirect user to authentication form"
// @Router /oauth/auth [get]
func (apiV1 *OauthApiV1) auth(ctx echo.Context) error {
	state := map[string]interface{}{
		"foo": "bar",
	}
	return ctx.Redirect(http.StatusFound, apiV1.oauthManager.GetAuthUrl(state))
}

// @Summary Get tokens by authentication code
// @Description Get full list of currencies or get list of currencies filtered by name
// @Tags Currency
// @Accept json
// @Produce json
// @Param name query string false "name or symbolic ISO 4217 code of currency"
// @Success 200 {object} jwtverifier.Token "OK"
// @Failure 400 {object} model.Error "Authentication code not found"
// @Failure 403 {object} model.Error "Unable to exchange cote to tokens"
// @Router /oauth/callback [get]
func (apiV1 *OauthApiV1) callback(ctx echo.Context) error {
	code := ctx.QueryParam("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "authentication code not found")
	}

	tokens, err := apiV1.oauthManager.ExchangeCodeToTokens(ctx, code)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return ctx.JSON(http.StatusOK, tokens)
}
