package transform

import (
	"encoding/json"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestModel(t *testing.T) {
	def := &ModelDef{
		Main: Model{
			Fields: []Field{
				{Name: "Username", Type: "string", Tag: "json:\"username\" mapstructure:\"username\""},
				{Name: "Profile", Type: "Profile", Tag: `json:"profile" mapstructure:"profile"`},
			},
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

	col := def.Main.Fields[0].GetColumn("json")
	assert.Equal(t, "username", col)
}
