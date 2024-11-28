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
	"net/http"
	"net/http/httptest"
	"testing"
)

func NewCustomerRepository() mongodb.MongoDb[model.Customer] {
	return mongodb.NewMongoDb(model.Customer{})
}

func TestCustomerCreateIntegration(t *testing.T) {
	testDB := SetupTestDatabase()
	defer testDB.TearDown()

	originalNewMongoDatabase := mongodb.NewMongoDatabase
	mongodb.NewMongoDatabase = NewMongoDatabaseWithInstance(testDB.DbInstance)

	originalCustomerRepo := service.CustomerRepository
	service.CustomerRepository = NewCustomerRepository()

	defer func() {
		mongodb.NewMongoDatabase = originalNewMongoDatabase
		service.CustomerRepository = originalCustomerRepo
	}()

	router := mux.NewRouter()

	route.AuthRoute(router)
	route.CustomerRoute(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()
	t.Run("customer/register", func(t *testing.T) {
		signupPayload := schema.SignupSchema{
			FirstName: "John",
			LastName:  "Doe",
			Email:     "john.doe@example.com",
			Phone:     "+1234567890",
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

		customerPayload := schema.CustomerRegister{
			Name:    "John",
			Phone:   "Doe",
			Email:   "john.doe@example.com",
			Address: "heaven's gate",
		}

		jsonPayload, err := json.Marshal(customerPayload)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", testServer.URL+"/v1/customer/register", bytes.NewBuffer(jsonPayload))
		assert.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var protectedResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&protectedResponse)
		assert.NoError(t, err)

		data = protectedResponse["data"].(map[string]interface{})
		expectedData, ok := data["email"].(string)
		assert.Equal(t, customerPayload.Email, expectedData)
	})

}

func TestCustomerFindAllIntegration(t *testing.T) {
	testDB := SetupTestDatabase()
	defer testDB.TearDown()

	originalNewMongoDatabase := mongodb.NewMongoDatabase
	mongodb.NewMongoDatabase = NewMongoDatabaseWithInstance(testDB.DbInstance)

	originalCustomerRepo := service.CustomerRepository
	service.CustomerRepository = NewCustomerRepository()

	defer func() {
		mongodb.NewMongoDatabase = originalNewMongoDatabase
		service.CustomerRepository = originalCustomerRepo
	}()

	router := mux.NewRouter()
	route.AuthRoute(router)
	route.CustomerRoute(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	t.Run("find all customers", func(t *testing.T) {
		// First, create a few customers
		customers := []schema.CustomerRegister{
			{
				Name:    "John Doe",
				Phone:   "+1234567890",
				Email:   "john.doe@example.com",
				Address: "123 Main St",
			},
			{
				Name:    "Jane Smith",
				Phone:   "+0987654321",
				Email:   "jane.smith@example.com",
				Address: "456 Elm St",
			},
		}

		token := loginAndGetToken(t, testServer)

		for _, customerPayload := range customers {
			jsonPayload, err := json.Marshal(customerPayload)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", testServer.URL+"/v1/customer/register", bytes.NewBuffer(jsonPayload))
			assert.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}

		req, err := http.NewRequest("GET", testServer.URL+"/v1/customer", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var findAllResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&findAllResponse)
		assert.NoError(t, err)

		data, ok := findAllResponse["data"].([]interface{})
		assert.True(t, ok, "Expected data to be an array of customers")
		assert.GreaterOrEqual(t, len(data), 2, "Should have at least 2 customers")

		// Verify customer details
		customerEmails := make(map[string]bool)
		for _, customer := range data {
			customerMap, ok := customer.(map[string]interface{})
			assert.True(t, ok, "Each customer should be a map")

			email, ok := customerMap["email"].(string)
			assert.True(t, ok, "Customer should have an email")
			customerEmails[email] = true
		}

		assert.True(t, customerEmails["john.doe@example.com"], "John Doe should be in customers")
		assert.True(t, customerEmails["jane.smith@example.com"], "Jane Smith should be in customers")
	})
}

func TestCustomerFindOneIntegration(t *testing.T) {
	testDB := SetupTestDatabase()
	defer testDB.TearDown()

	originalNewMongoDatabase := mongodb.NewMongoDatabase
	mongodb.NewMongoDatabase = NewMongoDatabaseWithInstance(testDB.DbInstance)

	originalCustomerRepo := service.CustomerRepository
	service.CustomerRepository = NewCustomerRepository()

	defer func() {
		mongodb.NewMongoDatabase = originalNewMongoDatabase
		service.CustomerRepository = originalCustomerRepo
	}()

	router := mux.NewRouter()
	route.AuthRoute(router)
	route.CustomerRoute(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	t.Run("find specific customer", func(t *testing.T) {
		// Create a customer
		customerPayload := schema.CustomerRegister{
			Name:    "John Doe",
			Phone:   "+1234567890",
			Email:   "john.doe@example.com",
			Address: "123 Main St",
		}

		// Login to get authentication token
		token := loginAndGetToken(t, testServer)

		// Register customer
		jsonPayload, err := json.Marshal(customerPayload)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", testServer.URL+"/v1/customer/register", bytes.NewBuffer(jsonPayload))
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResponse)
		assert.NoError(t, err)

		data, ok := createResponse["data"].(map[string]interface{})
		assert.True(t, ok, "Expected data in response")

		customerId, ok := data["_id"].(string)
		assert.True(t, ok, "Customer ID should be a string")

		// Now find the specific customer
		req, err = http.NewRequest("GET", testServer.URL+"/v1/customer/"+customerId, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err = client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var findOneResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&findOneResponse)
		assert.NoError(t, err)

		customerData, ok := findOneResponse["data"].(map[string]interface{})
		assert.True(t, ok, "Expected data in response")

		// Verify customer details
		assert.Equal(t, customerPayload.Name, customerData["name"])
		assert.Equal(t, customerPayload.Email, customerData["email"])
		assert.Equal(t, customerPayload.Phone, customerData["phone"])
		assert.Equal(t, customerPayload.Address, customerData["address"])
	})
}

func loginAndGetToken(t *testing.T, testServer *httptest.Server) string {
	signupPayload := schema.SignupSchema{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Phone:     "+1234567890",
		Password:  "StrongPassword123!",
	}

	jsonSignupPayload, err := json.Marshal(signupPayload)
	assert.NoError(t, err)

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

	loginResp, err := http.Post(testServer.URL+"/v1/auth/login", "application/json", bytes.NewBuffer(jsonLoginPayload))
	assert.NoError(t, err)
	defer loginResp.Body.Close()

	assert.Equal(t, http.StatusCreated, loginResp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(loginResp.Body).Decode(&responseBody)
	assert.NoError(t, err)

	data := responseBody["data"].(map[string]interface{})
	token, ok := data["token"].(string)
	assert.True(t, ok, "Expected token in response")
	assert.NotEmpty(t, token, "Token should not be empty")

	return token
}
