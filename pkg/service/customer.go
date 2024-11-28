package service

import (
	"fmt"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/mongodb"
	"github.com/e-harsley/numerisbook_test/pkg/model"
	"github.com/e-harsley/numerisbook_test/pkg/schema"
	"go.mongodb.org/mongo-driver/bson"
)

var CustomerRepository = mongodb.NewMongoDb(model.Customer{})

type (
	CustomerService struct {
	}
)

func (cls CustomerService) Register(payload schema.CustomerRegister) (customer model.Customer, resError apiLayer.ErrorDetail) {
	filter := bson.M{
		"user_id": payload.UserID,
		"$or": []bson.M{
			{"email": payload.Email},
			{"phone": payload.Phone},
		},
	}

	fmt.Println("payload.UserID", payload.UserID)

	count, err := CustomerRepository.Count(filter)

	if err != nil {
		return customer, apiLayer.FormatValidationError("email_or_phone", apiLayer.ValidateError, err.Error())
	}

	if count > 0 {
		return customer, apiLayer.FormatValidationError("email_or_phone", apiLayer.ValidateError, "you already have a customer with this email or phone")
	}

	customer, err = CustomerRepository.Save(payload)

	if err != nil {
		return customer, apiLayer.FormatValidationError("", apiLayer.CustomError, err.Error())
	}

	return customer, resError
}
