package provisioner

import (
	"context"
	"fmt"
	"strings"
)

func (p *Provisioner) getArchetypeFromTopics(topics []string, archetypeTopicPrefix string) (string, error) {
	for _, topic := range topics {
		if strings.HasPrefix(topic, archetypeTopicPrefix) {
			return strings.TrimPrefix(topic, archetypeTopicPrefix), nil
		}
	}
	return "", fmt.Errorf("archetype topic not found in topics: %v", topics)
}

// getArchetypeSubPath returns the key from the archetypeSubPaths map that contains the selectedArchetype.
// If it is not found, returns an error.
func (p *Provisioner) getArchetypeSubPath(selectedArchetype string) (string, error) {
	for k := range archetypeSubPaths {
		if strings.Contains(k, selectedArchetype) {
			return k, nil
		}
	}

	return "", fmt.Errorf("archetype not found")
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
