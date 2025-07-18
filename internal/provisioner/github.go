package provisioner

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/angelvargass/repository-provisioner/internal/utils"
	"github.com/google/go-github/v73/github"
)

func New(logger *slog.Logger, ghToken, archetypesDirectory string) *Github {
	client := github.NewClient(nil).WithAuthToken(ghToken)
	return &Github{
		Logger:              logger,
		Client:              client,
		ArchetypesDirectory: archetypesDirectory,
	}
}

// CreateRepoBasedOnArchetype creates a new GitHub repository based on the specified archetype.
// The archetype can be a template repository or a specific structure that you want to replicate.
func (gh *Github) CreateOrUpdateRepoBasedOnArchetype(ctx context.Context, owner, repoName, archetype string) error {
	gh.Logger.Info("creating or updating repository based on archetype", slog.String("repository name", repoName), slog.String("archetype", archetype))
	newRepo := false

	repo, err := gh.getRepository(ctx, owner, repoName)
	utils.HandleError(fmt.Sprintf("error fetching repository: %s", repoName), err)

	if repo == nil {
		repo, err = gh.createRepository(ctx, repoName, archetype)
		newRepo = true
		utils.HandleError("error creating repository", err)
	}

	if archetype == "" {
		archetype, err = utils.GetArchetypeFromTopics(repo.Topics, ArchetypeTopic)
		utils.HandleError("error getting archetype from topics", err)
	}

	gh.ConfigureRepository(ctx, owner, repoName, archetype, newRepo)
	//gh.AddArchetypeFiles(ctx, owner, repoName, archetype)

	return nil
}

func (gh *Github) ConfigureRepository(ctx context.Context, owner, repoName, archetype string, newRepo bool) {
	gh.Logger.Info("configuring repository", slog.String("repository name", repoName), slog.String("archetype", archetype))
	err := gh.configureRepositoryTopics(ctx, owner, repoName, archetype)
	utils.HandleError(fmt.Sprintf("error configuring repo topics for repo: %s", repoName), err)

	newRepo = true

	if newRepo {
		gh.createBranch(ctx, owner, repoName, "init")
	}

	//configure repo rules

	//configure repo secrets
}

// func (gh *Github) AddArchetypeFiles(ctx context.Context, owner, repoName, archetype string) {
// 	gh.Logger.Info("adding files", slog.String("repository name", repoName), slog.String("archetype", archetype))
// 	gh.addFilesToRepo(ctx, owner, repoName)
// }
