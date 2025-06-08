package config

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	testAppPrompt     = "アプリの種類はなんですか？"
	testConfigContent = `questions:
  definitions:
    app:
      prompt: "アプリの種類はなんですか？"
      choices:
        - deployment
        - job
    env:
      prompt: "環境名はなんですか？"
      type:
        multiple: true
      choices:
        - dev
        - staging
  order:
    - app
    - env
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
	configContent := `questions:
  definitions:
    app:
      prompt: "アプリの種類はなんですか？"
      choices:
        - deployment
        - job
  order:
    - app
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
    prompt: "アプリの種類はなんですか？"
    choices:
      - deployment
      - job
  env:
    prompt: "環境名はなんですか？"
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

func TestQuestionsGetOrderEmpty(t *testing.T) {
	// Test GetOrder when order is empty - should generate from definitions
	questions := &Questions{
		Definitions: map[string]Question{
			"z-last":  {Prompt: "Last"},
			"a-first": {Prompt: "First"},
			"m-mid":   {Prompt: "Middle"},
		},
	}

	order := questions.GetOrder()
	if len(order) != 3 {
		t.Errorf("Expected 3 items in generated order, got %d", len(order))
	}

	// Should contain all keys (order not guaranteed since it's from map iteration)
	orderMap := make(map[string]bool)
	for _, key := range order {
		orderMap[key] = true
	}

	expectedKeys := []string{"z-last", "a-first", "m-mid"}
	for _, key := range expectedKeys {
		if !orderMap[key] {
			t.Errorf("Expected key %s not found in generated order", key)
		}
	}
}

func TestQuestionsGetOrderWithOrder(t *testing.T) {
	// Test GetOrder when explicit order is provided
	questions := &Questions{
		Order: []string{"second", "first", "third"},
		Definitions: map[string]Question{
			"first":  {Prompt: "First"},
			"second": {Prompt: "Second"},
			"third":  {Prompt: "Third"},
		},
	}

	order := questions.GetOrder()
	expected := []string{"second", "first", "third"}

	if len(order) != len(expected) {
		t.Errorf("Expected order length %d, got %d", len(expected), len(order))
	}

	for i, expectedKey := range expected {
		if i >= len(order) || order[i] != expectedKey {
			t.Errorf("Expected order[%d] = %s, got %s", i, expectedKey, order[i])
		}
	}
}

func TestQuestionsGetQuestionsDirectMap(t *testing.T) {
	// Test GetQuestions with DirectMap (old format)
	questions := &Questions{
		DirectMap: map[string]Question{
			"app": {Prompt: "App type"},
			"env": {Prompt: "Environment"},
		},
	}

	result := questions.GetQuestions()
	if len(result) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(result))
	}

	if result["app"].Prompt != "App type" {
		t.Errorf("Expected app prompt 'App type', got %s", result["app"].Prompt)
	}
}

func TestQuestionsNormalizeOldFormat(t *testing.T) {
	// Test normalize with old format (DirectMap)
	questions := &Questions{
		DirectMap: map[string]Question{
			"app": {Prompt: "App type"},
			"env": {Prompt: "Environment"},
		},
	}

	questions.normalize()

	// Should move DirectMap to Definitions
	if questions.Definitions == nil {
		t.Error("Expected Definitions to be set after normalize")
	}

	if len(questions.Definitions) != 2 {
		t.Errorf("Expected 2 definitions, got %d", len(questions.Definitions))
	}

	// Should generate order
	if len(questions.Order) != 2 {
		t.Errorf("Expected 2 items in order, got %d", len(questions.Order))
	}

	// DirectMap should be cleared
	if questions.DirectMap != nil {
		t.Error("Expected DirectMap to be nil after normalize")
	}
}

func TestQuestionGetChoicesNonEnvDependency(t *testing.T) {
	// Test dynamic choices with non-env dependency
	question := Question{
		Type: &QuestionType{
			Dynamic: &DynamicType{
				DependencyQuestions: []string{"app"},
			},
		},
		Choices: map[string]interface{}{
			"deployment": []interface{}{"deploy-choice1", "deploy-choice2"},
			"job":        []interface{}{"job-choice1"},
		},
	}

	answers := map[string]interface{}{
		"app": "deployment",
	}

	choices, err := question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get choices for app dependency: %v", err)
	}

	expected := []string{"deploy-choice1", "deploy-choice2"}
	if len(choices) != len(expected) {
		t.Errorf("Expected %d choices, got %d", len(expected), len(choices))
	}

	for i, choice := range choices {
		if choice != expected[i] {
			t.Errorf("Expected choice %s, got %s", expected[i], choice)
		}
	}
}

func TestQuestionGetChoicesMultiLevelDependency(t *testing.T) {
	// Test single-level dynamic dependency (the current implementation doesn't support multi-level)
	question := Question{
		Type: &QuestionType{
			Dynamic: &DynamicType{
				DependencyQuestions: []string{"app"},
			},
		},
		Choices: map[string]interface{}{
			"deployment": []interface{}{"deploy-choice-1", "deploy-choice-2"},
			"job":        []interface{}{"job-choice-1"},
		},
	}

	answers := map[string]interface{}{
		"app": "deployment",
	}

	choices, err := question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get choices for dynamic dependency: %v", err)
	}

	expected := []string{"deploy-choice-1", "deploy-choice-2"}
	if len(choices) != len(expected) {
		t.Errorf("Expected %d choices, got %d", len(expected), len(choices))
	}

	for i, choice := range choices {
		if choice != expected[i] {
			t.Errorf("Expected choice %s, got %s", expected[i], choice)
		}
	}
}

func TestQuestionGetChoicesInvalidStructure(t *testing.T) {
	// Test invalid choices structure in dynamic resolution
	question := Question{
		Type: &QuestionType{
			Dynamic: &DynamicType{
				DependencyQuestions: []string{"app"},
			},
		},
		Choices: map[string]interface{}{
			"deployment": "invalid_structure", // Should be array or map
		},
	}

	answers := map[string]interface{}{
		"app": "deployment",
	}

	_, err := question.GetChoices(answers)
	if err == nil {
		t.Error("Expected error for invalid choices structure")
	}
}

func TestQuestionGetChoicesMultipleDependencies(t *testing.T) {
	// Test multiple environment selection affecting cluster choices
	question := Question{
		Type: &QuestionType{
			Dynamic: &DynamicType{
				DependencyQuestions: []string{"env"},
			},
		},
		Choices: map[string]interface{}{
			"dev":        []interface{}{"dev-cluster-1", "dev-cluster-2"},
			"staging":    []interface{}{"staging-cluster-1", "staging-cluster-2"},
			"production": []interface{}{"production-cluster-1"},
		},
	}

	// Test single environment selection
	answers := map[string]interface{}{
		"env": "dev",
	}

	choices, err := question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get choices: %v", err)
	}

	expected := []string{"dev-cluster-1", "dev-cluster-2"}
	if len(choices) != len(expected) {
		t.Errorf("Expected %d choices, got %d", len(expected), len(choices))
	}

	// Test multiple environment selection
	answers = map[string]interface{}{
		"env": []string{"dev", "staging"},
	}

	choices, err = question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get choices for multiple environments: %v", err)
	}

	// Should get combined choices from both environments (no duplicates)
	expectedMultiple := []string{"dev-cluster-1", "dev-cluster-2", "staging-cluster-1", "staging-cluster-2"}
	if len(choices) != len(expectedMultiple) {
		t.Errorf("Expected %d choices for multiple envs, got %d", len(expectedMultiple), len(choices))
	}

	// Verify all expected choices are present
	choiceMap := make(map[string]bool)
	for _, choice := range choices {
		choiceMap[choice] = true
	}
	for _, expected := range expectedMultiple {
		if !choiceMap[expected] {
			t.Errorf("Expected choice %s not found in results", expected)
		}
	}
}

func TestQuestionGetChoicesArbitraryKeyDependency(t *testing.T) {
	// Test arbitrary key dependency (not just env/cluster)
	question := Question{
		Type: &QuestionType{
			Dynamic: &DynamicType{
				DependencyQuestions: []string{"service-type"},
			},
		},
		Choices: map[string]interface{}{
			"web":      []interface{}{"nginx", "apache", "caddy"},
			"database": []interface{}{"mysql", "postgresql", "mongodb"},
			"cache":    []interface{}{"redis", "memcached"},
		},
	}

	// Test single service type
	answers := map[string]interface{}{
		"service-type": "web",
	}

	choices, err := question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get choices: %v", err)
	}

	expected := []string{"nginx", "apache", "caddy"}
	if len(choices) != len(expected) {
		t.Errorf("Expected %d choices, got %d", len(expected), len(choices))
	}

	// Test multiple service types
	answers = map[string]interface{}{
		"service-type": []string{"database", "cache"},
	}

	choices, err = question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get choices for multiple service types: %v", err)
	}

	expectedMultiple := []string{"mysql", "postgresql", "mongodb", "redis", "memcached"}
	if len(choices) != len(expectedMultiple) {
		t.Errorf("Expected %d choices for multiple service types, got %d", len(expectedMultiple), len(choices))
	}
}

func TestQuestionGetChoicesNestedDependency(t *testing.T) {
	// Test multi-level dependency: region -> env -> cluster
	question := Question{
		Type: &QuestionType{
			Dynamic: &DynamicType{
				DependencyQuestions: []string{"region", "env"},
			},
		},
		Choices: map[string]interface{}{
			"us-east": map[string]interface{}{
				"dev":  []interface{}{"us-east-dev-1", "us-east-dev-2"},
				"prod": []interface{}{"us-east-prod-1"},
			},
			"us-west": map[string]interface{}{
				"dev":  []interface{}{"us-west-dev-1"},
				"prod": []interface{}{"us-west-prod-1", "us-west-prod-2"},
			},
		},
	}

	answers := map[string]interface{}{
		"region": "us-east",
		"env":    "dev",
	}

	choices, err := question.GetChoices(answers)
	if err != nil {
		t.Fatalf("Failed to get choices: %v", err)
	}

	expected := []string{"us-east-dev-1", "us-east-dev-2"}
	if len(choices) != len(expected) {
		t.Errorf("Expected %d choices, got %d", len(expected), len(choices))
	}

	for i, choice := range choices {
		if choice != expected[i] {
			t.Errorf("Expected choice %s, got %s", expected[i], choice)
		}
	}
}

func TestLoadConfigInvalidYaml(t *testing.T) {
	// Test loading invalid YAML
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yg", "_templates")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	configFile := filepath.Join(configDir, ".yg-config.yaml")
	invalidYaml := `questions:
  app:
    prompt: "Test"
    choices: [
      - invalid yaml structure
    invalid_indentation
`

	err = os.WriteFile(configFile, []byte(invalidYaml), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid config file: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	_, err = LoadConfig("")
	if err == nil {
		t.Error("Expected error when loading invalid YAML")
	}
}

func TestLoadConfigWithTemplateQuestion(t *testing.T) {
	// Test loading config with template_question specification
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yg")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
	configContent := `questions:
  template_question: "app"
  definitions:
    app:
      prompt: "アプリの種類はなんですか？"
      choices:
        - deployment
        - job
        - microservice
    appName:
      prompt: "アプリ名は何ですか？"
      choices:
        - sample-app-1
        - sample-app-2
    env:
      prompt: "環境名はなんですか？"
      type:
        multiple: true
      choices:
        - dev
        - staging
        - production
  order:
    - app
    - appName
    - env
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

	// Test template_question setting
	templateQuestion := config.Questions.GetTemplateQuestion()
	if templateQuestion != "app" {
		t.Errorf("Expected template_question to be 'app', got '%s'", templateQuestion)
	}

	// Test that other functionality still works
	questions := config.Questions.GetQuestions()
	if len(questions) != 3 {
		t.Errorf("Expected 3 questions, got %d", len(questions))
	}

	order := config.Questions.GetOrder()
	expectedOrder := []string{"app", "appName", "env"}
	if len(order) != len(expectedOrder) {
		t.Errorf("Expected order length %d, got %d", len(expectedOrder), len(order))
	}
}

func TestLoadConfigWithoutTemplateQuestion(t *testing.T) {
	// Test backward compatibility - configs without template_question should work
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".yg")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")
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

	// Template question should be empty (using heuristics)
	templateQuestion := config.Questions.GetTemplateQuestion()
	if templateQuestion != "" {
		t.Errorf("Expected template_question to be empty, got '%s'", templateQuestion)
	}

	// Other functionality should still work
	questions := config.Questions.GetQuestions()
	if len(questions) != 2 {
		t.Errorf("Expected 2 questions, got %d", len(questions))
	}
}

func TestGetTemplateQuestion(t *testing.T) {
	// Test GetTemplateQuestion method directly
	questions := &Questions{
		TemplateQuestion: "custom-template-key",
	}

	result := questions.GetTemplateQuestion()
	if result != "custom-template-key" {
		t.Errorf("Expected 'custom-template-key', got '%s'", result)
	}

	// Test empty template question
	questions = &Questions{}
	result = questions.GetTemplateQuestion()
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}
