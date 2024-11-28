package api

import (
	"github.com/e-harsley/numerisbook_test/pkg/route"
	"github.com/gorilla/mux"
)

func APIRoute(r *mux.Router) {
	route.AuthRoute(r)
	route.InvoiceConfigurationRoute(r)
	route.CustomerRoute(r)
	route.InvoiceRoute(r)
	route.AuditLogRoute(r)
	route.InvoiceMetricRoute(r)
}
