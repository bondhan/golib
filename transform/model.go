package transform

import (
	"errors"
	"strings"
	"time"

	ds "github.com/ompluscator/dynamic-struct"
)

// Model sql model
type Model struct {
	Fields  []Field `json:"fields,omitempty" mapstructure:"fields"`
	builder ds.DynamicStruct
}

// ModelDef model definition
type ModelDef struct {
	Main       Model            `json:"main,omitempty" mapstructure:"main"`
	References map[string]Model `json:"references,omitempty" mapstructure:"references"`
	builder    ds.DynamicStruct
}

// Field struct field
type Field struct {
	Name   string `json:"name,omitempty" mapstructure:"name"`
	Type   string `json:"type,omitempty" mapstructure:"type"`
	Tag    string `json:"tag,omitempty" mapstructure:"tag"`
	Column string `json:"column,omitempty" mapstructure:"column"`
}

func (f *Field) GetColumn(tags ...string) string {
	if f.Column != "" {
		return f.Column
	}

	for _, t := range tags {
		ts := strings.Split(f.Tag, " ")
		if len(ts) == 0 {
			return ""
		}
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

func (m *Model) GetValue(fieldName string, obj interface{}) (interface{}, error) {
	reader := ds.NewReader(obj)

	df := m.GetField(fieldName)
	if df == nil {
		return nil, errors.New("field definition not found")
	}

	f := reader.GetField(fieldName)
	if f == nil {
		return nil, errors.New("field not found")
	}

	switch df.Type {
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

func (m *Model) getPrimaryKey() (string, error) {
	if m.Fields == nil || len(m.Fields) == 0 {
		return "", errors.New("empty fields definition")
	}

	for _, f := range m.Fields {
		if strings.Contains(f.Tag, `gorm:"primaryKey"`) {
			return f.Name, nil
		}
	}

	for _, f := range m.Fields {
		if f.Name == "ID" {
			return f.Name, nil
		}
	}

	return "", errors.New("primary key not found")
}

func (m *Model) ListColumn(tags ...string) []interface{} {
	out := make([]interface{}, 0)
	for _, f := range m.Fields {
		if col := f.GetColumn(tags...); col != "" {
			out = append(out, col)
		}
	}
	return out
}

// Build build model structure
func (d *ModelDef) Build() error {
	if d.builder != nil {
		return nil
	}

	refs := make(map[string]interface{})

	if d.References != nil && len(d.References) > 0 {
		for n, r := range d.References {
			i, err := r.newInstance()
			if err != nil {
				return err
			}
			refs[n] = i
		}
	}

	out := ds.NewStruct()
	for _, f := range d.Main.Fields {
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

	d.builder = out.Build()
	return nil
}

// GetInstance get instance of model definition
func (d *ModelDef) GetInstance() (interface{}, error) {
	if err := d.Build(); err != nil {
		return nil, err
	}
	return d.builder.New(), nil
}

// GetPrimaryKey get primary key
func (d *ModelDef) GetPrimaryKey() (string, error) {
	return d.Main.getPrimaryKey()
}
