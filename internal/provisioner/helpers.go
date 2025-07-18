package provisioner

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/google/go-github/v73/github"
)

const (
	DefaultBranch        = "main"
	InitialPRBranch      = "init"
	RepoProvisionerTopic = "repository-provisioner"
	ArchetypeTopic       = "archetype"
)

var AvailableArchetypes = []string{
	"golang",
}

var ManagedFiles = []string{
	".github/workflows/golang-ci-lint.yaml",
	".github/workflows/release-please.yml",
	".goreleaser.yaml",
	".golangci.yaml",
}

func (gh *Github) getRepository(ctx context.Context, owner, name string) (*github.Repository, error) {
	repo, res, err := gh.Client.Repositories.Get(ctx, owner, name)
	if res.StatusCode == 404 {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf(err.Error(), err)
	}

	return repo, nil
}

func (gh *Github) createRepository(ctx context.Context, repoName, description string) (*github.Repository, error) {
	repo, _, err := gh.Client.Repositories.Create(ctx, "", &github.Repository{
		Name:                      github.Ptr(repoName),
		Description:               github.Ptr(description),
		Private:                   github.Ptr(true),
		HasIssues:                 github.Ptr(true),
		HasProjects:               github.Ptr(false),
		AutoInit:                  github.Ptr(true),
		DefaultBranch:             github.Ptr(DefaultBranch),
		AllowSquashMerge:          github.Ptr(true),
		AllowForking:              github.Ptr(false),
		DeleteBranchOnMerge:       github.Ptr(true),
		UseSquashPRTitleAsDefault: github.Ptr(true),
	})

	if err != nil {
		gh.Logger.Debug("error creating repository", slog.String("error", err.Error()))
		return nil, err
	}
	gh.Logger.Debug(fmt.Sprintf("created repository with name: %s", *repo.FullName))

	return repo, nil
}

func (gh *Github) configureRepositoryTopics(ctx context.Context, owner, repoName, archetype string) error {
	topics := []string{RepoProvisionerTopic, fmt.Sprintf("%s-%s", ArchetypeTopic, archetype), repoName}
	gh.Logger.Debug("configuring repository topics", slog.String("repo name", repoName), slog.Any("topics", topics))

	if !slices.Contains(AvailableArchetypes, archetype) {
		return fmt.Errorf("archetype: %s does not exists", archetype)
	}
	_, resp, err := gh.Client.Repositories.ReplaceAllTopics(ctx, owner, repoName, topics)
	if resp.StatusCode == 404 {
		gh.Logger.Debug("error updating repository topics")
		fmt.Errorf("repository: %s not found", repoName)
	}

	if err != nil {
		gh.Logger.Debug("error updating topics", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (gh *Github) getLatestRefFromDefaultBranch(ctx context.Context, owner, repoName string) (*github.Reference, error) {
	gh.Logger.Debug("getting latest ref from default branch")
	refString := fmt.Sprintf("refs/heads/%s", DefaultBranch)
	ref, resp, err := gh.Client.Git.GetRef(ctx, owner, repoName, refString)
	if resp.StatusCode != 200 || err != nil {
		return nil, fmt.Errorf("error getting reference for branch %s", DefaultBranch, err)
	}

	return ref, nil
}

func (gh *Github) createBranch(ctx context.Context, owner, repoName, branchName string) error {
	gh.Logger.Debug("creating new branch", slog.String("name", branchName))

	latestDefaultBranchRef, err := gh.getLatestRefFromDefaultBranch(ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("error creating branch, cannot get latest ref from default branch", err)
	}

	_, _, err = gh.Client.Git.CreateRef(ctx, owner, repoName, &github.Reference{
		Ref: github.Ptr(fmt.Sprintf("refs/heads/%s", branchName)),
		Object: &github.GitObject{
			SHA: latestDefaultBranchRef.Object.SHA,
		},
	})

	if err != nil {
		return fmt.Errorf("error creating new branch %s", branchName, err)
	}

	return nil
}

// func (gh *Github) createInitialPRForArchetype() {
// 	pr := &github.NewPullRequest{}
// 	gh.Client.Git.CreateRef()
// }

// func (gh *Github) createOrUpdateFile(ctx context.Context, owner, repoName, filename, commitMessage string, buffer *bytes.Buffer) {
// 	gh.Client.PullRequests.Create(ctx, owner, repoName, &github.NewPullRequest{
// 		Title: github.Ptr("chore: initializing PR from repository-provisioner"),
// 		Head: github.Ptr(),
// 	})
// }

// func (gh *Github) addFilesToRepo(ctx context.Context, owner, repoName string) error {
// 	repo, err := gh.getRepository(ctx, owner, repoName)
// 	if repo == nil || err != nil {
// 		gh.Logger.Debug(fmt.Sprintf("error fetching repository %s", repoName))
// 		return fmt.Errorf("couldn't add files to repo %s, error fetching", repoName)
// 	}

// 	archetype, err := utils.GetArchetypeFromTopics(repo.Topics, ArchetypeTopic)
// 	utils.HandleError("archetype topic not found", err)

// 	archetypeFiles := filesystem.LoadFilesForArchetype(gh.ArchetypesDirectory, archetype)
// 	for _, file := range archetypeFiles {

// 	}

// 	return nil
// }
