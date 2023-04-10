package utils

import "fmt"

func ArrayFind(s string, sub []string) (int, bool) {

	index, b, _ := ArrayFinds(sub, func(m string) bool {
		return s == m
	})
	return index, b
}

func ArrayFinds(maps interface{}, f func(m string) bool) (int, bool, interface{}) {
	switch maps.(type) {
	case []string:
		strings := maps.([]string)
		for i := range strings {
			if b := f(strings[i]); b {
				return i, b, strings[i]
			}

		}
	case map[string]interface{}:
		m := maps.(map[string]interface{})
		for key, val := range m {
			if b := f(key); b {
				return 0, b, val
			}
		}
	default:
		fmt.Printf("%T\n", maps)
	}
	return -1, false, nil
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
	switch o.(type) {
	case []string:
		stringArray := o.([]string)
		if i, b := ArrayFind(s, stringArray); b {
			stringArray = append(stringArray[:i], stringArray[i+1:]...)
		}
		return stringArray
	}
	return nil
}
