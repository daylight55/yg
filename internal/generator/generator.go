// Package generator provides the main generation logic.
package generator

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/daylight55/yg/internal/config"
	"github.com/daylight55/yg/internal/prompt"
	"github.com/daylight55/yg/internal/template"
)

// Options holds CLI options for the generator.
type Options struct {
	Answers    map[string]interface{}
	SkipPrompt bool
}

// Generator handles the main generation workflow.
type Generator struct {
	config   *config.Config
	prompter prompt.PrompterInterface
	answers  map[string]interface{}
}

// New creates a new Generator instance.
func New() (*Generator, error) {
	return NewWithConfig("")
}

// NewWithConfig creates a new Generator instance with specified config path.
func NewWithConfig(configPath string) (*Generator, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &Generator{
		config:   cfg,
		prompter: prompt.NewPrompter(),
		answers:  make(map[string]interface{}),
	}, nil
}

// Run executes the generation workflow.
func (g *Generator) Run() error {
	return g.RunWithOptions(&Options{})
}

// RunWithOptions executes the generation workflow with CLI options.
func (g *Generator) RunWithOptions(options *Options) error {
	// Set up signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nOperation canceled by user")
		cancel()
		os.Exit(1)
	}()

	// Use CLI options if skip prompt is enabled
	if options.SkipPrompt {
		if err := g.validateOptions(options); err != nil {
			return fmt.Errorf("invalid options: %w", err)
		}
		// Copy all provided answers
		for key, value := range options.Answers {
			g.answers[key] = value
		}
	} else {
		// Pre-fill answers with CLI options if provided
		if options.Answers != nil {
			for key, value := range options.Answers {
				g.answers[key] = value
			}
		}

		// Process questions in the order defined in config
		questionOrder := g.config.Questions.GetOrder()
		questions := g.config.Questions.GetQuestions()

		for _, questionKey := range questionOrder {
			select {
			case <-ctx.Done():
				return fmt.Errorf("operation canceled")
			default:
			}

			// Skip if already answered via CLI option
			if _, exists := g.answers[questionKey]; exists && !options.SkipPrompt {
				continue
			}

			question, exists := questions[questionKey]
			if !exists {
				return fmt.Errorf("question %s not found in config", questionKey)
			}

			answer, err := g.askQuestion(questionKey, question)
			if err != nil {
				return fmt.Errorf("failed to ask question %s: %w", questionKey, err)
			}

			g.answers[questionKey] = answer
		}
	}

	// Generate and show preview
	if err := g.generatePreview(); err != nil {
		return fmt.Errorf("failed to generate preview: %w", err)
	}

	// Confirm generation (skip if using --yes flag)
	if !options.SkipPrompt {
		confirmed, err := g.prompter.Confirm("Do you want to proceed with file generation?")
		if err != nil {
			return fmt.Errorf("failed to get confirmation: %w", err)
		}

		if !confirmed {
			fmt.Println("Generation canceled")
			return nil
		}
	}

	// Generate files
	if err := g.generateFiles(); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}

	// Show CLI example if run interactively
	if !options.SkipPrompt {
		g.showCLIExample()
	}

	fmt.Println("generated!")
	return nil
}

func (g *Generator) validateOptions(options *Options) error {
	if options.Answers == nil {
		return fmt.Errorf("answers map is required")
	}

	// Validate that all required questions have answers
	questions := g.config.Questions.GetQuestions()
	for questionKey := range questions {
		if _, exists := options.Answers[questionKey]; !exists {
			return fmt.Errorf("answer for question '%s' is required", questionKey)
		}
	}

	return nil
}

// determineTemplateAndMultiValues determines which question provides the template type and which are multi-value.
func (g *Generator) determineTemplateAndMultiValues() (string, map[string][]string, error) {
	questions := g.config.Questions.GetQuestions()
	multiValueQuestions := make(map[string][]string)
	var templateType string

	// First, collect all multi-value questions
	for questionKey, question := range questions {
		answer := g.answers[questionKey]
		if question.IsMultiple() {
			// This is a multi-value question
			if strSlice, ok := answer.([]string); ok {
				multiValueQuestions[questionKey] = strSlice
			}
		}
	}

	// Determine template type based on configuration or heuristics
	templateQuestionKey := g.config.Questions.GetTemplateQuestion()
	if templateQuestionKey != "" {
		// Use configured template question
		answer, exists := g.answers[templateQuestionKey]
		if !exists {
			return "", nil, fmt.Errorf("template question '%s' not answered", templateQuestionKey)
		}
		if str, ok := answer.(string); ok {
			templateType = str
		} else {
			return "", nil, fmt.Errorf(
				"template question '%s' must have a single string answer, got %T",
				templateQuestionKey, answer,
			)
		}
	} else {
		// Fall back to heuristic: first non-multi question as template type
		questionOrder := g.config.Questions.GetOrder()
		for _, questionKey := range questionOrder {
			question, exists := questions[questionKey]
			if !exists {
				continue
			}

			answer := g.answers[questionKey]
			if !question.IsMultiple() && templateType == "" {
				// Use first single-value question as template type
				if str, ok := answer.(string); ok {
					templateType = str
				}
			}
		}
	}

	if templateType == "" {
		return "", nil, fmt.Errorf("no suitable template type found in answers")
	}

	return templateType, multiValueQuestions, nil
}

// generateCombinations generates all combinations of multi-value questions with single-value answers.
func (g *Generator) generateCombinations(multiValueQuestions map[string][]string) []map[string]interface{} {
	if len(multiValueQuestions) == 0 {
		// No multi-value questions, return single combination with all answers
		return []map[string]interface{}{g.copyAnswers()}
	}

	// Process hierarchical selections by parsing formatted choices
	processedQuestions := g.parseHierarchicalSelections(multiValueQuestions)

	// Extract keys and values for combination generation
	keys := make([]string, 0, len(processedQuestions))
	values := make([][]map[string]string, 0, len(processedQuestions))

	for key, vals := range processedQuestions {
		keys = append(keys, key)
		values = append(values, vals)
	}

	var combinations []map[string]interface{}
	g.generateHierarchicalCombinationsRecursive(keys, values, 0, make(map[string]string), &combinations)

	return combinations
}

func (g *Generator) copyAnswers() map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range g.answers {
		result[key] = value
	}
	return result
}

// parseHierarchicalSelections parses choices that may contain hierarchical format (parent: child)
// and returns a map of question keys to their possible value combinations
func (g *Generator) parseHierarchicalSelections(
	multiValueQuestions map[string][]string,
) map[string][]map[string]string {
	result := make(map[string][]map[string]string)

	for questionKey, selections := range multiValueQuestions {
		var parsedSelections []map[string]string

		for _, selection := range selections {
			// Check if this selection contains hierarchical format (parent: child)
			if strings.Contains(selection, ": ") {
				parts := strings.SplitN(selection, ": ", 2)
				if len(parts) == 2 {
					// This is a hierarchical selection
					parentKey := g.findParentQuestion(questionKey)
					parsedSelections = append(parsedSelections, map[string]string{
						parentKey:   parts[0], // parent value (e.g., "dev")
						questionKey: parts[1], // child value (e.g., "dev-cluster-1")
					})
				}
			} else {
				// This is a regular selection
				parsedSelections = append(parsedSelections, map[string]string{
					questionKey: selection,
				})
			}
		}

		result[questionKey] = parsedSelections
	}

	return result
}

// findParentQuestion finds the parent question for a given question based on dependency configuration
func (g *Generator) findParentQuestion(questionKey string) string {
	questions := g.config.Questions.GetQuestions()
	question, exists := questions[questionKey]
	if !exists {
		return ""
	}

	if question.Type != nil && question.Type.Dynamic != nil {
		deps := question.Type.Dynamic.DependencyQuestions
		if len(deps) > 0 {
			return deps[len(deps)-1] // Return the last (most immediate) dependency
		}
	}

	return ""
}

// generateHierarchicalCombinationsRecursive generates combinations considering hierarchical relationships
func (g *Generator) generateHierarchicalCombinationsRecursive(
	keys []string, values [][]map[string]string, index int,
	current map[string]string, result *[]map[string]interface{},
) {
	if index >= len(keys) {
		// Create a copy of the base answers and override with current combination
		combination := g.copyAnswers()
		for key, value := range current {
			combination[key] = value
		}
		*result = append(*result, combination)
		return
	}

	for _, valueMap := range values[index] {
		// Add all values from this valueMap to current
		for k, v := range valueMap {
			current[k] = v
		}

		g.generateHierarchicalCombinationsRecursive(keys, values, index+1, current, result)

		// Backtrack - remove the values we just added
		for k := range valueMap {
			delete(current, k)
		}
	}
}

func (g *Generator) askQuestion(_ string, question config.Question) (interface{}, error) {
	choices, err := question.GetChoices(g.answers)
	if err != nil {
		return nil, fmt.Errorf("failed to get choices: %w", err)
	}

	if question.IsMultiple() {
		return g.prompter.MultiSelect(question.Prompt, choices)
	}

	if question.Type != nil && question.Type.Interactive {
		return g.prompter.Search(question.Prompt, choices)
	}

	return g.prompter.Select(question.Prompt, choices)
}

func (g *Generator) generatePreview() error {
	fmt.Println("\nOutput:")
	fmt.Println()

	// Determine template type and multi-value questions
	templateType, multiValueQuestions, err := g.determineTemplateAndMultiValues()
	if err != nil {
		return fmt.Errorf("failed to determine template and multi-values: %w", err)
	}

	tmpl, err := template.LoadTemplate(templateType)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Generate all combinations for multi-value questions
	combinations := g.generateCombinations(multiValueQuestions)

	for _, combination := range combinations {
		// Create template data for this combination
		templateData := &template.Data{
			Questions: combination,
		}

		renderResult, err := tmpl.Render(templateData)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}

		// Show preview for all files in the result
		for _, file := range renderResult.Files {
			fullPath := filepath.Join(file.Path, file.Filename)
			fmt.Printf("* %s\n\n", fullPath)

			// Show the rendered content preview
			lines := strings.Split(file.Content, "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("%s\n", line)
				}
			}
			fmt.Println()
		}
	}

	return nil
}

func (g *Generator) generateFiles() error {
	// Determine template type and multi-value questions
	templateType, multiValueQuestions, err := g.determineTemplateAndMultiValues()
	if err != nil {
		return fmt.Errorf("failed to determine template and multi-values: %w", err)
	}

	tmpl, err := template.LoadTemplate(templateType)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Generate all combinations for multi-value questions
	combinations := g.generateCombinations(multiValueQuestions)

	for _, combination := range combinations {
		// Create template data for this combination
		templateData := &template.Data{
			Questions: combination,
		}

		renderResult, err := tmpl.Render(templateData)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}

		// Write all files in the result
		for _, file := range renderResult.Files {
			// Create directory if it doesn't exist
			if err := os.MkdirAll(file.Path, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", file.Path, err)
			}

			// Write file
			fullPath := filepath.Join(file.Path, file.Filename)
			if err := os.WriteFile(fullPath, []byte(file.Content), 0o600); err != nil {
				return fmt.Errorf("failed to write file %s: %w", fullPath, err)
			}
		}
	}

	return nil
}

// showCLIExample displays the CLI command equivalent of the interactive session.
func (g *Generator) showCLIExample() {
	fmt.Println("\nCLI Example:")
	fmt.Print("yg --yes")

	// Get question order from config
	questionOrder := g.config.Questions.GetOrder()
	questions := g.config.Questions.GetQuestions()

	// Generate --answer options in order
	for _, questionKey := range questionOrder {
		answer, exists := g.answers[questionKey]
		if !exists {
			continue
		}

		question, questionExists := questions[questionKey]
		if !questionExists {
			continue
		}

		var answerStr string
		if question.IsMultiple() {
			// Handle multiple selection questions - join with comma
			if strSlice, ok := answer.([]string); ok {
				answerStr = strings.Join(strSlice, ",")
			} else {
				continue // Skip if not string slice
			}
		} else {
			// Handle single selection questions
			if str, ok := answer.(string); ok {
				answerStr = str
			} else {
				continue // Skip if not string
			}
		}

		fmt.Printf(" --answer %s=%s", questionKey, answerStr)
	}

	fmt.Println()
	fmt.Println()
}
