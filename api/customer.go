package api

import "github.com/minio/minio-go"

type customerRoute struct {
	*Api
}

func (api *Api) initOnboardingRoutes() (*Api, error) {
	route := &customerRoute{Api: api}

	api.authUserRouteGroup.GET("/merchants", route.listMerchants)
	api.authUserRouteGroup.GET("/merchants/:id", route.getMerchant)
	api.authUserRouteGroup.GET("/merchants/user", route.getMerchantByUser)
	api.authUserRouteGroup.POST("/merchants", route.changeMerchant)
	api.authUserRouteGroup.PUT("/merchants", route.changeMerchant)
	api.authUserRouteGroup.PUT("/merchants/:id/change-status", route.changeMerchantStatus)
	api.authUserRouteGroup.PATCH("/merchants/:id", route.changeAgreement)

	api.authUserRouteGroup.GET("/merchants/:id/agreement", route.generateAgreement)
	api.authUserRouteGroup.GET("/merchants/:id/agreement/document", route.getAgreementDocument)
	api.authUserRouteGroup.POST("/merchants/:id/agreement/document", route.uploadAgreementDocument)

	api.authUserRouteGroup.POST("/merchants/:merchant_id/notifications", route.createNotification)
	api.authUserRouteGroup.GET("/merchants/:merchant_id/notifications/:notification_id", route.getNotification)
	api.authUserRouteGroup.GET("/merchants/:merchant_id/notifications", route.listNotifications)
	api.authUserRouteGroup.PUT("/merchants/:merchant_id/notifications/:notification_id/mark-as-read", route.markAsReadNotification)

	api.authUserRouteGroup.GET("/merchants/:merchant_id/methods/:method_id", route.getPaymentMethod)
	api.authUserRouteGroup.GET("/merchants/:merchant_id/methods", route.listPaymentMethods)
	api.authUserRouteGroup.PUT("/merchants/:merchant_id/methods/:method_id", route.changePaymentMethod)

	return api, nil
}
