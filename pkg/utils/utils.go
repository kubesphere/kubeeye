package utils

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

func ArrayFind(s string, sub []string) (int, bool) {

	index, b, _ := ArrayFinds[string](sub, func(m string) bool {
		return s == m
	})
	return index, b
}

func ArrayFilter[T any](filterData []T, filter func(v T) bool) ([]T, []T) {
	var where []T
	var notWhere []T
	for _, v := range filterData {
		if filter(v) {
			where = append(where, v)
		} else {
			notWhere = append(notWhere, v)
		}
	}
	return where, notWhere
}

func ArrayFinds[T any](maps []T, f func(m T) bool) (int, bool, T) {
	var val T
	for i, t := range maps {

		if f(t) {
			return i, true, t
		}
	}
	return -1, false, val
}

func ArrayDeduplication(sub []string) []string {
	var newSub []string
	for _, s := range sub {
		if _, b := ArrayFind(s, newSub); !b {
			newSub = append(newSub, s)
		}
	}

	return newSub
}

func SliceRemove(s string, o interface{}) interface{} {
	switch t := o.(type) {
	case []string:
		if i, b := ArrayFind(s, t); b {
			t = append(t[:i], t[i+1:]...)
		}
		return t
	case []int:
		return nil
	}
	return nil
}

func BoolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
func StringToBool(b string) bool {
	return b == "true"
}

func IsEmptyValue(val interface{}) bool {
	if val == nil {
		return true
	}
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.String:
		return len(strings.TrimSpace(v.String())) == 0
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	case reflect.Struct:
		return v.IsZero()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	default:
		return false
	}
}

func ArrayStructToArrayMap(obj interface{}) ([]map[string]interface{}, error) {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	err = json.Unmarshal(marshal, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func StructToMap(obj interface{}) map[string]interface{} {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return nil
	}
	var result map[string]interface{}
	err = json.Unmarshal(marshal, &result)
	if err != nil {
		return nil
	}
	return result
}

func MapToStruct[T any](maps ...map[string]interface{}) []T {
	var result []T
	marshal, err := json.Marshal(maps)
	if err != nil {
		return nil
	}
	err = json.Unmarshal(marshal, &result)
	if err != nil {
		return nil
	}
	return result
}

// ParseDateTime parse time string to time.Time (2006-01-02 15:04:05)
func ParseDateTime(timeStr *string) (time.Time, error) {
	parse, err := time.Parse("2006-01-02 15:04:05", *timeStr)
	if err != nil {
		return parse, err
	}
	return parse, nil

}

func MergeMap[T any](maps ...map[string]T) map[string]T {
	result := make(map[string]T)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
func MapValConvert[T any](mapV1 map[string]interface{}) map[string]T {
	var newMap = make(map[string]T)
	for k, v := range mapV1 {
		newMap[k] = v.(T)
	}
	return newMap
}

func ArrayValConvert[T any](arrayV1 []interface{}) []T {
	var newVal []T
	for _, v := range arrayV1 {
		newVal = append(newVal, v.(T))
	}
	return newVal
}

func DeepCopyMap[T any](obj map[string]T) map[string]T {
	if obj == nil {
		return nil
	}
	result := make(map[string]T, len(obj))
	for k, v := range obj {
		result[k] = v
	}
	return result
}
