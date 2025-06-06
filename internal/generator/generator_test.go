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

func (m *MockPrompter) Reset() {
	m.callIndex = 0
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
	if len(options) > 0 {
		return options[:1], nil
	}
	return []string{}, nil
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
  order:
    - app
    - appName
    - env
    - cluster
  definitions:
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
		Answers: map[string]interface{}{
			"app":     "deployment",
			"appName": "test-app",
			"env":     []string{"dev"},
			"cluster": []string{"dev-cluster-1"},
		},
		SkipPrompt: true,
	}

	err = generator.validateOptions(validOptions)
	if err != nil {
		t.Errorf("Valid options should not produce error: %v", err)
	}

	// Test missing answers map
	invalidOptions := &Options{
		Answers:    nil,
		SkipPrompt: true,
	}

	err = generator.validateOptions(invalidOptions)
	if err == nil {
		t.Error("Missing answers map should produce error")
	}

	// Test missing required question
	invalidOptions = &Options{
		Answers: map[string]interface{}{
			"app":     "deployment",
			"appName": "test-app",
			"env":     []string{"dev"},
			// missing cluster
		},
		SkipPrompt: true,
	}

	err = generator.validateOptions(invalidOptions)
	if err == nil {
		t.Error("Missing required question should produce error")
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
		Answers: map[string]interface{}{
			"app":     "deployment",
			"appName": "test-app",
			"env":     []string{"dev"},
			"cluster": []string{"dev-cluster-1"},
		},
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

	questions := generator.config.Questions.GetQuestions()
	question := questions["app"]
	answer, err := generator.askQuestion("app", question)
	if err != nil {
		t.Fatalf("Failed to ask question: %v", err)
	}

	if answer != "deployment" {
		t.Errorf("Expected answer 'deployment', got %s", answer)
	}
}

func TestRun(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Mock the prompter for interactive flow
	// Order: app (select) -> appName (search) -> env (multiselect) -> cluster (multiselect) -> confirm
	mockPrompter := &MockPrompter{
		selectResults:      []string{"deployment"},           // app selection
		searchResults:      []string{"sample-server-1"},     // appName search
		multiSelectResults: [][]string{{"dev"}, {"dev-cluster-1"}}, // env, then cluster
		confirmResults:     []bool{true},                    // final confirmation
	}
	generator.prompter = mockPrompter

	err = generator.Run()
	if err != nil {
		t.Fatalf("Failed to run generator: %v", err)
	}

	// Just verify that some YAML file was created (the exact path might vary)
	found := false
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".yaml" {
			found = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Error walking directory: %v", err)
	}
	
	if !found {
		t.Error("No YAML files were generated")
	}
}

func TestAskQuestionMultiSelect(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	mockPrompter := &MockPrompter{
		multiSelectResults: [][]string{{"dev", "staging"}},
	}
	generator.prompter = mockPrompter

	questions := generator.config.Questions.GetQuestions()
	question := questions["env"]
	answer, err := generator.askQuestion("env", question)
	if err != nil {
		t.Fatalf("Failed to ask multi-select question: %v", err)
	}

	if answerSlice, ok := answer.([]string); !ok {
		t.Errorf("Expected answer to be []string, got %T", answer)
	} else if len(answerSlice) != 2 {
		t.Errorf("Expected 2 selections, got %d", len(answerSlice))
	}
}

func TestAskQuestionInteractiveSearch(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Set up generator answers to satisfy dependencies
	generator.answers = map[string]interface{}{
		"app": "deployment",
	}

	mockPrompter := &MockPrompter{
		searchResults: []string{"sample-server-1"},
	}
	generator.prompter = mockPrompter

	questions := generator.config.Questions.GetQuestions()
	question := questions["appName"]
	answer, err := generator.askQuestion("appName", question)
	if err != nil {
		t.Fatalf("Failed to ask interactive search question: %v", err)
	}

	if answer != "sample-server-1" {
		t.Errorf("Expected answer 'sample-server-1', got %s", answer)
	}
}

func TestDetermineTemplateAndMultiValues(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Set up test answers
	generator.answers = map[string]interface{}{
		"app":     "deployment",
		"appName": "test-app",
		"env":     []string{"dev", "staging"},
		"cluster": []string{"dev-cluster-1", "staging-cluster-1"},
	}

	templateType, multiValues, err := generator.determineTemplateAndMultiValues()
	if err != nil {
		t.Fatalf("Failed to determine template and multi-values: %v", err)
	}

	if templateType != "deployment" {
		t.Errorf("Expected template type 'deployment', got %s", templateType)
	}

	if len(multiValues) != 2 {
		t.Errorf("Expected 2 multi-value questions, got %d", len(multiValues))
	}

	if _, exists := multiValues["env"]; !exists {
		t.Error("Expected 'env' in multi-value questions")
	}

	if _, exists := multiValues["cluster"]; !exists {
		t.Error("Expected 'cluster' in multi-value questions")
	}
}

func TestGenerateCombinations(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	generator.answers = map[string]interface{}{
		"app":     "deployment",
		"appName": "test-app",
	}

	multiValues := map[string][]string{
		"env":     {"dev", "staging"},
		"cluster": {"cluster-1", "cluster-2"},
	}

	combinations := generator.generateCombinations(multiValues)
	
	// Should generate 2 * 2 = 4 combinations
	expectedCount := 4
	if len(combinations) != expectedCount {
		t.Errorf("Expected %d combinations, got %d", expectedCount, len(combinations))
	}

	// Each combination should have all base answers plus the specific multi-value selections
	for i, combination := range combinations {
		if combination["app"] != "deployment" {
			t.Errorf("Combination %d: expected app='deployment', got %v", i, combination["app"])
		}
		if combination["appName"] != "test-app" {
			t.Errorf("Combination %d: expected appName='test-app', got %v", i, combination["appName"])
		}
		if _, exists := combination["env"]; !exists {
			t.Errorf("Combination %d: missing env", i)
		}
		if _, exists := combination["cluster"]; !exists {
			t.Errorf("Combination %d: missing cluster", i)
		}
	}
}

func TestGenerateCombinationsEmpty(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	generator.answers = map[string]interface{}{
		"app":     "deployment",
		"appName": "test-app",
	}

	// No multi-value questions
	combinations := generator.generateCombinations(map[string][]string{})
	
	// Should return single combination with all answers
	if len(combinations) != 1 {
		t.Errorf("Expected 1 combination, got %d", len(combinations))
	}

	if combinations[0]["app"] != "deployment" {
		t.Errorf("Expected app='deployment', got %v", combinations[0]["app"])
	}
}

func TestRunWithOptionsPrefilledAnswers(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Mock prompter for remaining questions
	mockPrompter := &MockPrompter{
		searchResults:      []string{"sample-server-1"},
		multiSelectResults: [][]string{{"dev-cluster-1"}},
		confirmResults:     []bool{true},
	}
	generator.prompter = mockPrompter

	// Pre-fill some answers
	options := &Options{
		Answers: map[string]interface{}{
			"app": "deployment",
			"env": []string{"dev"},
		},
		SkipPrompt: false,
	}

	err = generator.RunWithOptions(options)
	if err != nil {
		t.Fatalf("Failed to run generator with prefilled answers: %v", err)
	}

	// Check that pre-filled answers were used
	if generator.answers["app"] != "deployment" {
		t.Errorf("Expected app='deployment', got %v", generator.answers["app"])
	}
}
