package transform

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/bondhan/golib/config"
	"github.com/bondhan/golib/util"

	"github.com/Jeffail/gabs/v2"
	"github.com/mitchellh/mapstructure"
)

// Mapper maper
type Mapper interface {
	Map(obj interface{}) (interface{}, error)
}

// GetMapper get mapper
func GetMapper(ctx context.Context, conf config.Getter) (Mapper, error) {
	switch strings.ToUpper(conf.GetString("type")) {
	case "ENRICHER":
		return NewEnricher(ctx, conf)
	case "TEMPLATE":
		return NewMapTemplate(ctx, conf)
	default:
		return NewMapper(ctx, conf)
	}
}

// ObjectMapper object mapper
type ObjectMapper struct {
	MapDef    map[string]interface{} `json:"map,omitempty" mapstructure:"map"`
	ExistOnly bool                   `json:"exist_only,omitempty" mapstructure:"exist_only"`
	enrich    bool
	getter    map[string]*util.ValueGetter
}

// NewMapper new object mapper
func NewMapper(ctx context.Context, conf interface{}) (*ObjectMapper, error) {
	return newMapper(ctx, conf, false)
}

// NewEnricher new object enricher
func NewEnricher(ctx context.Context, conf interface{}) (*ObjectMapper, error) {
	return newMapper(ctx, conf, true)
}

func newMapper(ctx context.Context, conf interface{}, enrich bool) (*ObjectMapper, error) {
	var mp ObjectMapper

	switch cfg := conf.(type) {
	case config.Getter:
		if err := cfg.Unmarshal(&mp); err != nil {
			return nil, err
		}
	case ObjectMapper:
		mp = cfg
	case *ObjectMapper:
		mp = *cfg
	case map[string]interface{}:
		if err := util.DecodeJSON(cfg, &mp); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unknown configuration type")
	}

	if mp.MapDef == nil || len(mp.MapDef) == 0 {
		return nil, errors.New("missing map param")
	}

	mp.enrich = enrich

	if err := mp.init(); err != nil {
		return nil, err
	}

	return &mp, nil

}

func (c *ObjectMapper) init() error {
	c.getter = make(map[string]*util.ValueGetter)
	for k, v := range c.MapDef {
		vg, err := util.NewValueGetter(v)
		if err != nil {
			return err
		}
		c.getter[k] = vg
	}
	return nil
}

// Map map object
func (c *ObjectMapper) Map(obj interface{}) (interface{}, error) {
	if c.enrich {
		return c.mapData(obj, obj, c.enrich)
	}
	return c.mapData(obj, nil, c.enrich)
}

// Merge merge object
func (c *ObjectMapper) Merge(in interface{}, base interface{}) (interface{}, error) {
	return c.mapData(in, base, true)
}

func (c *ObjectMapper) mapData(in interface{}, base interface{}, enrich bool) (interface{}, error) {
	var obj *gabs.Container
	if base != nil {
		var tmp map[string]interface{}

		if err := mapstructure.Decode(base, &tmp); err != nil {
			return nil, err
		}

		obj = gabs.Wrap(tmp)
	} else {
		obj = gabs.New()
	}

	for p, v := range c.getter {
		val, err := v.Get(in)
		if err != nil {
			return nil, err
		}

		if p == "." && !enrich {
			return val, nil
		}

		if val == nil {
			if enrich {
				obj.Delete(strings.Split(p, ".")...)
				continue
			}
			if c.ExistOnly {
				return nil, errors.New("missing value")
			}
			continue
		}

		if err := initValue(obj, p); err != nil {
			return nil, err
		}

		_, err = obj.Set(val, strings.Split(p, ".")...)
		if err != nil {
			fmt.Println(err)
		}

	}

	return obj.Data(), nil
}

// MapTemplate map template
type MapTemplate struct {
	Filepath string `json:"filepath,omitempty" mapstructure:"filepath"`
	template *template.Template
	funcMap  template.FuncMap
}

// NewMapTemplate create new template mapper
func NewMapTemplate(ctx context.Context, conf config.Getter) (*MapTemplate, error) {
	fp := conf.GetString("filepath")
	if fp == "" {
		return nil, errors.New("missing filepath param")
	}

	mt := &MapTemplate{
		Filepath: fp,
	}

	if err := mt.init(); err != nil {
		return nil, err
	}
	return mt, nil
}

func (t *MapTemplate) init() error {
	t.funcMap = template.FuncMap{
		"marshal": marshal,
		"now":     now,
		"bound":   bound,
		"nbound":  nbound,
		"incr":    incr,
	}

	tmpl, err := template.New("mapper").Funcs(t.funcMap).ParseFiles(t.Filepath)
	if err != nil {
		return err
	}

	t.template = tmpl
	return nil
}

// Map map object
func (t *MapTemplate) Map(obj interface{}) (interface{}, error) {
	if obj == nil {
		return nil, errors.New("null object")
	}
	var buf bytes.Buffer
	if err := t.template.ExecuteTemplate(&buf, t.Filepath, obj); err != nil {
		return nil, err
	}

	var out interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		return nil, err
	}

	return out, nil

}

func marshal(v interface{}) string {
	a, _ := json.Marshal(v)
	return string(a)
}

func now() int64 {
	return time.Now().Unix()
}

func bound(i int, obj interface{}) bool {
	rv := reflect.ValueOf(obj)
	switch rv.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return i >= (rv.Len() - 1)
	default:
		return false
	}
}

func nbound(i int, obj interface{}) bool {
	return !bound(i, obj)
}

func incr(i int) int {
	return i + 1
}

func isNum(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return -1
}

func initValue(obj *gabs.Container, key string) error {
	ks := strings.Split(key, ".")
	if len(ks) <= 1 {
		if num := isNum(key); num >= 0 {
			if tmp, ok := obj.Data().([]interface{}); ok {
				if len(tmp) <= num {
					for x := len(tmp); x <= num; x++ {
						tmp = append(tmp, nil)
					}
					*obj = *gabs.Wrap(tmp)
					return nil
				}
				return nil
			}

			arr := make([]interface{}, 0)

			for x := 0; x < num+1; x++ {
				arr = append(arr, nil)
			}

			*obj = *gabs.Wrap(arr)
			return nil
		}
	}

	p := ""
	for i := range ks {
		p = strings.Join(ks[:i+1], ".")
		if num := isNum(p); i == 0 && num >= 0 { // root

			if tmp, ok := obj.Data().([]interface{}); ok {
				if len(tmp) <= num {
					for x := len(tmp); x <= num; x++ {
						tmp = append(tmp, make(map[string]interface{}))
					}
					*obj = *gabs.Wrap(tmp)
					continue
				}
				continue
			}

			arr := make([]interface{}, 0)

			for x := 0; x < num+1; x++ {
				arr = append(arr, make(map[string]interface{}))
			}

			*obj = *gabs.Wrap(arr)
			continue
		}
		if !obj.Exists(p) {
			if i > 0 {
				if num := isNum(ks[i]); num >= 0 {
					if obj.Exists(ks[:i]...) {
						tmp := obj.Search(ks[:i]...)

						if sv, ok := tmp.Data().([]interface{}); ok {
							if len(sv) <= num {

								for x := len(sv); x <= num; x++ {
									sv = append(sv, make(map[string]interface{}))
								}
								obj.Set(sv, ks[:i]...)
								continue
							}
							continue
						}
						continue
					}

					if i < len(ks)-1 {

						arr := make([]interface{}, 0)
						for x := 0; x < num+1; x++ {
							arr = append(arr, make(map[string]interface{}))
						}

						obj.Set(arr, ks[:i]...)
						continue
					}

					obj.Set(make([]interface{}, num+1), ks[:i]...)
				}
			}
		}

	}

	return nil
}
