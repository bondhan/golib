package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConfigTemplate_Render(t *testing.T) {
	// Template string template
	os.Setenv("DEFINED", "some_value")

	assert.Equal(t, "{'config_A': 'some_value', 'config_B': ''}", NewTemplate("{'config_A': '{$DEFINED}', 'config_B': '{$NOT_DEFINED}'}").Render())
	assert.Equal(t, "{'config_A': 'some_value', 'config_B': 'BAR'}", NewTemplate("{'config_A': '{$DEFINED}', 'config_B': '{$NOT_DEFINED|BAR}'}").Render())
	assert.Equal(t, "{'config_A': 'some_value', 'config_B': 'FOO'}", NewTemplate("{'config_A': '{$DEFINED}', 'config_B': 'FOO'}").Render())

}
