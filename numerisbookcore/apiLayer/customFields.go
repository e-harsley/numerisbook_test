package apiLayer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"
)

const (
	BindJSON  = "json_bind"
	BindQuery = "query_bind"
	BindURI   = "uri_bind"

	bindContext = "context_bind"
	ctxSep      = ":::"
	bindPrefix  = "context_bind:::"
)

type PaginationResponse struct {
	Page    int64 `json:"page"`
	PerPage int64 `json:"perPage"`
	Skip    int64 `json:"skip"`
	Count   int64 `json:"count"`
}

type C struct {
	W http.ResponseWriter
	R *http.Request
}

type GeneralResponse struct {
	Status  int         `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type GeneralGetResponse struct {
	Status     int                    `json:"status"`
	Pagination PaginationResponse     `json:"pagination"`
	FilterBy   map[string]interface{} `json:"filter_by"`
	SortBy     bson.D                 `json:"sort_by"`
	Data       interface{}            `json:"data,omitempty"`
	Message    string                 `json:"message,omitempty"`
}

func responseJSON(res http.ResponseWriter, status int, object interface{}) {
	res.Header().Set("Content-Resource", "application/json")
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(status)
	err := json.NewEncoder(res).Encode(object)

	if err != nil {
		return
	}
}

func (c *C) BindJSON(data interface{}) error {
	if err := json.NewDecoder(c.R.Body).Decode(data); err != nil {
		fmt.Println("ERROR: ", err)
		return err
	}
	return nil
}

func (c *C) Query(key string) string {
	query := c.R.URL.Query()
	return query.Get(key)
}

func (c *C) Response(status int, data interface{}, message string) {
	responseSuccess := GeneralResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
	responseJSON(c.W, status, responseSuccess)
}

func (c *C) GetResponse(status int, filterBy map[string]interface{}, sort bson.D, pagination PaginationResponse, data interface{}, message string) {
	responseSuccess := GeneralGetResponse{
		FilterBy:   filterBy,
		Pagination: pagination,
		SortBy:     sort,
		Status:     status,
		Message:    message,
		Data:       data,
	}
	responseJSON(c.W, status, responseSuccess)
}

func (c *C) Params(key string) string {
	return mux.Vars(c.R)[key]
}

type AuthToken struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func NewAuthToken(userId string) *AuthToken {
	return &AuthToken{
		UserID: userId,
	}
}

func (cls AuthToken) Token() (string, error) {

	atClaims := jwt.MapClaims{}
	atClaims["access_uuid"] = uuid.New().String()
	atClaims["user_id"] = cls.UserID
	atClaims["aud"] = GetEnv("JWT_AUDIENCE_CLAIM", "")
	atClaims["iss"] = GetEnv("JWT_ISSUER_CLAIM", "")
	atClaims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	return at.SignedString([]byte(GetEnv("JWT_SECRET_KEY", "")))
}

func (cls AuthToken) ParseAuthToken(tokenString string) (*AuthToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &cls, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetEnv("JWT_SECRET_KEY", "")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthToken); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("invalid token")
	}
}
