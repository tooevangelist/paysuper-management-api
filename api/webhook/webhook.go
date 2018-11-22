package webhook

import "github.com/ProtocolONE/p1pay.api/api"

type WebHook struct {
	*api.Api
}

func InitWebHook(api *api.Api) *WebHook {
	return &WebHook{Api: api}
}
