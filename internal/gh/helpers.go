package gh

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/go-github/v73/github"
)

// Gets a repository as specified by the owner/name parameters
func (gh *Github) GetRepository(ctx context.Context, owner, name string) (*github.Repository, error) {
	gh.Logger.Debug("get repository", slog.String("owner", owner), slog.String("repo name", name))
	repo, res, err := gh.Client.Repositories.Get(ctx, owner, name)
	if res.StatusCode == 404 {
		gh.Logger.Debug("repository not found", slog.String("owner", owner), slog.String("repo name", name))
	}

	if err != nil {
		return nil, fmt.Errorf(fmt.Sprintf("error getting repository %s/%s", owner, name), err)
	}

	return repo, nil
}

// Creates a new repository as specified by the organization/name.
// If authenticated as user, pass an empty organization string to create the repository under the authenticated user.
// Repositories created by this function are private by default.
// Default branch name is set to your configuration on Github.
// Branches are deleted when merged by default.
// A README.md file is created by default.
// Changes can have propagation time on GH's servers.
func (gh *Github) CreateRepository(ctx context.Context, organization, name string) (*github.Repository, error) {
	gh.Logger.Debug("creating repository", slog.String("organization", organization), slog.String("name", name))
	repo, res, err := gh.Client.Repositories.Create(ctx, "", &github.Repository{
		Name:                      github.Ptr(name),
		Private:                   github.Ptr(true),
		HasIssues:                 github.Ptr(true),
		HasProjects:               github.Ptr(false),
		HasWiki:                   github.Ptr(false),
		AutoInit:                  github.Ptr(true),
		HasDiscussions:            github.Ptr(true),
		DeleteBranchOnMerge:       github.Ptr(true),
		UseSquashPRTitleAsDefault: github.Ptr(true),
		AllowForking:              github.Ptr(true),
	})

	if res.StatusCode == 422 {
		gh.Logger.Debug("validation failed", slog.String("organization", organization), slog.String("name", name))
	}

	if err != nil {
		gh.Logger.Debug("error creating repository", slog.String("organization", organization), slog.String("name", name))
		return nil, err
	}

	return repo, nil
}

// Creates a new branch in the specified repository.
// Takes the last commit from the default branch and creates a new branch with the specified name.
func (gh *Github) CreateBranch(ctx context.Context, owner, repoName, branchName string) (*github.Reference, error) {
	gh.Logger.Debug("creating branch", slog.String("owner", owner), slog.String("repo name", repoName), slog.String("branch name", branchName))
	repo, err := gh.GetRepository(ctx, owner, repoName)
	if err != nil {
		gh.Logger.Debug("error getting repository", slog.String("owner", owner), slog.String("repo name", repoName))
		return nil, err
	}

	gh.Logger.Debug("getting latest reference from default branch", slog.String("owner", owner), slog.String("repo name", repoName), slog.String("default branch name", *repo.DefaultBranch))
	ref, _, err := gh.Client.Git.GetRef(ctx, owner, repoName, fmt.Sprintf("refs/heads/%s", *repo.DefaultBranch))
	if err != nil {
		gh.Logger.Debug("error getting latest reference from default branch", slog.String("owner", owner), slog.String("repo name", repoName), slog.String("default branch name", *repo.DefaultBranch))
	}

	branch, _, err := gh.Client.Git.CreateRef(ctx, owner, repoName, &github.Reference{
		Ref: github.Ptr(fmt.Sprintf("refs/heads/%s", branchName)),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	})

	if err != nil {
		gh.Logger.Debug("error creating new branch", slog.String("owner", owner), slog.String("repo name", repoName), slog.String("new branch name", branchName))
	}

	return branch, nil
}
