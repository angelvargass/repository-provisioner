package provisioner

import (
	"context"
	"fmt"

	"github.com/angelvargass/repository-provisioner/internal/utils"
)

// getUpdatingFilesSHA fetches the repository content from the default branch, Searches if a file
// from the archetypeFiles slice exists in the repository. If exists,
// returns the SHA associated with the repo's file.
func (p *Provisioner) getUpdatingFileSHA(ctx context.Context, owner, repoName, path string) {
	fileContent, _, err := p.GHClient.GetRepositoryContent(ctx, owner, repoName, path, "")
	utils.HandleError("failed to get repository content", err)

	fmt.Println(fileContent)
}
