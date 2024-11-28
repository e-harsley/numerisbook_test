package mongodb

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"
)

type (
	TestUser struct {
		ID        primitive.ObjectID `json:"_id" bson:"_id"`
		CreatedAt *time.Time         `json:"created_at" bson:"created_at"`
		UpdatedAt *time.Time         `json:"updated_at" bson:"updated_at"`
		Name      string             `json:"name" bson:"name"`
		Email     string             `json:"email" bson:"email"`
		Age       int64              `json:"age" bson:"age"`
	}
)

func (u TestUser) GetModelName() string {
	return "test_users"
}

func NewMongoDatabaseWithInstance(dbInstance *mongo.Database) func(string, string) (*mongo.Database, error) {
	return func(uri, database string) (*mongo.Database, error) {
		return dbInstance, nil
	}
}

func TestNewMongoDbOperations(t *testing.T) {
	testDB := SetupTestDatabase()
	defer testDB.TearDown()

	mongoDB := NewMongoDb(TestUser{})

	originalNewMongoDatabase := NewMongoDatabase
	NewMongoDatabase = NewMongoDatabaseWithInstance(testDB.DbInstance)
	defer func() {
		NewMongoDatabase = originalNewMongoDatabase
	}()

	t.Run("Save", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
			"age":   30,
		}
		savedUser, err := mongoDB.Save(userData)
		assert.NoError(t, err)
		assert.NotNil(t, savedUser)
		assert.Equal(t, "John Doe", savedUser.Name)
		assert.Equal(t, "john@example.com", savedUser.Email)
	})

	t.Run("FindOne", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":  "Jane Smith",
			"email": "jane@example.com",
			"age":   25,
		}

		_, err := mongoDB.Save(userData)

		assert.NoError(t, err)

		foundUser, err := mongoDB.FindOne(bson.M{"name": "Jane Smith"})

		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, "Jane Smith", foundUser.Name)
	})

	t.Run("Update", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":  "Update Test",
			"email": "update@example.com",
			"age":   35,
		}
		_, err := mongoDB.Save(userData)
		assert.NoError(t, err)

		// Update the user
		updateData := map[string]interface{}{
			"name": "Updated Name",
			"age":  40,
		}
		updatedUser, err := mongoDB.Update(bson.M{"name": "Update Test"}, updateData)

		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", updatedUser.Name)
		assert.Equal(t, int64(40), updatedUser.Age)
	})

	t.Run("Delete", func(t *testing.T) {
		userData := map[string]interface{}{
			"name":  "Delete Test",
			"email": "delete@email.com",
			"age":   45,
		}
		_, err := mongoDB.Save(userData)
		assert.NoError(t, err)

		result, err := mongoDB.Delete(bson.M{"name": "Delete Test"})
		assert.NoError(t, err)
		assert.NotNil(t, result)

		_, err = mongoDB.FindOne(bson.M{"name": "Delete Test"})
		assert.Error(t, err)
	})

	t.Run("Count", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			userData := map[string]interface{}{
				"name":  fmt.Sprintf("Count User %d", i),
				"email": fmt.Sprintf("count%d@example.com", i),
				"age":   20 + i,
			}
			_, err := mongoDB.Save(userData)

			assert.NoError(t, err)
		}
		count, err := mongoDB.Count(bson.M{"age": bson.M{"$gte": 22}})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(3))
	})

	t.Run("FindWithOptions", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			userData := map[string]interface{}{
				"name":  fmt.Sprintf("Options User %d", i),
				"email": fmt.Sprintf("options%d@example.com", i),
				"age":   20 + i,
			}
			_, err := mongoDB.Save(userData)
			assert.NoError(t, err)
		}

		limit := int64(5)

		findOptions := FindOptions{Limit: &limit}
		findOptions.SetFetchAllLinks(true)

		cursor, err := mongoDB.Find(bson.M{}, &findOptions)
		assert.NoError(t, err)

		models := []TestUser{}

		err = cursor.ToSlice(&models)
		fmt.Println("models", models)

		assert.NoError(t, err)
		assert.Equal(t, 5, len(models))
	})
}
