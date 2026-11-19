package migrations

import (
	"embed"
	"sort"
	"strings"
)

//go:embed *.sql
var files embed.FS

func Schemas() ([]string, error) {
	entries, err := files.ReadDir(".")
	if err != nil {
		return nil, err
	}

	names := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	schemas := make([]string, 0, len(names))
	for _, name := range names {
		content, err := files.ReadFile(name)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, string(content))
	}
	return schemas, nil
}
