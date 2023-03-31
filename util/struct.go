package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/gotoko/commonlib/constant"
)

func FieldExist(name string, val interface{}) bool {

	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		if value, ok := lookup(parts[0], val); ok {
			return FieldExist(parts[1], value)
		}
		return false
	}

	reflectValue := reflect.ValueOf(val)
	switch reflectValue.Kind() {
	// If the current context is a map, we'll look for a key in that map
	// that matches the name.
	case reflect.Map:
		item := reflectValue.MapIndex(reflect.ValueOf(name))
		return item.IsValid()
	// If the current context is a struct, we'll look for a property in that
	// struct that matches the name. In the near future I'd like to add
	// support for matching struct names to tags so we can use lower_case
	// names in our templates which makes it more mustache like.
	case reflect.Struct:
		field := reflectValue.FieldByName(name)
		if field.IsValid() {
			return true
		}
		method := reflectValue.MethodByName(name)
		if method.IsValid() {
			return true
		}

	case reflect.Slice, reflect.Array:
		idx, err := strconv.Atoi(name)
		if err != nil {
			return false
		}
		if idx >= reflectValue.Len() {
			return false
		}
		item := reflectValue.Index(idx)
		return item.IsValid()
	}

	return false
}

func SetValue(obj interface{}, key string, value interface{}) error {

	if strings.Contains(key, ".") {
		parts := strings.Split(key, ".")
		val, ok := Lookup(strings.Join(parts[:len(parts)-1], "."), obj)
		if ok {
			key = parts[len(parts)-1]
			obj = val
		}
	}

	reflectValue := reflect.ValueOf(obj)
	switch reflectValue.Kind() {

	case reflect.Ptr:
		if reflectValue.Elem().Kind() != reflect.Struct {
			return errors.New("should be pointer of struct or map")
		}
		field := reflectValue.Elem().FieldByName(key)
		if field.IsValid() && field.CanSet() {
			field.Set(reflect.ValueOf(value))
			//return setItem(field, value)
			return nil
		}
		return nil
		//return SetValue(reflectValue.Elem(), key, value)
	case reflect.Map:
		reflectValue.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
		return nil
		//}
	// If the current context is a struct, we'll look for a property in that
	// struct that matches the name. In the near future I'd like to add
	// support for matching struct names to tags so we can use lower_case
	// names in our templates which makes it more mustache like.
	case reflect.Struct:
		return errors.New("should be pointer of struct")
	}

	return errors.New("unsupported type")
}

func FindFieldByTag(obj interface{}, tag, key string) (string, error) {
	reflectType := reflect.TypeOf(obj)
	switch reflectType.Kind() {
	case reflect.Ptr:
		reflectType = reflectType.Elem()
		fallthrough
	case reflect.Struct:
		for i := 0; i < reflectType.NumField(); i++ {
			field := reflectType.Field(i)
			if ft := field.Tag.Get(tag); ft == key || strings.HasPrefix(ft, key+",") {
				return field.Name, nil
			}
		}
		return "", errors.New("field not found")
	case reflect.Map:
		return key, nil
	default:
		return "", errors.New("unsupported type")
	}
}

func FindFieldTypeByTag(obj interface{}, tag, key string) (reflect.Type, error) {
	reflectType := reflect.TypeOf(obj)
	switch reflectType.Kind() {
	case reflect.Ptr:
		reflectType = reflectType.Elem()
		fallthrough
	case reflect.Struct:
		for i := 0; i < reflectType.NumField(); i++ {
			field := reflectType.Field(i)
			if ft := field.Tag.Get(tag); ft == key || strings.HasPrefix(ft, key+",") {
				return field.Type, nil
			}
		}
		return nil, errors.New("field not found")
	default:
		return nil, errors.New("unsupported type")
	}
}

func IsMap(obj interface{}) bool {
	reflectType := reflect.TypeOf(obj)
	return reflectType.Kind() == reflect.Map
}

func IsStructOrPointerOf(obj interface{}) bool {
	return IsStruct(obj) || IsPointerOfStruct(obj)
}

func IsStruct(obj interface{}) bool {
	reflectType := reflect.TypeOf(obj)
	return reflectType.Kind() == reflect.Struct
}

func IsMapStringInterface(obj interface{}) bool {
	reflectType := reflect.TypeOf(obj)
	m := make(map[string]interface{})
	return reflectType == reflect.TypeOf(m) || reflectType == reflect.TypeOf(&m)
}

func IsTime(val interface{}) bool {
	reflectType := reflect.TypeOf(val)

	if reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}
	if reflectType == reflect.TypeOf(time.Time{}) {
		return true
	}
	return false
}

func IsPointerOfStruct(obj interface{}) bool {
	reflectType := reflect.TypeOf(obj)

	if reflectType.Kind() != reflect.Ptr {
		return false
	}

	if reflectType.Elem().Kind() != reflect.Struct {
		return false
	}

	return true
}

func IsPointerOfSlice(obj interface{}) bool {
	reflectType := reflect.TypeOf(obj)

	if reflectType.Kind() != reflect.Ptr {
		return false
	}

	if reflectType.Elem().Kind() != reflect.Slice {
		return false
	}

	return true
}

func IsSliceOrPointerOfSlice(obj interface{}) bool {
	return IsSlice(obj) || IsPointerOfSlice(obj)
}

func IsSlice(obj interface{}) bool {
	reflectType := reflect.TypeOf(obj)
	return reflectType.Kind() == reflect.Slice
}

func IsNumber(val interface{}) bool {
	reflectValue := reflect.ValueOf(val)
	switch reflectValue.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Float64, reflect.Float32:
		return true
	default:
		return false
	}
}

func GetSliceElem(val interface{}) (reflect.Value, error) {

	rv := reflect.ValueOf(val)

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Slice {
		return rv, errors.New("not a slice or pointer of slice")
	}

	return rv, nil

}

func CompareValue(src, dst interface{}, op string) (bool, error) {
	switch op {
	case constant.EQ:
		return fmt.Sprintf("%v", src) == fmt.Sprintf("%v", dst), nil
	case constant.SE:
		if reflect.ValueOf(src).Kind() != reflect.ValueOf(dst).Kind() {
			return false, nil
		}
		return fmt.Sprintf("%v", src) == fmt.Sprintf("%v", dst), nil
	case constant.GT:
		s, d, err := valsToNumber(src, dst)
		if err != nil {
			return false, err
		}
		return s > d, nil
	case constant.GE:
		s, d, err := valsToNumber(src, dst)
		if err != nil {
			return false, err
		}
		return s >= d, nil
	case constant.LT:
		s, d, err := valsToNumber(src, dst)
		if err != nil {
			return false, err
		}
		return s < d, nil
	case constant.LE:
		s, d, err := valsToNumber(src, dst)
		if err != nil {
			return false, err
		}
		return s <= d, nil
	case constant.RE:
		k, err := regexp.Match(fmt.Sprintf("%v", dst), []byte(fmt.Sprintf("%v", src)))
		if err != nil {
			return false, err
		}
		return k, nil
	case constant.NE:
		return fmt.Sprintf("%v", src) != fmt.Sprintf("%v", dst), nil
	case constant.SN:
		if reflect.ValueOf(src).Kind() != reflect.ValueOf(dst).Kind() {
			return true, nil
		}
		return fmt.Sprintf("%v", src) != fmt.Sprintf("%v", dst), nil
	case constant.IN:
		rv, err := GetSliceElem(dst)
		if err != nil {
			return false, err
		}
		for i := 0; i < rv.Len(); i++ {
			item := rv.Index(i)
			if item.IsValid() {
				if IsSliceOrPointerOfSlice(src) {
					sv, err := GetSliceElem(src)
					if err != nil {
						return false, nil
					}
					for j := 0; j < sv.Len(); j++ {
						sitem := sv.Index(j)
						if fmt.Sprintf("%v", item.Interface()) == fmt.Sprintf("%v", sitem.Interface()) {
							return true, nil
						}
					}
				} else {
					if fmt.Sprintf("%v", item.Interface()) == fmt.Sprintf("%v", src) {
						return true, nil
					}
				}

			}
		}

		return false, nil
	case constant.EVAL:
		script, ok := dst.(string)
		if !ok {
			return false, errors.New("eval script is not a string")
		}

		ev, err := NewEvaluator(script)
		if err != nil {
			return false, err
		}

		out, err := ev.Eval(src)
		if err != nil {
			return false, err
		}

		bout, ok := out.(bool)
		if !ok {
			return false, errors.New("eval script must return a boolean")
		}

		return bout, nil
	default:
		return false, errors.New("unsupported operand")
	}
}

func valsToNumber(src, dst interface{}) (float64, float64, error) {

	s, err := toNumber(src)
	if err != nil {
		return 0, 0, fmt.Errorf("error converting source to number %v", err)
	}
	d, err := toNumber(dst)
	if err != nil {
		return 0, 0, fmt.Errorf("error converting destination to number %v", err)
	}
	return s, d, nil
}

func toNumber(val interface{}) (float64, error) {
	switch v := val.(type) {
	case time.Time:
		return float64(v.UnixNano()), nil
	case *time.Time:
		return float64(v.UnixNano()), nil
	case time.Duration:
		return float64(v), nil
	default:
		if !IsNumber(val) {
			return 0, errors.New("value is not a number")
		}
		ns := fmt.Sprintf("%v", val)
		return strconv.ParseFloat(ns, 64)
	}
}

func Reverse(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

func ListTag(obj interface{}, tag string) ([]string, error) {
	reflectType := reflect.TypeOf(obj)
	switch reflectType.Kind() {
	case reflect.Ptr:
		reflectType = reflectType.Elem()
		fallthrough
	case reflect.Struct:
		out := make([]string, 0)
		for i := 0; i < reflectType.NumField(); i++ {
			field := reflectType.Field(i)
			if ft := field.Tag.Get(tag); ft != "" {
				out = append(out, ft)
			}
		}

		if len(out) == 0 {
			return nil, errors.New("empty tag")
		}

		return out, nil
	default:
		return nil, errors.New("unsupported type")
	}
}

func Stringify(obj interface{}) string {
	if obj == nil {
		return ""
	}

	switch v := obj.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, bool:
		return fmt.Sprintf("%v", obj)
	case []byte:
		return base64.StdEncoding.EncodeToString(v)
	default:
		b, err := json.Marshal(obj)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

func GetSliceLength(v interface{}) int {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.Len()
}

func GetSliceItem(v interface{}, i int) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	val := rv.Index(i)
	if val.IsValid() {
		return val.Interface()
	}
	return nil
}
