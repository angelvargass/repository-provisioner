package config

import (
	"log/slog"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	LogLevel     string `envconfig:"LOG_LEVEL" default:"info"`
	GithubConfig *Github
}

type Github struct {
	AccessToken         string `envconfig:"GITHUB_ACCESS_TOKEN" required:"true"`
	RepoName            string `envconfig:"GITHUB_REPO_NAME" required:"true"`
	RepoOwner           string `envconfig:"GITHUB_OWNER" required:"true"`
	Archetype           string `envconfig:"GITHUB_ARCHETYPE" required:"true"`
	ArchetypesDirectory string `envconfig:"GITHUB_ARCHETYPES_DIRECTORY" default:"internal/archetypes/"`
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
