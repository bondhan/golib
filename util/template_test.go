package util

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplate(t *testing.T) {
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

	tmp := NewTemplate("http://localhost:8080/{metadata.record-type}/{metadata.operation}/{metadata.sequence}")
	r := tmp.Render(data)
	assert.Equal(t, "http://localhost:8080/control/create-table/87", r)

}

func TestTemplateStringJSON(t *testing.T) {
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

	b, err := json.Marshal(data)
	require.Nil(t, err)
	require.NotNil(t, b)
	tmp := NewTemplate("http://localhost:8080/{metadata.record-type}/{metadata.operation}/{metadata.sequence}")
	r := tmp.RenderFromStringJSON(string(b))
	assert.Equal(t, "http://localhost:8080/control/create-table/87", r)

}

func TestDefaultValue(t *testing.T) {
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

	tmp := NewTemplate("http://localhost:8080/{metadata.record-type}/{metadata.operation}/{metadata.id|10}")
	r := tmp.Render(data)
	assert.Equal(t, "http://localhost:8080/control/create-table/10", r)

}

func TestDefaultValueStringJSON(t *testing.T) {
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

	b, err := json.Marshal(data)
	require.Nil(t, err)
	require.NotNil(t, b)
	tmp := NewTemplate("http://localhost:8080/{metadata.record-type}/{metadata.operation}/{metadata.id|10}")
	r := tmp.RenderFromStringJSON(string(b))
	assert.Equal(t, "http://localhost:8080/control/create-table/10", r)

}

func TestExpression(t *testing.T) {
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

	tmp := NewTemplate("http://localhost:8080/{metadata.record-type}/{metadata.operation}/{[inow()]}")
	r := tmp.Render(data)
	ts := time.Now().Unix()
	assert.Equal(t, "http://localhost:8080/control/create-table/"+fmt.Sprintf("%v", ts), r)

}

func TestExpressionStringJSON(t *testing.T) {
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

	b, err := json.Marshal(data)
	require.Nil(t, err)
	require.NotNil(t, b)
	tmp := NewTemplate("http://localhost:8080/{metadata.record-type}/{metadata.operation}/{[inow()]}")
	r := tmp.RenderFromStringJSON(string(b))
	ts := time.Now().Unix()
	assert.Equal(t, "http://localhost:8080/control/create-table/"+fmt.Sprintf("%v", ts), r)

}

func TestUnquoteValue(t *testing.T) {
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

	tmp := NewTemplate(`http://localhost:8080/{metadata.record-type}/{metadata.operation}/"{'metadata.id|10}"`)
	r := tmp.Render(data)
	assert.Equal(t, "http://localhost:8080/control/create-table/10", r)

}

func TestUnquoteValueStringJSON(t *testing.T) {
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
	b, err := json.Marshal(data)
	require.Nil(t, err)
	require.NotNil(t, b)
	tmp := NewTemplate(`http://localhost:8080/{metadata.record-type}/{metadata.operation}/"{'metadata.id|10}"`)
	r := tmp.RenderFromStringJSON(string(b))
	assert.Equal(t, "http://localhost:8080/control/create-table/10", r)

}

func TestEvalTemplate(t *testing.T) {

	str := "SELECT * FROM `peak-nimbus-307910.ops_function.getStockWarehouse`('{[ trim(ftime(now(), '2006-01-02 ')) ]}','06:00') LIMIT 5"

	tmp := NewTemplate(str)

	r := tmp.Render(nil)

	td := time.Now().Format("2006-01-02")

	assert.Equal(t, "SELECT * FROM `peak-nimbus-307910.ops_function.getStockWarehouse`('"+td+"','06:00') LIMIT 5", r)
}
