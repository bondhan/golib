package gojsonq

import (
	"io/ioutil"
	"testing"
)

func TestParser(t *testing.T) {
	query := "s: id, name, age; w: id == 1 && name == 'john'; f: users; l: 10; p: 1; o: id; d: asc; g:name; u:name|>sum: age"

	q := Parse(query)
	if q == nil {
		t.Error("failed to parse query")
		return
	}

	if q.Distinct != "name" {
		t.Error("failed to parse distinct")
	}

	if q.Pipe != "sum" {
		t.Error("failed to parse pipe")
	}

	if q.PipeParam != "age" {
		t.Error("failed to parse pipe param")
	}

	if len(q.Select) != 3 {
		t.Error("failed to parse select")
	}

	if q.Select[0] != "id" {
		t.Error("failed to parse select")
	}

	if q.Select[1] != "name" {
		t.Error("failed to parse select")
	}

	if q.Select[2] != "age" {
		t.Error("failed to parse select")
	}

	if len(q.Where) != 1 {
		t.Error("failed to parse where")
	}

	if len(q.Where[0]) != 2 {
		t.Error("failed to parse where")
	}

	if q.Where[0][0].key != "id" {
		t.Error("failed to parse where")
	}

	if q.Where[0][0].operator != "==" {
		t.Error("failed to parse where")
	}

	if q.Where[0][0].value != int64(1) {
		t.Error("failed to parse where ", q.Where[0][0].value)
	}

	if q.Where[0][1].key != "name" {
		t.Error("failed to parse where")
	}

	if q.Where[0][1].operator != "==" {
		t.Error("failed to parse where")
	}

	if q.Where[0][1].value != "john" {
		t.Error("failed to parse where")
	}

	if q.From != "users" {
		t.Error("failed to parse from")
	}

	if q.Limit != 10 {
		t.Error("failed to parse limit")
	}

	if q.Offset != 1 {
		t.Error("failed to parse offset")
	}

	if q.SortBy != "id" {
		t.Error("failed to parse sort by")
	}

	if q.Order != "asc" {
		t.Error("failed to parse order")
	}

	if q.GroupBy != "name" {
		t.Error("failed to parse group by")
	}

}

func TestOrParser(t *testing.T) {
	query := "s: id, name, age; w: id == 1 || name == 'john'; f: users; l: 10; p: 1; o: id; d: asc; distinct|>sum: age"
	q := Parse(query)
	if q == nil {
		t.Error("failed to parse query")
		return
	}

	if len(q.Where) != 2 {
		t.Error("failed to parse where", q.Where)
	}

	if len(q.Where[0]) != 1 {
		t.Error("failed to parse where")
	}

	if len(q.Where[1]) != 1 {
		t.Error("failed to parse where")
	}
}

func TestQueryParser(t *testing.T) {
	json := `{
		"name":"computers",
		"description":"List of computer products",
		"prices":[2400, 2100, 1200, 400.87, 89.90, 150.10],
		"names":["John Doe", "Jane Doe", "Tom", "Jerry", "Nicolas", "Abby"],
		"items":[
		   {
			  "id":1,
			  "name":"MacBook Pro 13 inch retina",
			  "price":1350
		   },
		   {
			  "id":2,
			  "name":"MacBook Pro 15 inch retina",
			  "price":1700
		   },
		   {
			  "id":3,
			  "name":"Sony VAIO",
			  "price":1200
		   },
		   {
			  "id":4,
			  "name":"Fujitsu",
			  "price":850
		   },
		   {
			  "id":null,
			  "name":"HP core i3 SSD",
			  "price":850
		   }
		]
	 }`

	query := "f: prices |> sum"
	res := Parse(query).Get(json)

	if res != 6340.87 {
		t.Error("failed to execute query", res)
	}

	res2 := Parse("s: price ; f: items |> sum : price").Get(json)
	if res2 != float64(5950) {
		t.Error("failed to execute query", res2)
	}

	res3 := Parse("s: price ; f: items ; w: name contains Macbook || name = 'Fujitsu' |> sum : price").Get(json)
	if res3 != float64(3900) {
		t.Error("failed to execute query", res3)
	}
}

func TestQueryFile(t *testing.T) {

	str, err := ioutil.ReadFile("order-test.json")
	if err != nil {
		t.Error("failed to read file", err)
		return
	}

	query := "e: itemsBreakdown.items; f:items ;s:- -> w: skuNo = 1A04AAU12352 |> sum:qty"
	res := Parse(query).Get(string(str))
	if res.(float64) != 36 {
		t.Error("failed to execute query", res)
	}

	query2 := "f:items; w: product.brand.name = Sun Kara; s:qty.ordered |> sum:ordered "
	q := Parse(query2)
	res2 := q.Get(string(str))
	if res2.(float64) != 36 {
		t.Error("failed to execute query", res2)
	}

	query3 := "e: itemsBreakdown.items; w: discountType = 0 -> e:. ;f:items ; s:- -> w: skuNo = 1A04AAU12352 |> sum:qty"
	res3 := Parse(query3).Get(string(str))
	if res3.(float64) != 18 {
		t.Error("failed to execute query", res)
	}

}
