package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTemplate(t *testing.T) {
	// Create temporary template directory and file
	tempDir := t.TempDir()
	templateDir := filepath.Join(tempDir, ".yg", "_templates")
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp template directory: %v", err)
	}

	templateFile := filepath.Join(templateDir, "deployment.yaml")
	templateContent := `path: {{.Questions.env}}/{{.Questions.cluster}}/deployment
filename: {{.Questions.appName}}-deployment.yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Questions.appName}}
spec:
  replicas: 3`

	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp template file: %v", err)
	}

	// Change working directory to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Test loading template
	tmpl, err := LoadTemplate("deployment")
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	expectedPath := "{{.Questions.env}}/{{.Questions.cluster}}/deployment"
	if tmpl.Path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, tmpl.Path)
	}

	expectedFilename := "{{.Questions.appName}}-deployment.yaml"
	if tmpl.Filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, tmpl.Filename)
	}

	if !strings.Contains(tmpl.Content, "kind: Deployment") {
		t.Error("Template content should contain 'kind: Deployment'")
	}
}

func TestLoadTemplateFileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	_, err := LoadTemplate("nonexistent")
	if err == nil {
		t.Error("Expected error when template file doesn't exist")
	}
}

func TestLoadTemplateInvalidFormat(t *testing.T) {
	// Create temporary template directory and file with invalid format
	tempDir := t.TempDir()
	templateDir := filepath.Join(tempDir, ".yg", "_templates")
	err := os.MkdirAll(templateDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp template directory: %v", err)
	}

	templateFile := filepath.Join(templateDir, "invalid.yaml")
	templateContent := `path: {{.Questions.env}}
filename: {{.Questions.appName}}.yaml
content without separator`

	err = os.WriteFile(templateFile, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp template file: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	_, err = LoadTemplate("invalid")
	if err == nil {
		t.Error("Expected error for invalid template format")
	}
}

func TestTemplateRender(t *testing.T) {
	tmpl := &Template{
		Path:     "{{.Questions.env}}/{{.Questions.cluster}}/deployment",
		Filename: "{{.Questions.appName}}-deployment.yaml",
		Content: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}
spec:
  replicas: 3`,
	}

	data := &TemplateData{
		Questions: map[string]interface{}{
			"appName": "test-app",
			"env":     "dev",
			"cluster": "dev-cluster-1",
		},
	}

	path, filename, content, err := tmpl.Render(data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	expectedPath := "dev/dev-cluster-1/deployment"
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}

	expectedFilename := "test-app-deployment.yaml"
	if filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, filename)
	}

	if !strings.Contains(content, "name: test-app") {
		t.Error("Rendered content should contain 'name: test-app'")
	}

	if !strings.Contains(content, "namespace: dev") {
		t.Error("Rendered content should contain 'namespace: dev'")
	}
}

func TestTemplateRenderInvalidTemplate(t *testing.T) {
	tmpl := &Template{
		Path:     "{{.InvalidField}}",
		Filename: "test.yaml",
		Content:  "content",
	}

	data := &TemplateData{
		Questions: map[string]interface{}{
			"appName": "test-app",
		},
	}

	_, _, _, err := tmpl.Render(data)
	if err == nil {
		t.Error("Expected error for invalid template field")
	}
}