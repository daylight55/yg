package generator

import (
	"os"
	"path/filepath"
	"testing"
)

// MockPrompter implements PrompterInterface for testing
type MockPrompter struct {
	selectResults      []string
	multiSelectResults [][]string
	searchResults      []string
	confirmResults     []bool
	callIndex          int
}

func (m *MockPrompter) Select(_ string, options []string) (string, error) {
	if m.callIndex < len(m.selectResults) {
		result := m.selectResults[m.callIndex]
		m.callIndex++
		return result, nil
	}
	return options[0], nil
}

func (m *MockPrompter) MultiSelect(_ string, options []string) ([]string, error) {
	if m.callIndex < len(m.multiSelectResults) {
		result := m.multiSelectResults[m.callIndex]
		m.callIndex++
		return result, nil
	}
	return options[:1], nil
}

func (m *MockPrompter) Search(_ string, options []string) (string, error) {
	if m.callIndex < len(m.searchResults) {
		result := m.searchResults[m.callIndex]
		m.callIndex++
		return result, nil
	}
	return options[0], nil
}

func (m *MockPrompter) Confirm(_ string) (bool, error) {
	if m.callIndex < len(m.confirmResults) {
		result := m.confirmResults[m.callIndex]
		m.callIndex++
		return result, nil
	}
	return true, nil
}

func setupTestEnvironment(t *testing.T) string {
	tempDir := t.TempDir()

	// Create .yg directory structure
	templateDir := filepath.Join(tempDir, ".yg", "_templates")
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp template directory: %v", err)
	}

	// Create config file
	configFile := filepath.Join(templateDir, ".yg-config.yaml")
	configContent := `questions:
  app:
    prompt: "アプリの種類はなんですか？"
    choices:
      - deployment
      - job
  appName:
    prompt: "アプリ名は何ですか？"
    type:
      dynamic:
        dependency_questions: ["app"]
      interactive: true
    choices:
      deployment:
        - sample-server-1
        - sample-server-2
      job:
        - sample-job-1
        - sample-job-2
  env:
    prompt: "環境名はなんですか？"
    multiple: true
    choices:
      - dev
      - staging
  cluster:
    prompt: "クラスターはどこですか？"
    multiple: true
    type:
      dynamic:
        dependency_questions: ["env"]
    choices:
      dev:
        - dev-cluster-1
        - dev-cluster-2
      staging:
        - staging-cluster-1`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create template files
	deploymentTemplate := filepath.Join(templateDir, "deployment.yaml")
	deploymentContent := `path: {{.Questions.env}}/{{.Questions.cluster}}/deployment
filename: {{.Questions.appName}}-deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}`

	err = os.WriteFile(deploymentTemplate, []byte(deploymentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write deployment template: %v", err)
	}

	jobTemplate := filepath.Join(templateDir, "job.yaml")
	jobContent := `path: {{.Questions.env}}/{{.Questions.cluster}}/job
filename: {{.Questions.appName}}-job.yaml
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}`

	err = os.WriteFile(jobTemplate, []byte(jobContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write job template: %v", err)
	}

	return tempDir
}

func TestNewGenerator(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if generator.config == nil {
		t.Error("Config should not be nil")
	}

	if generator.prompter == nil {
		t.Error("Prompter should not be nil")
	}

	if generator.answers == nil {
		t.Error("Answers should not be nil")
	}
}

func TestValidateOptions(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Test valid options
	validOptions := &Options{
		AppType:  "deployment",
		AppName:  "test-app",
		Envs:     []string{"dev"},
		Clusters: []string{"dev-cluster-1"},
	}

	err = generator.validateOptions(validOptions)
	if err != nil {
		t.Errorf("Valid options should not produce error: %v", err)
	}

	// Test missing app type
	invalidOptions := &Options{
		AppName:  "test-app",
		Envs:     []string{"dev"},
		Clusters: []string{"dev-cluster-1"},
	}

	err = generator.validateOptions(invalidOptions)
	if err == nil {
		t.Error("Missing app type should produce error")
	}

	// Test missing app name
	invalidOptions = &Options{
		AppType:  "deployment",
		Envs:     []string{"dev"},
		Clusters: []string{"dev-cluster-1"},
	}

	err = generator.validateOptions(invalidOptions)
	if err == nil {
		t.Error("Missing app name should produce error")
	}

	// Test missing environments
	invalidOptions = &Options{
		AppType:  "deployment",
		AppName:  "test-app",
		Clusters: []string{"dev-cluster-1"},
	}

	err = generator.validateOptions(invalidOptions)
	if err == nil {
		t.Error("Missing environments should produce error")
	}

	// Test missing clusters
	invalidOptions = &Options{
		AppType: "deployment",
		AppName: "test-app",
		Envs:    []string{"dev"},
	}

	err = generator.validateOptions(invalidOptions)
	if err == nil {
		t.Error("Missing clusters should produce error")
	}
}

func TestRunWithOptionsSkipPrompt(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Mock the prompter
	mockPrompter := &MockPrompter{
		confirmResults: []bool{true}, // Confirm generation
	}
	generator.prompter = mockPrompter

	options := &Options{
		AppType:    "deployment",
		AppName:    "test-app",
		Envs:       []string{"dev"},
		Clusters:   []string{"dev-cluster-1"},
		SkipPrompt: true,
	}

	err = generator.RunWithOptions(options)
	if err != nil {
		t.Fatalf("Failed to run generator with options: %v", err)
	}

	// Check if files were generated
	expectedFile := filepath.Join(tempDir, "dev", "dev-cluster-1", "deployment", "test-app-deployment.yaml")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s was not generated", expectedFile)
	}
}

func TestAskQuestion(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Mock the prompter for regular select
	mockPrompter := &MockPrompter{
		selectResults: []string{"deployment"},
	}
	generator.prompter = mockPrompter

	question := generator.config.Questions["app"]
	answer, err := generator.askQuestion("app", question)
	if err != nil {
		t.Fatalf("Failed to ask question: %v", err)
	}

	if answer != "deployment" {
		t.Errorf("Expected answer 'deployment', got %s", answer)
	}
}
