package route

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/pkg/handler"
	"github.com/e-harsley/numerisbook_test/pkg/service"
	"github.com/gorilla/mux"
)

var customerHandler = handler.CustomerHandler{}

var customerActions = apiLayer.CustomActions{
	apiLayer.Action{
		Name:        "/register",
		Handler:     apiLayer.Depend(customerHandler.Register, apiLayer.BindContext("primitive_user_context")),
		Method:      apiLayer.POST,
		Middlewares: []mux.MiddlewareFunc{privateMiddleware.AuthDeps},
	},
}

func CustomerRoute(r *mux.Router) {
	apiLayer.Crud(r, service.CustomerRepository, "/v1/customer", customerActions,
		apiLayer.WithoutCreate, []mux.MiddlewareFunc{privateMiddleware.AuthDeps},
		apiLayer.WithoutUpdate, apiLayer.BindContext("primitive_user_context"))
}
