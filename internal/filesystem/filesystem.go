package filesystem

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/angelvargass/repository-provisioner/internal/utils"
)

func LoadFilesForArchetype(archetypesDirectory, archetype string) []string {
	rootPath := archetypesDirectory + archetype
	files := []string{}

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		files = append(files, strings.TrimPrefix(path, rootPath))

		return nil
	})

	utils.HandleError(fmt.Sprintf("error loading files for archetype: %s", archetype), err)
	return files
}
