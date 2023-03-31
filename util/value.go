package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ValueGetter value getter
type ValueGetter struct {
	eval      *Evaluator
	temp      *Template
	lookupKey string
	value     interface{}
}

// NewValueGetter new value getter instance
func NewValueGetter(exp interface{}) (*ValueGetter, error) {
	vg := &ValueGetter{}
	ks := fmt.Sprintf("%v", exp)
	if ks == "" {
		vg.value = nil
		return vg, nil
	}
	if strings.HasPrefix(ks, "eval:") {
		ev, err := NewEvaluator(fmt.Sprintf("%v", ks))
		if err != nil {
			return nil, err
		}
		vg.eval = ev
		return vg, nil
	}

	if strings.Contains(ks, "{") {
		vg.temp = NewTemplate(ks)
		return vg, nil
	}

	if strings.Contains(ks, "$") {
		vg.lookupKey = strings.TrimPrefix(ks, "$")
		return vg, nil
	}

	vg.value = exp
	return vg, nil
}

// Get get value
func (v *ValueGetter) Get(data interface{}) (interface{}, error) {
	if v.value != nil {
		if strings.Contains(fmt.Sprintf("%v", v.value), "now") {
			tv, err := GetTimeValue(fmt.Sprintf("%v", v.value))
			if err == nil {
				v.value = tv
			}
		}
		return v.value, nil
	}

	if v.eval != nil {
		out, err := v.eval.Eval(data)
		if err != nil {
			return nil, err
		}
		return out, nil
	}

	if v.temp != nil {
		return v.temp.Render(data), nil
	}

	if v.lookupKey == "" {
		return nil, nil
	}

	marshal := false
	unmarshal := false
	stringify := false
	num := false

	lkey := v.lookupKey
	if strings.HasPrefix(lkey, "'") {
		stringify = true
		lkey = strings.TrimPrefix(lkey, "'")
	}

	if strings.HasPrefix(lkey, ":") {
		marshal = true
		lkey = strings.TrimPrefix(lkey, ":")
	}

	if strings.HasPrefix(lkey, "-") {
		unmarshal = true
		lkey = strings.TrimPrefix(lkey, "-")
	}

	if strings.HasPrefix(lkey, "%") {
		num = true
		lkey = strings.TrimPrefix(lkey, "%")
	}

	out, ok := Lookup(lkey, data)
	if !ok {
		return nil, nil
	}

	if stringify {
		return Stringify(out), nil
	}

	if marshal {
		b, err := json.Marshal(out)
		return string(b), err
	}

	if unmarshal {
		var b interface{}
		err := json.Unmarshal([]byte(out.(string)), &b)
		return b, err
	}

	if num {
		fl, err := strconv.ParseFloat(fmt.Sprintf("%v", out), 64)
		if err != nil {
			return nil, err
		}
		if fl == float64(int64(fl)) {
			return int64(fl), nil
		}
		return fl, nil
	}

	return out, nil
}

// GetTimeValue get time value from expression
func GetTimeValue(exp string) (interface{}, error) {
	if !strings.Contains(exp, "now") {
		return nil, nil
	}

	num := false

	if strings.Contains(exp, "inow") {
		num = true
		exp = strings.ReplaceAll(exp, "inow", "now")
	}

	t := time.Now()
	arg := strings.TrimSuffix(strings.TrimPrefix(exp, "now("), ")")
	if arg != "" {
		d, err := strconv.Atoi(arg)
		if err == nil {
			t = t.Add(time.Duration(d) * time.Second)
		}
	}

	if num {
		return t.Unix(), nil
	}

	return t, nil
}
