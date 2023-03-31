package util

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bondhan/golib/constant"
)

func TestAssertOptions(t *testing.T) {

	data := map[string]interface{}{
		"name": "sahal",
		"age":  31,
	}

	opts := AssertOpts{
		{Path: "name", Op: constant.EQ, Value: "sahal"},
		{Path: "age", Op: constant.GT, Value: 30},
	}

	ok, err := opts.Assert(data)
	require.Nil(t, err)
	require.True(t, ok)

	opts = AssertOpts{
		{Path: "age", Op: constant.GT, Value: 30},
		{Op: constant.OR, Value: AssertOpts{
			{Path: "name", Op: constant.EQ, Value: "sahal"},
			{Path: "name", Op: constant.EQ, Value: "zain"},
		}},
	}

	ok, err = opts.Assert(data)
	require.Nil(t, err)
	require.True(t, ok)
}
