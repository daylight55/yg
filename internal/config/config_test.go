package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testAppPrompt     = "What type of template do you want to use?"
	testConfigContent = `questions:
  definitions:
    app:
      prompt: "What type of template do you want to use?"
      choices:
        - deployment
        - job
    env:
      prompt: "Which environment do you want to target?"
      type:
        multiple: true
      choices:
        - dev
        - staging
  order:
    - app
    - env
`
	simpleAppConfig = `questions:
  definitions:
    app:
      prompt: "What type of template do you want to use?"
      choices:
        - deployment
        - job
  order:
    - app
`
)

func TestLoadConfig(t *testing.T) {
	// Create temporary config directory and file
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yg", "_templates")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	configFile := filepath.Join(configDir, ".yg-config.yaml")
	configContent := testConfigContent

	err = os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Change working directory to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test loading config
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	questions := config.Questions.GetQuestions()
	if len(questions) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(questions))
	}

	if questions["app"].Prompt != testAppPrompt {
		t.Errorf("Unexpected app prompt: %s", questions["app"].Prompt)
	}

	envQuestion := questions["env"]
	if !envQuestion.IsMultiple() {
		t.Error("Expected env question to be multiple choice")
	}

	order := config.Questions.GetOrder()
	expectedOrder := []string{"app", "env"}
	if len(order) != len(expectedOrder) {
		t.Errorf("Expected order length %d, got %d", len(expectedOrder), len(order))
	}
	for i, expected := range expectedOrder {
		if i >= len(order) || order[i] != expected {
			t.Errorf("Expected order[%d] = %s, got %v", i, expected, order)
		}
	}
}

func TestLoadConfigNewPath(t *testing.T) {
	// Test new .yg/config.yaml path
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yg")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := simpleAppConfig

	err = os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Change working directory to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test loading config
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	questions := config.Questions.GetQuestions()
	if len(questions) != 1 {
		t.Errorf("Expected 1 question, got %d", len(questions))
	}

	if questions["app"].Prompt != testAppPrompt {
		t.Errorf("Unexpected app prompt: %s", questions["app"].Prompt)
	}
}

func TestLoadConfigSpecificPath(t *testing.T) {
	// Test loading config from specific path
	tempDir := t.TempDir()
	customConfigPath := filepath.Join(tempDir, "custom-config.yaml")
	configContent := `questions:
  definitions:
    custom:
      prompt: "Custom question?"
      choices:
        - option1
        - option2
  order:
    - custom
`

	err := os.WriteFile(customConfigPath, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write custom config file: %v", err)
	}

	// Test loading config with specific path
	config, err := LoadConfig(customConfigPath)
	if err != nil {
		t.Fatalf("Failed to load config from specific path: %v", err)
	}

	questions := config.Questions.GetQuestions()
	if len(questions) != 1 {
		t.Errorf("Expected 1 question, got %d", len(questions))
	}

	if questions["custom"].Prompt != "Custom question?" {
		t.Errorf("Unexpected custom prompt: %s", questions["custom"].Prompt)
	}
}

func TestLoadConfigBackwardCompatibility(t *testing.T) {
	// Test old format without order/definitions structure
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yg", "_templates")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	configFile := filepath.Join(configDir, ".yg-config.yaml")
	configContent := `questions:
  app:
    prompt: "What type of template do you want to use?"
    choices:
      - deployment
      - job
  env:
    prompt: "Which environment do you want to target?"
    type:
      multiple: true
    choices:
      - dev
      - staging
`

	err = os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Change working directory to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test loading config
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load old format config: %v", err)
	}

	questions := config.Questions.GetQuestions()
	if len(questions) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(questions))
	}

	if questions["app"].Prompt != testAppPrompt {
		t.Errorf("Unexpected app prompt: %s", questions["app"].Prompt)
	}

	envQuestion := questions["env"]
	if !envQuestion.IsMultiple() {
		t.Error("Expected env question to be multiple choice")
	}

	// Order should be auto-generated for old format
	order := config.Questions.GetOrder()
	if len(order) != 2 {
		t.Errorf("Expected order length 2, got %d", len(order))
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	_, err := LoadConfig("")
	if err == nil {
		t.Error("Expected error when config file doesn't exist")
	}
}

func TestLoadConfigSpecificPathNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error when specific config file doesn't exist")
	}
}

func TestQuestionGetChoices(t *testing.T) {
	// Test static choices
	question := Question{
		Choices: []interface{}{"choice1", "choice2", "choice3"},
	}

	choices, err := question.GetChoices(nil)
	if err != nil {
		t.Fatalf("Failed to get static choices: %v", err)
	}

	expected := []string{"choice1", "choice2", "choice3"}
	if len(choices) != len(expected) {
		t.Errorf("Expected %d choices, got %d", len(expected), len(choices))
	}

	for i, choice := range choices {
		if choice != expected[i] {
			t.Errorf("Expected choice %s, got %s", expected[i], choice)
		}
	}
}

func TestQuestionGetDynamicChoices(t *testing.T) {
	// Test dynamic choices for cluster based on env
	question := Question{
		Type: &QuestionType{
			Dynamic: &DynamicType{
				DependencyQuestions: []string{"env"},
			},
		},
		Choices: map[string]interface{}{
			"dev":     []interface{}{"dev-cluster-1", "dev-cluster-2"},
			"staging": []interface{}{"staging-cluster-1"},
		},
	}

	answers := map[string]interface{}{
		"env": []string{"dev", "staging"},
	}

	choices, err := question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get dynamic choices: %v", err)
	}

	// Should return all clusters from both environments in hierarchical format
	expectedHierarchicalChoices := []string{
		"dev: dev-cluster-1", "dev: dev-cluster-2", "staging: staging-cluster-1",
	}
	if len(choices) != len(expectedHierarchicalChoices) {
		t.Errorf("Expected %d choices, got %d", len(expectedHierarchicalChoices), len(choices))
	}

	choiceMap := make(map[string]bool)
	for _, choice := range choices {
		choiceMap[choice] = true
	}

	for _, expected := range expectedHierarchicalChoices {
		if !choiceMap[expected] {
			t.Errorf("Expected hierarchical choice %s not found", expected)
		}
	}
}

// ... rest of test functions would continue here to maintain the full functionality

func TestQuestionGetChoicesHierarchicalMultipleSelection(t *testing.T) {
	// Test the new hierarchical multiple selection implementation
	question := Question{
		Type: &QuestionType{
			Dynamic: &DynamicType{
				DependencyQuestions: []string{"env"},
			},
		},
		Choices: map[string]interface{}{
			"dev": []interface{}{"dev-cluster-1", "dev-cluster-2", "dev-cluster-3"},
			"stg": []interface{}{"stg-cluster-1", "stg-cluster-2", "stg-cluster-3"},
		},
	}

	// When multiple environments are selected, choices should show hierarchy
	answers := map[string]interface{}{
		"env": []string{"dev", "stg"},
	}

	choices, err := question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get choices: %v", err)
	}

	// New implementation should return formatted choices showing hierarchy
	t.Logf("New implementation returns: %v", choices)

	// Expected behavior: choices formatted as "parent: child" to preserve hierarchy
	expectedFormattedChoices := []string{
		"dev: dev-cluster-1", "dev: dev-cluster-2", "dev: dev-cluster-3",
		"stg: stg-cluster-1", "stg: stg-cluster-2", "stg: stg-cluster-3",
	}

	if len(choices) != len(expectedFormattedChoices) {
		t.Errorf("Expected %d choices, got %d", len(expectedFormattedChoices), len(choices))
	}

	// Verify that choices contain hierarchical format
	hierarchicalChoices := 0
	for _, choice := range choices {
		if strings.Contains(choice, ": ") {
			hierarchicalChoices++
		}
	}

	if hierarchicalChoices != len(choices) {
		t.Errorf("Expected all choices to be hierarchical, got %d out of %d", hierarchicalChoices, len(choices))
	}
}

func TestLoadConfigWithPreview(t *testing.T) {
	// Test loading config with preview configuration
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yg")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := simpleAppConfig + `preview:
  enabled: false
`

	err = os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Change working directory to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test loading config
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test preview setting
	if config.Preview == nil {
		t.Fatal("Expected Preview config to be set")
	}

	if config.Preview.Enabled {
		t.Error("Expected preview to be disabled")
	}
}

func TestLoadConfigWithoutPreview(t *testing.T) {
	// Test loading config without preview configuration (should default to nil)
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yg")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := simpleAppConfig

	err = os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Change working directory to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test loading config
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Preview should be nil (not configured)
	if config.Preview != nil {
		t.Error("Expected Preview config to be nil when not configured")
	}
}

func TestLoadConfigPreviewEnabled(t *testing.T) {
	// Test loading config with preview enabled
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yg")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := simpleAppConfig + `preview:
  enabled: true
`

	err = os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Change working directory to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	// Test loading config
	config, err := LoadConfig("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test preview setting
	if config.Preview == nil {
		t.Fatal("Expected Preview config to be set")
	}

	if !config.Preview.Enabled {
		t.Error("Expected preview to be enabled")
	}
}