package util

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// IRegoRulesReader

func ListRegoRuleFileName(fsys fs.FS) []string {

	files := make([]string, 0)
	visitRegoFile := func(fp string, fi os.DirEntry, err error) error {
		if err != nil {
			fmt.Println(err) // can't walk here,
			return nil       // but continue walking elsewhere
		}
		if !fi.IsDir() && strings.HasSuffix(fi.Name(), ".rego") {
			files = append(files, fp)
		}
		return nil
	}
	err := fs.WalkDir(fsys, ".", visitRegoFile)

	if err != nil {
		fmt.Println("Failed to list the dir of rego rule files.")
		os.Exit(1)
	}
	return files
}

// GetRegoRulesfiles get rego rules , put it into pointer of RegoRulesList
func GetRegoRules(regoRulesFiles []string, reader fs.FS) []string {
	regoRulesContent := make([]string, 0)
	for _, file := range regoRulesFiles {
		getregoRule, _ := readRegoFile(reader, file)
		regoRulesContent = append(regoRulesContent, string(getregoRule))
	}
	return regoRulesContent
}

// readRegoFile
func readRegoFile(fsys fs.FS, name string) ([]byte, error) {
	file, err := fs.ReadFile(fsys, name)
	if err != nil {
		fmt.Println("Failed to read the dir of rego rule files.")
		os.Exit(1)
	}
	return file, err
}
