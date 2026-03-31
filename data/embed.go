package sqldata

import (
	"embed"
	"fmt"
	"path"
	"strings"
)

//go:embed migration/*.sql query/*.sql
var files embed.FS

func Query(name string) (string, error) {
	return read(path.Join("query", name))
}

func MustQuery(name string) string {
	query, err := Query(name)
	if err != nil {
		panic(err)
	}
	return query
}

func Migration(name string) (string, error) {
	return read(path.Join("migration", name))
}

func MustMigration(name string) string {
	migration, err := Migration(name)
	if err != nil {
		panic(err)
	}
	return migration
}

func Render(template string, values map[string]string) string {
	rendered := template
	for key, value := range values {
		rendered = strings.ReplaceAll(rendered, "{{"+key+"}}", value)
	}
	return rendered
}

func read(name string) (string, error) {
	content, err := files.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("read embedded sql %s: %w", name, err)
	}
	return string(content), nil
}
