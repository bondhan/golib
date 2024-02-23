package log

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestLogAdditionalFields(t *testing.T) {
	const key = "header"
	fields := map[string]interface{}{
		"header": "value",
	}

	config := LogConfig{}
	LogAdditionalFields(fields)(&config)
	if _, ok := config.Fields[key]; !ok {
		t.Errorf("fields are wrong, expected %v but got '%v'", fields, config.Fields)
	}
}

func TestNewLogInstanceTextFormatter(t *testing.T) {
	const (
		logFile = "file.log"
	)

	tests := []struct {
		name       string
		logOptions []LogOption
		withLog    bool
		want       string
	}{
		{
			name: "default configuration",
			logOptions: []LogOption{
				LogAdditionalFields(map[string]interface{}{"user_id": "123"}),
			},
			withLog: false,
			want:    "user_id=123",
		},
		{
			name: "text formatter with log",
			logOptions: []LogOption{
				LogToFile(logFile),
			},
			withLog: true,
			want:    `"msg":"test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogInstance(tt.logOptions...)
			logger.Out = &buf

			logger.Warn("test")

			if !tt.withLog {
				assert.Contains(t, buf.String(), tt.want)
			} else {
				dt := time.Now().UTC()
				fName := "logs/" + dt.Format("20060102") + "_" + logFile
				file, err := os.Open(fName)
				require.NoError(t, err)
				defer func() {
					file.Close()
					os.Remove(fName)
				}()

				data, err := io.ReadAll(file)
				require.NoError(t, err)

				assert.Contains(t, string(data), tt.want)
			}
		})
	}
}

func TestNewLogInstanceJSONFormatter(t *testing.T) {
	const (
		logFile = "file.log"
	)

	tests := []struct {
		name       string
		logOptions []LogOption
		withLog    bool
		want       string
	}{
		{
			name: "json formatter configuration",
			logOptions: []LogOption{
				IsJSONFormatter(true),
				LogAdditionalFields(map[string]interface{}{"user_id": "123"}),
			},
			withLog: false,
			want:    `"user_id": "123"`,
		},
		{
			name: "text formatter with log",
			logOptions: []LogOption{
				IsJSONFormatter(true),
				LogToFile(logFile),
			},
			withLog: true,
			want:    `"msg":"test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogInstance(tt.logOptions...)
			logger.Out = &buf

			logger.Warn("test")

			if !tt.withLog {
				assert.Contains(t, buf.String(), tt.want)

				m := make(map[string]interface{})
				err := json.Unmarshal([]byte(buf.String()), &m)
				assert.NoError(t, err)
			} else {
				dt := time.Now().UTC()
				fName := "logs/" + dt.Format("20060102") + "_" + logFile
				file, err := os.Open(fName)
				require.NoError(t, err)
				defer func() {
					file.Close()
					os.Readlink(fName)
				}()

				data, err := io.ReadAll(file)
				require.NoError(t, err)

				assert.Contains(t, string(data), tt.want)
			}
		})
	}
}

func TestGetLevel(t *testing.T) {
	tests := []struct {
		name string
		want string
		lvl  logrus.Level
	}{
		{
			name: "default is PANIC",
			want: "PANIC",
		},
		{
			name: "DEBUG",
			want: "DEBUG",
			lvl:  logrus.DebugLevel,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger.SetLevel(tt.lvl)
			assert.Equalf(t, tt.want, GetLevel(), "GetLevel()")
		})
	}
}
