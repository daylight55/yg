package config

import (
	"os"
	"path/filepath"
	"testing"
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
	configContent := `questions:
  app:
    prompt: "アプリの種類はなんですか？"
    choices:
      - deployment
      - job
  env:
    prompt: "環境名はなんですか？"
    multiple: true
    choices:
      - dev
      - staging
`

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp config file: %v", err)
	}

	// Change working directory to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Test loading config
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.Questions) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(config.Questions))
	}

	if config.Questions["app"].Prompt != "アプリの種類はなんですか？" {
		t.Errorf("Unexpected app prompt: %s", config.Questions["app"].Prompt)
	}

	if !config.Questions["env"].Multiple {
		t.Error("Expected env question to be multiple choice")
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error when config file doesn't exist")
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
			"dev": []interface{}{"dev-cluster-1", "dev-cluster-2"},
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

	// Should return all clusters from both environments
	expectedChoices := []string{"dev-cluster-1", "dev-cluster-2", "staging-cluster-1"}
	if len(choices) != len(expectedChoices) {
		t.Errorf("Expected %d choices, got %d", len(expectedChoices), len(choices))
	}

	choiceMap := make(map[string]bool)
	for _, choice := range choices {
		choiceMap[choice] = true
	}

	for _, expected := range expectedChoices {
		if !choiceMap[expected] {
			t.Errorf("Expected choice %s not found", expected)
		}
	}
}

func TestQuestionGetChoicesInvalidType(t *testing.T) {
	question := Question{
		Choices: "invalid_choice_type",
	}

	_, err := question.GetChoices(nil)
	if err == nil {
		t.Error("Expected error for invalid choices type")
	}
}

func TestQuestionGetDynamicChoicesErrorCases(t *testing.T) {
	// Test missing dependency answer
	question := Question{
		Type: &QuestionType{
			Dynamic: &DynamicType{
				DependencyQuestions: []string{"env"},
			},
		},
		Choices: map[string]interface{}{
			"dev": []interface{}{"dev-cluster-1"},
		},
	}

	_, err := question.GetChoices(map[string]interface{}{})
	if err == nil {
		t.Error("Expected error when dependency answer is missing")
	}

	// Test invalid env answer type
	answers := map[string]interface{}{
		"env": 123, // Invalid type
	}

	_, err = question.GetChoices(answers)
	if err == nil {
		t.Error("Expected error for invalid env answer type")
	}

	// Test single string env answer
	answers = map[string]interface{}{
		"env": "dev",
	}

	choices, err := question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get choices for single string env: %v", err)
	}

	expected := []string{"dev-cluster-1"}
	if len(choices) != len(expected) {
		t.Errorf("Expected %d choices, got %d", len(expected), len(choices))
	}
}

func TestQuestionGetChoicesStaticMap(t *testing.T) {
	// Test map choices without dynamic type (should fail)
	question := Question{
		Choices: map[string]interface{}{
			"dev": []interface{}{"dev-cluster-1"},
		},
	}

	_, err := question.GetChoices(nil)
	if err == nil {
		t.Error("Expected error for map choices without dynamic type")
	}
}