package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type (
	User struct {
		ID        primitive.ObjectID `json:"_id" bson:"_id"`
		CreatedAt *time.Time         `json:"created_at" bson:"created_at"`
		UpdatedAt *time.Time         `json:"updated_at" bson:"updated_at"`
		Name      string             `json:"name" bson:"name"`
		Phone     string             `json:"phone" bson:"phone"`
		Email     string             `json:"email" bson:"email"`
		Password  string             `json:"password" bson:"password"`
	}

	Customer struct {
		ID        primitive.ObjectID `json:"_id" bson:"_id"`
		CreatedAt *time.Time         `json:"created_at" bson:"created_at"`
		UpdatedAt *time.Time         `json:"updated_at" bson:"updated_at"`
		UserID    primitive.ObjectID `json:"user_id" bson:"user_id" lookup:"user:user_id:_id:user:struct"`
		User      User               `json:"user" bson:"-"`
		Name      string             `json:"name" bson:"name"`
		Phone     string             `json:"phone" bson:"phone"`
		Email     string             `json:"email" bson:"email"`
		Address   string             `json:"address" bson:"address"`
	}
)

func (user User) GetModelName() string {
	return "user"
}

func (user Customer) GetModelName() string {
	return "customer"
}

func (cls *User) SetPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cls.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	cls.Password = string(hashedPassword)

	return nil
}

func (cls *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(cls.Password), []byte(password))
	return err == nil
}
