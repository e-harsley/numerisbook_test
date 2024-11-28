package route

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/pkg/handler"
	"github.com/e-harsley/numerisbook_test/pkg/service"
	"github.com/gorilla/mux"
)

var invoiceHandler = handler.InvoiceHandler{}

var invoiceActions = apiLayer.CustomActions{
	apiLayer.Action{
		Name:        "/register",
		Handler:     apiLayer.Depend(invoiceHandler.Register, apiLayer.BindContext("primitive_user_context")),
		Method:      apiLayer.POST,
		Middlewares: []mux.MiddlewareFunc{privateMiddleware.AuthDeps},
	},
	apiLayer.Action{
		Name:        "/invoice",
		Handler:     apiLayer.Depend(invoiceHandler.Metrics, apiLayer.BindContext("primitive_user_context")),
		Method:      apiLayer.GET,
		Middlewares: []mux.MiddlewareFunc{privateMiddleware.AuthDeps},
	},
}

func InvoiceRoute(r *mux.Router) {
	apiLayer.Crud(r, service.InvoiceRepository, "/v1/invoice", invoiceActions,
		apiLayer.WithoutCreate,
		apiLayer.WithoutUpdate)
}

var invoiceMetricsActions = apiLayer.CustomActions{
	apiLayer.Action{
		Name:        "/invoice",
		Handler:     apiLayer.Depend(invoiceHandler.Metrics, apiLayer.BindContext("primitive_user_context")),
		Method:      apiLayer.GET,
		Middlewares: []mux.MiddlewareFunc{privateMiddleware.AuthDeps},
	},
}

func InvoiceMetricRoute(r *mux.Router) {
	apiLayer.Crud(r, service.InvoiceRepository, "/v1/metric", invoiceMetricsActions,
		apiLayer.WithoutCreate, apiLayer.WithoutFetch, apiLayer.WithoutList,
		apiLayer.WithoutUpdate)
}
