package mongodb

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestFindOptions_SetFetchAllLinks(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"SetFetchAllLinks True", true},
		{"SetFetchAllLinks False", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Find()
			opts.SetFetchAllLinks(tt.value)
			assert.Equal(t, &tt.value, opts.FetchAllLinks)
		})
	}
}

func TestFindOptions_SetLimit(t *testing.T) {
	opts := Find()
	opts.SetLimit(10)
	assert.Equal(t, int64(10), *opts.Limit)
}

func TestFindOptions_SetSkip(t *testing.T) {
	opts := Find()
	opts.SetSkip(5)
	assert.Equal(t, int64(5), *opts.Skip)
}

func TestFindOptions_SetSort(t *testing.T) {
	opts := Find()
	opts.SetSort(bson.D{
		{Key: "created_at", Value: -1},
	})
	assert.Equal(t, bson.D{
		{Key: "created_at", Value: -1},
	}, opts.Sort)
}
