package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testAppTypeDeployment = "deployment"
	testDeploymentContent = `path: {{.Questions.env}}/{{.Questions.cluster}}/deployment
filename: {{.Questions.appName}}-deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}`

	testJobContent = `path: {{.Questions.env}}/{{.Questions.cluster}}/job
filename: {{.Questions.appName}}-job.yaml
---
apiVersion: batch/v1
kind: Job
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}`
)

// MockPrompter implements PrompterInterface for testing
type MockPrompter struct {
	selectResults      []string
	multiSelectResults [][]string
	searchResults      []string
	confirmResults     []bool
	selectIndex        int
	multiSelectIndex   int
	searchIndex        int
	confirmIndex       int
}

func (m *MockPrompter) Reset() {
	m.selectIndex = 0
	m.multiSelectIndex = 0
	m.searchIndex = 0
	m.confirmIndex = 0
}

func (m *MockPrompter) Select(_ string, options []string) (string, error) {
	if m.selectIndex < len(m.selectResults) {
		result := m.selectResults[m.selectIndex]
		m.selectIndex++
		return result, nil
	}
	return options[0], nil
}

func (m *MockPrompter) MultiSelect(_ string, options []string) ([]string, error) {
	if m.multiSelectIndex < len(m.multiSelectResults) {
		result := m.multiSelectResults[m.multiSelectIndex]
		m.multiSelectIndex++
		return result, nil
	}
	if len(options) > 0 {
		return options[:1], nil
	}
	return []string{}, nil
}

func (m *MockPrompter) Search(_ string, options []string) (string, error) {
	if m.searchIndex < len(m.searchResults) {
		result := m.searchResults[m.searchIndex]
		m.searchIndex++
		return result, nil
	}
	return options[0], nil
}

func (m *MockPrompter) Confirm(_ string) (bool, error) {
	if m.confirmIndex < len(m.confirmResults) {
		result := m.confirmResults[m.confirmIndex]
		m.confirmIndex++
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
      prompt: "What type of template do you want to use?"
      choices:
        - deployment
        - job
    appName:
      prompt: "What is the name of your item?"
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
      prompt: "Which environment do you want to target?"
      type:
        multiple: true
      choices:
        - dev
        - staging
    cluster:
      prompt: "Which target destinations do you want to deploy to?"
      type:
        multiple: true
        dynamic:
          dependency_questions: ["env"]
      choices:
        dev:
          - dev-cluster-1
          - dev-cluster-2
          - dev-cluster-3
        staging:
          - staging-cluster-1
          - staging-cluster-2
          - staging-cluster-3`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create template files
	deploymentTemplate := filepath.Join(templateDir, "deployment.yaml")
	deploymentContent := testDeploymentContent

	err = os.WriteFile(deploymentTemplate, []byte(deploymentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write deployment template: %v", err)
	}

	jobTemplate := filepath.Join(templateDir, "job.yaml")
	jobContent := testJobContent

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
			"app":     testAppTypeDeployment,
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
			"app":     testAppTypeDeployment,
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
			"app":     testAppTypeDeployment,
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
		selectResults: []string{testAppTypeDeployment},
	}
	generator.prompter = mockPrompter

	questions := generator.config.Questions.GetQuestions()
	question := questions["app"]
	answer, err := generator.askQuestion("app", question)
	if err != nil {
		t.Fatalf("Failed to ask question: %v", err)
	}

	if answer != testAppTypeDeployment {
		t.Errorf("Expected answer '%s', got %s", testAppTypeDeployment, answer)
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
		selectResults:      []string{testAppTypeDeployment},                         // app selection
		searchResults:      []string{"sample-server-1"},                             // appName search
		multiSelectResults: [][]string{{"dev"}, {"dev-cluster-1", "dev-cluster-2"}}, // env, then cluster
		confirmResults:     []bool{true},                                            // final confirmation
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
		"app": testAppTypeDeployment,
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
		"app":     testAppTypeDeployment,
		"appName": "test-app",
		"env":     []string{"dev", "staging"},
		"cluster": []string{"dev-cluster-1", "staging-cluster-1"},
	}

	templateType, multiValues, err := generator.determineTemplateAndMultiValues()
	if err != nil {
		t.Fatalf("Failed to determine template and multi-values: %v", err)
	}

	if templateType != testAppTypeDeployment {
		t.Errorf("Expected template type '%s', got %s", testAppTypeDeployment, templateType)
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
		"app":     testAppTypeDeployment,
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
		if combination["app"] != testAppTypeDeployment {
			t.Errorf("Combination %d: expected app='%s', got %v", i, testAppTypeDeployment, combination["app"])
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
		"app":     testAppTypeDeployment,
		"appName": "test-app",
	}

	// No multi-value questions
	combinations := generator.generateCombinations(map[string][]string{})

	// Should return single combination with all answers
	if len(combinations) != 1 {
		t.Errorf("Expected 1 combination, got %d", len(combinations))
	}

	if combinations[0]["app"] != testAppTypeDeployment {
		t.Errorf("Expected app='%s', got %v", testAppTypeDeployment, combinations[0]["app"])
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
			"app": testAppTypeDeployment,
			"env": []string{"dev"},
		},
		SkipPrompt: false,
	}

	err = generator.RunWithOptions(options)
	if err != nil {
		t.Fatalf("Failed to run generator with prefilled answers: %v", err)
	}

	// Check that pre-filled answers were used
	if generator.answers["app"] != testAppTypeDeployment {
		t.Errorf("Expected app='%s', got %v", testAppTypeDeployment, generator.answers["app"])
	}
}

func TestDynamicChoicesInRun(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	defer func() { _ = os.Chdir(tempDir) }()

	// Change to temp directory before creating generator
	err := os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Set up answers simulating dynamic choice selection
	// First select multiple environments, then verify cluster choices are combined
	generator.answers = map[string]interface{}{
		"app": testAppTypeDeployment,
		"env": []string{"dev", "staging"}, // Multiple environments
	}

	// Mock the prompter to simulate user selections
	mockPrompter := &MockPrompter{
		selectResults: []string{testAppTypeDeployment}, // app selection
		searchResults: []string{"dynamic-app"},         // appName search
		multiSelectResults: [][]string{
			{"dev", "staging"}, // env multiselect
			{"dev-cluster-1", "dev-cluster-2", "staging-cluster-1"}, // cluster multiselect (combined from multiple envs)
		},
		confirmResults: []bool{true}, // final confirmation
	}
	generator.prompter = mockPrompter

	// Test getting dynamic choices for cluster question
	questions := generator.config.Questions.GetQuestions()
	clusterQuestion, exists := questions["cluster"]
	if !exists {
		t.Fatal("cluster question should exist in config")
	}

	choices, err := clusterQuestion.GetChoices(generator.answers)
	if err != nil {
		t.Fatalf("Failed to get dynamic choices: %v", err)
	}

	// Should get choices from both dev and staging environments in hierarchical format
	if len(choices) < 3 { // At least dev choices should be present
		t.Errorf("Expected at least 3 dynamic choices, got %d: %v", len(choices), choices)
	}

	// Verify that choices contain hierarchical format and expected values from both environments
	devChoicesFound := 0
	stagingChoicesFound := 0
	hierarchicalChoicesFound := 0

	for _, choice := range choices {
		// Check if choice is in hierarchical format
		if strings.Contains(choice, ": ") {
			hierarchicalChoicesFound++

			// Parse hierarchical choice to check parent environment
			parts := strings.SplitN(choice, ": ", 2)
			if len(parts) == 2 {
				parent := parts[0]
				child := parts[1]

				if parent == "dev" && strings.Contains(child, "dev-cluster") {
					devChoicesFound++
				}
				if parent == "staging" && strings.Contains(child, "staging-cluster") {
					stagingChoicesFound++
				}
			}
		}
	}

	if hierarchicalChoicesFound == 0 {
		t.Error("Expected hierarchical choices with format 'parent: child', but none found")
	}
	if devChoicesFound == 0 {
		t.Error("No dev cluster choices found in dynamic selection")
	}
	if stagingChoicesFound == 0 {
		t.Error("No staging cluster choices found in dynamic selection")
	}
}

func setupTestEnvironmentWithTemplateQuestion(t *testing.T) string {
	tempDir := t.TempDir()

	// Create .yg directory structure
	templateDir := filepath.Join(tempDir, ".yg", "_templates")
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp template directory: %v", err)
	}

	// Create config file with template_question specified
	configFile := filepath.Join(templateDir, ".yg-config.yaml")
	configContent := `questions:
  template_question: "appType"  # Explicitly set which question provides template
  order:
    - appType
    - appName
    - env
    - cluster
  definitions:
    appType:
      prompt: "Select application type"
      choices:
        - deployment
        - job
        - microservice
    appName:
      prompt: "What is the name of your item?"
      type:
        interactive: true
      choices:
        - sample-app-1
        - sample-app-2
        - sample-service-1
    env:
      prompt: "Which environment do you want to target?"
      type:
        multiple: true
      choices:
        - dev
        - staging
        - production
    cluster:
      prompt: "Which target destinations do you want to deploy to?"
      type:
        multiple: true
        dynamic:
          dependency_questions: ["env"]
      choices:
        dev:
          - dev-cluster-1
          - dev-cluster-2
        staging:
          - staging-cluster-1
          - staging-cluster-2`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create template files
	deploymentTemplate := filepath.Join(templateDir, "deployment.yaml")
	deploymentContent := testDeploymentContent

	err = os.WriteFile(deploymentTemplate, []byte(deploymentContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write deployment template: %v", err)
	}

	jobTemplate := filepath.Join(templateDir, "job.yaml")
	jobContent := testJobContent

	err = os.WriteFile(jobTemplate, []byte(jobContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write job template: %v", err)
	}

	microserviceTemplate := filepath.Join(templateDir, "microservice.yaml")
	microserviceContent := `path: {{.Questions.env}}/{{.Questions.cluster}}/microservice
filename: {{.Questions.appName}}-microservice.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}
  labels:
    type: microservice`

	err = os.WriteFile(microserviceTemplate, []byte(microserviceContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write microservice template: %v", err)
	}

	return tempDir
}

func TestDetermineTemplateAndMultiValuesWithTemplateQuestion(t *testing.T) {
	tempDir := setupTestEnvironmentWithTemplateQuestion(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Set up test answers - note that order is different from typical config
	generator.answers = map[string]interface{}{
		"appName": "test-app",                // This would be first non-multi in order
		"appType": "microservice",            // This is specified as template_question
		"env":     []string{"dev"},           // Multi-value
		"cluster": []string{"dev-cluster-1"}, // Multi-value
	}

	templateType, multiValues, err := generator.determineTemplateAndMultiValues()
	if err != nil {
		t.Fatalf("Failed to determine template and multi-values: %v", err)
	}

	// Should use appType (configured template_question) not appName (first non-multi)
	if templateType != "microservice" {
		t.Errorf("Expected template type 'microservice', got '%s'", templateType)
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

func TestDetermineTemplateAndMultiValuesWithMissingTemplateQuestion(t *testing.T) {
	tempDir := setupTestEnvironmentWithTemplateQuestion(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Missing the template question answer
	generator.answers = map[string]interface{}{
		"appName": "test-app",
		"env":     []string{"dev"},
		"cluster": []string{"dev-cluster-1"},
		// appType is missing
	}

	_, _, err = generator.determineTemplateAndMultiValues()
	if err == nil {
		t.Error("Expected error when template question is not answered")
	}

	if !strings.Contains(err.Error(), "template question 'appType' not answered") {
		t.Errorf("Expected error about missing template question, got: %v", err)
	}
}

func TestDetermineTemplateAndMultiValuesWithInvalidTemplateAnswer(t *testing.T) {
	tempDir := setupTestEnvironmentWithTemplateQuestion(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Template question has non-string answer
	generator.answers = map[string]interface{}{
		"appName": "test-app",
		"appType": []string{"deployment", "job"}, // Array instead of string
		"env":     []string{"dev"},
		"cluster": []string{"dev-cluster-1"},
	}

	_, _, err = generator.determineTemplateAndMultiValues()
	if err == nil {
		t.Error("Expected error when template question has non-string answer")
	}

	if !strings.Contains(err.Error(), "template question 'appType' must have a single string answer") {
		t.Errorf("Expected error about invalid template answer type, got: %v", err)
	}
}

func TestDetermineTemplateAndMultiValuesHeuristicFallback(t *testing.T) {
	// Test that heuristic fallback still works when template_question is not set
	tempDir := setupTestEnvironment(t) // This config has no template_question
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
		"env":     []string{"dev"},
		"cluster": []string{"dev-cluster-1"},
	}

	templateType, multiValues, err := generator.determineTemplateAndMultiValues()
	if err != nil {
		t.Fatalf("Failed to determine template and multi-values: %v", err)
	}

	// Should use first non-multi question (app) via heuristic
	if templateType != "deployment" {
		t.Errorf("Expected template type 'deployment', got '%s'", templateType)
	}

	if len(multiValues) != 2 {
		t.Errorf("Expected 2 multi-value questions, got %d", len(multiValues))
	}
}

func TestRunWithTemplateQuestion(t *testing.T) {
	tempDir := setupTestEnvironmentWithTemplateQuestion(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Mock the prompter
	mockPrompter := &MockPrompter{
		selectResults:      []string{"microservice"},               // appType selection
		searchResults:      []string{"sample-service-1"},           // appName search
		multiSelectResults: [][]string{{"dev"}, {"dev-cluster-1"}}, // env, then cluster
		confirmResults:     []bool{true},                           // final confirmation
	}
	generator.prompter = mockPrompter

	err = generator.Run()
	if err != nil {
		t.Fatalf("Failed to run generator: %v", err)
	}

	// Check if the microservice template was used
	expectedFile := filepath.Join(tempDir, "dev", "dev-cluster-1", "microservice", "sample-service-1-microservice.yaml")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected microservice file %s was not generated", expectedFile)
	}

	// Read the generated file to verify it used the microservice template
	content, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "type: microservice") {
		t.Error("Generated file should contain 'type: microservice' from template")
	}
}

func TestShowCLIExample(t *testing.T) {
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

	// Capture output
	// Note: Since showCLIExample prints to stdout, we'd need to capture it
	// For this test, we'll just verify it doesn't panic
	generator.showCLIExample()

	// The function should execute without error
	// Visual verification would show:
	// CLI Example:
	// yg --yes --answer app=deployment --answer appName=test-app --answer env=dev,staging --answer cluster=dev-cluster-1,staging-cluster-1
}

func TestShowCLIExampleNoAnswers(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Empty answers
	generator.answers = map[string]interface{}{}

	// Should not panic with empty answers
	generator.showCLIExample()
}

func TestShouldShowPreview(t *testing.T) {
	tempDir := setupTestEnvironment(t)
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test with no preview config (should default to enabled)
	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	options := &Options{}
	if !generator.shouldShowPreview(options) {
		t.Error("Expected preview to be enabled by default")
	}

	// Test with CLI flag disabled
	options.NoPreview = true
	if generator.shouldShowPreview(options) {
		t.Error("Expected preview to be disabled with NoPreview flag")
	}
}

func TestShouldShowPreviewWithConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create config with preview disabled
	configDir := filepath.Join(tempDir, ".yg")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configContent := `questions:
  definitions:
    app:
      prompt: "What type of template?"
      choices:
        - deployment
  order:
    - app
preview:
  enabled: false
`
	configFile := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Test with preview disabled in config
	options := &Options{}
	if generator.shouldShowPreview(options) {
		t.Error("Expected preview to be disabled per config")
	}

	// Test CLI override - CLI should take precedence
	options.NoPreview = false // CLI doesn't disable preview
	if generator.shouldShowPreview(options) {
		t.Error("Expected preview to be disabled per config even with CLI not set")
	}
}

func TestShouldShowPreviewConfigEnabled(t *testing.T) {
	tempDir := t.TempDir()

	// Create config with preview enabled
	configDir := filepath.Join(tempDir, ".yg")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configContent := `questions:
  definitions:
    app:
      prompt: "What type of template?"
      choices:
        - deployment
  order:
    - app
preview:
  enabled: true
`
	configFile := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	generator, err := New()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Test with preview enabled in config
	options := &Options{}
	if !generator.shouldShowPreview(options) {
		t.Error("Expected preview to be enabled per config")
	}

	// Test CLI override takes precedence over config
	options.NoPreview = true
	if generator.shouldShowPreview(options) {
		t.Error("Expected CLI NoPreview to override config enabled setting")
	}
}
