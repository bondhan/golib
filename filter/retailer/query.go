package retailer

import (
	"time"

	"github.com/bondhan/golib/constant"
	"github.com/bondhan/golib/docstore"
	domain "github.com/bondhan/golib/domain/retailer"
)

type QueryBuilder struct {
	noFilterSymbol string
	retailer       *domain.RetailerContext
	query          *docstore.QueryOpt
}

func NewQueryBuilder(retailer *domain.RetailerContext, noFilterSymbol string) *QueryBuilder {
	return &QueryBuilder{
		retailer:       retailer,
		noFilterSymbol: noFilterSymbol,
		query:          &docstore.QueryOpt{Filter: []docstore.FilterOpt{}},
	}
}

func (q *QueryBuilder) Get() *docstore.QueryOpt {
	return q.query
}

func (q *QueryBuilder) WithProvince(provinceField string) *QueryBuilder {
	val := []string{q.noFilterSymbol, q.retailer.Address.Province}
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: provinceField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithDistrict(districtField string) *QueryBuilder {
	val := []string{q.noFilterSymbol, q.retailer.Address.District}
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: districtField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithDistrictSub(districtSubField string) *QueryBuilder {
	val := []string{q.noFilterSymbol, q.retailer.Address.DistrictSub}
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: districtSubField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithCity(cityField string) *QueryBuilder {
	val := []string{q.noFilterSymbol, q.retailer.Address.City}
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: cityField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithZip(zipField string) *QueryBuilder {
	val := []string{q.noFilterSymbol, q.retailer.Address.ZipCode}
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: zipField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithRegion(regionField string) *QueryBuilder {
	val := []string{q.noFilterSymbol, q.retailer.Area.Region}
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: regionField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithSubRegion(subRegionField string) *QueryBuilder {
	val := []string{q.noFilterSymbol, q.retailer.Area.RegionSub}
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: subRegionField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithCluster(clusterField string) *QueryBuilder {
	val := []string{q.noFilterSymbol, q.retailer.Area.Cluster}
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: clusterField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithWarehouse(warehouseField string) *QueryBuilder {
	val := []string{q.noFilterSymbol, q.retailer.Area.Warehouse}
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: warehouseField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithSegment(segmentField string) *QueryBuilder {
	val := q.retailer.Segments
	if val == nil {
		val = []string{}
	}
	val = append(val, q.noFilterSymbol)
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: segmentField, Ops: constant.IN, Value: val})
	return q
}

func (q *QueryBuilder) WithActiveTimes(activeTimeField, startField, endField string) *QueryBuilder {
	val := []docstore.FilterOpt{
		{Field: startField, Ops: constant.LE, Value: time.Now()},
		{Field: endField, Ops: constant.GE, Value: time.Now()},
	}

	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: activeTimeField, Ops: constant.EM, Value: val})
	return q
}

func (q *QueryBuilder) WithActiveUnixTimes(activeTimeField, startField, endField string) *QueryBuilder {
	val := []docstore.FilterOpt{
		{Field: startField, Ops: constant.LE, Value: time.Now().Unix()},
		{Field: endField, Ops: constant.GE, Value: time.Now().Unix()},
	}

	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: activeTimeField, Ops: constant.EM, Value: val})
	return q
}

func (q *QueryBuilder) AddFilter(field string, ops string, value interface{}) *QueryBuilder {
	q.query.Filter = append(q.query.Filter, docstore.FilterOpt{Field: field, Ops: ops, Value: value})
	return q
}

func (q *QueryBuilder) Sort(fieldName string, isAscend bool) *QueryBuilder {
	q.query.OrderBy = fieldName
	q.query.IsAscend = isAscend
	return q
}

func (q *QueryBuilder) Limit(limit int) *QueryBuilder {
	q.query.Limit = limit
	return q
}

func (q *QueryBuilder) Skip(skip int) *QueryBuilder {
	q.query.Skip = skip
	return q
}
