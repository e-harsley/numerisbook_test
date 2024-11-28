package service

import (
	"context"
	"fmt"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/mongodb"
	"github.com/e-harsley/numerisbook_test/pkg"
	"github.com/e-harsley/numerisbook_test/pkg/model"
	"github.com/e-harsley/numerisbook_test/pkg/schema"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

var InvoiceRepository = mongodb.NewMongoDb(model.Invoice{})
var InvoiceReminderRepository = mongodb.NewMongoDb(model.InvoiceReminder{})
var InvoiceActivityLogRepository = mongodb.NewMongoDb(model.InvoiceActivityLog{})

type (
	InvoiceService struct {
	}
)

func (cls InvoiceService) Metrics(payload schema.UserBaseSchema) (map[string]interface{}, apiLayer.ErrorDetail) {

	pipeline := mongo.Pipeline{
		{{
			"$match", bson.D{
				{"user_id", payload.UserID},
			},
		}},
		{{
			"$group", bson.D{
				{"_id", "$status"},
				{"total_amount_due", bson.D{{"$sum", "$amount_due"}}},
			},
		}},
	}

	cursor, err := InvoiceRepository.Raw().Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Fatalf("Failed to execute aggregation: %v", err)
	}
	defer cursor.Close(context.Background())

	var results []map[string]interface{}
	if err = cursor.All(context.Background(), &results); err != nil {
		log.Fatalf("Failed to parse aggregation results: %v", err)
	}

	response := map[string]interface{}{
		"drafted":  0,
		"overdue":  0,
		"paid":     0,
		"not_paid": 0,
		"pending":  0,
	}

	for _, result := range results {
		if status, ok := result["_id"].(string); ok {
			if amount, ok := result["total_amount_due"]; ok {
				response[status] = amount
			}
		}
	}

	return response, apiLayer.ErrorDetail{}
}

func (cls InvoiceService) Register(payload schema.InvoiceRequestSchema) (invoice model.Invoice, resError apiLayer.ErrorDetail) {

	customerID, _ := primitive.ObjectIDFromHex(payload.CustomerID)

	customer, err := CustomerRepository.FindOne(bson.M{"_id": customerID})

	if err != nil {
		return invoice, apiLayer.FormatValidationError("customer_id", apiLayer.ValidateError, err.Error())
	}

	invoice, err = InvoiceRepository.BindModel(payload)

	if err != nil {
		return invoice, apiLayer.FormatValidationError("", apiLayer.CustomError, err.Error())
	}

	invoice.SetDefaultInvoice()

	invoice.CalculatePayment()

	invoice, err = InvoiceRepository.Save(invoice)

	if err != nil {
		return invoice, apiLayer.FormatValidationError("", apiLayer.CustomError, err.Error())
	}

	invConfigCursor, err := InvoiceConfigRepository.Find(bson.M{"active": true, "user_id": payload.UserID})

	if err != nil {
		return invoice, apiLayer.FormatValidationError("", apiLayer.CustomError, err.Error())
	}
	invConfigs := []model.InvoiceConfiguration{}
	err = invConfigCursor.ToSlice(invConfigs)

	for _, invConfig := range invConfigs {

		buildTimeForReminder := invConfig.GetReminderTriggerDate()

		payload := map[string]interface{}{
			"user_id":           payload.UserID,
			"invoice_id":        invoice.ID,
			"active":            true,
			"time_for_reminder": buildTimeForReminder,
		}

		_, _ = InvoiceReminderRepository.Save(payload)

	}

	activityLogMap := map[string]interface{}{
		"user_id":    payload.UserID,
		"action":     pkg.InvoiceCreated,
		"message":    fmt.Sprintf("invoice with id %s has be created for %s", invoice.InvoiceID, customer.Name),
		"invoice_id": invoice.ID,
	}

	_, _ = InvoiceActivityLogRepository.Save(activityLogMap)
	return invoice, resError
}
