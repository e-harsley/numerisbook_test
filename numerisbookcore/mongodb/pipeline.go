package mongodb

import "go.mongodb.org/mongo-driver/bson"

type Pipeline struct {
	pipeline []bson.M
}

func (p *Pipeline) SetPipelineSort(sort interface{}) {
	skipStage := bson.M{"$sort": sort}
	p.pipeline = append(p.pipeline, skipStage)
}

func (p *Pipeline) SetPipelineSkip(skip int64) {
	limitStage := bson.M{"$skip": skip}
	p.pipeline = append(p.pipeline, limitStage)
}

func (p *Pipeline) SetPipelineLimit(limit int64) {
	limitStage := bson.M{"$limit": limit}
	p.pipeline = append(p.pipeline, limitStage)
}
