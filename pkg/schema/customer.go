package schema

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	CustomerRegister struct {
		Name    string             `json:"name" validate:"required"`
		Phone   string             `json:"phone" validate:"required"`
		Email   string             `json:"email" validate:"required"`
		Address string             `json:"address" validate:"required"`
		UserID  primitive.ObjectID `json:"user_id"`
	}

	CustomerResponse struct {
		ID      string `json:"_id"`
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Email   string `json:"email"`
		Address string `json:"address"`
		UserID  string `json:"userID"`
	}
)

func (cls CustomerRegister) Validate() *apiLayer.ValidationError {
	validate := validator.New()
	err := validate.Struct(cls)
	if err != nil {
		return apiLayer.NewValidationError(err.(validator.ValidationErrors))
	}
	return nil
}
