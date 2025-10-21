package provisioner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/angelvargass/go-common/gh"
	"github.com/angelvargass/repository-provisioner/internal/templateengine"
	"github.com/angelvargass/repository-provisioner/internal/utils"
	"github.com/google/go-github/v73/github"
)

const (
	GoTemplateExtension        = ".go.tmpl"
	InitBranchName             = "init"
	RepositoryProvisionerTopic = "repository-provisioner"
	ArchetypeTopicPrefix       = "archetype-%s"
	DefaultRulesetName         = "default-branch-ruleset"
	DefaultBranch              = "main"
	ReconcileBranchName        = "repository-provisioner-reconcile-%d"
)

// archetypeSubPaths maps archetype names to their subdirectory paths.
// It also maps the files within the archetype that must not be reconciled.
var archetypeSubPaths = map[string][]string{
	"golang/": {"main.go", "go.mod", "README.md"},
}

func New(logger *slog.Logger, ghClient *gh.Github, archetypesDirectory, developmentFilesPath, goReleaserToken, releasePleaseToken string, reconciling bool) *Provisioner {
	return &Provisioner{
		Logger:   logger.With(slog.String("internal", "provisioner")),
		GHClient: *ghClient,
		Config: &Config{
			ArchetypesDirectory:  archetypesDirectory,
			DevelopmentFilesPath: developmentFilesPath,
			GoReleaserToken:      goReleaserToken,
			ReleasePleaseToken:   releasePleaseToken,
			Reconciling:          reconciling,
		},
	}
}

// ProvisionOrReconcileRepository creates or reconciles a given repository.
func (p *Provisioner) ProvisionOrReconcileRepository(ctx context.Context, owner, repoName, archetype string) {
	p.Logger.Info("provisioning repository", slog.String("owner", owner), slog.String("repoName", repoName))
	randomInt, err := utils.GenerateRandomInteger()
	utils.HandleError("error generating random integer", err)
	branchName := fmt.Sprintf(ReconcileBranchName, randomInt)

	if !p.Config.Reconciling {
		branchName = InitBranchName
		_, err := p.GHClient.CreateRepository(ctx, "", repoName)
		utils.HandleError("failed to create repository", err)
	}

	p.Logger.Info("creating branch", slog.String("branch name", branchName))
	ref, err := p.GHClient.CreateBranch(ctx, owner, repoName, branchName)
	utils.HandleError("failed to create initial branch", err)

	archetypeSubPath, err := p.getArchetypeSubPath(archetype)
	utils.HandleError("failed get archetype sub path", err)

	archetypeFiles, err := templateengine.RenderTemplates(p.Config.ArchetypesDirectory, archetypeSubPath, map[string]any{
		"RepositoryOwner": owner,
		"RepositoryName":  repoName,
		"DefaultBranch":   DefaultBranch,
	})
	utils.HandleError("failed to load archetype files", err)
	for _, file := range archetypeFiles {
		if p.Config.Reconciling && slices.Contains(archetypeSubPaths[archetypeSubPath], file.Name) {
			continue
		}

		replacingFileSHA := ""
		fileContent, _, err := p.GHClient.GetRepositoryContent(ctx, owner, repoName, file.Name, "")
		utils.HandleError(fmt.Sprintf("failed to fetch content for file %s", file.Name), err)

		if fileContent != nil {
			replacingFileSHA = *fileContent.SHA
		}

		p.Logger.Debug("commiting file", slog.String("branch name", *ref.Ref), slog.String("file path", file.Name))
		_, err = p.GHClient.CreateOrUpdateFile(ctx, owner, repoName, *ref.Ref, fmt.Sprintf("chore: add %s file", file.Name), file.Name, replacingFileSHA, file.Content)
		utils.HandleError("error commiting file", err)
	}

	p.Logger.Info("configuring repository")
	p.configureRepository(ctx, owner, repoName, archetype)

	p.Logger.Info("opening PR")
	p.createPullRequest(ctx, owner, repoName, branchName, archetype)
}

// configureRepository configures the specified repository.
// It creates or updates the topics, configures the rulesets and creates or updates the repository secrets based on the archetype.
// The default PR method used is squash and merge.
// On a reconcile event, it overrides all the topics, and the necessary secrets for the specified archetype.
//
// For more information about squash and merge, please visit: https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/configuring-pull-request-merges/configuring-commit-squashing-for-pull-requests
func (p *Provisioner) configureRepository(ctx context.Context, owner, repoName, archetype string) {
	topics := []string{RepositoryProvisionerTopic, fmt.Sprintf(ArchetypeTopicPrefix, archetype), repoName}
	rules := &github.RepositoryRulesetRules{
		PullRequest: &github.PullRequestRuleParameters{
			AllowedMergeMethods:            []github.PullRequestMergeMethod{github.PullRequestMergeMethodSquash},
			DismissStaleReviewsOnPush:      true,
			RequireCodeOwnerReview:         true,
			RequiredApprovingReviewCount:   0,
			RequiredReviewThreadResolution: true,
		},
		RequiredStatusChecks: &github.RequiredStatusChecksRuleParameters{
			RequiredStatusChecks: []*github.RuleStatusCheck{
				{
					Context: "validate-commits",
				},
			},
		},
	}

	p.Logger.Info("replacing topics for repo", slog.String("repo name", repoName), slog.String("archetype", archetype))
	_, err := p.GHClient.ReplaceTopics(ctx, owner, repoName, topics)
	utils.HandleError(fmt.Sprintf("error replacing topics for repository %s", repoName), err)

	if !p.Config.Reconciling {
		p.Logger.Info("configuring rulesets", slog.String("repo name", repoName))
		_, err = p.GHClient.CreateRepositoryRuleset(ctx, owner, repoName, DefaultRulesetName, rules)
		utils.HandleError(fmt.Sprintf("error configuring rulesets for repository %s", repoName), err)
	}

	p.Logger.Info("configuring repository secrets", slog.String("archetype", archetype))
	switch archetype {
	case "golang":
		err := p.configureGolangArchetypeRepositorySecrets(ctx, owner, repoName, p.Config.ReleasePleaseToken, p.Config.GoReleaserToken)
		utils.HandleError("error configuring repo secrets for golang archetype", err)
	}
}

// createPullRequest opens a pull request for the specified repository.
func (p *Provisioner) createPullRequest(ctx context.Context, owner, repoName, branchName, archetype string) {
	p.Logger.Info("opening PR", slog.String("repo name", repoName), slog.String("archetype", archetype))
	title := fmt.Sprintf("chore: commit from repository-provisioner for %s archetype", archetype)
	bodyContent := fmt.Sprintf("This PR was automatically created by the repository-provisioner for the %s archetype.\n\nPlease review the changes and merge them to start using the repository.", archetype)

	repo, err := p.GHClient.GetRepository(ctx, owner, repoName)
	utils.HandleError(fmt.Sprintf("error getting repository %s", repoName), err)

	_, err = p.GHClient.CreatePullRequest(ctx, owner, repoName, title, bodyContent, branchName, *repo.DefaultBranch)
	utils.HandleError(fmt.Sprintf("error creating initial PR for repo %s", repoName), err)
}

// CreateArchetypeLocally is used when the DEVELOPMENT environment variable = true, it creates the files for the specified
// archetype at the root path of the main.go program.
func (p *Provisioner) CreateArchetypeLocally(owner, repoName, archetype string) {
	archetypeSubPath, err := p.getArchetypeSubPath(archetype)
	utils.HandleError("failed get archetype sub path", err)

	files, err := templateengine.RenderTemplates(p.Config.ArchetypesDirectory, archetypeSubPath, map[string]any{
		"RepositoryOwner": owner,
		"RepositoryName":  repoName,
		"DefaultBranch":   DefaultBranch,
	})
	utils.HandleError(fmt.Sprintf("error rendering archetype: %s template", archetype), err)

	err = os.RemoveAll(p.Config.DevelopmentFilesPath)
	utils.HandleError("error cleaning up development files path", err)

	if err := os.MkdirAll(p.Config.DevelopmentFilesPath, 0755); err != nil {
		utils.HandleError(fmt.Sprintf("failed to create development files dir: %s", p.Config.DevelopmentFilesPath), err)
	}

	for _, file := range files {
		targetPath := filepath.Join(p.Config.DevelopmentFilesPath, file.Name)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			utils.HandleError(fmt.Sprintf("failed to create parent dirs for %s", targetPath), err)
		}

		if err := os.WriteFile(targetPath, file.Content, 0600); err != nil {
			utils.HandleError(fmt.Sprintf("failed to create file for %s", targetPath), err)
		}
	}
}
