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
	chain               = "->"
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
	Each                = "e:"
	PipeSum             = "sum"
	PipeCount           = "count"
	PipeMax             = "max"
	PipeMin             = "min"
	PipeAvg             = "avg"
	PipeFirst           = "first"
	PipeLast            = "last"
	PipeNth             = "nth"
	PipeFind            = "find"
	PipePluck           = "pluck"
	PipeOnly            = "only"
)

type Query struct {
	Raw              string
	Select           []string
	SelectMapAsArray bool
	Each             string
	From             string
	Where            [][]query
	Limit            int
	Offset           int
	SortBy           string
	Order            string
	Distinct         string
	Pipe             string
	PipeParam        string
	GroupBy          string
	Next             *Query
}

func Parse(str string) *Query {
	pipes := strings.Split(str, pipe)

	queries := strings.Split(pipes[0], chain)

	q := toQuery(strings.TrimSpace(queries[0]))

	if len(queries) > 1 {
		qn := strings.Join(queries[1:], chain)
		q.Next = Parse(qn)
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

func toQuery(queryStr string) *Query {
	parts := strings.Split(queryStr, delim)
	q := &Query{Raw: queryStr}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) < 2 {
			continue
		}
		switch part[:2] {
		case Select:
			if q.Select == nil {
				q.Select = make([]string, 0)
			}
			slc := strings.Split(strings.TrimSpace(strings.TrimPrefix(part, Select)), ",")
			for _, s := range slc {
				if strings.TrimSpace(s) == "-" {
					q.SelectMapAsArray = true
					continue
				}
				q.Select = append(q.Select, strings.TrimSpace(s))
			}
			continue
		case Each:
			q.Each = strings.TrimSpace(strings.TrimPrefix(part, Each))
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
			if len(parts) < 3 {
				continue
			}
			qe := query{
				key:      parts[0],
				operator: parts[1],
				value:    parseValue(strings.Join(parts[2:], " ")),
			}
			q = append(q, qe)

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

	if q.Each != "" {
		ctx := jq.jsonContent
		if q.Each != "." {
			ctx = jq.Find(q.Each)
		}

		var inputs []interface{}
		switch v := ctx.(type) {
		case []interface{}:
			inputs = v
		case map[string]interface{}:
			for _, m := range v {
				inputs = append(inputs, m)
			}
		default:
			fmt.Println("not map or slice")
		}
		q.Each = ""
		res := make([]interface{}, 0)
		for _, input := range inputs {
			if len(q.Where) > 0 && q.From == "" {
				if m, ok := input.(map[string]interface{}); ok {
					input = []interface{}{m}
				}
			}
			r := q.prepare(New().FromInterface(input)).Get()
			if r != nil {
				switch v := r.(type) {
				case map[string]interface{}:
					if len(v) == 0 {
						continue
					}
					if q.SelectMapAsArray {
						for _, m := range v {
							res = append(res, m)
						}
						continue
					}
					res = append(res, v)
				case []interface{}:
					if len(v) > 0 {
						res = append(res, v...)
					}
				default:
					res = append(res, v)
				}
			}
		}
		if len(res) > 0 {
			jq = New().FromInterface(res)
		}

		return jq
	}

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
	jq := q.prepare(New().FromInterface(data))
	if q.Next != nil {
		tmp := jq.Get()
		switch v := tmp.(type) {
		case map[string]interface{}:
			if q.SelectMapAsArray {
				out := make([]interface{}, 0, len(v))
				for _, vv := range v {
					out = append(out, vv)
				}
				tmp = out
			} else {
				if len(q.Next.Where) > 0 && q.Next.From == "" {
					tmp = []interface{}{v}
				}
			}

		}

		//jq = q.Next.prepare(New().FromInterface(tmp))
		jq = q.Next.Prepare(tmp)
	}
	return jq
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
		out := jq.First()
		if q.PipeParam != "" {
			return New().FromInterface(out).Find(q.PipeParam)
		}
		return out
	case PipeLast:
		out := jq.Last()
		if q.PipeParam != "" {
			return New().FromInterface(out).Find(q.PipeParam)
		}
		return out
	case PipeNth:
		params := strings.Split(q.PipeParam, ",")
		i, _ := strconv.Atoi(strings.TrimSpace(params[0]))
		out := jq.Nth(i)

		if len(params) > 1 {
			return New().FromInterface(out).Find(strings.TrimSpace(params[1]))
		}
		return out
	case PipeFind:
		out := jq.Get()
		if q.PipeParam != "" {
			return New().FromInterface(out).Find(q.PipeParam)
		}
		return out
	case PipePluck:
		return jq.Pluck(q.PipeParam)
	case PipeOnly:
		return jq.Only(strings.Split(q.PipeParam, ",")...)
	default:
		return jq.Get()
	}
}
