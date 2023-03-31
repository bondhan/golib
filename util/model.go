package util

import (
	"errors"
	"strings"
	"time"

	ds "github.com/ompluscator/dynamic-struct"
)

// Model sql model
type Model struct {
	Fields     []Field          `json:"fields,omitempty" mapstructure:"fields"`
	References map[string]Model `json:"references,omitempty" mapstructure:"references"`
	builder    ds.DynamicStruct
}

// Field struct field
type Field struct {
	Name string `json:"name,omitempty" mapstructure:"name"`
	Type string `json:"type,omitempty" mapstructure:"type"`
	Tag  string `json:"tag,omitempty" mapstructure:"tag"`
}

func (f *Field) GetNameByTag(tags ...string) string {
	for _, t := range tags {
		if !strings.Contains(f.Tag, t) {
			return ""
		}
		ts := strings.Split(f.Tag, " ")
		for _, s := range ts {
			if strings.HasPrefix(s, t) {
				s = s[len(t)+1:]
				s = strings.Trim(s, `"`)
				return s
			}
		}
	}

	return ""
}

func getSampleType(t string) interface{} {
	switch strings.ToLower(t) {
	case "string":
		return ""
	case "int", "uint":
		return 0
	case "float":
		return 0.0
	case "bool":
		return false
	case "sliceint":
		return []int{}
	case "slicestring":
		return []string{}
	case "time":
		return time.Time{}
	case "map":
		return map[string]interface{}{}
	default:
		return nil
	}
}

func (m *Model) GetField(name string) *Field {
	for _, f := range m.Fields {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func (m *Model) GetFieldByTag(name, tag string) *Field {
	for _, f := range m.Fields {
		if f.GetNameByTag(tag) == name {
			return &f
		}
	}
	return nil
}

func (m *Model) getValue(field *Field, obj interface{}) (interface{}, error) {
	reader := ds.NewReader(obj)
	f := reader.GetField(field.Name)
	if f == nil {
		return nil, errors.New("field not found")
	}

	switch field.Type {
	case "string":
		return f.String(), nil
	case "int":
		return f.Int(), nil
	case "uint":
		return f.Uint(), nil
	case "float":
		return f.Float64(), nil
	case "bool":
		return f.Bool(), nil
	default:
		return f.Interface(), nil
	}
}

func (m *Model) GetValue(fieldName string, obj interface{}) (interface{}, error) {
	df := m.GetField(fieldName)
	if df == nil {
		return nil, errors.New("[util/map] field definition not found")
	}

	return m.getValue(df, obj)

}

func (m *Model) GetValueByTag(name, tag string, obj interface{}) (interface{}, error) {
	df := m.GetFieldByTag(name, tag)
	if df == nil {
		return nil, errors.New("[util/map] field definition not found")
	}
	return m.getValue(df, obj)
}

func (m *Model) newInstance() (interface{}, error) {
	if m.builder != nil {
		return m.builder.New(), nil
	}
	if m.Fields == nil || len(m.Fields) == 0 {
		return nil, errors.New("empty fields definition")
	}

	out := ds.NewStruct()
	for _, f := range m.Fields {
		s := getSampleType(f.Type)
		if s == nil {
			return nil, errors.New("invalid field type")
		}
		out = out.AddField(f.Name, s, f.Tag)
	}
	m.builder = out.Build()
	return m.builder.New(), nil
}

func (m *Model) ListFieldsByTag(tags ...string) []string {
	out := make([]string, 0)
	for _, f := range m.Fields {
		if col := f.GetNameByTag(tags...); col != "" {
			out = append(out, col)
		}
	}
	return out
}

// Build build model structure
func (m *Model) Build() error {
	if m.builder != nil {
		return nil
	}

	refs := make(map[string]interface{})

	for n, r := range m.References {
		if err := r.Build(); err != nil {
			return err
		}
		i, err := r.newInstance()
		if err != nil {
			return err
		}
		refs[n] = i
	}

	out := ds.NewStruct()
	if len(m.Fields) == 0 {
		return errors.New("[util/model] empty fields definition")
	}
	for _, f := range m.Fields {
		var s interface{}
		if i, ok := refs[f.Type]; ok {
			s = i
		} else {
			s = getSampleType(f.Type)
		}

		if s == nil {
			return errors.New("invalid field type")
		}
		out = out.AddField(f.Name, s, f.Tag)
	}

	m.builder = out.Build()
	return nil
}

// GetInstance get instance of model definition
func (m *Model) GetInstance() (interface{}, error) {
	if err := m.Build(); err != nil {
		return nil, err
	}
	return m.builder.New(), nil
}

func NewModel(conf interface{}) (*Model, error) {
	var md Model
	switch c := conf.(type) {
	case map[string]interface{}:
		if err := DecodeJSON(c, &md); err != nil {
			return nil, err
		}
	case Model:
		md = c
	case *Model:
		md = *c
	default:
		return nil, errors.New("[util/model] unsupported config scheme")
	}

	return &md, nil
}
