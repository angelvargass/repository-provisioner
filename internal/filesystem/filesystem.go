package filesystem

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/angelvargass/repository-provisioner/internal/utils"
)

func LoadFilesForArchetype(archetypesDirectory, archetypePath string) []string {
	rootPath := archetypesDirectory + archetypePath
	files := []string{}

	err := filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			files = append(files, strings.TrimPrefix(path, rootPath))
		}

		return nil
	})

	utils.HandleError(fmt.Sprintf("error loading files for archetype: %s", archetypePath), err)
	return files
}

func LoadTemplateFile(filePath string, data any) []byte {
	tmpl, err := template.ParseFiles(filePath)
	utils.HandleError(fmt.Sprintf("error loading template file: %s", filePath), err)

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	utils.HandleError(fmt.Sprintf("error executing template file: %s", filePath), err)

	return buf.Bytes()
}
