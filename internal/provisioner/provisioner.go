package provisioner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/angelvargass/repository-provisioner/internal/gh"
	"github.com/angelvargass/repository-provisioner/internal/utils"
)

const (
	GoTemplateExtension        = ".go.tmpl"
	InitBranchName             = "init"
	RepositoryProvisionerTopic = "repository-provisioner"
	ArchetypeTopicPrefix       = "archetype-%s"
)

var archetypesSubPaths = []string{"golang/"}

func New(logger *slog.Logger, ghClient *gh.Github, archetypesDirectory, goReleaserToken, releasePleaseToken string) *Provisioner {
	return &Provisioner{
		Logger:   logger.With(slog.String("internal", "provisioner")),
		GHClient: *ghClient,
		Config: &Config{
			ArchetypesDirectory: archetypesDirectory,
		},
	}
}

func (p *Provisioner) ProvisionRepository(ctx context.Context, owner, repoName, archetype string) {
	// p.Logger.Info("provisioning repository", slog.String("owner", owner), slog.String("repoName", repoName))
	// archetypeSubPath := p.getArchetypeSubPath(archetype)

	// // _, err := p.GHClient.CreateRepository(ctx, "", repoName)
	// // utils.HandleError("failed to create repository", err)

	// p.Logger.Info("creating default branch", slog.String("branch name", InitBranchName))
	// ref, err := p.GHClient.CreateBranch(ctx, owner, repoName, InitBranchName)
	// utils.HandleError("failed to create initial branch", err)

	// archetypeFiles := filesystem.LoadFilesForArchetype(p.Config.ArchetypesDirectory, archetypeSubPath)
	// p.Logger.Info("committing archetype to created branch")
	// for _, file := range archetypeFiles {
	// 	parsedFileName := file
	// 	replacingFileSHA := ""
	// 	archetypeFilePath := p.Config.ArchetypesDirectory + archetypeSubPath + file
	// 	contents, err := os.ReadFile(archetypeFilePath)
	// 	utils.HandleError("failed to load file contents", err)

	// 	if strings.Contains(file, GoTemplateExtension) {
	// 		parsedFileName = strings.TrimSuffix(file, GoTemplateExtension)
	// 		contents = filesystem.LoadTemplateFile(archetypeFilePath, map[string]any{
	// 			"RepositoryName": repoName,
	// 			"DefaultBranch":  "main",
	// 		})
	// 	}

	// 	fileContent, _, err := p.GHClient.GetRepositoryContent(ctx, owner, repoName, parsedFileName, "")
	// 	utils.HandleError(fmt.Sprintf("failed to fetch content for file %s", parsedFileName), err)

	// 	if fileContent != nil {
	// 		replacingFileSHA = *fileContent.SHA
	// 	}

	// 	p.Logger.Debug("commiting file", slog.String("branch name", *ref.Ref), slog.String("file path", file))
	// 	_, err = p.GHClient.CreateOrUpdateFile(ctx, owner, repoName, *ref.Ref, fmt.Sprintf("chore: add %s file", parsedFileName), parsedFileName, replacingFileSHA, contents)
	// 	utils.HandleError("error commiting file", err)
	// }

	p.Logger.Info("configuring repository")
	p.ConfigureRepository(ctx, owner, repoName, archetype)
}

func (p *Provisioner) ConfigureRepository(ctx context.Context, owner, repoName, archetype string) {
	p.Logger.Info("replacing topics for repo", slog.String("repo name", repoName), slog.String("archetype", archetype))
	_, err := p.GHClient.ReplaceTopics(ctx, owner, repoName, []string{RepositoryProvisionerTopic, fmt.Sprintf(ArchetypeTopicPrefix, archetype)})
	utils.HandleError(fmt.Sprintf("error replacing topics for repository %s", repoName), err)

	p.Logger.Info("configuring repository secrets", slog.String("archetype", archetype))
	switch archetype {
	case "golang":
		err := p.configureGolangArchetypeRepositorySecrets(ctx, owner, repoName, p.Config.ReleasePleaseToken, p.Config.GoReleaserToken)
		utils.HandleError("error configuring repo secrets for golang archetype", err)
	}
}
