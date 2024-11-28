package schema

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	BaseSchema struct {
	}
	UserBaseSchema struct {
		BaseSchema
		UserID primitive.ObjectID `json:"user_id"`
	}
)

func (bs BaseSchema) Validate() *apiLayer.ValidationError {
	return nil
}
