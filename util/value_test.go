package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTime(t *testing.T) {
	ti, err := GetTimeValue("now()")
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Unix(), ti.(time.Time).Unix())

	ti, err = GetTimeValue("now(10)")
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Add(time.Duration(10)*time.Second).Unix(), ti.(time.Time).Unix())

	ti, err = GetTimeValue("now(-100)")
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Add(time.Duration(-100)*time.Second).Unix(), ti.(time.Time).Unix())

	ti, err = GetTimeValue("inow()")
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Unix(), ti)

	ti, err = GetTimeValue("inow(10)")
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Add(time.Duration(10)*time.Second).Unix(), ti)

	ti, err = GetTimeValue("inow(-100)")
	assert.Nil(t, err)
	assert.Equal(t, time.Now().Add(time.Duration(-100)*time.Second).Unix(), ti)
}

func TestNumGetter(t *testing.T) {

	obj := map[string]interface{}{
		"time":  "123456",
		"ftime": "1234.54",
	}

	vg, err := NewValueGetter("$%time")
	assert.Nil(t, err)
	assert.NotNil(t, vg)

	out, err := vg.Get(obj)
	assert.Nil(t, err)
	assert.Equal(t, int64(123456), out)

	fg, err := NewValueGetter("$%ftime")
	assert.Nil(t, err)
	assert.NotNil(t, vg)

	out, err = fg.Get(obj)
	assert.Nil(t, err)
	assert.Equal(t, 1234.54, out)

}
