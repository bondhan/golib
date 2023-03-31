package mongo

import (
	"context"
	"fmt"
	"testing"

	"github.com/bondhan/golib/client"
	"github.com/bondhan/golib/constant"
	"github.com/bondhan/golib/docstore"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestAtlas(t *testing.T) {
	t.Skip()
	cfg := &AtlasSearch{
		Database:   "s-product",
		Collection: "product",
		IndexName:  "default",
		Indexes: map[string]string{
			"productName": "fuzzy",
			"description": "fuzzy",
		},
		Connection: client.MongoClient{
			URI:     "mongodb+srv://root:9g2w9HXVo7b75z18@staging-pri.ycvc3.mongodb.net/?retryWrites=true&w=majority",
			AppName: "default",
		},
	}

	keyword := "shampoo"

	as, err := NewAtlasSearch(cfg)
	assert.Nil(t, err)
	assert.NotNil(t, as)

	opt := &docstore.QueryOpt{
		Filter: []docstore.FilterOpt{
			{Field: "isActive", Ops: constant.EQ, Value: true},
		},
		Limit: 10,
	}

	out, err := as.Search(context.Background(), keyword, opt)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	//assert.Equal(t, 63, out.Count())
	res := make([]interface{}, 0)
	out.Decode(context.Background(), &res)
	//fmt.Println(res)

	count, err := as.Count(context.Background(), keyword, opt)
	assert.Nil(t, err)
	assert.Equal(t, out.Count(), count)
	//assert.True(t, false)
	/*
		var doc map[string]interface{}
		score, err := out.Next(context.Background(), &doc)
		assert.Nil(t, err)
		assert.NotEqual(t, 0, score)
	*/
}

func TestAggregate(t *testing.T) {
	//t.Skip()
	cfg := &AtlasSearch{
		Database:   "s-product",
		Collection: "product",
		IndexName:  "default",
		Indexes: map[string]string{
			"address.detail": "autocomplete",
			"name":           "autocomplete",
		},
		Connection: client.MongoClient{
			URI:     "mongodb+srv://root:password@staging-pri.ycvc3.mongodb.net/test",
			AppName: "default",
		},
	}

	pipeline := cfg.buildAggregate("shampoo", &docstore.QueryOpt{Limit: 10, Filter: []docstore.FilterOpt{{Field: "user", Ops: constant.NE, Value: nil}}}, false)
	//fmt.Println(pipeline)
	pipe := []bson.D(*pipeline)
	b, _ := bson.MarshalExtJSON(pipe[0], false, false)
	fmt.Println(string(b))
	//t.Error()
}
