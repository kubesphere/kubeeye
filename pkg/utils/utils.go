package utils

func ArrayFind(s string, sub []string) bool {
	for i := range sub {
		if sub[i] == s {
			return true
		}
	}
	return false
}
