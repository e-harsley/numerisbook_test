package utils

import (
	"crypto/rand"
	"fmt"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numbers  = "0123456789"
	alphaNum = alphabet + numbers
)

func GenerateAlphNumericID(length int) string {
	return randomStringFromAlphabet(length, alphaNum)
}

func GenerateNumericID(length, fill int, prefix string) string {
	id := randomStringFromAlphabet(length, numbers)
	return prefix + fmt.Sprintf("%0*s", fill, id)
}

func GenerateAlphaID(length, fill int, prefix string) string {
	id := randomStringFromAlphabet(length, alphabet)
	return prefix + fmt.Sprintf("%0*s", fill, id)
}

func randomStringFromAlphabet(length int, alphabet string) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	for i, b := range bytes {
		bytes[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(bytes)
}
