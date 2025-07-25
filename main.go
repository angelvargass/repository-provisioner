package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/angelvargass/go-common/gh"
	"github.com/angelvargass/repository-provisioner/internal/config"
	"github.com/angelvargass/repository-provisioner/internal/logger"
	"github.com/angelvargass/repository-provisioner/internal/provisioner"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		slog.Error("failed to load configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger := logger.New(cfg.LogLevel)
	logger.Info("repository-provisioner started", slog.String("logLevel", cfg.LogLevel))

	gh := gh.New(logger, cfg.GithubConfig.AccessToken)
	provisioner := provisioner.New(logger, gh, cfg.ArchetypesDirectory, cfg.GithubConfig.GoReleaserToken, cfg.GithubConfig.ReleasePleaseToken)

	if cfg.Reconciling {
		logger.Info("reconciling mode enabled, skipping repository provisioning")
		//implement reonciling logic here
		return
	}

	provisioner.ProvisionRepository(ctx, cfg.RepoOwner, cfg.RepoName, cfg.Archetype)
	logger.Info("repository-provisioner ended", slog.String("logLevel", cfg.LogLevel))
}
