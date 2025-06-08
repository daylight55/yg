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
	Questions Questions                 `yaml:"questions"`
	Templates map[string]TemplateConfig `yaml:"templates,omitempty"`
}

// TemplateConfig represents template configuration.
type TemplateConfig struct {
	Type string `yaml:"type"` // "file" or "directory"
	Path string `yaml:"path"` // path to template file or directory
}

// Questions represents the questions configuration with order and definitions.
type Questions struct {
	Order            []string            `yaml:"order,omitempty"`
	TemplateQuestion string              `yaml:"template_question,omitempty"`
	Definitions      map[string]Question `yaml:"definitions,omitempty"`
	// For backward compatibility, support the old direct map format
	DirectMap map[string]Question `yaml:",inline"`
}

// Question represents a single question configuration.
type Question struct {
	Prompt  string        `yaml:"prompt"`
	Type    *QuestionType `yaml:"type,omitempty"`
	Choices interface{}   `yaml:"choices"`
}

// QuestionType defines the type of question.
type QuestionType struct {
	Dynamic     *DynamicType `yaml:"dynamic,omitempty"`
	Interactive bool         `yaml:"interactive,omitempty"`
	Multiple    bool         `yaml:"multiple,omitempty"`
}

// DynamicType defines dynamic question dependencies.
type DynamicType struct {
	DependencyQuestions []string `yaml:"dependency_questions"`
}

// IsMultiple returns whether the question supports multiple selections.
func (q *Question) IsMultiple() bool {
	return q.Type != nil && q.Type.Multiple
}

// LoadConfig loads the configuration from the specified path or default locations.
// If configPath is empty, it tries default paths: ./.yg/config.yaml and ./.yg/config.yml
func LoadConfig(configPath string) (*Config, error) {
	var paths []string

	if configPath != "" {
		// Use specified config path
		paths = []string{configPath}
	} else {
		// Try default paths in order
		paths = []string{
			filepath.Join(".yg", "config.yaml"),
			filepath.Join(".yg", "config.yml"),
			// Keep backward compatibility with old path
			filepath.Join(".yg", "_templates", ".yg-config.yaml"),
		}
	}

	var lastErr error
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			lastErr = err
			continue
		}

		var config Config
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
		}

		// Normalize the config to handle both new and old formats
		config.Questions.normalize()

		return &config, nil
	}

	if configPath != "" {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, lastErr)
	}
	return nil, fmt.Errorf("no config file found in default locations (./.yg/config.yaml, ./.yg/config.yml): %w", lastErr)
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

// GetTemplateQuestion returns the question key that provides the template name.
// If not specified, returns empty string and the caller should use heuristics.
func (q *Questions) GetTemplateQuestion() string {
	return q.TemplateQuestion
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
	if q.Type == nil || q.Type.Dynamic == nil {
		return nil, fmt.Errorf("dynamic type configuration missing")
	}

	var current interface{} = choices

	// Process each dependency question in order
	for _, dep := range q.Type.Dynamic.DependencyQuestions {
		answer, exists := answers[dep]
		if !exists {
			return nil, fmt.Errorf("dependency answer for %s not found", dep)
		}

		// Handle both single and multiple selections
		var answerValues []string
		switch answerValue := answer.(type) {
		case []string:
			answerValues = answerValue
		case string:
			answerValues = []string{answerValue}
		case []interface{}:
			for _, v := range answerValue {
				answerValues = append(answerValues, fmt.Sprintf("%v", v))
			}
		default:
			answerValues = []string{fmt.Sprintf("%v", answer)}
		}

		// If multiple values are provided, collect choices from all values
		if len(answerValues) > 1 {
			allChoices := make(map[string]bool) // To avoid duplicates
			var result []string

			for _, answerStr := range answerValues {
				currentMap, ok := current.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("expected map for dependency lookup, got %T", current)
				}
				next, exists := currentMap[answerStr]
				if !exists {
					continue // Skip missing choices
				}

				switch nextValue := next.(type) {
				case []interface{}:
					// Direct choice list
					for _, choice := range nextValue {
						choiceStr := fmt.Sprintf("%v", choice)
						if !allChoices[choiceStr] {
							result = append(result, choiceStr)
							allChoices[choiceStr] = true
						}
					}
				case map[string]interface{}:
					// Nested structure - collect all choices from nested maps
					for _, subChoices := range nextValue {
						if choiceList, ok := subChoices.([]interface{}); ok {
							for _, choice := range choiceList {
								choiceStr := fmt.Sprintf("%v", choice)
								if !allChoices[choiceStr] {
									result = append(result, choiceStr)
									allChoices[choiceStr] = true
								}
							}
						}
					}
				}
			}

			return result, nil
		}

		// Single value - navigate to next level
		answerStr := answerValues[0]
		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected map for dependency lookup, got %T", current)
		}
		next, exists := currentMap[answerStr]
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
			return nil, fmt.Errorf("invalid next value type: %T", nextValue)
		}
	}

	// If we reach here, current should be a final choice list
	switch finalChoices := current.(type) {
	case []interface{}:
		result := make([]string, len(finalChoices))
		for i, choice := range finalChoices {
			result[i] = fmt.Sprintf("%v", choice)
		}
		return result, nil
	case map[string]interface{}:
		// If we have a map, it means we didn't process all dependencies
		return nil, fmt.Errorf("incomplete dependency resolution, remaining structure: %T", finalChoices)
	default:
		return nil, fmt.Errorf("final choices must be an array, got %T", finalChoices)
	}
}
