package utils

func ArrayFind(s string, sub []string) bool {
	for i := range sub {
		if sub[i] == s {
			return true
		}
	}
	return false
}

func ArrayDeduplication(sub []string) []string {
	var newSub []string
	for _, s := range sub {
		if !ArrayFind(s, newSub) {
			newSub = append(newSub, s)
		}
	}
	return newSub
}
