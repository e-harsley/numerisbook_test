package apiLayer

type IRequest interface {
	Validate() *ValidationError
}

type SerializerHandler struct {
	createDto      interface{}
	updateDto      interface{}
	response       interface{}
	updateResponse interface{}
	fetchResponse  interface{}
}

func defaultSerializerHandler() *SerializerHandler {
	return &SerializerHandler{}
}

type SerializerHandlerFunc func(*SerializerHandler)

func CreateSchema(schema SerializerHandler) SerializerHandlerFunc {
	return func(opts *SerializerHandler) {
		opts.createDto = schema
	}
}

func UpdateSchema(schema SerializerHandler) SerializerHandlerFunc {
	return func(opts *SerializerHandler) {
		opts.updateDto = schema
	}
}

func ResponseSchema(schema SerializerHandler) SerializerHandlerFunc {
	return func(opts *SerializerHandler) {
		opts.response = schema
	}
}

func updateResponseSchema(schema SerializerHandler) SerializerHandlerFunc {
	return func(opts *SerializerHandler) {
		opts.updateResponse = schema
	}
}
func fetchResponseSchema(schema SerializerHandler) SerializerHandlerFunc {
	return func(opts *SerializerHandler) {
		opts.fetchResponse = schema
	}
}
