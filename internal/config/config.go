// Package config provides configuration management for the YAML template generator.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure.
type Config struct {
	Questions map[string]Question `yaml:"questions"`
}

// Question represents a single question configuration.
type Question struct {
	Prompt   string        `yaml:"prompt"`
	Type     *QuestionType `yaml:"type,omitempty"`
	Choices  interface{}   `yaml:"choices"`
	Multiple bool          `yaml:"multiple,omitempty"`
}

// QuestionType defines the type of question.
type QuestionType struct {
	Dynamic     *DynamicType `yaml:"dynamic,omitempty"`
	Interactive bool         `yaml:"interactive,omitempty"`
}

// DynamicType defines dynamic question dependencies.
type DynamicType struct {
	DependencyQuestions []string `yaml:"dependency_questions"`
}

// LoadConfig loads the configuration from .yg-config.yaml file.
func LoadConfig() (*Config, error) {
	configPath := filepath.Join(".yg", "_templates", ".yg-config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetChoices resolves choices for a question based on dependencies.
func (q *Question) GetChoices(answers map[string]interface{}) ([]string, error) {
	switch choices := q.Choices.(type) {
	case []interface{}:
		result := make([]string, len(choices))
		for i, choice := range choices {
			result[i] = fmt.Sprintf("%v", choice)
		}
		return result, nil
	case map[string]interface{}:
		return q.resolveDynamicChoices(choices, answers)
	default:
		return nil, fmt.Errorf("invalid choices type: %T", choices)
	}
}

func (q *Question) resolveDynamicChoices(choices, answers map[string]interface{}) ([]string, error) {
	// Special handling for cluster question based on selected environments
	if q.Type != nil && q.Type.Dynamic != nil {
		for _, dep := range q.Type.Dynamic.DependencyQuestions {
			if dep == "env" {
				// Handle multiple environment selection for cluster choices
				envAnswer, exists := answers["env"]
				if !exists {
					return nil, fmt.Errorf("dependency answer for env not found")
				}

				var envs []string
				switch envValue := envAnswer.(type) {
				case []string:
					envs = envValue
				case string:
					envs = []string{envValue}
				default:
					return nil, fmt.Errorf("invalid env answer type: %T", envValue)
				}

				// Collect all cluster choices from selected environments
				var allClusters []string
				clusterMap := make(map[string]bool) // To avoid duplicates

				for _, env := range envs {
					envChoices, exists := choices[env]
					if !exists {
						continue
					}
					clusterList, ok := envChoices.([]interface{})
					if !ok {
						continue
					}
					for _, cluster := range clusterList {
						clusterStr := fmt.Sprintf("%v", cluster)
						if !clusterMap[clusterStr] {
							allClusters = append(allClusters, clusterStr)
							clusterMap[clusterStr] = true
						}
					}
				}

				return allClusters, nil
			}
		}
	}

	// Handle other dynamic dependencies
	current := choices

	if q.Type != nil && q.Type.Dynamic != nil {
		for _, dep := range q.Type.Dynamic.DependencyQuestions {
			answer, exists := answers[dep]
			if !exists {
				return nil, fmt.Errorf("dependency answer for %s not found", dep)
			}

			answerStr := fmt.Sprintf("%v", answer)
			next, exists := current[answerStr]
			if !exists {
				return nil, fmt.Errorf("no choices found for %s = %s", dep, answerStr)
			}

			switch nextValue := next.(type) {
			case map[string]interface{}:
				current = nextValue
			case []interface{}:
				result := make([]string, len(nextValue))
				for i, choice := range nextValue {
					result[i] = fmt.Sprintf("%v", choice)
				}
				return result, nil
			default:
				return nil, fmt.Errorf("invalid choices structure at %s", dep)
			}
		}
	}

	return nil, fmt.Errorf("unable to resolve choices")
}
