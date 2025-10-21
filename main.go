package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/angelvargass/go-common/gh"
	"github.com/angelvargass/repository-provisioner/internal/config"
	"github.com/angelvargass/repository-provisioner/internal/logger"
	"github.com/angelvargass/repository-provisioner/internal/provisioner"
	"github.com/angelvargass/repository-provisioner/internal/templateengine"
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

	if cfg.Development {
		logger.Info("development mode enabled, creating archetype at root main.go program path")
		files, _ := templateengine.RenderTemplates(cfg.ArchetypesDirectory, "golang", map[string]any{
			"RepositoryOwner": "test",
			"RepositoryName":  "test",
			"DefaultBranch":   "main",
		})

		os.RemoveAll(cfg.DevelopmentFilesPath)

		if err := os.MkdirAll(cfg.DevelopmentFilesPath, 0755); err != nil {
			logger.Error(fmt.Sprintf("failed to create development files dir: %s", cfg.DevelopmentFilesPath), slog.String("message", err.Error()))
			return
		}
		for _, file := range files {
			targetPath := filepath.Join(cfg.DevelopmentFilesPath, file.Name)
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				logger.Error(fmt.Sprintf("failed to create parent dirs for %s", targetPath), slog.String("message", err.Error()))
				continue
			}

			if err := os.WriteFile(targetPath, file.Content, 0600); err != nil {
				logger.Error(fmt.Sprintf("failed to create file for %s", targetPath), slog.String("message", err.Error()))
			}
		}

		return
	}

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
