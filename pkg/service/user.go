package service

import (
	"github.com/e-harsley/numerisbook_test/numerisbookcore/apiLayer"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/mongodb"
	"github.com/e-harsley/numerisbook_test/pkg/model"
	"github.com/e-harsley/numerisbook_test/pkg/schema"
	"go.mongodb.org/mongo-driver/bson"
)

var UserRepository = mongodb.NewMongoDb(model.User{})

type (
	UserService struct {
	}
)

func (cls UserService) Register(payload schema.SignupSchema) (user model.User, resError apiLayer.ErrorDetail) {
	filter := bson.M{
		"$or": []bson.M{
			{"email": payload.Email},
			{"phone": payload.Phone},
		},
	}

	count, err := UserRepository.Count(filter)
	if err != nil {
		return user, apiLayer.FormatValidationError("email_or_password", apiLayer.ValidateError, err.Error())
	}

	if count > 0 {
		return user, apiLayer.FormatValidationError("email_or_password", apiLayer.ValidateError, "email or phone already exist")
	}

	user, err = UserRepository.BindModel(payload)

	if err != nil {
		return user, apiLayer.FormatValidationError("", apiLayer.CustomError, err.Error())

	}

	err = user.SetPassword()

	if err != nil {
		return user, apiLayer.FormatValidationError("", apiLayer.CustomError, err.Error())
	}

	user, err = UserRepository.Save(user)

	if err != nil {
		return user, apiLayer.FormatValidationError("", apiLayer.CustomError, err.Error())
	}
	return user, resError
}

func (cls UserService) Authenticate(payload schema.LoginSchema) (response map[string]interface{}, resError apiLayer.ErrorDetail) {
	filter := bson.M{
		"$or": []bson.M{
			{"email": payload.Username},
			{"phone": payload.Username},
		},
	}
	user, err := UserRepository.FindOne(filter)

	if err != nil {
		return response, apiLayer.FormatValidationError("email_or_password", apiLayer.ValidateError, "invalid username")
	}

	if !user.CheckPassword(payload.Password) {
		return nil, apiLayer.FormatValidationError("password", apiLayer.ValidateError, "invalid password")
	}

	authToken := apiLayer.NewAuthToken(user.ID.Hex())

	accessToken, err := authToken.Token()
	if err != nil {
		return response, apiLayer.FormatValidationError("", apiLayer.CustomError, err.Error())
	}

	data, err := apiLayer.MapDump(user)

	data["token"] = accessToken

	return data, resError
}
