package handler

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/pkg/schema"
	"github.com/e-harsley/numerisbook_test/pkg/service"
	"net/http"
)

type InvoiceHandler struct {
	service service.InvoiceService
}

func (cls InvoiceHandler) Register(req schema.InvoiceRequestSchema, c apiLayer.C) *apiLayer.Response {
	invoice, err := cls.service.Register(req)

	if err.Type != "" {
		if err.Type == apiLayer.ValidateError {
			return apiLayer.FormatErrRes(err, http.StatusUnprocessableEntity)
		}
		return apiLayer.FormatErrRes(err, http.StatusBadRequest)
	}
	return apiLayer.ResWithBinding(invoice, schema.CustomerResponse{})
}

func (cls InvoiceHandler) Metrics(req schema.UserBaseSchema, c apiLayer.C) *apiLayer.Response {
	metrics, err := cls.service.Metrics(req)

	if err.Type != "" {
		if err.Type == apiLayer.ValidateError {
			return apiLayer.FormatErrRes(err, http.StatusUnprocessableEntity)
		}
		return apiLayer.FormatErrRes(err, http.StatusBadRequest)
	}
	return apiLayer.ResWithBinding(metrics, map[string]interface{}{})
}
