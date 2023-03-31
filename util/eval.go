package util

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	ev "github.com/Knetic/govaluate"
	"github.com/spatial-go/geoos/geoencoding"
	"github.com/spatial-go/geoos/space"
	"github.com/tidwall/gjson"
	"github.com/wI2L/jsondiff"

	"github.com/bondhan/golib/gojsonqv2/v2"
)

// Evaluator value evaluator
type Evaluator struct {
	exp    *ev.EvaluableExpression
	params map[string]string
}

// NewEvaluator create new evaluator instance
func NewEvaluator(exp string) (*Evaluator, error) {
	exp = strings.TrimPrefix(exp, "eval:")
	params := make(map[string]string)
	re, err := regexp.Compile(`\$[^\s,)]*`)
	if err != nil {
		return nil, err
	}

	match := re.FindAllString(exp, -1)
	if len(match) > 0 {
		for _, m := range match {
			k := strings.TrimPrefix(m, "$")
			if k == "." {
				params["ctx"] = "."
				exp = strings.ReplaceAll(exp, m, "ctx")
				continue
			}
			n := strings.ReplaceAll(k, ".", "_")
			n = strings.ReplaceAll(n, "-", "_")
			n = strings.ReplaceAll(n, "$", "")
			params[n] = k
			exp = strings.Replace(exp, m, n, 1)
		}
	}

	fns := map[string]ev.ExpressionFunction{
		"match": func(args ...interface{}) (interface{}, error) {
			if len(args) != 3 {
				return nil, errors.New("invalid argument length")
			}
			n, ok := args[0].(string)
			if !ok {
				return nil, errors.New("invalid argument type")
			}
			return Match(n, args[1], args[2]), nil
		},
		"regex": func(args ...interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, errors.New("invalid argument length")
			}
			p, ok := args[0].(string)
			if !ok {
				return nil, errors.New("invalid pattern type")
			}
			val, ok := args[1].(string)
			if !ok {
				return nil, errors.New("invalid argument type")
			}
			return regexp.MatchString(p, val)
		},
		"now": func(args ...interface{}) (interface{}, error) {
			t := time.Now()
			if len(args) > 0 {
				d, err := strconv.Atoi(fmt.Sprintf("%v", args[0]))
				if err == nil {
					t = t.Add(time.Duration(d) * time.Second)
				}
			}

			if len(args) > 1 {
				opt := fmt.Sprintf("%v", args[1])
				switch strings.ToUpper(opt) {
				case "NANO":
					return t.UnixNano(), nil
				default:
					return t.Unix(), nil
				}
			}

			return t, nil
		},
		"inow": func(args ...interface{}) (interface{}, error) {
			return time.Now().Unix(), nil
		},
		"null": func(args ...interface{}) (interface{}, error) {
			return nil, nil
		},
		"hash": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}
			h := sha256.Sum256([]byte(fmt.Sprintf("%v", args[0])))
			return base64.StdEncoding.EncodeToString(h[:]), nil
		},
		"int": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}
			res, err := strconv.Atoi(fmt.Sprintf("%v", args[0]))
			if err != nil {
				if len(args) > 1 {
					return strconv.Atoi(fmt.Sprintf("%v", args[1]))
				}
				return int64(0), err
			}
			return res, nil
		},
		"float": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}

			res, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
			if err != nil {
				if len(args) > 1 {
					return strconv.ParseFloat(fmt.Sprintf("%v", args[1]), 64)
				}
				return float64(0), err
			}
			return res, nil
		},
		"bool": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}

			res, err := strconv.ParseBool(fmt.Sprintf("%v", args[0]))
			if err != nil {
				if len(args) > 1 {
					return strconv.ParseBool(fmt.Sprintf("%v", args[1]))
				}
				return false, err
			}
			return res, nil
		},
		"multibool": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}

			for _, arg := range args {
				b, err := strconv.ParseBool(fmt.Sprintf("%v", arg))
				if err != nil {
					return nil, err
				}
				if !b {
					return false, nil
				}
			}

			return true, nil
		},
		"todate": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}

			layout := time.RFC3339
			if len(args) > 1 {
				layout = fmt.Sprintf("%v", args[1])
			}

			ts := fmt.Sprintf("%v", args[0])

			t, err := time.Parse(layout, ts)
			if err != nil {
				return nil, err
			}

			return t, nil
		},
		"date": func(args ...interface{}) (interface{}, error) {
			f := "2006/01/02"
			if len(args) > 0 {
				switch args[0].(type) {
				case time.Time:
					t := args[0].(time.Time)
					return t.Format(f), nil
				default:
					ts := fmt.Sprintf("%v", args[0])

					layout := time.RFC3339
					if len(args) > 1 {
						layout = fmt.Sprintf("%v", args[1])
					}

					t, err := time.Parse(layout, ts)
					if err != nil {
						return nil, err
					}
					return t.Format(f), nil
				}
			}

			return time.Now().Format(f), nil
		},
		"ftime": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, errors.New("not enough parameter length")
			}

			lout := strings.TrimPrefix(fmt.Sprintf("%v", args[1]), "T")
			fmt.Println(args, lout)
			switch t := args[0].(type) {
			case time.Time:
				return t.Format(lout), nil
			case string:
				ts := fmt.Sprintf("%v", t)
				lin := time.RFC3339
				if len(args) > 2 {
					lin = strings.TrimPrefix(fmt.Sprintf("%v", args[2]), "T")
				}

				to, err := time.Parse(lin, ts)
				if err != nil {
					return nil, err
				}
				return to.Format(lout), nil
			case int64:
				to := time.Unix(t, 0)
				return to.Format(lout), nil
			case float64:
				to := time.Unix(int64(t), 0)
				return to.Format(lout), nil
			default:
				return nil, errors.New("unsupported type")
			}
		},
		"len": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}

			if !IsSliceOrPointerOfSlice(args[0]) {
				return len(args), nil
			}
			rv := reflect.ValueOf(args[0])
			return rv.Len(), nil
		},
		"env": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}

			return os.Getenv(strings.ToUpper(fmt.Sprintf("%v", args[0]))), nil
		},
		"upper": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}
			return strings.ToUpper(fmt.Sprintf("%v", args[0])), nil
		},
		"lower": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}
			return strings.ToLower(fmt.Sprintf("%v", args[0])), nil
		},
		"time": func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errors.New("not enough parameter length")
			}

			ti, err := strconv.ParseFloat(fmt.Sprintf("%v", args[0]), 64)
			if err != nil {
				return nil, err
			}

			return normalizeTime(int64(ti)), nil
		},
		"diff": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, errors.New("parameter length not match")
			}

			b1, err := json.Marshal(args[0])
			if err != nil {
				return nil, fmt.Errorf("marshal first parameter failed: %v", err)
			}

			b2, err := json.Marshal(args[1])
			if err != nil {
				return nil, fmt.Errorf("marshal second parameter failed: %v", err)
			}

			patch, err := jsondiff.CompareJSON(b1, b2)
			if err != nil {
				return nil, fmt.Errorf("compare json failed: %v", err)
			}

			if len(args) > 2 && args[2] == "all" {
				return map[string]interface{}{
					"source": args[0],
					"target": args[1],
					"diff":   patch,
				}, nil
			}

			return patch, nil

		},
		"query": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, errors.New("parameter length not match")
			}

			def := interface{}("")
			if len(args) > 2 {
				def = args[2]
			}

			q := gojsonq.Parse(fmt.Sprintf("%v", args[1]))
			out := q.Get(args[0])
			if out == nil {
				return def, nil
			}
			return out, nil
		},
		"round": func(args ...interface{}) (interface{}, error) {
			if len(args) < 1 {
				return nil, errors.New("parameter length not match")
			}

			switch v := args[0].(type) {
			case float64:
				return math.Round(v), nil
			case float32:
				return float64(math.Round(float64(v))), nil
			case int:
				return float64(v), nil
			case int64:
				return float64(v), nil
			case int32:
				return float64(v), nil
			case string:
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				return math.Round(f), nil
			default:
				f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
				if err != nil {
					return nil, err
				}
				return math.Round(f), nil
			}
		},
		"floor": func(args ...interface{}) (interface{}, error) {
			if len(args) < 1 {
				return nil, errors.New("parameter length not match")
			}

			switch v := args[0].(type) {
			case float64:
				return math.Floor(v), nil
			case float32:
				return float64(math.Floor(float64(v))), nil
			case int:
				return float64(v), nil
			case int64:
				return float64(v), nil
			case int32:
				return float64(v), nil
			case string:
				f, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return nil, err
				}
				return math.Floor(f), nil
			default:
				f, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
				if err != nil {
					return nil, err
				}
				return math.Floor(f), nil
			}
		},
		"trim": func(args ...interface{}) (interface{}, error) {
			if len(args) < 1 {
				return nil, errors.New("parameter length not match")
			}
			val := fmt.Sprintf("%v", args[0])

			if len(args) == 1 {
				return strings.TrimSpace(val), nil
			}

			if len(args) > 1 {
				sub := fmt.Sprintf("%v", args[1])
				mode := ""
				if len(args) > 2 {
					mode = strings.ToLower(fmt.Sprintf("%v", args[2]))
				}
				switch mode {
				case "prefix":
					return strings.TrimPrefix(val, sub), nil
				case "suffix":
					return strings.TrimSuffix(val, sub), nil
				default:
					return strings.Trim(val, sub), nil
				}
			}

			return strings.TrimSpace(val), nil
		},
		"replace": func(args ...interface{}) (interface{}, error) {
			if len(args) < 3 {
				return nil, errors.New("parameter length not match")
			}
			return strings.ReplaceAll(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1]), fmt.Sprintf("%v", args[2])), nil
		},
		"case": func(args ...interface{}) (interface{}, error) {
			if len(args) < 1 {
				return nil, errors.New("parameter length not match")
			}

			mode := ""
			if len(args) > 1 {
				mode = strings.ToLower(fmt.Sprintf("%v", args[1]))
			}

			switch mode {
			case "upper":
				return strings.ToUpper(fmt.Sprintf("%v", args[0])), nil
			case "lower":
				return strings.ToLower(fmt.Sprintf("%v", args[0])), nil
			default:
				return cases.Title(language.English).String(strings.ToLower(fmt.Sprintf("%v", args[0]))), nil
				//return strings.Title(strings.ToLower(fmt.Sprintf("%v", args[0]))), nil
			}
		},
		"contains": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, errors.New("parameter length not match")
			}

			mode := ""
			if len(args) > 2 {
				mode = strings.ToLower(fmt.Sprintf("%v", args[2]))
			}

			switch v := args[0].(type) {
			case []interface{}:
				for _, vv := range v {
					if fmt.Sprintf("%v", vv) == fmt.Sprintf("%v", args[1]) {
						return true, nil
					}
				}
				return false, nil
			}

			switch mode {
			case "prefix":
				return strings.HasPrefix(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1])), nil
			case "suffix":
				return strings.HasSuffix(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1])), nil
			default:
				return strings.Contains(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1])), nil
			}
		},
		"split": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, errors.New("parameter length not match")
			}

			slice := strings.Split(fmt.Sprintf("%v", args[0]), fmt.Sprintf("%v", args[1]))
			if len(args) > 2 {
				idx, err := strconv.Atoi(fmt.Sprintf("%v", args[2]))
				if err == nil {
					if len(slice) > idx {
						return slice[idx], nil
					}
				}
			}
			return slice, nil
		},
		"unwrapslice": func(args ...interface{}) (interface{}, error) {
			if len(args) < 1 {
				return nil, errors.New("parameter length not match")
			}
			slice := make([]interface{}, 0)
			for _, arg := range args {
				unwrapslice(arg, &slice)
			}
			return slice, nil
		},
		"join": func(args ...interface{}) (interface{}, error) {
			if len(args) < 2 {
				return nil, errors.New("parameter length not match")
			}

			switch v := args[0].(type) {
			case []interface{}:
				tmp := make([]string, len(v))
				for i, vv := range v {
					tmp[i] = fmt.Sprintf("%v", vv)
				}
				return strings.Join(tmp, fmt.Sprintf("%v", args[1])), nil
			case []string:
				return strings.Join(v, fmt.Sprintf("%v", args[1])), nil
			}

			if len(args) >= 2 {
				sep := fmt.Sprintf("%v", args[len(args)-1])
				tmp := make([]string, len(args)-1)
				for i := 0; i < len(args)-1; i++ {
					tmp[i] = fmt.Sprintf("%v", args[i])
				}
				return strings.Join(tmp, sep), nil
			}
			return nil, errors.New("parameter type not match")
		},
		"sub": func(args ...interface{}) (interface{}, error) {
			if len(args) < 3 {
				return nil, errors.New("parameter length not match")
			}

			start, err := strconv.Atoi(fmt.Sprintf("%v", args[1]))
			if err != nil {
				return nil, err
			}

			end, err := strconv.Atoi(fmt.Sprintf("%v", args[2]))
			if err != nil {
				return nil, err
			}

			switch v := args[0].(type) {
			case []interface{}:
				return v[start:end], nil
			default:
				return fmt.Sprintf("%v", args[0])[start:end], nil
			}
		},
		"sum": func(args ...interface{}) (interface{}, error) {
			if len(args) < 1 {
				return nil, errors.New("parameter length not match")
			}

			slice := make([]float64, 0)

			switch v := args[0].(type) {
			case []int:
				for _, vv := range v {
					slice = append(slice, float64(vv))
				}
			case []int64:
				for _, vv := range v {
					slice = append(slice, float64(vv))
				}
			case []float64:
				slice = v
			default:
				if !IsSliceOrPointerOfSlice(v) {
					return nil, errors.New("parameter type not match")
				}
				len := GetSliceLength(v)
				for i := 0; i < len; i++ {
					tmp, err := strconv.ParseFloat(fmt.Sprintf("%v", GetSliceItem(v, i)), 64)
					if err != nil {
						return nil, err
					}
					slice = append(slice, tmp)
				}

			}

			if len(slice) == 0 {
				return 0, nil
			}

			sum := float64(0)
			max := slice[0]
			min := slice[0]
			for _, v := range slice {
				if max < v {
					max = v
				}
				if min > v {
					min = v
				}
				sum += v
			}

			if len(args) > 1 {
				switch fmt.Sprintf("%v", args[1]) {
				case "max":
					return max, nil
				case "min":
					return min, nil
				case "avg":
					return sum / float64(len(slice)), nil
				default:
					return sum, nil
				}
			}

			return sum, nil
		},
		"geoconvert": func(args ...interface{}) (interface{}, error) {
			if len(args) < 3 {
				return nil, errors.New("parameter length not match")
			}
			from := fmt.Sprintf("%v", args[1])
			to := fmt.Sprintf("%v", args[2])

			return geoconvert(from, to, args[0])
		},
	}

	ve, err := ev.NewEvaluableExpressionWithFunctions(exp, fns)
	if err != nil {
		return nil, err
	}

	return &Evaluator{
		exp:    ve,
		params: params,
	}, nil
}

// Eval evaluate object
func (e *Evaluator) Eval(obj interface{}) (interface{}, error) {
	if e.exp == nil {
		return nil, errors.New("empty evaluator instance")
	}

	if len(e.params) > 0 && obj == nil {
		return nil, errors.New("null object")
	}

	var params map[string]interface{}
	if len(e.params) > 0 {
		params = make(map[string]interface{})
		for m, k := range e.params {
			v, _ := Lookup(k, obj)
			if v == nil {
				v = string("")
			}
			params[m] = v
		}
	}

	return e.exp.Evaluate(params)
}

func (e *Evaluator) EvalStringJSON(json string) (interface{}, error) {
	if e.exp == nil {
		return nil, errors.New("empty evaluator instance")
	}

	if len(e.params) > 0 && json == "" {
		return nil, errors.New("null object")
	}

	var params map[string]interface{}
	if len(e.params) > 0 {
		params = make(map[string]interface{})
		for m, k := range e.params {
			if k == "'." {
				params[m] = json
				continue
			}
			if k == "." {
				k = "@this"
			}
			v := gjson.Get(json, k).Value()
			if v == nil {
				v = string("")
			}

			params[m] = v
		}
	}

	return e.exp.Evaluate(params)
}

func (e *Evaluator) EvalBytesJSON(json []byte) (interface{}, error) {
	if e.exp == nil {
		return nil, errors.New("empty evaluator instance")
	}

	if len(e.params) > 0 && json == nil {
		return nil, errors.New("null object")
	}

	var params map[string]interface{}
	if len(e.params) > 0 {
		params = make(map[string]interface{})
		for m, k := range e.params {
			if k == "'." {
				params[m] = json
				continue
			}
			if k == "." {
				k = "@this"
			}
			v := gjson.GetBytes(json, k).Value()
			if v == nil {
				v = string("")
			}

			params[m] = v
		}
	}

	return e.exp.Evaluate(params)
}

func normalizeTime(t int64) time.Time {
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

func unwrapslice(val interface{}, slice *[]interface{}) {
	if IsSliceOrPointerOfSlice(val) {
		len := GetSliceLength(val)
		for i := 0; i < len; i++ {
			unwrapslice(GetSliceItem(val, i), slice)
		}
		return
	}
	*slice = append(*slice, val)

}

func geoconvert(from, to string, obj interface{}) (interface{}, error) {
	var geom space.Geometry

	switch from {
	case "wkt":
		buf := new(bytes.Buffer)
		buf.Write([]byte(fmt.Sprintf("%v", obj)))
		got, err := geoencoding.Read(buf, geoencoding.WKT)
		if err != nil {
			return nil, err
		}
		geom = got
	case "geojson":
		buf := new(bytes.Buffer)
		buf.Write([]byte(fmt.Sprintf("%v", obj)))
		got, err := geoencoding.Read(buf, geoencoding.GeoJSON)
		if err != nil {
			return nil, err
		}
		geom = got
	case "geojsonobj":
		b, err := json.Marshal(obj)
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		buf.Write(b)
		got, err := geoencoding.Read(buf, geoencoding.GeoJSON)
		if err != nil {
			return nil, err
		}
		geom = got
	default:
		return nil, errors.New("unsupported source type")
	}
	buf := new(bytes.Buffer)
	switch to {
	case "wkt":
		err := geoencoding.Write(buf, geom, geoencoding.WKT)
		if err != nil {
			return nil, err
		}
	case "geojson":
		err := geoencoding.Write(buf, geom, geoencoding.GeoJSON)
		if err != nil {
			return nil, err
		}
	case "geojsonobj":
		err := geoencoding.Write(buf, geom, geoencoding.GeoJSON)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported target type")
	}

	if to == "geojsonobj" {
		v := make(map[string]interface{})
		err := json.Unmarshal(buf.Bytes(), &v)
		if err != nil {
			return nil, err
		}
		return v, nil
	}

	return buf.String(), nil
}
