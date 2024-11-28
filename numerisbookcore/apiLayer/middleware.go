package apiLayer

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strings"
)

type ExtraParameters struct {
	UserID string `json:"user_id"`
}

type PrimitiveExtraParameters struct {
	UserID primitive.ObjectID `json:"user_id"`
}

type MiddlewareFunc func(http.Handler) http.Handler

type Middleware struct {
	PublicRoute bool
}

func (m Middleware) AuthDeps(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("AuthDeps middleware invoked. PublicRoute: %v, Path: %s\n", m.PublicRoute, r.URL.Path)

		if m.PublicRoute {
			fmt.Println("Public route detected. Skipping authentication.")
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errorResponse := map[string]interface{}{
				"code":    "ENDPOINT.OPERATION.RESTRICTED",
				"message": "This activity cannot be carried out because the endpoint is restricted",
			}

			// Convert the map to a JSON string
			errorJSON, err := json.Marshal(errorResponse)
			if err != nil {
				// Handle the error if JSON marshalling fails
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Write the JSON error response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(errorJSON)
			return
		}
		if !strings.HasPrefix(authHeader, "Bearer ") {
			fmt.Println("Invalid token format.")
			errorResponse := map[string]interface{}{
				"code":    "ENDPOINT.OPERATION.RESTRICTED",
				"message": "This activity cannot be carried out because the endpoint is restricted",
			}
			errorJSON, err := json.Marshal(errorResponse)
			if err != nil {
				// Handle the error if JSON marshalling fails
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(errorJSON)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		auth := AuthToken{}
		authToken, err := auth.ParseAuthToken(token)

		if err != nil {
			fmt.Println("Error parsing token:", err)
			errorResponse := map[string]interface{}{
				"code":    "ENDPOINT.OPERATION.RESTRICTED",
				"message": "This activity cannot be carried out because the endpoint is restricted",
			}
			errorJSON, err := json.Marshal(errorResponse)
			if err != nil {
				// Handle the error if JSON marshalling fails
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(errorJSON)
			return
		}
		fmt.Println("Token parsed successfully. UserID:", authToken.UserID)
		primitiveUserHex, _ := primitive.ObjectIDFromHex(authToken.UserID)
		ctx := r.Context()
		ctx = context.WithValue(ctx, "primitive_user_context", PrimitiveExtraParameters{UserID: primitiveUserHex})
		ctx = context.WithValue(ctx, "user_context", ExtraParameters{UserID: authToken.UserID})

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
