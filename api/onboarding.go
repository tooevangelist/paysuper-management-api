package api

import "github.com/labstack/echo"

type onboardingRoute struct {
	*Api
}

func (api *Api) initOnboardingRoutes() *Api {
	route := onboardingRoute{Api: api}

	api.accessRouteGroup.GET("/merchant", route.getMerchant)
	api.accessRouteGroup.POST("/merchant", route.createMerchant)
	api.accessRouteGroup.PUT("/merchant", route.updateMerchant)

	return api
}

func (r *onboardingRoute) getMerchant(ctx echo.Context) error {
	return nil
}

func (r *onboardingRoute) listMerchants(ctx echo.Context) error {
	return nil
}

func (r *onboardingRoute) createMerchant(ctx echo.Context) error {
	return nil
}

func (r *onboardingRoute) updateMerchant(ctx echo.Context) error {
	return nil
}

func (r *onboardingRoute) getPaymentMethod(ctx echo.Context) error {
	return nil
}

func (r *onboardingRoute) listPaymentMethods(ctx echo.Context) error {
	return nil
}

func (r *onboardingRoute) createPaymentMethod(ctx echo.Context) error {
	return nil
}

func (r *onboardingRoute) updatePaymentMethod(ctx echo.Context) error {
	return nil
}

func (r *onboardingRoute) deletePaymentMethod(ctx echo.Context) error {
	return nil
}
