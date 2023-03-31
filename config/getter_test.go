package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefault(t *testing.T) {
	def := map[string]interface{}{
		"env":     "dev",
		"address": "localhost",
		"port":    "8080",
		"number":  10,
	}

	conf, err := Load(def, "")
	assert.NotNil(t, conf)
	assert.Nil(t, err)

	assert.Equal(t, def["env"], conf.GetString("env"))
	assert.Equal(t, def["number"], conf.GetInt("number"))
	assert.Equal(t, "", conf.GetString("any"))
}

func TestLoadEnvVar(t *testing.T) {
	type Endpoint struct {
		Path string `json:"path"`
		Host string `json:"host"`
	}

	type Server struct {
		Name string `json:"name"`
	}
	type Sender struct {
		Type   string      `json:"type"`
		Config interface{} `json:"config"`
	}

	type EventConfig struct {
		EventMap map[string]string `json:"event_map"`
	}

	type Logger struct {
		Sender      Sender      `json:"sender"`
		EventConfig EventConfig `json:"event_config"`
	}

	type MyConf struct {
		Env       string     `json:"env"`
		Address   string     `json:"address"`
		Port      string     `json:"port"`
		Number    int        `json:"number"`
		Server    Server     `json:"server"`
		Logger    Logger     `json:"logger"`
		Endpoints []Endpoint `json:"endpoints"`
		Roles     []string   `json:"roles"`
		Backend   Endpoint   `json:"backend"`
	}

	def := map[string]interface{}{
		"env":     "dev",
		"address": "localhost",
		"port":    "8080",
		"number":  10,
		"server": map[string]interface{}{
			"name": "test",
		},
		"logger": map[string]interface{}{
			"sender": map[string]interface{}{
				"type": "logger",
				"config": map[string]interface{}{
					"schema":     "gcppubsub://projects/myproject/topics/",
					"credential": "google-cred.json",
				},
			},
			"event_config": map[string]interface{}{
				"event_map": map[string]string{
					"command_log": "s-telebot-command-log",
				},
			},
		},
		"endpoints": []map[string]interface{}{
			{
				"path": "/v1/hello",
				"host": "localhost:8080",
			},
			{
				"path": "/v1/me",
				"host": "localhost:3000",
			},
		},
		"roles": []string{"admin", "user"},
		"backend": Endpoint{
			Host: "localhost:8181",
			Path: "/v1/*",
		},
	}

	os.Setenv("ENV", "development")
	os.Setenv("ADDRESS", "127.0.0.1")
	os.Setenv("NUMBER", "100")
	os.Setenv("SERVER_NAME", "local")
	os.Setenv("LOGGER_SENDER_TYPE", "pubsub")
	os.Setenv("LOGGER_SENDER_CONFIG_SCHEMA", "mem://")
	os.Setenv("LOGGER_EVENT_CONFIG_EVENT_MAP_COMMAND_LOG", "mytopic")
	os.Setenv("ENDPOINTS_0_HOST", "localhost:8000")
	os.Setenv("ENDPOINTS_1_HOST", "localhost:8000")
	os.Setenv("ROLES_0", "root")
	os.Setenv("BACKEND_HOST", "localhost:5000")
	conf, err := Load(def, "")
	assert.NotNil(t, conf)
	assert.Nil(t, err)

	var myconf MyConf
	assert.Nil(t, conf.Unmarshal(&myconf))

	assert.Equal(t, "development", conf.GetString("env"))
	assert.Equal(t, "127.0.0.1", conf.GetString("address"))
	assert.Equal(t, 100, conf.GetInt("number"))
	assert.Equal(t, "local", myconf.Server.Name)
	assert.Equal(t, "pubsub", myconf.Logger.Sender.Type)
	assert.Equal(t, "mem://", myconf.Logger.Sender.Config.(map[string]interface{})["schema"])
	assert.Equal(t, "mytopic", myconf.Logger.EventConfig.EventMap["command_log"])
	assert.Equal(t, "localhost:8000", myconf.Endpoints[0].Host)
	assert.Equal(t, "root", myconf.Roles[0])
	assert.Equal(t, "localhost:5000", myconf.Backend.Host)

	os.Unsetenv("ENV")
	os.Unsetenv("ADDRESS")
	os.Unsetenv("NUMBER")
	os.Unsetenv("SERVER_NAME")
	os.Unsetenv("LOGGER_SENDER_TYPE")
	os.Unsetenv("LOGGER_SENDER_CONFIG_SCHEMA")
	os.Unsetenv("LOGGER_EVENT_CONFIG_EVENT_MAP_COMMAND_LOG")
	os.Unsetenv("ENDPOINTS_0_HOST")
	os.Unsetenv("ENDPOINTS_1_HOST")
	os.Unsetenv("ROLES_0")
	os.Unsetenv("BACKEND_HOST")
}

func TestConfigFile(t *testing.T) {
	//t.Skip()
	confstr := `{"env" : "testing","port" : "8000","address" : "testhost","number": 99,"type" : "file","server":{ "cluster": { "type":"slave"}}}`
	err := ioutil.WriteFile("./config.json", []byte(confstr), 0644)
	require.Nil(t, err)

	def := map[string]interface{}{
		"env":     "dev",
		"address": "localhost",
		"port":    "8080",
		"number":  10,
		"server": map[string]interface{}{
			"host": "localhost",
			"port": "8000",
			"cluster": map[string]interface{}{
				"ID":   1,
				"type": "master",
			},
		},
	}

	conf, err := Load(def, "file://config.json")
	require.NotNil(t, conf)
	assert.Nil(t, err)

	assert.Equal(t, "testing", conf.GetString("env"))
	assert.Equal(t, "testhost", conf.GetString("address"))
	assert.Equal(t, 99, conf.GetInt("number"))
	assert.Equal(t, "", conf.GetString("type"))
	assert.Equal(t, "8000", conf.GetString("server.port"))
	assert.Equal(t, "slave", conf.GetString("server.cluster.type"))

	os.Remove("./config.json")
}

func TestEnvToStruct(t *testing.T) {
	type myconfig struct {
		Name        string `json:"name"`
		ServiceName string `json:"service_name"`
	}

	os.Setenv("NAME", "testing")
	os.Setenv("SERVICE_NAME", "golib")

	var conf myconfig

	err := EnvToStruct(&conf)
	require.Nil(t, err)

	assert.Equal(t, "testing", conf.Name)
	assert.Equal(t, "golib", conf.ServiceName)
}

func TestConfigTemplateFile(t *testing.T) {
	//t.Skip()
	confstr := `{"env" : "testing","port" : "{$PORT}","address" : "{$HOST}", "logger" : { "level" : "{$LOG_LEVEL}", "duration" : {$DURATION} }}`
	err := ioutil.WriteFile("./config.tmp.json", []byte(confstr), 0644)
	require.Nil(t, err)

	def := map[string]interface{}{
		"env":     "dev",
		"address": "localhost",
		"port":    "8080",
		"number":  10,
		"logger": map[string]interface{}{
			"level":    "INFO",
			"duration": 10,
		},
	}

	os.Setenv("PORT", "3000")
	os.Setenv("HOST", "localtest")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("DURATION", "100")

	conf, err := Load(def, "file://config.tmp.json")
	require.NotNil(t, conf)
	assert.Nil(t, err)

	assert.Equal(t, "3000", conf.GetString("port"))
	assert.Equal(t, "localtest", conf.GetString("address"))
	assert.Equal(t, "DEBUG", conf.Get("logger.level"))
	assert.Equal(t, 100, conf.GetInt("logger.duration"))

	os.Remove("./config.tmp.json")
}

func TestConfigStruct(t *testing.T) {
	type Endpoint struct {
		Path string `json:"path"`
		Host string `json:"host"`
	}

	type Server struct {
		Name string `json:"name"`
	}
	type Sender struct {
		Type   string                 `json:"type"`
		Config map[string]interface{} `json:"config"`
	}

	type Logger struct {
		Sender Sender `json:"sender"`
	}

	def := map[string]interface{}{
		"env":     "dev",
		"address": "localhost",
		"port":    "8080",
		"number":  10,
		"server": &Server{
			Name: "testserver",
		},
		"logger": &Logger{},
		"roles":  []string{"admin", "user"},
		"backend": Endpoint{
			Host: "localhost:8181",
			Path: "/v1/*",
		},
	}

	confstr := `{"env" : "testing","port" : "8000","address" : "testhost","number": 99,"server" : {"name" : "myserver"}, "logger" : {"sender" : {"type" : "pubsub", "config" : {"dir" : "ASC"}}}}`
	err := ioutil.WriteFile("./config.struct.json", []byte(confstr), 0644)
	require.Nil(t, err)

	conf, err := Load(def, "file://config.struct.json")
	require.NotNil(t, conf)
	assert.Nil(t, err)

	assert.Equal(t, "testing", conf.GetString("env"))
	assert.Equal(t, "myserver", conf.GetString("server.name"))
	assert.Equal(t, "localhost:8181", conf.GetString("backend.host"))
	assert.Equal(t, "ASC", conf.GetString("logger.sender.config.dir"))

	os.Remove("./config.struct.json")
}
