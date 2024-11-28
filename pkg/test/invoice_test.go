package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Rhymond/go-money"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/mongodb"
	"github.com/e-harsley/numerisbook_test/pkg"
	"github.com/e-harsley/numerisbook_test/pkg/route"
	"github.com/e-harsley/numerisbook_test/pkg/schema"
	"github.com/e-harsley/numerisbook_test/pkg/service"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestInvoiceCreateIntegration(t *testing.T) {
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
	route.InvoiceRoute(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()
	t.Run("invoice/register", func(t *testing.T) {
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

		items := []schema.Item{
			schema.Item{
				Name:        "shoe",
				Description: "shoe payment invoice",
				Quantity:    3,
				Value:       1000,
				Currency:    money.NGN,
			},
		}
		paymentInformation := schema.PaymentInformation{
			AccountNumber: "22939939393",
			AccountName:   "United Bank of Africa",
			AchRoutingNo:  "routing ach",
			BankName:      "janet doe",
			BankAddress:   "heaven's gate",
		}

		invoicePayload := schema.InvoiceRequestSchema{
			CustomerID:         data["_id"].(string),
			Name:               "invoice for shoe",
			DueDate:            time.Now().Add(10),
			Currency:           money.NGN,
			Items:              items,
			DiscountType:       pkg.Percentage,
			DiscountValue:      20,
			PaymentInformation: paymentInformation,
			Note:               "please pay",
		}

		jsonPayload, err = json.Marshal(invoicePayload)
		assert.NoError(t, err)

		req, err = http.NewRequest("POST", testServer.URL+"/v1/invoice/register", bytes.NewBuffer(jsonPayload))
		assert.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+token)

		client = &http.Client{}
		resp, err = client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		err = json.NewDecoder(resp.Body).Decode(&protectedResponse)
		assert.NoError(t, err)

		data = protectedResponse["data"].(map[string]interface{})
		expectedData, ok = data["name"].(string)
		assert.Equal(t, invoicePayload.Name, expectedData)

	})

}

func TestInvoiceFindAllIntegration(t *testing.T) {
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
	route.InvoiceRoute(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	t.Run("find all invoices", func(t *testing.T) {
		token, _ := signupAndLogin(t, testServer)

		_, customerID := createCustomer(t, testServer, token)

		invoicePayloads := []schema.InvoiceRequestSchema{
			{
				CustomerID: customerID,
				Name:       "Invoice 1",
				DueDate:    time.Now().Add(10),
				Currency:   money.NGN,
				Items: []schema.Item{
					{
						Name:        "Item 1",
						Description: "First item description",
						Quantity:    2,
						Value:       1000,
						Currency:    money.NGN,
					},
				},
				DiscountType:  pkg.Percentage,
				DiscountValue: 10,
				PaymentInformation: schema.PaymentInformation{
					AccountNumber: "1111111111",
					AccountName:   "Bank Account 1",
					AchRoutingNo:  "routing1",
					BankName:      "Bank 1",
					BankAddress:   "Address 1",
				},
				Note: "First invoice note",
			},
			{
				CustomerID: customerID,
				Name:       "Invoice 2",
				DueDate:    time.Now().Add(20),
				Currency:   money.NGN,
				Items: []schema.Item{
					{
						Name:        "Item 2",
						Description: "Second item description",
						Quantity:    3,
						Value:       2000,
						Currency:    money.NGN,
					},
				},
				DiscountType:  pkg.Percentage,
				DiscountValue: 15,
				PaymentInformation: schema.PaymentInformation{
					AccountNumber: "2222222222",
					AccountName:   "Bank Account 2",
					AchRoutingNo:  "routing2",
					BankName:      "Bank 2",
					BankAddress:   "Address 2",
				},
				Note: "Second invoice note",
			},
		}

		// Create invoices
		createdInvoiceIDs := make([]string, 0)
		for _, payload := range invoicePayloads {
			invoiceID := createInvoice(t, testServer, token, payload)
			createdInvoiceIDs = append(createdInvoiceIDs, invoiceID)
		}

		// Find all invoices
		req, err := http.NewRequest("GET", testServer.URL+"/v1/invoice", nil)
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
		assert.True(t, ok, "Expected data to be an array of invoices")
		assert.GreaterOrEqual(t, len(data), 2, "Should have at least 2 invoices")

		// Verify invoice details
		invoiceNames := make(map[string]bool)
		for _, invoice := range data {
			invoiceMap, ok := invoice.(map[string]interface{})
			assert.True(t, ok, "Each invoice should be a map")

			name, ok := invoiceMap["name"].(string)
			assert.True(t, ok, "Invoice should have a name")
			invoiceNames[name] = true
		}

		assert.True(t, invoiceNames["Invoice 1"], "Invoice 1 should be in invoices")
		assert.True(t, invoiceNames["Invoice 2"], "Invoice 2 should be in invoices")
	})
}

func TestInvoiceFindOneIntegration(t *testing.T) {
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
	route.InvoiceRoute(router)
	testServer := httptest.NewServer(router)
	defer testServer.Close()

	t.Run("find specific invoice", func(t *testing.T) {
		// Signup and login
		token, _ := signupAndLogin(t, testServer)

		// Create customer
		_, customerID := createCustomer(t, testServer, token)

		// Create invoice
		invoicePayload := schema.InvoiceRequestSchema{
			CustomerID: customerID,
			Name:       "Specific Invoice",
			DueDate:    time.Now().Add(15),
			Currency:   money.NGN,
			Items: []schema.Item{
				{
					Name:        "Test Item",
					Description: "Test item description",
					Quantity:    2,
					Value:       1500,
					Currency:    money.NGN,
				},
			},
			DiscountType:  pkg.Percentage,
			DiscountValue: 20,
			PaymentInformation: schema.PaymentInformation{
				AccountNumber: "3333333333",
				AccountName:   "Test Bank Account",
				AchRoutingNo:  "routing3",
				BankName:      "Test Bank",
				BankAddress:   "Test Address",
			},
			Note: "Specific invoice note",
		}

		// Create invoice and get its ID
		invoiceID := createInvoice(t, testServer, token, invoicePayload)

		// Find the specific invoice
		req, err := http.NewRequest("GET", testServer.URL+"/v1/invoice/"+invoiceID, nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var findOneResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&findOneResponse)
		assert.NoError(t, err)

		invoiceData, ok := findOneResponse["data"].(map[string]interface{})
		assert.True(t, ok, "Expected data in response")

		// Verify invoice details
		assert.Equal(t, invoicePayload.Name, invoiceData["name"])
		assert.Equal(t, invoicePayload.Note, invoiceData["note"])
	})
}

// Helper function to signup and login
func signupAndLogin(t *testing.T, testServer *httptest.Server) (string, schema.SignupSchema) {
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

	return token, signupPayload
}

// Helper function to create a customer
func createCustomer(t *testing.T, testServer *httptest.Server, token string) (schema.CustomerRegister, string) {
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

	data := protectedResponse["data"].(map[string]interface{})
	customerID, ok := data["_id"].(string)
	assert.True(t, ok, "Customer ID should be a string")

	return customerPayload, customerID
}

// Helper function to create an invoice
func createInvoice(t *testing.T, testServer *httptest.Server, token string, invoicePayload schema.InvoiceRequestSchema) string {
	jsonPayload, err := json.Marshal(invoicePayload)
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", testServer.URL+"/v1/invoice/register", bytes.NewBuffer(jsonPayload))
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

	data := protectedResponse["data"].(map[string]interface{})
	invoiceID, ok := data["_id"].(string)
	assert.True(t, ok, "Invoice ID should be a string")

	return invoiceID
}
