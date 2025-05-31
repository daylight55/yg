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
	AppType    string
	AppName    string
	Envs       []string
	Clusters   []string
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
	cfg, err := config.LoadConfig()
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
		g.answers["app"] = options.AppType
		g.answers["appName"] = options.AppName
		g.answers["env"] = options.Envs
		g.answers["cluster"] = options.Clusters
	} else {
		// Pre-fill answers with CLI options if provided
		if options.AppType != "" {
			g.answers["app"] = options.AppType
		}
		if options.AppName != "" {
			g.answers["appName"] = options.AppName
		}
		if len(options.Envs) > 0 {
			g.answers["env"] = options.Envs
		}
		if len(options.Clusters) > 0 {
			g.answers["cluster"] = options.Clusters
		}

		// Process questions in order
		questionOrder := []string{"app", "appName", "env", "cluster"}

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

			question, exists := g.config.Questions[questionKey]
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
		confirmed, err := g.prompter.Confirm("出力して問題ないですか?")
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

	fmt.Println("generated!")
	return nil
}

func (g *Generator) validateOptions(options *Options) error {
	if options.AppType == "" {
		return fmt.Errorf("app type is required")
	}
	if options.AppName == "" {
		return fmt.Errorf("app name is required")
	}
	if len(options.Envs) == 0 {
		return fmt.Errorf("at least one environment is required")
	}
	if len(options.Clusters) == 0 {
		return fmt.Errorf("at least one cluster is required")
	}
	return nil
}

func (g *Generator) askQuestion(_ string, question config.Question) (interface{}, error) {
	choices, err := question.GetChoices(g.answers)
	if err != nil {
		return nil, fmt.Errorf("failed to get choices: %w", err)
	}

	if question.Multiple {
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

	appType := g.answers["app"].(string)
	appName := g.answers["appName"].(string)
	envs := g.answers["env"].([]string)
	clusters := g.answers["cluster"].([]string)

	tmpl, err := template.LoadTemplate(appType)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	for _, env := range envs {
		for _, cluster := range clusters {
			// Create template data for this combination
			templateData := &template.Data{
				Questions: map[string]interface{}{
					"app":     appType,
					"appName": appName,
					"env":     env,
					"cluster": cluster,
				},
			}

			path, filename, content, err := tmpl.Render(templateData)
			if err != nil {
				return fmt.Errorf("failed to render template: %w", err)
			}

			fullPath := filepath.Join(path, filename)
			fmt.Printf("* %s\n\n", fullPath)

			// Show the rendered content preview
			lines := strings.Split(content, "\n")
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
	appType := g.answers["app"].(string)
	appName := g.answers["appName"].(string)
	envs := g.answers["env"].([]string)
	clusters := g.answers["cluster"].([]string)

	tmpl, err := template.LoadTemplate(appType)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	for _, env := range envs {
		for _, cluster := range clusters {
			// Create template data for this combination
			templateData := &template.Data{
				Questions: map[string]interface{}{
					"app":     appType,
					"appName": appName,
					"env":     env,
					"cluster": cluster,
				},
			}

			path, filename, content, err := tmpl.Render(templateData)
			if err != nil {
				return fmt.Errorf("failed to render template: %w", err)
			}

			// Create directory if it doesn't exist
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}

			// Write file
			fullPath := filepath.Join(path, filename)
			if err := os.WriteFile(fullPath, []byte(content), 0600); err != nil {
				return fmt.Errorf("failed to write file %s: %w", fullPath, err)
			}
		}
	}

	return nil
}
