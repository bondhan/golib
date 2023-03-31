package transform

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapObject(t *testing.T) {
	conf := &ObjectMapper{
		MapDef: map[string]interface{}{
			"username":      "$username",
			"id":            "$'id",
			"profile.age":   "$age",
			"profile.type":  "person",
			"profile.id":    "$profile_id",
			"profile.score": "eval:$age * 2",
		},
	}

	err := conf.init()
	assert.Nil(t, err)

	obj := map[string]interface{}{
		"username": "sahalzain",
		"id":       123,
		"age":      37,
	}

	out, err := conf.Map(obj)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	assert.Equal(t, "123", out.(map[string]interface{})["id"])

	prof := out.(map[string]interface{})["profile"].(map[string]interface{})
	assert.NotNil(t, prof)
	assert.Equal(t, 37, prof["age"])
	assert.Equal(t, float64(37*2), prof["score"])
}

func TestRootMap(t *testing.T) {
	conf := &ObjectMapper{
		MapDef: map[string]interface{}{
			".": "$profile",
		},
	}

	err := conf.init()
	assert.Nil(t, err)

	obj := map[string]interface{}{
		"profile": map[string]interface{}{
			"username": "sahalzain",
			"id":       123,
			"age":      37,
		},
	}

	out, err := conf.Map(obj)
	assert.Nil(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, obj["profile"], out)
}

func TestEnrichObject(t *testing.T) {
	conf := &ObjectMapper{
		enrich: true,
		MapDef: map[string]interface{}{
			"id":            "$'id",
			"age":           "",
			"profile.age":   "$age",
			"profile.type":  "person",
			"profile.id":    "$profile_id",
			"profile.score": "eval:$age * 2",
		},
	}

	err := conf.init()
	assert.Nil(t, err)

	obj := map[string]interface{}{
		"username": "sahalzain",
		"id":       123,
		"age":      37,
	}

	out, err := conf.Map(obj)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	fmt.Println(out)

	assert.Equal(t, "123", out.(map[string]interface{})["id"])
	assert.Equal(t, "sahalzain", out.(map[string]interface{})["username"])
	assert.Nil(t, out.(map[string]interface{})["age"])

	prof := out.(map[string]interface{})["profile"].(map[string]interface{})
	assert.NotNil(t, prof)
	assert.Equal(t, 37, prof["age"])
	assert.Equal(t, float64(37*2), prof["score"])
}

func TestMapSlice(t *testing.T) {
	conf := &ObjectMapper{
		MapDef: map[string]interface{}{
			"id":                  "$username",
			"addresses.0.address": "$phone",
			"addresses.0.channel": "sms",
			"addresses.1.address": "$phone",
			"addresses.1.channel": "wa",
		},
	}

	err := conf.init()
	assert.Nil(t, err)

	obj := map[string]interface{}{
		"username": "sahalzain",
		"phone":    "08231323213",
	}

	out, err := conf.Map(obj)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	fout := out.(map[string]interface{})["addresses"]
	assert.Equal(t, 2, len(fout.([]interface{})))
}

func TestMapSliceRoot(t *testing.T) {
	conf := &ObjectMapper{
		MapDef: map[string]interface{}{
			"0.address": "$phone",
			"0.channel": "sms",
			"1.address": "$phone",
			"1.channel": "wa",
		},
	}

	err := conf.init()
	assert.Nil(t, err)

	obj := map[string]interface{}{
		"username": "sahalzain",
		"phone":    "08231323213",
	}

	out, err := conf.Map(obj)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	assert.Equal(t, 2, len(out.([]interface{})))
	fmt.Println(out)

}

func TestMapSliceStringRoot(t *testing.T) {
	conf := &ObjectMapper{
		MapDef: map[string]interface{}{
			"0": "$phone",
			"1": "$username",
		},
	}

	err := conf.init()
	assert.Nil(t, err)

	obj := map[string]interface{}{
		"username": "sahalzain",
		"phone":    "08231323213",
	}

	out, err := conf.Map(obj)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	assert.Equal(t, 2, len(out.([]interface{})))
	assert.Equal(t, "08231323213", out.([]interface{})[0])
	assert.Equal(t, "sahalzain", out.([]interface{})[1])

}

func TestMapSliceEnd(t *testing.T) {
	conf := &ObjectMapper{
		MapDef: map[string]interface{}{
			"data.0": "$phone",
			"data.1": "$username",
		},
	}

	err := conf.init()
	assert.Nil(t, err)

	obj := map[string]interface{}{
		"username": "sahalzain",
		"phone":    "08231323213",
	}

	out, err := conf.Map(obj)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	data := out.(map[string]interface{})["data"]
	assert.Equal(t, 2, len(data.([]interface{})))
	assert.Equal(t, "08231323213", data.([]interface{})[0])
	assert.Equal(t, "sahalzain", data.([]interface{})[1])
}

func TestMapSliceFloat(t *testing.T) {
	conf := &ObjectMapper{
		MapDef: map[string]interface{}{
			"location.type":          "Point",
			"location.coordinates.0": "$job.br.lon",
			"location.coordinates.1": "$job.br.lat",
		},
	}

	err := conf.init()
	assert.Nil(t, err)

	obj := map[string]interface{}{
		"job": map[string]interface{}{
			"br": map[string]interface{}{
				"lon": 1.20321332432,
				"lat": -1.3234234234,
			},
		},
	}

	out, err := conf.Map(obj)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	data := out.(map[string]interface{})["location"].(map[string]interface{})
	assert.Equal(t, "Point", data["type"])
	fmt.Println(data["coordinates"])
	assert.Equal(t, 2, len(data["coordinates"].([]interface{})))

}

func TestMapTemplate(t *testing.T) {
	obj := []map[string]interface{}{
		{
			"name": "sahal",
			"age":  37,
		},
		{
			"name": "zain",
			"age":  36,
		},
	}
	conf := &MapTemplate{
		Filepath: "test.tmpl",
	}
	assert.Nil(t, conf.init())
	out, err := conf.Map(obj)
	assert.Nil(t, err)
	assert.NotNil(t, out)

	assert.Equal(t, 2, len(out.(map[string]interface{})["result"].([]interface{})))
}
