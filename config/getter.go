package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bondhan/commonlib/constant"
	"github.com/bondhan/commonlib/util"
	"gopkg.in/yaml.v2"
)

//Getter config getter interface
//go:generate mockery --with-expecter --name Getter

type Getter interface {
	Get(k string) interface{}
	GetString(k string) string
	GetBool(k string) bool
	GetInt(k string) int
	GetFloat64(k string) float64
	GetStringSlice(k string) []string
	GetStringMap(k string) map[string]interface{}
	GetStringMapString(k string) map[string]string
	Unmarshal(rawVal interface{}) error
}

type Decoder func(data []byte, v interface{}) error

func JsonDecoder(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func YamlDecoder(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

func getDecoder(ext string) Decoder {
	switch ext {
	case ".yaml":
		return YamlDecoder
	case ".json":
		return JsonDecoder
	default:
		return nil
	}
}

func Load(defaultConfig map[string]interface{}, uri string) (Getter, error) {
	if uri == "" {
		readEnvVar(defaultConfig, "")
		return NewEmbedConfig(defaultConfig), nil
	}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "env":
		readEnvVar(defaultConfig, u.Query().Get("prefix"))
	case "file":
		if err := readConfigFile(defaultConfig, u); err != nil {
			return nil, err
		}
		readEnvVar(defaultConfig, u.Query().Get("prefix"))
	default:
		return nil, errors.New("unsupported scheme")
	}

	return NewEmbedConfig(defaultConfig), nil
}

func readEnvVar(defaultConfig map[string]interface{}, prefix string) {
	for k, v := range defaultConfig {
		if v == nil {
			if val := os.Getenv(strings.ToUpper(prefix + k)); val != "" {
				defaultConfig[k] = val
			}
			continue
		}
		if obj, ok := v.(map[string]interface{}); ok {
			readEnvVar(obj, prefix+k+"_")
			defaultConfig[k] = obj
			continue
		}
		if obj, ok := v.(map[string]string); ok {
			for km := range obj {
				if val := os.Getenv(strings.ToUpper(prefix + k + "_" + km)); val != "" {
					obj[km] = val
				}
			}
			defaultConfig[k] = obj
			continue
		}

		if util.IsSliceOrPointerOfSlice(v) {
			l := util.GetSliceLength(v)
			arr := make([]interface{}, l)

			for i := 0; i < l; i++ {
				item := util.GetSliceItem(v, i)
				if item != nil {
					switch vi := item.(type) {
					case map[string]interface{}:
						key := prefix + k + "_" + fmt.Sprintf("%v", i) + "_"
						readEnvVar(vi, key)
						arr[i] = item
					case string:
						arr[i] = vi
						if val := os.Getenv(strings.ToUpper(prefix + k + "_" + fmt.Sprintf("%v", i))); val != "" {
							//obj[km] = val
							arr[i] = val
						}
					default:
						fmt.Println("not match")
					}
				}
			}
			defaultConfig[k] = arr
			continue
		}

		if util.IsStructOrPointerOf(v) {
			tmp := make(map[string]interface{})
			if err := util.DecodeJSON(v, &tmp); err != nil {
				fmt.Println(err)
				continue
			}
			readEnvVar(tmp, prefix+k+"_")
			defaultConfig[k] = tmp
			continue
		}

		if val := os.Getenv(strings.ToUpper(prefix + k)); val != "" {
			defaultConfig[k] = val
			continue
		}
	}
}

func readConfigFile(defaultConfig map[string]interface{}, uri *url.URL) error {
	path := filepath.Join(uri.Host, uri.Path)
	ext := filepath.Ext(path)
	environ := os.Getenv(constant.EnvKey)
	if environ != "" {
		environ = "." + environ
	}
	fname := strings.TrimSuffix(path, ext) + environ + ext

	var out map[string]interface{}

	decoder := getDecoder(ext)
	if decoder == nil {
		return errors.New("unsupported file format")
	}
	bfile, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}

	sfile := NewTemplate(string(bfile)).Render()

	if err := decoder([]byte(sfile), &out); err != nil {
		return err
	}
	mergeMap(defaultConfig, out)
	return nil
	//return mergo.Merge(&defaultConfig, out, mergo.WithOverride)
}

func mergeMap(dst, src map[string]interface{}) {
	for k, v := range dst {
		if v == nil {
			if val, ok := src[k]; ok {
				dst[k] = val
			}
			continue
		}

		if obj, ok := v.(map[string]interface{}); ok {
			if val, ok := src[k]; ok && len(obj) == 0 {
				dst[k] = val
				continue
			}
			if sobj, ok := util.LookupMap(k, src); ok {
				mergeMap(obj, sobj)
				dst[k] = obj
			}
			continue
		}

		if obj, ok := v.(map[string]string); ok {
			if val, ok := src[k]; ok && len(obj) == 0 {
				dst[k] = val
				continue
			}
			for km := range obj {
				if val, ok := util.LookupString(k+"."+km, src); ok {
					obj[km] = val
				}
			}
			dst[k] = obj
			continue
		}

		if util.IsStructOrPointerOf(v) {
			tmp := make(map[string]interface{})
			if err := util.DecodeJSON(v, &tmp); err != nil {
				fmt.Println(err)
				continue
			}
			dst[k] = tmp
			if sobj, ok := util.LookupMap(k, src); ok {
				mergeMap(tmp, sobj)
				dst[k] = tmp
			}
			continue
		}

		if util.IsSliceOrPointerOfSlice(v) {
			l := util.GetSliceLength(v)
			if val, ok := src[k]; ok && l == 0 {
				dst[k] = val
				continue
			}
			arr := make([]interface{}, l)

			for i := 0; i < l; i++ {
				item := util.GetSliceItem(v, i)
				key := k + "." + fmt.Sprintf("%v", i)
				if item != nil {
					switch vi := item.(type) {
					case map[string]interface{}:
						arr[i] = vi
						if val, ok := util.LookupMap(key, src); ok {
							mergeMap(vi, val)
							arr[i] = vi
						}

					case string:
						arr[i] = vi
						if val, ok := util.LookupString(key, src); ok {
							arr[i] = val
						}
					default:
						fmt.Println("not match")
					}
				}
			}
			dst[k] = arr
			continue
		}

		if val, ok := src[k]; ok {
			dst[k] = val
		}

	}
}

func EnvToStruct(obj interface{}) error {
	tags, err := util.ListTag(obj, "json")
	if err != nil {
		return err
	}

	conf := make(map[string]interface{})
	for _, t := range tags {
		if strings.Contains(t, ",") {
			t = strings.Split(t, ",")[0]
		}
		conf[t] = os.Getenv(strings.ToUpper(t))
	}

	return util.DecodeJSON(conf, obj)
}

func MergeConfig(obj interface{}, tag string) error {
	tags, err := util.ListTag(obj, tag)
	if err != nil {
		return err
	}

	for _, t := range tags {
		if strings.Contains(t, ",") {
			t = strings.Split(t, ",")[0]
		}

		if ev := os.Getenv(strings.ToUpper(t)); ev != "" {
			fn, err := util.FindFieldByTag(obj, tag, t)
			if err != nil {
				return err
			}

			if val, err := util.DecodeString(ev); err == nil {
				if err := util.SetValue(obj, fn, val); err != nil {
					return err
				}
			}

		}

	}

	return nil
}
