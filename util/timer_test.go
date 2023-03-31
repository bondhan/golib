package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimer(t *testing.T) {
	tm := NewTimer(1)
	assert.False(t, tm.IsTimeout())

	time.Sleep(1 * time.Second)
	assert.True(t, tm.IsTimeout())

	tm = NewTimer(1)
	assert.False(t, tm.IsTimeout())
	tm.Extend(1)

	time.Sleep(1 * time.Second)
	assert.False(t, tm.IsTimeout())

	time.Sleep(1 * time.Second)
	assert.True(t, tm.IsTimeout())
}

func TestNormalize(t *testing.T) {
	ts := time.Now()

	assert.Equal(t, ts.Unix(), NormalizeTime(ts.Unix(), ""))
	assert.Equal(t, ts.Unix(), NormalizeTime(ts.UnixMilli(), ""))
	assert.Equal(t, ts.Unix(), NormalizeTime(ts.UnixMicro(), ""))
	assert.Equal(t, ts.Unix(), NormalizeTime(ts.UnixNano(), ""))
}

func TestMongoDateFormat(t *testing.T) {
	ts := "2022-11-27T09:05:30.847+00:00"
	tm, err := DateStringToTime(ts)
	assert.Nil(t, err)
	fmt.Println(tm)
	assert.False(t, tm.IsZero())
}
