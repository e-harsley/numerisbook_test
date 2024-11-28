package mongodb

import (
	"context"
	"fmt"
	"github.com/e-harsley/numerisbook_test/numerisbookcore/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	LookupOptions string

	DocumentModel interface {
		GetModelName() string
	}

	Model[M DocumentModel] interface {
	}

	Cursor struct {
		cursor *mongo.Cursor
	}

	FindOptions struct {
		FetchAllLinks *bool
		Limit         *int64
		Skip          *int64
		Sort          bson.D
	}
)

const (
	STRUCTLOOKUP LookupOptions = "struct"
	SLICELOOKUP  LookupOptions = "slice"
)

func Find() *FindOptions {
	return &FindOptions{}
}

func (f *FindOptions) SetFetchAllLinks(b bool) *FindOptions {
	f.FetchAllLinks = &b
	return f
}

func (f *FindOptions) SetLimit(i int64) *FindOptions {
	f.Limit = &i
	return f
}

func (f *FindOptions) SetSkip(i int64) *FindOptions {
	f.Skip = &i
	return f
}

func (f *FindOptions) SetSort(i bson.D) *FindOptions {
	f.Sort = i
	return f
}

func MergeFindOption(opts ...*FindOptions) *FindOptions {
	fo := Find()
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if opt.FetchAllLinks != nil {
			fo.FetchAllLinks = opt.FetchAllLinks
		}
		if opt.Limit != nil {
			fo.Limit = opt.Limit
		}
		if opt.Skip != nil {
			fo.Skip = opt.Skip
		}
		if opt.Sort != nil {
			fo.Sort = opt.Sort
		}
	}
	return fo
}

func (c *Cursor) ToSlice(payload interface{}) error {
	ctx := context.Background()
	res := []map[string]interface{}{}
	err := c.cursor.All(ctx, &res)
	if err != nil {
		return err
	}
	fmt.Println(res)
	err = utils.BindDataOperationStruct(res, payload)
	if err != nil {
		return err
	}
	return nil
}
