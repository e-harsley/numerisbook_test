package apiLayer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/mongodb"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// Mock Cursor Implementation
type MockCursor struct {
	items []interface{}
}

func (m *MockCursor) ToSlice(v interface{}) error {
	val := reflect.ValueOf(v).Elem()
	val.Set(reflect.ValueOf(m.items))
	return nil
}

// TestModel Definition
type TestModel struct {
	ID   primitive.ObjectID `json:"_id" bson:"_id"`
	Name string             `json:"name"`
}

func (t TestModel) GetModelName() string {
	return "test_model"
}

// TestRequest Definition
type TestRequest struct {
	Name string `json:"name" validate:"required"`
}

func (t *TestRequest) Validate() *ValidationError {
	if t.Name == "" {
		return &ValidationError{
			Errors: map[string]interface{}{
				"name": "name is required",
			},
		}
	}
	return nil
}

// Comprehensive Mock Implementation
type MockMongoDb[M mongodb.DocumentModel] struct {
	mock.Mock
}

// Implement Find method
func (m *MockMongoDb[M]) Find(filter interface{}, options ...*mongodb.FindOptions) (mongodb.Cursor, error) {
	args := m.Called(filter, options)
	if args.Get(0) == nil {
		return mongodb.Cursor{}, args.Error(1)
	}
	return args.Get(0).(mongodb.Cursor), args.Error(1)
}

// Implement FindOne method
func (m *MockMongoDb[M]) FindOne(filter interface{}, options ...*mongodb.FindOptions) (M, error) {
	args := m.Called(filter, options)

	// Handle zero value case
	if args.Get(0) == nil {
		var zeroValue M
		return zeroValue, args.Error(1)
	}

	result, ok := args.Get(0).(M)
	if !ok {
		var zeroValue M
		return zeroValue, fmt.Errorf("type assertion failed")
	}

	return result, args.Error(1)
}

// Implement Save method
func (m *MockMongoDb[M]) Save(data interface{}) (M, error) {
	args := m.Called(data)

	// Handle zero value case
	if args.Get(0) == nil {
		var zeroValue M
		return zeroValue, args.Error(1)
	}

	result, ok := args.Get(0).(M)
	if !ok {
		var zeroValue M
		return zeroValue, fmt.Errorf("type assertion failed")
	}

	return result, args.Error(1)
}

// Implement Update method
func (m *MockMongoDb[M]) Update(filter bson.M, data interface{}) (M, error) {
	args := m.Called(filter, data)

	// Handle zero value case
	if args.Get(0) == nil {
		var zeroValue M
		return zeroValue, args.Error(1)
	}

	result, ok := args.Get(0).(M)
	if !ok {
		var zeroValue M
		return zeroValue, fmt.Errorf("type assertion failed")
	}

	return result, args.Error(1)
}

// Implement Delete method
func (m *MockMongoDb[M]) Delete(filter bson.M) (interface{}, error) {
	args := m.Called(filter)
	return args.Get(0), args.Error(1)
}

// Implement Count method
func (m *MockMongoDb[M]) Count(filter bson.M) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

// Test Function
func TestCrudFunctions(t *testing.T) {
	router := mux.NewRouter()
	mockRepo := mongodb.NewMongoDb(TestModel{})
	// Create mock repository

	testCases := []struct {
		name string
		//setupMock      func(*MockMongoDb[TestModel])
		requestBody    interface{}
		method         string
		path           string
		expectedStatus int
	}{
		{
			name: "Successful List Endpoint",
			//setupMock: func(mr *MockMongoDb[TestModel]) {
			//	mockCursor := &MockCursor{
			//		items: []interface{}{
			//			TestModel{Name: "Test1"},
			//			TestModel{Name: "Test2"},
			//		},
			//	}
			//	mr.On("Find", mock.Anything, mock.Anything).Return(mockCursor, nil)
			//	mr.On("Count", mock.Anything).Return(int64(2), nil)
			//},
			method:         "GET",
			path:           "/test",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Successful Create Endpoint",
			//setupMock: func(mr *MockMongoDb[TestModel]) {
			//	mr.On("Save", mock.Anything).Return(TestModel{Name: "NewTest"}, nil)
			//},
			requestBody:    &TestRequest{Name: "NewTest"},
			method:         "POST",
			path:           "/test",
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			//tc.setupMock(mockRepo)

			// Create subrouter with CRUD
			subRouter := Crud(router, mockRepo, "/test")

			// Prepare request
			var req *http.Request
			var err error
			if tc.method == "POST" {
				bodyBytes, _ := json.Marshal(tc.requestBody)
				req, err = http.NewRequest(tc.method, tc.path, bytes.NewBuffer(bodyBytes))
			} else {
				req, err = http.NewRequest(tc.method, tc.path, nil)
			}
			fmt.Println("ERROR OCCURE", err)
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create ResponseRecorder
			recorder := httptest.NewRecorder()
			subRouter.ServeHTTP(recorder, req)

			// Assert status code
			fmt.Println("STATUSE gerrr", recorder.Code)
			assert.Equal(t, tc.expectedStatus, recorder.Code)

			// Reset mocks after each test case
			//mockRepo.AssertExpectations(t)
		})
	}
}
