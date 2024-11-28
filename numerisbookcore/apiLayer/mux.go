package apiLayer

import (
	"fmt"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/mongodb"
	"github.com/gorilla/mux"
	"net/http"
	"reflect"
	"strings"
)

type routeInfo struct {
	path   string
	method string
}

func applyMiddleware(subRouter *mux.Router, middleware []mux.MiddlewareFunc) {
	for _, m := range middleware {
		subRouter.Use(m)
	}
}

type Action struct {
	Name        string
	Handler     http.HandlerFunc
	Middlewares []mux.MiddlewareFunc
	Method      EndpointActionTypes
}

type CustomActions []Action

func (ca *CustomActions) AddAction(action Action) {
	*ca = append(*ca, action)
}

func Crud[M mongodb.DocumentModel](router *mux.Router, repository mongodb.MongoDb[M], relativePath string, options ...interface{}) *mux.Router {
	subRouter := router.PathPrefix(relativePath).Subrouter()
	fmt.Printf("Adding CRUD routes for path: %s\n", relativePath)

	defaultEndpointOptions := defaultEndpointOpts()
	defaultSerializer := defaultSerializerHandler()

	var customActions CustomActions
	var crudOpts MuxOptions
	var middleware []mux.MiddlewareFunc

	if len(options) < 1 {
		options = append(options, defaultSerializer, defaultEndpointOptions)
	}

	for _, option := range options {
		fmt.Println("Processing option:", reflect.TypeOf(option))
		switch opt := option.(type) {
		case []mux.MiddlewareFunc:
			middleware = append(middleware, opt...)
		case func(*EndpointOption):
			fmt.Println("Configuring EndpointOption")
			opt(defaultEndpointOptions)
		case Action:
			customActions.AddAction(opt)
		case CustomActions:
			for _, customAction := range opt {
				customActions.AddAction(customAction)
			}
		case string:
			if strings.HasPrefix(opt, bindPrefix) {
				customBindFunc := EndpointContextBind(opt)
				customBindFunc(defaultEndpointOptions)
			}
		case SerializerHandlerFunc:
			opt(defaultSerializer)
		}
	}

	applyMiddleware(subRouter, middleware)

	crudOpts = append(crudOpts, muxOptions(repository, *defaultEndpointOptions, *defaultSerializer)...)
	for _, opt := range crudOpts {
		fmt.Println("Applying CRUD Option")
		subRouter = opt(subRouter)
	}

	for _, action := range customActions {
		handler := action.Handler
		for _, middleware := range action.Middlewares {
			handler = middleware(handler).ServeHTTP
		}

		path := fmt.Sprintf("%s/%s", relativePath, action.Name)
		fmt.Printf("Registering route: %s %s\n", action.Method, path)
		switch action.Method {
		case POST:
			subRouter.HandleFunc(action.Name, handler).Methods("POST")
		case PUT:
			subRouter.HandleFunc(action.Name, handler).Methods("PUT")
		case DELETE:
			subRouter.HandleFunc(action.Name, handler).Methods("DELETE")
		case GET:
			subRouter.HandleFunc(action.Name, handler).Methods("GET")
		}
	}

	fmt.Println("Routes have been registered. Listing all routes:")
	err := subRouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		methods, err := route.GetMethods()
		if err != nil {
			return err
		}
		fmt.Printf("Route path: %s, Methods: %v\n", path, methods)
		return nil
	})
	if err != nil {
		fmt.Println("Error walking routes:", err)
	}

	return subRouter
}
