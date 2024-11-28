package schema

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/pkg"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type (
	Item struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Quantity    int64   `json:"quantity"`
		Value       float64 `json:"value"`
		Currency    string  `json:"currency"`
		Total       float64 `json:"total"`
	}

	PaymentInformation struct {
		AccountNumber string `json:"account_number"`
		AccountName   string `json:"account_name"`
		AchRoutingNo  string `json:"ach_routing_no"`
		BankName      string `json:"bank_name"`
		BankAddress   string `json:"bank_address"`
	}

	InvoiceRequestSchema struct {
		CustomerID         string             `json:"customer_id" validate:"required"`
		Name               string             `json:"name" validate:"required"`
		DueDate            time.Time          `json:"due_date" validate:"required"`
		Currency           string             `json:"currency" validate:"required"`
		Items              []Item             `json:"items" validate:"required"`
		DiscountType       pkg.DiscountType   `json:"discount_type" validate:"required"`
		DiscountValue      float64            `json:"discount_value" validate:"required"`
		SubTotal           float64            `json:"sub_total"`
		AmountDue          float64            `json:"amount_due"`
		PaymentInformation PaymentInformation `json:"payment_information"  validate:"required"`
		Note               string             `json:"note" validate:"required"`
		UserID             primitive.ObjectID `json:"user_id"`
	}
)

func (cls InvoiceRequestSchema) Validate() *apiLayer.ValidationError {
	validate := validator.New()
	err := validate.Struct(cls)
	if err != nil {
		return apiLayer.NewValidationError(err.(validator.ValidationErrors))
	}
	return nil
}
