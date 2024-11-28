package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/mongodb"
	"github.com/e-harsley/numerisbook_test/pkg/model"
	"github.com/e-harsley/numerisbook_test/pkg/route"
	"github.com/e-harsley/numerisbook_test/pkg/schema"
	"github.com/e-harsley/numerisbook_test/pkg/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewUserRepository() mongodb.MongoDb[model.User] {
	return mongodb.NewMongoDb(model.User{})
}

func TestAuthIntegration(t *testing.T) {
	testDB := SetupTestDatabase()
	defer testDB.TearDown()

	originalNewMongoDatabase := mongodb.NewMongoDatabase
	mongodb.NewMongoDatabase = NewMongoDatabaseWithInstance(testDB.DbInstance)

	originalUserRepo := service.UserRepository
	service.UserRepository = NewUserRepository()

	defer func() {
		mongodb.NewMongoDatabase = originalNewMongoDatabase
		service.UserRepository = originalUserRepo
	}()

	router := mux.NewRouter()

	route.AuthRoute(router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	t.Run("signup", func(t *testing.T) {
		signupPayload := schema.SignupSchema{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Phone:     "+1234567890",
			Password:  "StrongPassword123!",
		}

		jsonPayload, err := json.Marshal(signupPayload)

		assert.NoError(t, err)
		fmt.Println("url here", testServer.URL+"/v1/auth/signup")
		resp, err := http.Post(testServer.URL+"/v1/auth/signup", "application/json", bytes.NewBuffer(jsonPayload))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		user, err := service.UserRepository.FindOne(bson.M{"email": signupPayload.Email})
		assert.NoError(t, err)
		assert.Equal(t, signupPayload.Email, user.Email)
		assert.NotEmpty(t, user.Password)
	})
	t.Run("login", func(t *testing.T) {
		signupPayload := schema.SignupSchema{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe3@example.com",
			Phone:     "+12345678390",
			Password:  "StrongPassword123!",
		}

		jsonSignupPayload, err := json.Marshal(signupPayload)
		assert.NoError(t, err)

		fmt.Println("url here", testServer.URL+"/v1/auth/signup")
		signupResp, err := http.Post(testServer.URL+"/v1/auth/signup", "application/json", bytes.NewBuffer(jsonSignupPayload))
		assert.NoError(t, err)
		defer signupResp.Body.Close()

		assert.Equal(t, http.StatusCreated, signupResp.StatusCode)

		loginPayload := schema.LoginSchema{
			Username: signupPayload.Email,
			Password: signupPayload.Password,
		}

		jsonLoginPayload, err := json.Marshal(loginPayload)
		assert.NoError(t, err)

		fmt.Println("url here", testServer.URL+"/v1/auth/login")
		loginResp, err := http.Post(testServer.URL+"/v1/auth/login", "application/json", bytes.NewBuffer(jsonLoginPayload))
		assert.NoError(t, err)
		defer loginResp.Body.Close()

		assert.Equal(t, http.StatusCreated, loginResp.StatusCode)

		var responseBody map[string]interface{}
		err = json.NewDecoder(loginResp.Body).Decode(&responseBody)
		assert.NoError(t, err)

		fmt.Println("responseBody", responseBody)
		data := responseBody["data"].(map[string]interface{})
		token, ok := data["token"].(string)
		assert.True(t, ok, "Expected token in response")
		assert.NotEmpty(t, token, "Token should not be empty")
	})

}
