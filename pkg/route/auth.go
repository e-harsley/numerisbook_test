package route

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/pkg/handler"
	"github.com/e-harsley/numerisbook_test/pkg/service"
	"github.com/gorilla/mux"
)

var authHandler = handler.AuthHandler{}

var authActions = apiLayer.CustomActions{
	apiLayer.Action{
		Name:    "/signup",
		Handler: apiLayer.Depend(authHandler.Signup),
		Method:  apiLayer.POST,
	},
	apiLayer.Action{
		Name:    "/login",
		Handler: apiLayer.Depend(authHandler.Authenticate),
		Method:  apiLayer.POST,
	},
}

func AuthRoute(r *mux.Router) {
	apiLayer.Crud(r, service.UserRepository, "/v1/auth", authActions,
		apiLayer.WithoutCreate, apiLayer.WithoutList, apiLayer.WithoutFetch,
		apiLayer.WithoutUpdate)
}
