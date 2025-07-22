package provisioner

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/angelvargass/repository-provisioner/internal/utils"
)

func (p *Provisioner) getArchetypeSubPath(selectedArchetype string) string {
	archetypeIndex := slices.IndexFunc(archetypesSubPaths, func(archetypeSubPath string) bool {
		return strings.Contains(archetypeSubPath, selectedArchetype)
	})

	if archetypeIndex == -1 {
		utils.HandleError(fmt.Sprintf("archetype %s does not exists", selectedArchetype), fmt.Errorf("archetype not found"))
	}

	return archetypesSubPaths[archetypeIndex]
}

// configureGolangArchetypeRepositorySecrets creates or updates the required secrets for the
// golang archetype on the target repository
func (p *Provisioner) configureGolangArchetypeRepositorySecrets(ctx context.Context, owner, repoName, releasePleaseToken, goReleaserToken string) error {
	secrets := map[string]string{
		"RELEASE_PLEASE_TOKEN": releasePleaseToken,
		"GORELEASER_TOKEN":     goReleaserToken,
	}

	for secretName, secretValue := range secrets {
		err := p.GHClient.CreateOrUpdateRepositorySecret(ctx, owner, repoName, secretName, secretValue)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("error creating or updating repository secret %s", secretName), err)
		}

		p.Logger.Debug("created or updated repository secret", "secret name", secretName)
	}
	return nil
}
