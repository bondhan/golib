package transform

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bondhan/golib/config"
	"github.com/bondhan/golib/util"
)

// Filter filter interface
type Filter interface {
	Match(obj interface{}) bool
}

// ANDFilter filter AND
type ANDFilter struct {
	Conditions []Condition `json:"conditions,omitempty" mapstructure:"conditions"`
}

// ORFilter filter OR
type ORFilter struct {
	Conditions []Condition `json:"conditions,omitempty" mapstructure:"conditions"`
}

type RegexFilter struct {
	Conditions []Condition `json:"conditions,omitempty" mapstructure:"conditions"`
}

// ExpFilter expression filter
type ExpFilter struct {
	Expression string `json:"expression,omitempty" mapstructure:"expression"`
	eval       *util.Evaluator
}

// Condition filter condition
type Condition struct {
	Field string      `json:"field,omitempty" mapstructure:"field"`
	Ops   string      `json:"ops,omitempty" mapstructure:"ops"`
	Value interface{} `json:"value,omitempty" mapstructure:"value"`
}

// NewExpFilter new expression filter
func NewExpFilter(exp string) (*ExpFilter, error) {
	f := &ExpFilter{Expression: exp}
	if err := f.init(); err != nil {
		return nil, err
	}
	return f, nil
}

// Match check if object match with the condition
func (a *ANDFilter) Match(obj interface{}) bool {
	if a.Conditions == nil || len(a.Conditions) == 0 {
		return true
	}

	for _, c := range a.Conditions {
		if !c.assert(obj) {
			return false
		}
	}
	return true
}

// Match check if object match with the condition
func (o *ORFilter) Match(obj interface{}) bool {
	if o.Conditions == nil || len(o.Conditions) == 0 {
		return false
	}

	for _, c := range o.Conditions {
		if c.assert(obj) {
			return true
		}
	}
	return false
}

// Match check if object match with expression
func (e *ExpFilter) Match(obj interface{}) bool {
	r, err := e.eval.Eval(obj)
	if err != nil {
		return false
	}

	if b, ok := r.(bool); ok {
		return b
	}
	return false
}

// Match check if object match with expression
func (r *RegexFilter) Match(obj interface{}) bool {
	if r.Conditions == nil || len(r.Conditions) == 0 {
		return true
	}

	for _, c := range r.Conditions {
		v, _ := util.Lookup(c.Field, obj)
		k, _ := regexp.Match(fmt.Sprintf("%v", c.Value), []byte(fmt.Sprintf("%v", v)))
		if c.Ops == "!=" && k {
			return false
		}
		if !k {
			return false
		}
	}
	return true
}

func (e *ExpFilter) init() error {
	eval, err := util.NewEvaluator(e.Expression)
	if err != nil {
		return err
	}
	e.eval = eval
	return nil
}

func (c *Condition) assert(obj interface{}) bool {
	v, _ := util.Lookup(c.Field, obj)

	if c.Value == nil {
		switch c.Ops {
		case "=":
			return v == nil
		case "!=":
			return v != nil
		default:
			return false
		}
	}

	if c.Ops == "=" {
		return util.Match(c.Field, obj, c.Value)
	}

	switch c.Value.(type) {
	case int, int32, int64, float32, float64:
		ref, err := strconv.ParseFloat(fmt.Sprintf("%v", c.Value), 64)
		if err != nil {
			return false
		}
		val, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
		if err != nil {
			return false
		}
		return compareNum(val, ref, c.Ops)
	case bool:
		ref, err := strconv.ParseBool(fmt.Sprintf("%v", c.Value))
		if err != nil {
			return false
		}
		val, err := strconv.ParseBool(fmt.Sprintf("%v", v))
		if err != nil {
			return false
		}
		return compareBool(val, ref, c.Ops)
	default:
		return compareString(fmt.Sprintf("%v", v), fmt.Sprintf("%v", c.Value), c.Ops)
	}
}

func compareString(a, b, ops string) bool {
	switch ops {
	case "=":
		return a == b
	case "!=":
		return a != b
	case "~":
		return strings.Contains(a, b)
	default:
		return false
	}
}

func compareBool(a, b bool, ops string) bool {
	switch ops {
	case "=":
		return a == b
	case "!=":
		return a != b
	default:
		return false
	}
}

func compareNum(a, b float64, ops string) bool {
	switch ops {
	case "=":
		return a == b
	case "!=":
		return a != b
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "<":
		return a < b
	case "<=":
		return a <= b
	default:
		return false
	}
}

func toFilter(conf interface{}, filter interface{}) error {
	switch cfg := conf.(type) {
	case config.Getter:
		return cfg.Unmarshal(filter)
	default:
		return util.DecodeJSON(conf, filter)
	}
}

// GetFilter get filter from config
func GetFilter(conf interface{}) (Filter, error) {

	var kind string
	switch cfg := conf.(type) {
	case config.Getter:
		kind = cfg.GetString("type")
		if kind == "" {
			return nil, errors.New("unknown type")
		}
	default:
		k, ok := util.Lookup("type", conf)
		if !ok {
			return nil, errors.New("unknown type")
		}
		kind = fmt.Sprintf("%v", k)
	}

	switch strings.ToUpper(kind) {
	case "AND":
		var f ANDFilter
		if err := toFilter(conf, &f); err != nil {
			return nil, err
		}
		return &f, nil
	case "OR":
		var f ORFilter
		if err := toFilter(conf, &f); err != nil {
			return nil, err
		}
		return &f, nil
	case "REGEX":
		var f RegexFilter
		if err := toFilter(conf, &f); err != nil {
			return nil, err
		}
		return &f, nil
	case "EXP":
		var f ExpFilter
		if err := toFilter(conf, &f); err != nil {
			return nil, err
		}
		if err := f.init(); err != nil {
			return nil, err
		}
		return &f, nil
	default:
		return nil, errors.New("unsupported filter")
	}
}
