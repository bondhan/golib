package transform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCondition(t *testing.T) {

	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp":          "2021-02-15T15:00:03.686239Z",
			"record-type":        "control",
			"operation":          "create-table",
			"partition-key-type": "task-id",
			"schema-name":        "",
			"sequence":           87,
			"table-name":         "awsdms_apply_exceptions",
			"active":             true,
		},
	}

	assert.True(t, (&Condition{Field: "metadata.record-type", Ops: "=", Value: "control"}).assert(data))
	assert.False(t, (&Condition{Field: "metadata.record-type", Ops: "!=", Value: "control"}).assert(data))
	assert.True(t, (&Condition{Field: "metadata.record-type", Ops: "!=", Value: "ctrl"}).assert(data))
	assert.True(t, (&Condition{Field: "metadata.record-type", Ops: "~", Value: "cont"}).assert(data))

	assert.True(t, (&Condition{Field: "metadata.schema-name", Ops: "=", Value: ""}).assert(data))
	assert.True(t, (&Condition{Field: "metadata.operation", Ops: "!=", Value: ""}).assert(data))

	assert.True(t, (&Condition{Field: "metadata.sequence", Ops: "=", Value: 87}).assert(data))
	assert.True(t, (&Condition{Field: "metadata.sequence", Ops: ">", Value: 80}).assert(data))
	assert.False(t, (&Condition{Field: "metadata.sequence", Ops: "<", Value: 80}).assert(data))

	assert.True(t, (&Condition{Field: "metadata.active", Ops: "=", Value: true}).assert(data))

	assert.True(t, (&Condition{Field: "data", Ops: "="}).assert(data))
	assert.False(t, (&Condition{Field: "data", Ops: "!="}).assert(data))
	assert.True(t, (&Condition{Field: "metadata", Ops: "!="}).assert(data))
}

func TestAndFilter(t *testing.T) {

	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp":          "2021-02-15T15:00:03.686239Z",
			"record-type":        "control",
			"operation":          "create-table",
			"partition-key-type": "task-id",
			"schema-name":        "",
			"sequence":           87,
			"table-name":         "awsdms_apply_exceptions",
			"active":             "true",
		},
	}

	f := &ANDFilter{
		Conditions: []Condition{
			{Field: "metadata.record-type", Ops: "=", Value: "control"},
			{Field: "metadata.schema-name", Ops: "=", Value: ""},
			{Field: "metadata.operation", Ops: "~", Value: "create"},
		},
	}

	assert.True(t, f.Match(data))

	f = &ANDFilter{
		Conditions: []Condition{
			{Field: "metadata.record-type", Ops: "=", Value: "control"},
			{Field: "metadata.schema-name", Ops: "=", Value: ""},
			{Field: "metadata.operation", Ops: "=", Value: "update"},
		},
	}

	assert.False(t, f.Match(data))
}

func TestOrFilter(t *testing.T) {
	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp":          "2021-02-15T15:00:03.686239Z",
			"record-type":        "control",
			"operation":          "create-table",
			"partition-key-type": "task-id",
			"schema-name":        "",
			"sequence":           87,
			"table-name":         "awsdms_apply_exceptions",
			"active":             "true",
		},
	}

	f := &ORFilter{
		Conditions: []Condition{
			{Field: "metadata.record-type", Ops: "=", Value: "control"},
			{Field: "metadata.schema-name", Ops: "=", Value: ""},
			{Field: "metadata.operation", Ops: "~", Value: "create"},
		},
	}

	assert.True(t, f.Match(data))

	f = &ORFilter{
		Conditions: []Condition{
			{Field: "metadata.record-type", Ops: "=", Value: "control"},
			{Field: "metadata.schema-name", Ops: "=", Value: ""},
			{Field: "metadata.operation", Ops: "=", Value: "update"},
		},
	}

	assert.True(t, f.Match(data))

	f = &ORFilter{
		Conditions: []Condition{
			{Field: "metadata.record-type", Ops: "!=", Value: "control"},
			{Field: "metadata.schema-name", Ops: "!=", Value: ""},
			{Field: "metadata.operation", Ops: "=", Value: "update"},
		},
	}

	assert.False(t, f.Match(data))
}

func TestExpFilter(t *testing.T) {
	data := map[string]interface{}{
		"metadata": map[string]interface{}{
			"timestamp":          "2021-02-15T15:00:03.686239Z",
			"record-type":        "control",
			"operation":          "update",
			"partition-key-type": "task-id",
			"schema-name":        "",
			"sequence":           87,
			"table-name":         "awsdms_apply_exceptions",
			"active":             "true",
		},
	}

	f := &ExpFilter{Expression: "$metadata.record-type == 'control' && ($metadata.operation == 'create' || $metadata.operation == 'update') && $metadata.schema-name == ''"}
	assert.Nil(t, f.init())
	assert.True(t, f.Match(data))

	f = &ExpFilter{Expression: "$data == ''"}
	assert.Nil(t, f.init())
	assert.True(t, f.Match(data))

	f = &ExpFilter{Expression: "$. == 'string'"}
	assert.Nil(t, f.init())
	assert.True(t, f.Match("string"))
}
