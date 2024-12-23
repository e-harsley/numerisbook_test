package apiLayer

import (
	"fmt"
	"net/http"
)

var (
	MethodToStatusCode = map[string]int{
		http.MethodGet:     http.StatusOK,
		http.MethodPost:    http.StatusCreated,
		http.MethodPut:     http.StatusOK,
		http.MethodPatch:   http.StatusOK,
		http.MethodDelete:  http.StatusOK,
		http.MethodHead:    http.StatusNoContent,
		http.MethodOptions: http.StatusOK,
	}
)

type StandardResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"messsage"`
}

func genDefaultMsg(statusCode int) string {
	switch statusCode {
	case http.StatusOK:
		return "Ok"
	case http.StatusCreated:
		return "Created resource"
	case http.StatusAccepted:
		return "Accepted"
	case http.StatusUnprocessableEntity:
		return "Bad Request"
	default:
		return "Failed"
	}
}

var (
	rspWrapper = func(r *Response) interface{} {
		return &StandardResponse{
			Data:    r.Data,
			Message: genDefaultMsg(r.HTTPStatusCode),
		}
	}
	rspWrapNoOp = func(r *Response) interface{} {
		return r.Data
	}
)

func NoResponseWrap() {
	rspWrapper = rspWrapNoOp
}

func SetResponseWrapper(f func(r *Response) interface{}) {
	rspWrapper = f
}

// Response : response type for easy gin , it's a wrapper around the data with status code override
type Response struct {
	Data           interface{}
	HTTPStatusCode int
}

func (r *Response) JSONDATA() interface{} {
	return rspWrapper(r)
}

// Status : override wrapper with a status code instead of the default mapped ones
//
// Refer to : MethodToStatusCode to see the default ones
func (r *Response) Status(code int) *Response {
	r.HTTPStatusCode = code
	return r
}

// Ok : 200
func (r *Response) Ok() *Response {
	r.HTTPStatusCode = http.StatusOK
	return r
}

// Created : 201
func (r *Response) Created() *Response {
	r.HTTPStatusCode = http.StatusCreated
	return r
}

// Accepted : 202
func (r *Response) Accepted() *Response {
	r.HTTPStatusCode = http.StatusAccepted
	return r
}

// BadRequest : 400
func (r *Response) BadRequest() *Response {
	r.HTTPStatusCode = http.StatusBadRequest
	return r
}

// UnAuth : 401
func (r *Response) UnAuth() *Response {
	r.HTTPStatusCode = http.StatusUnauthorized
	return r
}

// Forebidden : 403
func (r *Response) Forebidden() *Response {
	r.HTTPStatusCode = http.StatusForbidden
	return r
}

// Conflict : 409
func (r *Response) Conflict() *Response {
	r.HTTPStatusCode = http.StatusConflict
	return r
}

// Fatal : 500
func (r *Response) Fatal() *Response {
	r.HTTPStatusCode = http.StatusInternalServerError
	return r
}

func ResWithBinding(data any, response interface{}) *Response {
	resp, err := SerializerFunc(data, response)

	fmt.Println("resp", &resp)

	if err != nil {
		return Err(err, 409)
	}
	return Res(resp)

}

// Res : response
func Res(data any) *Response {
	return &Response{
		Data: data,
	}
}

// Err : error Handler
func Err(err error, status int) *Response {

	return Res(err.Error()).Status(status)

	//return Res(fmt.Errorf("unexpected error fatal ... -> %w", err)).Status(500)
}

func FormatErrRes(err ErrorDetail, status int) *Response {

	return Res(err).Status(status)

	//return Res(fmt.Errorf("unexpected error fatal ... -> %w", err)).Status(500)
}
