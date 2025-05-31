package cmd

import (
	"fmt"
	"os"

	"github.com/daylight55/yg/internal/generator"
	"github.com/spf13/cobra"
)

var (
	appType    string
	appName    string
	envs       []string
	clusters   []string
	skipPrompt bool
)

var rootCmd = &cobra.Command{
	Use:   "yg",
	Short: "YAML template generator",
	Long:  `A CLI tool to generate YAML files from templates based on interactive prompts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		options := &generator.Options{
			AppType:    appType,
			AppName:    appName,
			Envs:       envs,
			Clusters:   clusters,
			SkipPrompt: skipPrompt,
		}
		return runGenerator(options)
	},
}

func init() {
	rootCmd.Flags().StringVar(&appType, "app", "", "Application type (deployment, job)")
	rootCmd.Flags().StringVar(&appName, "name", "", "Application name")
	rootCmd.Flags().StringSliceVar(&envs, "env", []string{}, "Environments (dev, staging, production)")
	rootCmd.Flags().StringSliceVar(&clusters, "cluster", []string{}, "Clusters")
	rootCmd.Flags().BoolVar(&skipPrompt, "yes", false, "Skip prompts and use provided values")
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
