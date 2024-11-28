package utils

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
	"path/filepath"
	"time"
)

func BindDataOperationStruct(data interface{}, output interface{}) error {
	jsonString, _ := json.Marshal(data)

	err := json.Unmarshal([]byte(jsonString), &output)

	if err != nil {

		return err
	}

	return nil

}

func init() {
	// Get the root directory of the project
	rootDir, err := findRootDir()
	if err != nil {
		log.Fatalf("Error finding root directory: %v", err)
	}

	// Load .env file from the root directory
	envPath := filepath.Join(rootDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
}

// findRootDir attempts to find the project root directory
func findRootDir() (string, error) {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree until we find a .env file or reach the filesystem root
	for {
		// Check if .env exists in the current directory
		envPath := filepath.Join(currentDir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			return currentDir, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)

		// If we've reached the root, stop searching
		if parentDir == currentDir {
			return "", fmt.Errorf("could not find .env file")
		}

		currentDir = parentDir
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

func CleanMap(data interface{}, includeKeys []string) (map[string]interface{}, error) {

	dataMap := map[string]interface{}{}

	err := BindDataOperationStruct(data, &dataMap)

	if err != nil {
		return nil, err
	}
	for _, key := range includeKeys {

		switch key {
		case "_id":
			dataMap["_id"] = primitive.NewObjectID()
		case "created_at":
			dataMap["created_at"] = time.Now()
		case "updated_at":
			dataMap["updated_at"] = time.Now()
		}

	}
	return dataMap, nil

}
