package test

import (
	"encoding/json"
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
	"net/url"
	"testing"
	"time"
)

func TestAuditLogFindAllIntegration(t *testing.T) {
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
		req, err := http.NewRequest("GET", testServer.URL+"/v1/activity-log", nil)
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

			name, ok := invoiceMap["action"].(string)
			assert.True(t, ok, "Invoice should have a name")
			invoiceNames[name] = true
		}

		assert.True(t, invoiceNames["Invoice 1"], "invoice_created")
		assert.True(t, invoiceNames["Invoice 2"], "invoice_created")
	})
}

func TestAuditLogFilterByInvoiceID(t *testing.T) {
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

	t.Run("filter activity log by invoice ID", func(t *testing.T) {
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
		}

		// Create invoice
		invoiceID := createInvoice(t, testServer, token, invoicePayloads[0])

		// Prepare filter request
		filterQuery := map[string]interface{}{
			"invoice_id": map[string]interface{}{
				"$eq": invoiceID,
			},
		}
		filterJSON, err := json.Marshal(filterQuery)
		assert.NoError(t, err)

		// Create request with filter
		req, err := http.NewRequest("GET", testServer.URL+"/v1/activity-log?filter_by="+url.QueryEscape(string(filterJSON)), nil)
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
		assert.True(t, ok, "Expected data to be an array of activity logs")
		assert.GreaterOrEqual(t, len(data), 1, "Should have at least 1 activity log for the invoice")

		// Verify activity log details
		for _, log := range data {
			logMap, ok := log.(map[string]interface{})
			assert.True(t, ok, "Each log should be a map")

			logInvoiceID, ok := logMap["invoice_id"].(string)
			assert.True(t, ok, "Log should have an invoice_id")
			assert.Equal(t, invoiceID, logInvoiceID, "Log should be for the specific invoice")
		}
	})
}
