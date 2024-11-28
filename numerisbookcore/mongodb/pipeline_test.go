package mongodb

import (
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestPipeline(t *testing.T) {
	pipeline := &Pipeline{}

	// Test SetPipelineSort
	pipeline.SetPipelineSort(bson.M{"name": 1})
	assert.Len(t, pipeline.pipeline, 1)
	assert.Contains(t, pipeline.pipeline[0], "$sort")

	// Test SetPipelineSkip
	pipeline.SetPipelineSkip(5)
	assert.Len(t, pipeline.pipeline, 2)
	assert.Contains(t, pipeline.pipeline[1], "$skip")

	// Test SetPipelineLimit
	pipeline.SetPipelineLimit(10)
	assert.Len(t, pipeline.pipeline, 3)
	assert.Contains(t, pipeline.pipeline[2], "$limit")
}
