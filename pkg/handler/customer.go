package handler

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/pkg/schema"
	"github.com/e-harsley/numerisbook_test/pkg/service"
	"net/http"
)

type CustomerHandler struct {
	service service.CustomerService
}

func (cls CustomerHandler) Register(req schema.CustomerRegister, c apiLayer.C) *apiLayer.Response {
	customer, err := cls.service.Register(req)

	if err.Type != "" {
		if err.Type == apiLayer.ValidateError {
			return apiLayer.FormatErrRes(err, http.StatusUnprocessableEntity)
		}
		return apiLayer.FormatErrRes(err, http.StatusBadRequest)
	}
	return apiLayer.ResWithBinding(customer, schema.CustomerResponse{})
}
