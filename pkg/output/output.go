package output

import (
	"path"
	"strings"
)

func ParseFileName(p string, defaultFileName string) string {
	if strings.LastIndex(p, ".html") > 0 {
		return p
	}
	return path.Join(p, defaultFileName)
}
