/*
odm.go

@Author: Harsley Ekhorutomwen
@Date: November 26, 2024


*/

package mongodb

import (
	"context"
	"fmt"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"reflect"
	"strings"
)

type (
	MongoDb[M DocumentModel] struct {
		model M
	}

	PipelineStages struct {
		Stages []bson.M
	}

	LookupConfig struct {
		FromCollection string
		LocalField     string
		ForeignField   string
		Alias          string
		Type           LookupOptions
	}
	MongoRepository[M DocumentModel] interface {
		Find(filter interface{}, options ...*FindOptions) (Cursor, error)
		FindOne(filter interface{}, options ...*FindOptions) (M, error)
		Save(data interface{}) (M, error)
		Update(filter bson.M, data interface{}) (M, error)
		Delete(filter bson.M) (interface{}, error)
		Count(filter bson.M) (int64, error)
	}
)

func (p *PipelineStages) SetPipelineSort(sort interface{}) {
	skipStage := bson.M{"$sort": sort}
	p.Stages = append(p.Stages, skipStage)
}

func (p *PipelineStages) SetPipelineSkip(skip int64) {
	limitStage := bson.M{"$skip": skip}
	p.Stages = append(p.Stages, limitStage)
}

func (p *PipelineStages) SetPipelineLimit(limit int64) {
	limitStage := bson.M{"$limit": limit}
	p.Stages = append(p.Stages, limitStage)
}

func NewMongoDb[M DocumentModel](model M) MongoDb[M] {
	return MongoDb[M]{model: model}
}

func (m MongoDb[M]) conn() *mongo.Collection {

	url := utils.GetEnv("MONGO_URL", "mongodb://localhost:27017")
	fmt.Println("url ", url)

	if url == "" {
		log.Fatal("MONGO_URL not found")
	}

	dbName := utils.GetEnv("DB_NAME", "test_numerisbook")

	if dbName == "" {
		log.Fatal("DB_NAME not found")
	}

	conn, err := NewMongoDatabase(url, dbName)

	if err != nil {
		log.Fatalln(err)
	}
	collectionName := m.model.GetModelName()
	return conn.Collection(collectionName)
}

func (m MongoDb[M]) Raw() *mongo.Collection {
	return m.conn()
}

func (m MongoDb[M]) findOneWithoutPrefetch(filter interface{}) (model M, err error) {

	coll := m.conn()

	err = coll.FindOne(context.Background(), filter).Decode(&model)

	if err != nil {
		return model, err
	}

	return model, err
}

func (m MongoDb[M]) BindModel(data interface{}) (M, error) {
	var model M
	err := utils.BindDataOperationStruct(data, &model)

	if err != nil {
		return model, err
	}

	return model, nil
}

func (m MongoDb[M]) findWithoutPrefetch(filter interface{}) (model Cursor, err error) {
	coll := m.conn()

	mongoCursor, err := coll.Find(context.Background(), filter)

	if err != nil {
		return Cursor{}, err
	}

	return Cursor{mongoCursor}, nil
}

func (m MongoDb[M]) Save(data interface{}) (model M, err error) {
	payload, err := utils.CleanMap(data, []string{"_id", "created_at", "updated_at"})

	if err != nil {
		return model, err
	}

	obj, err := m.BindModel(payload)

	if err != nil {
		return model, err
	}

	coll := m.conn()

	_, err = coll.InsertOne(context.Background(), obj)
	if err != nil {
		return model, err
	}

	return obj, nil
}

func (m MongoDb[M]) Count(filter bson.M) (count int64, err error) {
	coll := m.conn()

	count, err = coll.CountDocuments(context.Background(), filter)

	if err != nil {
		return count, err
	}
	return count, nil
}

func (m MongoDb[M]) findOneWithPrefetch(filter interface{}) (model M, err error) {
	coll := m.conn()

	pipeline, err := m.generatePipelineStages(filter)

	if err != nil {
		return model, err
	}

	opts := options.Aggregate().SetAllowDiskUse(true)

	mongoCursor, err := coll.Aggregate(context.Background(), pipeline.Stages, opts)
	if err != nil {
		return model, err
	}

	defer mongoCursor.Close(context.Background())

	// Decode the result into modelClass
	if mongoCursor.Next(context.Background()) {

		res := map[string]interface{}{}
		err = mongoCursor.Decode(&res)
		if err != nil {
			return model, err
		}
		model, err = m.BindModel(res)
		if err != nil {
			return model, err
		}
	}

	return model, nil
}

func (m MongoDb[M]) FindOne(filter interface{}, options ...*FindOptions) (model M, err error) {
	opts := MergeFindOption(options...)

	if opts.FetchAllLinks == nil {
		return m.findOneWithoutPrefetch(filter)
	}
	return m.findOneWithPrefetch(filter)
}

func (m MongoDb[M]) findWithPrefetch(filter interface{}, ops ...*FindOptions) (model Cursor, err error) {
	fmt.Println("here cool")
	coll := m.conn()

	pipeline, err := m.generatePipelineStages(filter)

	if len(ops) > 0 {
		opt := ops[0]
		if opt.Limit != nil {
			pageSize := *opt.Limit
			pipeline.SetPipelineLimit(pageSize)
		}
		if opt.Sort != nil {
			sort := opt.Sort
			pipeline.SetPipelineSort(sort)
		}
		if opt.Skip != nil {
			skip := *opt.Skip
			pipeline.SetPipelineSkip(skip)
		}
	}

	if err != nil {
		return model, err
	}

	opts := options.Aggregate().SetAllowDiskUse(true)

	fmt.Println("pipeline.Stages", pipeline.Stages)
	mongoCursor, err := coll.Aggregate(context.Background(), pipeline.Stages, opts)

	if err != nil {
		return model, err
	}

	fmt.Println("mongoCursor", mongoCursor)

	return Cursor{cursor: mongoCursor}, nil
}

func (m MongoDb[M]) Find(filter interface{}, options ...*FindOptions) (model Cursor, err error) {
	opts := MergeFindOption(options...)

	if opts.FetchAllLinks == nil {
		return m.findWithoutPrefetch(filter)
	}
	fmt.Println("alright cool i got here")
	return m.findWithPrefetch(filter, options...)
}

func generateLookupStages(modelType reflect.Type, parentAlias string, processedTypes map[string]struct{}, isNested bool) ([]bson.M, error) {

	var stages []bson.M

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		lookupTag := field.Tag.Get("lookup")
		if lookupTag == "" {
			continue
		}

		parts := strings.Split(lookupTag, ":")
		fmt.Println("parts", len(parts))
		if len(parts) < 3 || len(parts) > 5 {
			return nil, fmt.Errorf("invalid lookup tag format for field %s", field.Name)
		}

		fromCollection, localField, foreignField := parts[0], parts[1], parts[2]

		alias := parts[3]
		fmt.Println("parts parts", parts)

		if parentAlias != "" {
			alias = parentAlias + "." + alias
			localField = parentAlias + "." + localField
		}

		lookupStage := bson.M{
			"$lookup": bson.M{
				"from":         fromCollection,
				"localField":   localField,
				"foreignField": foreignField,
				"as":           alias,
			},
		}
		stages = append(stages, lookupStage)
		if len(parts) == 5 && (parts[4] == "struct" || parts[4] == "slice") {
			stages = append(stages, bson.M{
				"$unwind": bson.M{
					"path":                       "$" + alias,
					"preserveNullAndEmptyArrays": true,
				},
			})

			// Recursively handle nested structs
			if field.Type.Kind() == reflect.Struct ||
				(field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct) {
				nestedType := field.Type
				if nestedType.Kind() == reflect.Slice {
					nestedType = nestedType.Elem()
				}

				nestedStages, err := generateLookupStages(nestedType, alias, processedTypes, true)
				if err != nil {
					return nil, err
				}
				stages = append(stages, nestedStages...)
			}
		}
	}

	return stages, nil
}

func (m MongoDb[M]) generatePipelineStages(filter interface{}) (*PipelineStages, error) {

	var model M
	fmt.Println("olay wrighe", reflect.TypeOf(model))
	modelType := reflect.TypeOf(model)

	fmt.Println("olay wrighe", modelType.Kind())
	if modelType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model must be a struct type, got %v", modelType.Kind())
	}
	fmt.Println("aaldkkdkkd")
	pipeline := &PipelineStages{
		Stages: []bson.M{},
	}

	if filter != nil {
		pipeline.Stages = append(pipeline.Stages, bson.M{"$match": filter})
	}
	fmt.Println("errr herree to test i go here")
	processedTypes := make(map[string]struct{})
	lookupStages, err := generateLookupStages(modelType, "", processedTypes, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate lookup stages: %w", err)
	}
	pipeline.Stages = append(pipeline.Stages, lookupStages...)

	groupStage, err := generateGroupStage(modelType)
	if err != nil {
		return nil, fmt.Errorf("failed to generate group stage: %w", err)
	}
	pipeline.Stages = append(pipeline.Stages, groupStage)

	return pipeline, nil
}

func (m MongoDb[M]) Update(filter bson.M, data interface{}) (model M, err error) {
	coll := m.conn()
	res, err := m.FindOne(filter)

	if err != nil {
		return res, err
	}

	payload, err := utils.CleanMap(data, []string{"updated_at"})
	if err != nil {
		return model, err
	}
	option := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err = coll.FindOneAndUpdate(context.Background(), filter, bson.M{"$set": payload}, option).Decode(&model)

	if err != nil {
		return model, err
	}
	return model, nil
}

func (m MongoDb[M]) Delete(filter bson.M) (interface{}, error) {
	coll := m.conn()

	res, err := m.FindOne(filter)

	if err != nil {
		return res, err
	}
	_, err = coll.DeleteOne(context.Background(), filter)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"message": "deleted"}, nil
}

func (m MongoDb[M]) UpdateByID(idHex string, data interface{}) (M, error) {
	var model M

	objectID, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return model, err
	}

	return m.Update(bson.M{"_id": objectID}, data)
}

func generateGroupStage(modelType reflect.Type) (bson.M, error) {
	groupStage := bson.M{
		"$group": bson.M{
			"_id": "$_id",
		},
	}

	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		jsonTag := field.Tag.Get("json")

		if jsonTag != "" && jsonTag != "_id" {
			if field.Type.Kind() == reflect.Slice && field.Tag.Get("lookup") != "" {
				// Handle slice fields differently based on element type
				if field.Type.Elem().Kind() == reflect.Struct {
					groupStage["$group"].(bson.M)[jsonTag] = bson.M{"$push": "$" + jsonTag}
				} else {
					groupStage["$group"].(bson.M)[jsonTag] = bson.M{"$first": "$" + jsonTag}
				}
			} else {
				groupStage["$group"].(bson.M)[jsonTag] = bson.M{"$first": "$" + jsonTag}
			}
		}
	}

	return groupStage, nil
}
