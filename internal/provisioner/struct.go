package provisioner

import (
	"log/slog"

	"github.com/angelvargass/go-common/gh"
)

// Provisioner contains the necessary dependencies for the package to work.
type Provisioner struct {
	Logger   *slog.Logger
	GHClient gh.Github
	Config   *Config
}

// Config contains the necesary configurations for the provisioner package to work.
type Config struct {
	ArchetypesDirectory  string
	DevelopmentFilesPath string
	GoReleaserToken      string
	ReleasePleaseToken   string
	Reconciling          bool
}
