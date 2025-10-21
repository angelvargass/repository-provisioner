package config

import (
	"log/slog"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	LogLevel             string `envconfig:"LOG_LEVEL" default:"info"`
	RepoOwner            string `envconfig:"REPO_OWNER" required:"true"`
	RepoName             string `envconfig:"REPO_NAME" required:"true"`
	Archetype            string `envconfig:"ARCHETYPE" required:"true"`
	ArchetypesDirectory  string `envconfig:"ARCHETYPES_DIRECTORY" default:"internal/archetypes/"`
	Reconciling          bool   `envconfig:"RECONCILE" default:"false"`
	Development          bool   `envconfig:"DEVELOPMENT" default:"false"`
	DevelopmentFilesPath string `envconfig:"DEVELOPMENT_FILES_PATH" default:"development-files/"`

	GithubConfig *Github
}

type Github struct {
	AccessToken        string `envconfig:"GITHUB_ACCESS_TOKEN" required:"true"`
	GoReleaserToken    string `envconfig:"GO_RELEASER_TOKEN" required:"true"`
	ReleasePleaseToken string `envconfig:"RELEASE_PLEASE_TOKEN" required:"true"`
}

func New() (*Config, error) {
	c := new(Config)
	err := envconfig.Process("repository-provisioner", c)
	if err != nil {
		slog.Error("error reading env variables", slog.String("error", err.Error()))
		return nil, err
	}

	return c, nil
}
