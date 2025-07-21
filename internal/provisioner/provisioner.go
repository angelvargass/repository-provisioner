package provisioner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/angelvargass/repository-provisioner/internal/filesystem"
	"github.com/angelvargass/repository-provisioner/internal/gh"
	"github.com/angelvargass/repository-provisioner/internal/utils"
)

const (
	GoTemplateExtension = ".go.tmpl"
)

var availableArchetypeSubPaths = []string{"golang/"}

func New(logger *slog.Logger, ghClient *gh.Github) *Provisioner {
	return &Provisioner{
		Logger:   logger.With(slog.String("internal", "provisioner")),
		GHClient: *ghClient,
	}
}

func (p *Provisioner) ProvisionRepository(ctx context.Context, owner, repoName, archetypesDirectory, archetype string) error {
	p.Logger.Info("provisioning repository", slog.String("owner", owner), slog.String("repoName", repoName))

	archetypeIndex := slices.IndexFunc(availableArchetypeSubPaths, func(archetypeSubPath string) bool {
		return strings.Contains(archetypeSubPath, archetype)
	})

	if archetypeIndex == -1 {
		return fmt.Errorf(fmt.Sprintf("archetype %s does not exists", archetype), "archetype not found")
	}

	// _, err := p.GHClient.CreateRepository(ctx, "", repoName)
	// utils.HandleError("failed to create repository", err)

	p.Logger.Info("creating default branch", slog.String("branch name", "init"))
	ref, err := p.GHClient.CreateBranch(ctx, owner, repoName, "init")
	utils.HandleError("failed to create initial branch", err)

	archetypeFiles := filesystem.LoadFilesForArchetype(archetypesDirectory, availableArchetypeSubPaths[archetypeIndex])
	p.Logger.Info("committing archetype to created branch")
	for _, file := range archetypeFiles {
		parsedFileName := file
		replacingFileSHA := ""
		contents, err := os.ReadFile(archetypesDirectory + availableArchetypeSubPaths[archetypeIndex] + file)
		utils.HandleError("failed to load file contents", err)

		if strings.Contains(file, GoTemplateExtension) {
			parsedFileName = strings.TrimSuffix(file, GoTemplateExtension)
			contents = filesystem.LoadTemplateFile(archetypesDirectory+availableArchetypeSubPaths[archetypeIndex]+file, map[string]any{
				"RepositoryName": repoName,
				"DefaultBranch":  "main",
			})
		}

		fileContent, _, err := p.GHClient.GetRepositoryContent(ctx, owner, repoName, parsedFileName, "")
		utils.HandleError(fmt.Sprintf("failed to fetch content for file %s", parsedFileName), err)

		if fileContent != nil {
			replacingFileSHA = *fileContent.SHA
		}

		p.Logger.Debug("commiting file", slog.String("branch name", *ref.Ref), slog.String("file path", file))
		_, err = p.GHClient.CreateOrUpdateFile(ctx, owner, repoName, *ref.Ref, fmt.Sprintf("chore: add %s file", parsedFileName), parsedFileName, replacingFileSHA, contents)
		utils.HandleError("error commiting file", err)

	}

	return nil
}
