package util

import (
	"encoding/json"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestModel(t *testing.T) {
	def := &Model{
		Fields: []Field{
			{Name: "Username", Type: "string", Tag: "json:\"username\" mapstructure:\"username\""},
			{Name: "Profile", Type: "Profile", Tag: `json:"profile" mapstructure:"profile"`},
		},
		References: map[string]Model{
			"Profile": {
				Fields: []Field{
					{Name: "Name", Type: "string", Tag: `json:"name" mapstructure:"name"`},
					{Name: "Age", Type: "int", Tag: `json:"age" mapstructure:"age"`},
				},
			},
		},
	}

	data := []byte(`{"username":"sahalzain","profile":{"name":"Sahal Zain","age":37}}`)
	ins, err := def.GetInstance()
	assert.Nil(t, err)
	assert.NotNil(t, ins)
	err = json.Unmarshal(data, ins)
	assert.Nil(t, err)

	d, err := json.Marshal(ins)
	assert.Nil(t, err)
	assert.Equal(t, data, d)

	obj := map[string]interface{}{
		"username": "sahalzain",
		"profile": map[string]interface{}{
			"name": "Sahal Zain",
			"age":  37,
		},
	}

	im, err := def.GetInstance()
	assert.Nil(t, err)

	err = mapstructure.Decode(obj, im)
	assert.Nil(t, err)

	d, err = json.Marshal(ins)
	assert.Nil(t, err)
	assert.Equal(t, data, d)

	col := def.Fields[0].GetNameByTag("json")
	assert.Equal(t, "username", col)

	val, err := def.GetValue("Username", im)
	assert.Nil(t, err)
	assert.Equal(t, "sahalzain", val)

	val, err = def.GetValueByTag("username", "json", im)
	assert.Nil(t, err)
	assert.Equal(t, "sahalzain", val)
}

func TestModelFromMap(t *testing.T) {
	conf := map[string]interface{}{
		"fields": []map[string]interface{}{
			{"name": "Username", "type": "string", "tag": "json:\"username\" mapstructure:\"username\""},
			{"name": "Profile", "type": "Profile", "tag": `json:"profile" mapstructure:"profile"`},
		},
		"references": map[string]interface{}{
			"Profile": map[string]interface{}{
				"fields": []map[string]interface{}{
					{"name": "Name", "type": "string", "tag": `json:"name" mapstructure:"name"`},
					{"name": "Age", "type": "int", "tag": `json:"age" mapstructure:"age"`},
				},
			},
		},
	}

	def, err := NewModel(conf)
	assert.Nil(t, err)
	assert.NotNil(t, def)

	data := []byte(`{"username":"sahalzain","profile":{"name":"Sahal Zain","age":37}}`)
	ins, err := def.GetInstance()
	assert.Nil(t, err)
	assert.NotNil(t, ins)
	err = json.Unmarshal(data, ins)
	assert.Nil(t, err)

	d, err := json.Marshal(ins)
	assert.Nil(t, err)
	assert.Equal(t, data, d)

	obj := map[string]interface{}{
		"username": "sahalzain",
		"profile": map[string]interface{}{
			"name": "Sahal Zain",
			"age":  37,
		},
	}

	im, err := def.GetInstance()
	assert.Nil(t, err)

	err = mapstructure.Decode(obj, im)
	assert.Nil(t, err)

	d, err = json.Marshal(ins)
	assert.Nil(t, err)
	assert.Equal(t, data, d)

	col := def.Fields[0].GetNameByTag("json")
	assert.Equal(t, "username", col)
}

func TestModelWithMap(t *testing.T) {
	conf := map[string]interface{}{
		"fields": []map[string]interface{}{
			{"name": "Username", "type": "string", "tag": "json:\"username\" mapstructure:\"username\""},
			{"name": "Profile", "type": "Profile", "tag": `json:"profile" mapstructure:"profile"`},
			{"name": "Info", "type": "map", "tag": `json:"info" mapstructure:"info"`},
		},
		"references": map[string]interface{}{
			"Profile": map[string]interface{}{
				"fields": []map[string]interface{}{
					{"name": "Name", "type": "string", "tag": `json:"name" mapstructure:"name"`},
					{"name": "Age", "type": "int", "tag": `json:"age" mapstructure:"age"`},
				},
			},
		},
	}

	def, err := NewModel(conf)
	assert.Nil(t, err)
	assert.NotNil(t, def)

	data := []byte(`{"username":"sahalzain","profile":{"name":"Sahal Zain","age":37},"info":{"age":37,"name":"Sahal Zain"}}`)
	ins, err := def.GetInstance()
	assert.Nil(t, err)
	assert.NotNil(t, ins)
	err = json.Unmarshal(data, ins)
	assert.Nil(t, err)

	d, err := json.Marshal(ins)
	assert.Nil(t, err)
	assert.Equal(t, data, d)

	obj := map[string]interface{}{
		"username": "sahalzain",
		"profile": map[string]interface{}{
			"name": "Sahal Zain",
			"age":  37,
		},
		"info": map[string]interface{}{
			"name": "Sahal Zain",
			"age":  37,
		},
	}

	im, err := def.GetInstance()
	assert.Nil(t, err)

	err = mapstructure.Decode(obj, im)
	assert.Nil(t, err)

	d, err = json.Marshal(ins)
	assert.Nil(t, err)
	assert.Equal(t, data, d)

	col := def.Fields[0].GetNameByTag("json")
	assert.Equal(t, "username", col)
}
