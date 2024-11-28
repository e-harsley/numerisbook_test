package route

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/pkg/service"
	"github.com/gorilla/mux"
)

func InvoiceConfigurationRoute(r *mux.Router) {
	apiLayer.Crud(r, service.InvoiceConfigRepository, "/v1/invoice-configuration",
		[]mux.MiddlewareFunc{privateMiddleware.AuthDeps},
		apiLayer.WithoutUpdate, apiLayer.BindContext("primitive_user_context"))
}
