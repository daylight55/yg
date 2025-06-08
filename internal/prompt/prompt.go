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

// Search prompts the user with a searchable interface and text input support.
func (p *Prompter) Search(message string, options []string) (string, error) {
	var result string

	// First, try text input to check for exact match
	textPrompt := &survey.Input{
		Message: message + " (入力するか↓↑で選択):",
	}

	if err := survey.AskOne(textPrompt, &result); err != nil {
		return "", fmt.Errorf("failed to get text input: %w", err)
	}

	// Check if the input exactly matches any option
	for _, option := range options {
		if strings.EqualFold(result, option) {
			return option, nil
		}
	}

	// If no exact match found, show filtered selection
	filteredOptions := make([]string, 0)
	inputLower := strings.ToLower(result)

	for _, option := range options {
		optionLower := strings.ToLower(option)
		if strings.Contains(optionLower, inputLower) {
			filteredOptions = append(filteredOptions, option)
		}
	}

	// If no matches found, show all options
	if len(filteredOptions) == 0 {
		filteredOptions = options
	}

	// If only one match, return it directly
	if len(filteredOptions) == 1 {
		return filteredOptions[0], nil
	}

	// Show selection from filtered options
	selectPrompt := &survey.Select{
		Message: "選択してください:",
		Options: filteredOptions,
	}

	if err := survey.AskOne(selectPrompt, &result); err != nil {
		return "", fmt.Errorf("failed to get selection: %w", err)
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
