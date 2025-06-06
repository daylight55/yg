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
)

var rootCmd = &cobra.Command{
	Use:   "yg",
	Short: "YAML template generator",
	Long:  `A CLI tool to generate YAML files from templates based on interactive prompts.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Load config to get available questions for validation
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Convert CLI answers to the format expected by generator
		generatorAnswers := make(map[string]interface{})
		questions := cfg.Questions.GetQuestions()

		for questionKey, question := range questions {
			if answerStr, exists := answers[questionKey]; exists {
				if question.Multiple {
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

	// Keep backward compatibility flags for common questions
	rootCmd.Flags().StringVar(new(string), "app", "", "Application type (deprecated: use --answer app=value)")
	rootCmd.Flags().StringVar(new(string), "name", "", "Application name (deprecated: use --answer appName=value)")
	rootCmd.Flags().StringSlice("env", []string{}, "Environments (deprecated: use --answer env=value1,value2)")
	rootCmd.Flags().StringSlice("cluster", []string{}, "Clusters (deprecated: use --answer cluster=value1,value2)")

	// Mark deprecated flags as deprecated
	_ = rootCmd.Flags().MarkDeprecated("app", "use --answer app=value instead")
	_ = rootCmd.Flags().MarkDeprecated("name", "use --answer appName=value instead")
	_ = rootCmd.Flags().MarkDeprecated("env", "use --answer env=value1,value2 instead")
	_ = rootCmd.Flags().MarkDeprecated("cluster", "use --answer cluster=value1,value2 instead")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runGenerator(options *generator.Options) error {
	generator, err := generator.New()
	if err != nil {
		return fmt.Errorf("failed to initialize generator: %w", err)
	}

	return generator.RunWithOptions(options)
}
