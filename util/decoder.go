package util

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"time"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	DateTimeFormat      = "2006-01-02T15:04:05"
	DateTimeSpaceFormat = "2006-01-02 15:04:05"
	DateFormat          = "2006-01-02"
)

func Decode(input, output interface{}) error {
	return decodeVal(input, output, "", customTagDecoder)
	//return decode(input, defaultDecoderConfig(output))
}

func DecodeJSON(input, output interface{}) error {
	return decodeVal(input, output, "json", customTagDecoder)
	//return decode(input, customTagDecoder(output, "json"))
}

func DecodeTag(input, output interface{}, tag string) error {
	return decodeVal(input, output, tag, customTagDecoder)
	//return decode(input, customTagDecoder(output, tag))
}

func decodeVal(input, output interface{}, tag string, fn func(interface{}, string) *mapstructure.DecoderConfig) error {

	if IsMapStringInterface(output) && IsStructOrPointerOf(input) {

		s := structs.New(input)
		s.TagName = tag

		if reflect.ValueOf(output).Kind() == reflect.Ptr {
			if reflect.ValueOf(output).Elem().IsNil() {
				return errors.New("map should not nil")
			}
			s.FillMap(*output.(*map[string]interface{}))
			return nil
		}

		if reflect.ValueOf(output).IsNil() {
			return errors.New("map should not nil")
		}

		s.FillMap(output.(map[string]interface{}))

		return nil
	}

	return decode(input, fn(output, tag))
}

func decode(input interface{}, config *mapstructure.DecoderConfig) error {
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

func customTagDecoder(output interface{}, tag string) *mapstructure.DecoderConfig {
	return structDecoder(output, tag, mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		ToTimeHookFunc(""),
		FromTimeHookFunc(),
		FromStringHook(),
		//FromTimeHook(),
	))
}

func structDecoder(output interface{}, tag string, hook mapstructure.DecodeHookFunc) *mapstructure.DecoderConfig {
	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook:       hook,
	}

	if tag != "" {
		c.TagName = tag
	}

	return c
}

func ToTimeHookFunc(format string) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if t != reflect.TypeOf(time.Time{}) {
			return data, nil
		}

		if f == reflect.TypeOf(primitive.DateTime(0)) {
			return (data.(primitive.DateTime)).Time(), nil
		}

		switch f.Kind() {
		case reflect.String:
			if format != "" {
				return time.Parse(format, data.(string))
			}
			return parseTime(data.(string))
		case reflect.Float64:
			return getTime(int64(data.(float64))), nil
		case reflect.Int64:
			return getTime(data.(int64)), nil
		default:
			return data, nil
		}
		// Convert it by parsing
	}
}

func FromTimeHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		if f != reflect.TypeOf(time.Time{}) && f != reflect.PtrTo(reflect.TypeOf(time.Time{})) {
			return data, nil
		}

		tf, ok := data.(time.Time)

		if !ok {
			tmp, ok := data.(*time.Time)
			if !ok {
				return nil, errors.New("error converting time")
			}
			tf = *tmp
		}

		if t == reflect.TypeOf(time.Time{}) {
			return tf, nil
		}

		switch t.Kind() {
		case reflect.String:
			return tf.Format(time.RFC3339), nil
		case reflect.Float64:
			return float64(tf.UnixNano()), nil
		case reflect.Int64:
			return tf.UnixNano(), nil
		default:
			return tf.Format(time.RFC3339), nil
		}
		// Convert it by parsing
	}
}

func FromStringHook() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {

		if f != reflect.TypeOf("") {
			return data, nil
		}

		switch t.Kind() {
		case reflect.String:
			return data.(string), nil
		case reflect.Int:
			v, err := strconv.ParseInt(data.(string), 10, 64)
			if err != nil {
				return nil, err
			}
			return int(v), nil
		case reflect.Int32:
			v, err := strconv.ParseInt(data.(string), 10, 32)
			if err != nil {
				return nil, err
			}
			return int32(v), nil
		case reflect.Int64:
			return strconv.ParseInt(data.(string), 10, 64)
		case reflect.Uint:
			v, err := strconv.ParseUint(data.(string), 10, 64)
			if err != nil {
				return nil, err
			}
			return uint(v), nil
		case reflect.Uint32:
			v, err := strconv.ParseUint(data.(string), 10, 32)
			if err != nil {
				return nil, err
			}
			return uint32(v), nil
		case reflect.Uint64:
			return strconv.ParseUint(data.(string), 10, 64)
		case reflect.Float32:
			v, err := strconv.ParseFloat(data.(string), 32)
			if err != nil {
				return nil, err
			}
			return float32(v), nil
		case reflect.Float64:
			return strconv.ParseFloat(data.(string), 64)
		case reflect.Bool:
			return strconv.ParseBool(data.(string))
		case reflect.Map:
			out := make(map[string]interface{})
			if err := json.Unmarshal([]byte(data.(string)), &out); err != nil {
				return data, nil
			}
			return out, nil
		default:
			return data, nil
		}
	}
}

func DecodeString(str string) (interface{}, error) {
	if v, err := strconv.ParseBool(str); err == nil {
		return v, nil
	}

	if v, err := strconv.ParseFloat(str, 64); err == nil {
		if v == float64(int64(v)) {
			return int64(v), nil
		}
		return v, nil
	}

	return str, nil
}

func parseTime(v string) (time.Time, error) {
	return DateStringToTime(v)
}
