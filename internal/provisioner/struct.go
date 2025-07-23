package provisioner

import (
	"log/slog"

	"github.com/angelvargass/repository-provisioner/internal/gh"
)

type Provisioner struct {
	Logger   *slog.Logger
	GHClient gh.Github
	Config   *Config
}

type Config struct {
	ArchetypesDirectory string
	GoReleaserToken     string
	ReleasePleaseToken  string
}
