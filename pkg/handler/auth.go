package handler

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/pkg/schema"
	"github.com/e-harsley/numerisbook_test/pkg/service"
	"net/http"
)

type AuthHandler struct {
	service service.UserService
}

func (cls AuthHandler) Signup(req schema.SignupSchema, c apiLayer.C) *apiLayer.Response {

	req.Bind()
	user, err := cls.service.Register(req)

	if err.Type != "" {
		if err.Type == apiLayer.ValidateError {
			return apiLayer.FormatErrRes(err, http.StatusUnprocessableEntity)
		}
		return apiLayer.FormatErrRes(err, http.StatusBadRequest)
	}
	return apiLayer.ResWithBinding(user, schema.SignupResponseSchema{})
}

func (cls AuthHandler) Authenticate(req schema.LoginSchema, c apiLayer.C) *apiLayer.Response {

	user, err := cls.service.Authenticate(req)
	if err.Type != "" {
		if err.Type == apiLayer.ValidateError {
			return apiLayer.FormatErrRes(err, http.StatusUnprocessableEntity)
		}
		return apiLayer.FormatErrRes(err, http.StatusBadRequest)
	}
	return apiLayer.ResWithBinding(user, schema.LoginResponseSchema{})
}
