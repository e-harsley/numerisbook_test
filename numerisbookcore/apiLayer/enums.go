package apiLayer

type EndpointActionTypes string

const (
	POST   EndpointActionTypes = "post"
	PUT    EndpointActionTypes = "put"
	DELETE EndpointActionTypes = "DELETE"
	GET    EndpointActionTypes = "get"
)

type ErrorType string

const (
	ValidateError ErrorType = "validation_error"
	CustomError   ErrorType = "custom_error"
)
