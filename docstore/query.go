package docstore

import "github.com/bondhan/golib/util"

// FilterOpt filter option
type FilterOpt struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
	Ops   string      `json:"op"`
}

// QueryOpt query option
type QueryOpt struct {
	Limit    int         `json:"limit"`
	Skip     int         `json:"skip"`
	Page     int         `json:"page"`
	OrderBy  string      `json:"order_by"`
	IsAscend bool        `json:"is_ascend"`
	Filter   []FilterOpt `json:"filter"`
}

func (q *QueryOpt) AddFilter(filter FilterOpt) *QueryOpt {
	for _, f := range q.Filter {
		if f.Field == filter.Field {
			return q
		}
	}

	q.Filter = append(q.Filter, filter)
	return q
}

func (q *QueryOpt) Hash() string {
	return util.Hash64(q)
}

type InsertManyOpt struct {
	// If true, writes executed as part of the operation will opt out of document-level validation on the server. This
	// option is valid for MongoDB versions >= 3.2 and is ignored for previous server versions. The default value is
	// false. See https://www.mongodb.com/docs/manual/core/schema-validation/ for more information about document
	// validation.
	BypassDocumentValidation *bool

	// A string or document that will be included in server logs, profiling logs, and currentOp queries to help trace
	// the operation.  The default value is nil, which means that no comment will be included in the logs.
	Comment interface{}

	// If true, no writes will be executed after one fails. The default value is true.
	Ordered *bool
}
