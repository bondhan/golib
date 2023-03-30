package log

import (
	"testing"
)

func TestLogToFile(t *testing.T) {
	const filename = "file.log"
	config := LogConfig{}
	LogToFile(filename)(&config)
	if config.LogFileName != filename {
		t.Errorf("LogToFile did not set LogFileName correctly, expected %s but got '%s'", filename, config.LogFileName)
	}
}

func TestIsProduction(t *testing.T) {
	const isProd = true

	config := LogConfig{}
	IsProduction(isProd)(&config)
	if config.IsProduction != isProd {
		t.Errorf("IsProduction is wrong, expected %t but got '%t'", isProd, config.IsProduction)
	}
}

func TestIsJSONFormatter(t *testing.T) {
	const isJSON = true

	config := LogConfig{}
	IsJSONFormatter(isJSON)(&config)
	if config.IsJSON != isJSON {
		t.Errorf("IsJSONFormatter is wrong, expected %t but got '%t'", isJSON, config.IsJSON)
	}
}
