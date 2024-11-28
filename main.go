package main

import (
	"github.com/e-harsley/numerisbook_test/api"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func CORS(next http.Handler) http.Handler {
	// Just incase a frontend wants to connect to it.... used to prevent cors error
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()
	api.APIRoute(r)
	handler := CORS(r)
	err := http.ListenAndServe(apiLayer.GetEnv("PORT", ":8082"), handler)
	if err != nil {
		log.Fatal(err)
	}
}
