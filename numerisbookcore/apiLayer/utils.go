package apiLayer

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
)

// ErrorDetail represents the structure of the validation error.
type ErrorDetail struct {
	Type ErrorType   `json:"type"`
	Loc  string      `json:"loc"`
	Msg  interface{} `json:"msg"`
}

func extractContextBinds(bindFroms []string) (ctxKeys []string) {
	for _, v := range bindFroms {
		if strings.Contains(v, bindContext) {
			ctxKeys = append(ctxKeys, strings.Split(v, ctxSep)[1])
		}
	}
	return ctxKeys
}

func BindContext(ctxKey string) string {
	return bindContext + ctxSep + ctxKey
}

func withoutContexts(bindFroms []string) (nonCtx []string) {
	for _, v := range bindFroms {
		if !strings.Contains(v, bindContext) {
			nonCtx = append(nonCtx, v)
		}
	}
	return nonCtx
}

func SliceContains[T comparable](haystack []T, needle T) bool {
	for _, v := range haystack {
		if needle == v {
			return true
		}
	}
	return false
}

func bindContextAndQueryParams(r *http.Request, c *C, bindFrom []string, varsToBind map[string]string, dtoCastFrom interface{}) error {
	contextKeysToBind := extractContextBinds(bindFrom)

	canBindAll := len(withoutContexts(bindFrom)) == 0

	fmt.Println("herllp", canBindAll, SliceContains(bindFrom, BindJSON))

	if canBindAll || SliceContains(bindFrom, BindJSON) {
		err := json.NewDecoder(r.Body).Decode(&dtoCastFrom)
		if err != nil {
			return err
		}
	}

	if canBindAll || SliceContains(bindFrom, BindQuery) {
		queryParams := r.URL.Query()
		for key, value := range queryParams {
			varsToBind[key] = value[0]
		}
	}

	for _, cKey := range contextKeysToBind {
		val := c.R.Context().Value(cKey)
		if val == nil {
			fmt.Println("warning!! context key == " + cKey + " does not exist")
			continue
		}
		err := copier.Copy(&dtoCastFrom, val)
		if err != nil {
			return fmt.Errorf("context key %s made a bad cast: %v", cKey, err)
		}
	}
	return nil
}

func buildQuery(filter map[string]interface{}) bson.M {
	fmt.Println("????? param", filter)
	query := bson.M{}
	fmt.Println(query)
	for key, value := range filter {
		fmt.Println("????????????", key, value)
		if val, ok := value.(string); ok {
			if primitiveUserID, err := primitive.ObjectIDFromHex(val); err == nil {
				fmt.Printf("Converted %s to ObjectID: %v\n", key, primitiveUserID)
				value = primitiveUserID
			}
		}
		switch key {
		case "$or", "$and":
			subqueries := make([]bson.M, 0)
			subfilter, ok := value.([]map[string]interface{})
			if !ok {
				continue
			}

			for _, sub := range subfilter {
				subqueries = append(subqueries, buildQuery(sub))
			}

			if len(subqueries) > 0 {
				query[key] = subqueries
			}
		default:
			valueType := reflect.TypeOf(value)
			fmt.Println("chill value value", value)
			if valueType.Kind() == reflect.Map {
				subquery := buildQuery(value.(map[string]interface{}))
				query[key] = subquery
			} else {

				fmt.Println("yes d value", value)
				query[key] = value
			}
		}
	}
	fmt.Println("query", query)

	return query
}

func buildSort(m map[string]interface{}) bson.D {
	var bsonD bson.D
	fmt.Println(bson.D{{Key: "created_at", Value: -1}})
	for key, value := range m {
		sortOrder := 1
		if s, ok := value.(string); ok && strings.ToLower(s) == "desc" {
			sortOrder = -1 // Set to descending
		}
		bsonD = append(bsonD, bson.E{Key: key, Value: sortOrder})
	}
	return bsonD
}

func SerializerFunc(result interface{}, obj interface{}) (interface{}, error) {

	resultType := reflect.TypeOf(result)
	objType := reflect.TypeOf(obj)
	schemaInstance := reflect.New(reflect.TypeOf(obj)).Interface()

	hydrate := reflect.ValueOf(schemaInstance).MethodByName("Hydrate")

	if hydrate.IsValid() {
		newObj := map[string]interface{}{}
		if resultType.Kind() == reflect.Slice {
			var resp []interface{}
			resultValue := reflect.ValueOf(result)
			for i := 0; i < resultValue.Len(); i++ {
				element := resultValue.Index(i)
				elementJSON, err := json.Marshal(element.Interface())
				if err != nil {
					return nil, err
				}
				var newItem map[string]interface{}
				err = json.Unmarshal(elementJSON, &newItem)
				if err != nil {
					return nil, err
				}
				hydrateResult := hydrate.Call([]reflect.Value{reflect.ValueOf(newItem)})

				if len(hydrateResult) < 1 {
					return nil, errors.New("hydrate should return an interface")
				}
				resp = append(resp, hydrateResult[0].Interface())
			}
			result = resp
		} else {
			elementJSON, err := json.Marshal(result)
			if err != nil {
				return nil, err
			}

			err = json.Unmarshal(elementJSON, &newObj)
			if err != nil {
				return nil, err
			}
			hydrateResult := hydrate.Call([]reflect.Value{reflect.ValueOf(newObj)})

			if len(hydrateResult) < 1 {
				return nil, errors.New("hydrate should return an interface")
			}
			result = hydrateResult[0].Interface()
		}

	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	if resultType.Kind() == reflect.Slice {
		sliceType := reflect.SliceOf(objType)
		newObj := reflect.New(sliceType).Interface()
		if err := json.Unmarshal(resultJSON, newObj); err != nil {
			return nil, err
		}
		return newObj, nil
	} else {
		newObj := reflect.New(objType).Interface()
		if err := json.Unmarshal(resultJSON, newObj); err != nil {
			return nil, err
		}
		return newObj, nil
	}
}

func MapDump(data interface{}) (map[string]interface{}, error) {
	var mod map[string]interface{}
	jsonString, _ := json.Marshal(data)
	if err := json.Unmarshal(jsonString, &mod); err != nil {
		return mod, err
	}
	return mod, nil
}

func DeriveKey(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

func Encrypt(plaintext string, passphrase string) (string, error) {
	key := DeriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
func Decrypt(ciphertext string, passphrase string) (string, error) {
	key := DeriveKey(passphrase)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// FormatValidationError formats a validation error into a slice of ErrorDetail.
func FormatValidationError(key string, errorType ErrorType, message interface{}) ErrorDetail {
	if errorType == "" {
		errorType = "unknown_err"
	}
	return ErrorDetail{
		Type: errorType,
		Loc:  key,
		Msg:  message,
	}
}

func GetEnv(key, fallback string) string {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file:", err)
		return fallback
	}
	value, exists := os.LookupEnv(key)
	if !exists {
		value = fallback
	}
	return value
}

func ArrayHasValue[T comparable](value T, array []T) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

func PopByValue[T comparable](array []T, value T) ([]T, T, error) {
	for i, v := range array {
		if v == value {
			array = append(array[:i], array[i+1:]...)
			return array, v, nil
		}
	}
	var zeroValue T
	return array, zeroValue, errors.New("value not found in the array")
}
