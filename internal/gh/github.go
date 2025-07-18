package gh

import (
	"log/slog"

	"github.com/google/go-github/v73/github"
)

func New(logger *slog.Logger, token string) *Github {
	client := github.NewClient(nil).WithAuthToken(token)
	return &Github{
		Logger: logger.With("internal", "github"),
		Client: client,
	}
}
