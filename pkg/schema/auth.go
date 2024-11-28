package schema

import (
	"fmt"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type (
	SignupSchema struct {
		BaseSchema
		FirstName string `json:"first_name" validate:"required"`
		Email     string `json:"email" validate:"required"`
		LastName  string `json:"last_name" validate:"required"`
		Phone     string `json:"phone" validate:"required"`
		Password  string `json:"password" validate:"required"`
		Name      string `json:"name"`
	}

	LoginSchema struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required"`
	}

	SignupResponseSchema struct {
		ID        primitive.ObjectID `json:"_id"`
		CreatedAt *time.Time         `json:"created_at"`
		UpdatedAt *time.Time         `json:"updated_at"`
		Name      string             `json:"name"`
		Phone     string             `json:"phone"`
		Email     string             `json:"email"`
	}

	LoginResponseSchema struct {
		ID        primitive.ObjectID `json:"_id"`
		CreatedAt *time.Time         `json:"created_at"`
		UpdatedAt *time.Time         `json:"updated_at"`
		Name      string             `json:"name"`
		Phone     string             `json:"phone"`
		Email     string             `json:"email"`
		Token     string             `json:"token"`
	}
)

func (cls *SignupSchema) Bind() {
	cls.Name = fmt.Sprintf("%s %s", cls.FirstName, cls.LastName)

}

func (cls SignupSchema) Validate() *apiLayer.ValidationError {
	validate := validator.New()
	err := validate.Struct(cls)
	if err != nil {
		return apiLayer.NewValidationError(err.(validator.ValidationErrors))
	}
	return nil
}

func (cls LoginSchema) Validate() *apiLayer.ValidationError {
	validate := validator.New()
	err := validate.Struct(cls)
	if err != nil {
		return apiLayer.NewValidationError(err.(validator.ValidationErrors))
	}
	return nil
}
