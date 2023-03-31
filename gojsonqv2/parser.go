package gojsonq

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	DateTimeFormat      = "2006-01-02T15:04:05"
	DateTimeSpaceFormat = "2006-01-02 15:04:05"
	DateFormat          = "2006-01-02"
	delim               = ";"
	pipe                = "|>"
	and                 = "&&"
	or                  = "||"
	Select              = "s:"
	Where               = "w:"
	From                = "f:"
	Limit               = "l:"
	Offset              = "p:"
	SortBy              = "o:"
	Order               = "d:"
	Distinct            = "u:"
	GroupBy             = "g:"
	Time                = "t:"
	PipeSum             = "sum"
	PipeCount           = "count"
	PipeMax             = "max"
	PipeMin             = "min"
	PipeAvg             = "avg"
	PipeFirst           = "first"
	PipeLast            = "last"
	PipeNth             = "nth"
)

type Query struct {
	Select    []string
	From      string
	Where     [][]query
	Limit     int
	Offset    int
	SortBy    string
	Order     string
	Distinct  string
	Pipe      string
	PipeParam string
	GroupBy   string
}

func Parse(str string) *Query {
	pipes := strings.Split(str, pipe)

	q := &Query{}

	parts := strings.Split(pipes[0], delim)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch part[:2] {
		case Select:
			q.Select = strings.Split(strings.TrimSpace(strings.TrimPrefix(part, Select)), ",")
			for i, s := range q.Select {
				q.Select[i] = strings.TrimSpace(s)
			}
			continue
		case From:
			q.From = strings.TrimSpace(strings.TrimPrefix(part, From))
			continue
		case Distinct:
			q.Distinct = strings.TrimSpace(strings.TrimPrefix(part, Distinct))
			continue
		case Where:
			q.Where = parseWhere(strings.TrimSpace(strings.TrimPrefix(part, Where)))
			continue
		case Limit:
			l, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(part, Limit)))
			if err == nil {
				q.Limit = l
			}
			continue
		case Offset:
			o, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(part, Offset)))
			if err == nil {
				q.Offset = o
			}
			continue
		case SortBy:
			q.SortBy = strings.TrimSpace(strings.TrimPrefix(part, SortBy))
			continue
		case Order:
			q.Order = strings.TrimSpace(strings.TrimPrefix(part, Order))
			continue
		case GroupBy:
			q.GroupBy = strings.TrimSpace(strings.TrimPrefix(part, GroupBy))
			continue
		}
	}

	if len(pipes) > 1 {
		parts := strings.Split(pipes[1], ":")
		q.Pipe = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			q.PipeParam = strings.TrimSpace(parts[1])
		}
	}

	return q
}

func parseWhere(str string) [][]query {
	out := make([][]query, 0)
	ors := strings.Split(str, or)
	for _, o := range ors {
		q := make([]query, 0)
		ands := strings.Split(o, and)
		for _, a := range ands {
			parts := strings.Split(strings.TrimSpace(a), " ")
			if len(parts) != 3 {
				continue
			}
			q = append(q, query{
				key:      parts[0],
				operator: parts[1],
				value:    parseValue(parts[2]),
			})
		}
		out = append(out, q)
	}

	return out
}

func parseValue(str string) interface{} {
	str = strings.TrimSpace(str)
	i, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		return i
	}

	f, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return f
	}

	b, err := strconv.ParseBool(str)
	if err == nil {
		return b
	}

	if strings.HasPrefix(str, Time) {
		t, err := parseTime(strings.TrimPrefix(str, Time))
		if err == nil {
			return t
		}
	}

	if strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'") {
		return strings.Trim(str, "'")
	}

	if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
		return strings.Trim(str, `"`)
	}

	return str
}

func parseTime(t interface{}) (time.Time, error) {
	switch v := t.(type) {
	case string:
		if vi, err := strconv.ParseInt(v, 10, 64); err == nil {
			return getTime(vi), nil
		}
		switch len(v) {
		case len(DateFormat):
			return time.Parse(DateFormat, v)
		case len(DateTimeFormat):
			if strings.Contains(v, "T") {
				return time.Parse(DateTimeFormat, v)
			}
			return time.Parse(DateTimeSpaceFormat, v)
		default:
			return time.Parse(time.RFC3339, v)
		}
	case float64:
		return getTime(int64(v)), nil
	case int64:
		return getTime(v), nil
	case int:
		return getTime(int64(v)), nil
	default:
		return time.Time{}, fmt.Errorf("invalid time format %v", t)
	}
}

func getTime(t int64) time.Time {
	tstring := fmt.Sprintf("%v", t)

	//Nano second
	if len(tstring) > 18 {
		return time.Unix(t/1000000000, t%1000000000)
	}
	//Micro second
	if len(tstring) > 15 && len(tstring) <= 18 {
		return time.UnixMicro(t)
	}

	//Mili second
	if len(tstring) > 12 && len(tstring) <= 15 {
		return time.UnixMilli(t)
	}

	return time.Unix(t, 0)
}

func (q *Query) Chain(jq *JSONQ) *JSONQ {
	return q.prepare(jq.More())
}

func (q *Query) prepare(jq *JSONQ) *JSONQ {
	if q.Select != nil {
		jq = jq.Select(q.Select...)
	}

	if q.From != "" {
		jq = jq.From(q.From)
	}

	if q.Distinct != "" {
		jq = jq.Distinct(q.Distinct)
	}

	if q.Where != nil {
		jq.queries = q.Where
	}

	if q.Limit > 0 {
		jq = jq.Limit(q.Limit)
	}

	if q.Offset > 0 {
		jq = jq.Offset(q.Offset)
	}

	if q.Order != "" {
		if q.SortBy != "" {
			jq = jq.SortBy(q.SortBy, q.Order)
		} else {
			jq = jq.Sort(q.Order)
		}
	}

	if q.GroupBy != "" {
		jq = jq.GroupBy(q.GroupBy)
	}
	return jq
}

func (q *Query) Prepare(data interface{}) *JSONQ {
	jq := New().FromInterface(data)
	return q.prepare(jq)
}

func (q *Query) Get(data interface{}) interface{} {
	jq := q.Prepare(data)
	if q.Pipe != "" {
		return q.PipeResult(jq)
	}

	return jq.Get()
}

func (q *Query) PipeResult(jq *JSONQ) interface{} {
	switch q.Pipe {
	case PipeSum:
		if q.PipeParam == "" {
			return jq.Sum()
		}
		return jq.Sum(q.PipeParam)
	case PipeCount:
		return jq.Count()
	case PipeMax:
		if q.PipeParam == "" {
			return jq.Max()
		}
		return jq.Max(q.PipeParam)
	case PipeMin:
		if q.PipeParam == "" {
			return jq.Min()
		}
		return jq.Min(q.PipeParam)
	case PipeAvg:
		if q.PipeParam == "" {
			return jq.Avg()
		}
		return jq.Avg(q.PipeParam)
	case PipeFirst:
		return jq.First()
	case PipeLast:
		return jq.Last()
	case PipeNth:
		i, err := strconv.Atoi(q.PipeParam)
		if err != nil {
			return jq.Nth(0)
		}
		return jq.Nth(i)
	default:
		return jq.Get()
	}
}
