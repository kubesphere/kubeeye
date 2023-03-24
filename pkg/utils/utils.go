package utils

func ArrayFind(s string, sub []string) (int, bool) {
	for i := range sub {
		if sub[i] == s {
			return i, true
		}
	}
	return -1, false
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
