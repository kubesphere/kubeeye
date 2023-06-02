package utils

import (
	"bufio"
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"strings"
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

func DiffString(base1 string, base2 string) []string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(base1, base2, false)
	scan := bufio.NewScanner(strings.NewReader(dmp.DiffPrettyText(diffs)))
	lineNum := 1
	var isseus []string
	for scan.Scan() {
		line := scan.Text()
		if strings.Contains(line, "\x1b[3") {
			isseus = append(isseus, fmt.Sprintf("%d行 %s\n", lineNum, line))
		}
		lineNum++
	}
	return isseus
}
func FormatBool(b *bool) string {
	if b == nil {
		return "false"
	}
	if *b {
		return "true"
	}
	return "false"
}
