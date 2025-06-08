// Package prompt provides interactive prompt functionality.
package prompt

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
)

// PrompterInterface defines the interface for prompting users.
type PrompterInterface interface {
	Select(message string, options []string) (string, error)
	MultiSelect(message string, options []string) ([]string, error)
	Search(message string, options []string) (string, error)
	Confirm(message string) (bool, error)
}

// Prompter implements PrompterInterface using survey.
type Prompter struct{}

// NewPrompter creates a new Prompter instance.
func NewPrompter() *Prompter {
	// Disable color for consistent output
	core.DisableColor = false
	return &Prompter{}
}

// Select prompts the user to select a single option.
func (p *Prompter) Select(message string, options []string) (string, error) {
	var result string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return "", fmt.Errorf("failed to get selection: %w", err)
	}

	return result, nil
}

// MultiSelect prompts the user to select multiple options.
func (p *Prompter) MultiSelect(message string, options []string) ([]string, error) {
	var result []string
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return nil, fmt.Errorf("failed to get multi-selection: %w", err)
	}

	return result, nil
}

// Search prompts the user with a searchable interface supporting text input and filtering.
func (p *Prompter) Search(message string, options []string) (string, error) {
	var result string

	prompt := &survey.Select{
		Message: message + " (type to search, ↓↑ to select):",
		Options: options,
		Filter: func(filterValue string, optionValue string, _ int) bool {
			// If no filter input, show all options
			if filterValue == "" {
				return true
			}

			// Check for exact match first (case insensitive)
			if strings.EqualFold(filterValue, optionValue) {
				return true
			}

			// Then check for partial match (contains, case insensitive)
			return strings.Contains(
				strings.ToLower(optionValue),
				strings.ToLower(filterValue),
			)
		},
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return "", fmt.Errorf("failed to get search result: %w", err)
	}

	return result, nil
}

// Confirm prompts the user for confirmation.
func (p *Prompter) Confirm(message string) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: false,
	}

	if err := survey.AskOne(prompt, &result); err != nil {
		return false, fmt.Errorf("failed to get confirmation: %w", err)
	}

	return result, nil
}
