package mongo

import (
	"context"

	"github.com/bondhan/golib/docstore"
	"github.com/bondhan/golib/util"
	"go.mongodb.org/mongo-driver/mongo"
)

type AtlasResults struct {
	cursor    *mongo.Cursor
	scorePath string
	count     int
}

func (r *AtlasResults) Count() int {
	if r.count >= 0 {
		return r.count
	}
	return r.cursor.RemainingBatchLength()
}

func (r *AtlasResults) Decode(ctx context.Context, docs interface{}) error {
	return r.cursor.All(ctx, docs)
}

func (r *AtlasResults) Next(ctx context.Context, doc interface{}) (float64, error) {
	var out map[string]interface{}
	if r.cursor.Next(ctx) {
		if err := r.cursor.Decode(&out); err != nil {
			return 0, err
		}

		if err := util.DecodeJSON(out, doc); err != nil {
			return 0, err
		}

		score, _ := util.LookupFloat(r.scorePath, out)
		if score < 0 {
			score = 0
		}
		return score, nil
	}

	return 0, docstore.EndOfDoc
}
