package firestore

import (
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/bondhan/golib/constant"
	"github.com/bondhan/golib/docstore"
	"github.com/bondhan/golib/util"
)

func getFireQuery(q *docstore.QueryOpt, query firestore.Query) (*docstore.QueryOpt, firestore.Query, error) {
	if q == nil {
		return nil, query, nil
	}

	if !hasComplexQuery(q) {
		return nil, toFirestoreQuery(q, query), nil
	}

	qs := &docstore.QueryOpt{
		Limit: q.Limit,
		Skip:  q.Skip,
		Page:  q.Page,
	}
	if q.Page > 0 && q.Limit > 0 {
		qs.Skip = q.Page * q.Limit
	}
	qp := &docstore.QueryOpt{
		OrderBy:  q.OrderBy,
		IsAscend: q.IsAscend,
	}

	qs.Filter = q.Filter

	for i, f := range q.Filter {
		if f.Ops == constant.OR {
			return nil, query, fmt.Errorf("unsupported filter operation: %v", f.Ops)
		}
		if f.Ops == constant.IN || f.Ops == constant.AM || f.Ops == constant.AIN {
			if i == 0 {
				qp.Filter = q.Filter[0:1]
				qs.Filter = q.Filter[1:]
				break
			}
			qp.Filter = q.Filter[0:i]
			qs.Filter = q.Filter[i:]
			break
		}
	}

	return qs, toFirestoreQuery(qp, query), nil
}

func hasComplexQuery(q *docstore.QueryOpt) bool {
	if len(q.Filter) == 1 {
		return false
	}

	for _, f := range q.Filter {
		if f.Ops == constant.IN || f.Ops == constant.AM || f.Ops == constant.AIN ||
			f.Ops == constant.RE || f.Ops == constant.OR {
			return true
		}
	}
	return false
}

func isMatch(doc interface{}, q *docstore.QueryOpt) bool {
	match := false
	for _, f := range q.Filter {
		if f.Ops == constant.AIN || f.Ops == constant.AM {
			f.Ops = constant.IN
		}
		if !util.Assert(f.Field, doc, f.Value, f.Ops) {
			match = false
			break
		}
		match = true
	}
	return match
}

func toFirestoreQuery(q *docstore.QueryOpt, query firestore.Query) firestore.Query {
	if q == nil {
		return query
	}
	//fmt.Println(q)
	for _, f := range q.Filter {
		if f.Ops == constant.EQ {
			f.Ops = "=="
		}

		if f.Ops == constant.IN {
			f.Ops = "in"
		}

		if f.Ops == constant.AIN {
			f.Ops = "array-contains"
		}

		if f.Ops == constant.AM {
			f.Ops = "array-contains-any"
		}

		if f.Ops == constant.RE || f.Ops == constant.EM {
			continue
		}

		query = query.Where(f.Field, f.Ops, f.Value)
	}

	if q.Limit > 0 {
		query = query.Limit(q.Limit)
		if q.Page > 0 {
			q.Skip = q.Page * q.Limit
		}
	}

	if q.OrderBy != "" {
		dir := firestore.Desc
		if q.IsAscend {
			dir = firestore.Asc
		}
		query = query.OrderBy(q.OrderBy, dir)
	}

	if q.Skip > 0 {
		query = query.Offset(q.Skip)
	}

	return query
}
