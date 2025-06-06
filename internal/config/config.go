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
	Questions Questions `yaml:"questions"`
}

// Questions represents the questions configuration with order and definitions.
type Questions struct {
	Order       []string            `yaml:"order,omitempty"`
	Definitions map[string]Question `yaml:"definitions,omitempty"`
	// For backward compatibility, support the old direct map format
	DirectMap map[string]Question `yaml:",inline"`
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

	// Normalize the config to handle both new and old formats
	config.Questions.normalize()

	return &config, nil
}

// GetQuestions returns the questions map, handling both new and old formats.
func (q *Questions) GetQuestions() map[string]Question {
	if len(q.Definitions) > 0 {
		return q.Definitions
	}
	return q.DirectMap
}

// GetOrder returns the question order, generating one if not specified.
func (q *Questions) GetOrder() []string {
	if len(q.Order) > 0 {
		return q.Order
	}

	// Generate order from available questions
	questions := q.GetQuestions()
	order := make([]string, 0, len(questions))
	for key := range questions {
		order = append(order, key)
	}
	return order
}

// normalize handles backward compatibility by moving direct map to definitions if needed.
func (q *Questions) normalize() {
	// If using old format (direct map), convert to new format
	if q.Definitions == nil && q.DirectMap != nil {
		q.Definitions = q.DirectMap
		q.DirectMap = nil
	}

	// If using new format but no order specified, generate from keys
	if q.Definitions != nil && len(q.Order) == 0 {
		q.Order = make([]string, 0, len(q.Definitions))
		for key := range q.Definitions {
			q.Order = append(q.Order, key)
		}
	}
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
