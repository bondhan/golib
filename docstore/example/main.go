package main

import (
	"context"
	"fmt"

	"github.com/bondhan/golib/cache"
	_ "github.com/bondhan/golib/cache/mem"
	"github.com/bondhan/golib/constant"
	"github.com/bondhan/golib/docstore"
	"github.com/bondhan/golib/docstore/firestore"
	_ "github.com/bondhan/golib/docstore/mongo"
)

func main() {
	search()
}

func findOne() {
	config := &docstore.Config{
		Database:   "peak-nimbus-307910",
		Collection: "s-greenseer-areaRegion",
		IDField:    "id",
		Driver:     "firestore",
		CacheURL:   "mem://ms",
		Credential: "google.json",
		Connection: "",
	}
	ctx := context.Background()

	fs, err := firestore.NewFireStore(config)
	if err != nil {
		panic(err)
	}

	cache, _ := cache.New("mem://")
	cs := docstore.NewDocstore(fs, cache, config)

	query := &docstore.QueryOpt{
		Filter: []docstore.FilterOpt{
			{Field: "id", Ops: constant.EQ, Value: "h1ExOl6PnkTy4OHvyxh6"},
		},
	}

	ok, err := cs.IsExists(ctx, query)
	fmt.Println(ok)
	fmt.Println(err)

}

func memQuery() {
	config := &docstore.Config{
		Database:   "goretailer-infra-prod",
		Collection: "p-stormsend-skuDiscount",
		IDField:    "id",
		Driver:     "firestore",
		CacheURL:   "mem://ms",
		Credential: "google_prod_cred.json",
		Connection: "",
	}
	ctx := context.Background()

	fs, err := firestore.NewFireStore(config)
	if err != nil {
		panic(err)
	}

	out := make([]map[string]interface{}, 0)

	if err := fs.Find(ctx, nil, &out); err != nil {
		panic(err)
	}

	ms := docstore.NewMemoryStore("test", "id")
	in := make([]interface{}, 0)
	for _, o := range out {
		in = append(in, o)
	}
	// opt := &docstore.InsertManyOpt{}
	//fmt.Println(in)
	if err := ms.BulkCreate(ctx, in, nil); err != nil {
		panic(err)
	}

	query := &docstore.QueryOpt{
		Filter: []docstore.FilterOpt{
			{Field: "ruleKeyword.areaWarehouse", Ops: constant.IN, Value: []string{"bitung", "[all]"}},
			{Field: "ruleKeyword.locationCity", Ops: constant.IN, Value: []string{"tangerang", "[all]"}},
			{Field: "ruleKeyword.product", Ops: constant.IN, Value: []string{"1B04BAD12350"}},
		},
	}

	res := make([]map[string]interface{}, 0)
	ms.Find(ctx, query, &res)
	/*
		for _, r := range res {
			fmt.Println(r["validProducts"])
		}
	*/

	fmt.Println(len(res))
}

func fireQuery() {
	config := &docstore.Config{
		Database:   "goretailer-infra-prod",
		Collection: "p-stormsend-skuDiscount",
		IDField:    "id",
		Driver:     "firestore",
		CacheURL:   "mem://ms",
		Credential: "google_prod_cred.json",
		Connection: "",
	}
	ctx := context.Background()

	fs, err := firestore.NewFireStore(config)
	if err != nil {
		panic(err)
	}

	query := &docstore.QueryOpt{
		Filter: []docstore.FilterOpt{
			{Field: "ruleKeyword.product", Ops: constant.AM, Value: []string{"1B04BAD12350"}},
			{Field: "ruleKeyword.areaWarehouse", Ops: constant.AM, Value: []string{"bitung", "[all]"}},
			{Field: "ruleKeyword.locationCity", Ops: constant.AM, Value: []string{"tangerang", "[all]"}},
		},
	}

	out := make([]map[string]interface{}, 0)

	if err := fs.Find(ctx, query, &out); err != nil {
		panic(err)
	}

	fmt.Println(len(out))
}

func search() {
	/*
		config := &docstore.Config{
			Database:   "peak-nimbus-307910",
			Collection: "s-citadel-inventory",
			IDField:    "id",
			Driver:     "firestore",
			CacheURL:   "mem://ms",
			Credential: "google.json",
			Connection: "",
		}
	*/

	config := &docstore.Config{
		Database:   "test",
		Collection: "inventory",
		IDField:    "id",
		Driver:     "mongo",
		CacheURL:   "mem://ms",
		Credential: "google.json",
		Connection: map[string]interface{}{
			"uri":  "mongodb://localhost:27017",
			"name": "default",
		},
	}

	ctx := context.Background()

	ds, err := docstore.New(config)
	if err != nil {
		panic(err)
	}
	query := &docstore.QueryOpt{
		Filter: []docstore.FilterOpt{
			{Field: "warehouse.name", Ops: constant.EQ, Value: "karawaci"},
			{Field: "product.productName", Ops: constant.RE, Value: "Black"},
		},
	}

	out := make([]map[string]interface{}, 0)

	if err := ds.Find(ctx, query, &out); err != nil {
		panic(err)
	}

	fmt.Println(len(out))
}
