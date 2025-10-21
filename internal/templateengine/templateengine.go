package templateengine

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func RenderTemplates(archetypesDirectory, archetypePath string, data map[string]any) ([]ArchetypeFile, error) {
	rootPath := filepath.Join(archetypesDirectory, archetypePath)
	var results []ArchetypeFile

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if strings.HasSuffix(d.Name(), ".go.tmpl") {
			tmpl, err := template.ParseFiles(path)
			if err != nil {
				return err
			}

			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, filepath.Base(path), data); err != nil {
				return err
			}

			results = append(results, ArchetypeFile{
				Name:    strings.TrimSuffix(rel, ".go.tmpl"),
				Content: buf.Bytes(),
			})
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		results = append(results, ArchetypeFile{
			Name:    rel,
			Content: b,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}
	return results, nil
}
