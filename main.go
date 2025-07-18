package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/angelvargass/repository-provisioner/internal/config"
	"github.com/angelvargass/repository-provisioner/internal/gh"
	"github.com/angelvargass/repository-provisioner/internal/logger"
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
	logger.Info("application started", slog.String("logLevel", cfg.LogLevel))

	gh := gh.New(logger, cfg.GithubConfig.AccessToken, cfg.GithubConfig.ArchetypesDirectory)
	gh.CreateOrUpdateRepoBasedOnArchetype(ctx, cfg.GithubConfig.RepoOwner, cfg.GithubConfig.RepoName, cfg.GithubConfig.Archetype)
}
