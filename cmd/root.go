package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/daylight55/yg/internal/config"
	"github.com/daylight55/yg/internal/generator"
	"github.com/spf13/cobra"
)

var (
	answers    map[string]string
	skipPrompt bool
	configPath string
)

var rootCmd = &cobra.Command{
	Use:   "yg",
	Short: "YAML template generator",
	Long:  `A CLI tool to generate YAML files from templates based on interactive prompts.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Load config to get available questions for validation
		cfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Convert CLI answers to the format expected by generator
		generatorAnswers := make(map[string]interface{})
		questions := cfg.Questions.GetQuestions()

		for questionKey, question := range questions {
			if answerStr, exists := answers[questionKey]; exists {
				if question.IsMultiple() {
					// Split comma-separated values for multi-select questions
					generatorAnswers[questionKey] = strings.Split(answerStr, ",")
				} else {
					generatorAnswers[questionKey] = answerStr
				}
			}
		}

		options := &generator.Options{
			Answers:    generatorAnswers,
			SkipPrompt: skipPrompt,
		}
		return runGenerator(options)
	},
}

func init() {
	answers = make(map[string]string)

	// Dynamic flag creation based on config
	// For now, use StringToString flag to accept arbitrary key-value pairs
	rootCmd.Flags().StringToStringVar(&answers, "answer", map[string]string{}, "Answers for questions in format key=value")
	rootCmd.Flags().BoolVar(&skipPrompt, "yes", false, "Skip prompts and use provided values")
	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: ./.yg/config.yaml or ./.yg/config.yml)")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runGenerator(options *generator.Options) error {
	gen, err := generator.NewWithConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to initialize generator: %w", err)
	}

	return gen.RunWithOptions(options)
}
