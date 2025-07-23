package utils

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

func HandleError(errorMessage string, err error) {
	if err != nil {
		slog.Error(errorMessage, slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func GetArchetypeFromTopics(topics []string, archetypeTopicPrefix string) (string, error) {
	for _, topic := range topics {
		if strings.HasPrefix(topic, archetypeTopicPrefix) {
			return strings.TrimPrefix(topic, archetypeTopicPrefix), nil
		}
	}
	return "", fmt.Errorf("archetype topic not found in topics: %v", topics)
}
