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

	err = os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write temp template file: %v", err)
	}

	// Change working directory to temp directory
	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

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
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

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

	err = os.WriteFile(templateFile, []byte(templateContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write temp template file: %v", err)
	}

	originalWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalWd) }()
	_ = os.Chdir(tempDir)

	_, err = LoadTemplate("invalid")
	if err == nil {
		t.Error("Expected error for invalid template format")
	}
}

func TestTemplateRender(t *testing.T) {
	tmpl := &Template{
		Type:     TemplateTypeFile,
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

	data := &Data{
		Questions: map[string]interface{}{
			"appName": "test-app",
			"env":     "dev",
			"cluster": "dev-cluster-1",
		},
	}

	result, err := tmpl.Render(data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(result.Files))
	}

	file := result.Files[0]
	expectedPath := "dev/dev-cluster-1/deployment"
	if file.Path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, file.Path)
	}

	expectedFilename := "test-app-deployment.yaml"
	if file.Filename != expectedFilename {
		t.Errorf("Expected filename %s, got %s", expectedFilename, file.Filename)
	}

	if !strings.Contains(file.Content, "name: test-app") {
		t.Error("Rendered content should contain 'name: test-app'")
	}

	if !strings.Contains(file.Content, "namespace: dev") {
		t.Error("Rendered content should contain 'namespace: dev'")
	}
}

func TestTemplateRenderInvalidTemplate(t *testing.T) {
	tmpl := &Template{
		Type:     TemplateTypeFile,
		Path:     "{{.InvalidField}}",
		Filename: "test.yaml",
		Content:  "content",
	}

	data := &Data{
		Questions: map[string]interface{}{
			"appName": "test-app",
		},
	}

	_, err := tmpl.Render(data)
	if err == nil {
		t.Error("Expected error for invalid template field")
	}
}

func TestLoadDirectoryTemplate(t *testing.T) {
	// Create test directory
	testDir := t.TempDir()
	templateDir := filepath.Join(testDir, ".yg", "_templates", "microservice")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Change to test directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(testDir)

	// Create directory template config
	configContent := `output:
  base_path: "{{.Questions.env}}/{{.Questions.cluster}}/{{.Questions.appName}}"
files:
  deployment.yaml:
    filename: "{{.Questions.appName}}-deployment.yaml"
  service.yaml:
    filename: "{{.Questions.appName}}-service.yaml"
    enabled: "{{if eq .Questions.needsService \"yes\"}}true{{else}}false{{end}}"`

	configPath := filepath.Join(templateDir, ".template-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write template config: %v", err)
	}

	// Create template files
	deploymentContent := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Questions.appName}}`

	deploymentPath := filepath.Join(templateDir, "deployment.yaml")
	if err := os.WriteFile(deploymentPath, []byte(deploymentContent), 0o600); err != nil {
		t.Fatalf("Failed to write deployment template: %v", err)
	}

	serviceContent := `apiVersion: v1
kind: Service
metadata:
  name: {{.Questions.appName}}`

	servicePath := filepath.Join(templateDir, "service.yaml")
	if err := os.WriteFile(servicePath, []byte(serviceContent), 0o600); err != nil {
		t.Fatalf("Failed to write service template: %v", err)
	}

	// Test loading directory template
	tmpl, err := loadDirectoryTemplate("microservice")
	if err != nil {
		t.Fatalf("Failed to load directory template: %v", err)
	}

	if tmpl.Type != TemplateTypeDirectory {
		t.Errorf("Expected template type to be directory, got %s", tmpl.Type)
	}

	expectedBasePath := "{{.Questions.env}}/{{.Questions.cluster}}/{{.Questions.appName}}"
	if tmpl.BasePath != expectedBasePath {
		t.Errorf("Expected base path %s, got %s", expectedBasePath, tmpl.BasePath)
	}

	if len(tmpl.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(tmpl.Files))
	}
}

func TestRenderDirectory(t *testing.T) {
	tmpl := &Template{
		Type:     TemplateTypeDirectory,
		BasePath: "{{.Questions.env}}/{{.Questions.cluster}}/{{.Questions.appName}}",
		Files: map[string]*FileTemplate{
			"deployment.yaml": {
				Filename: "{{.Questions.appName}}-deployment.yaml",
				Content:  "appName: {{.Questions.appName}}",
			},
			"service.yaml": {
				Filename: "{{.Questions.appName}}-service.yaml",
				Content:  "serviceName: {{.Questions.appName}}",
				Enabled:  "{{if eq .Questions.needsService \"yes\"}}true{{else}}false{{end}}",
			},
		},
	}

	t.Run("with service enabled", func(t *testing.T) {
		data := &Data{
			Questions: map[string]interface{}{
				"appName":      "test-app",
				"env":          "dev",
				"cluster":      "dev-cluster-1",
				"needsService": "yes",
			},
		}

		result, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		if len(result.Files) != 2 {
			t.Fatalf("Expected 2 files, got %d", len(result.Files))
		}

		// Check that both files have correct base path
		expectedPath := "dev/dev-cluster-1/test-app"
		for i, file := range result.Files {
			if file.Path != expectedPath {
				t.Errorf("File %d: Expected path %s, got %s", i, expectedPath, file.Path)
			}
		}
	})

	t.Run("with service disabled", func(t *testing.T) {
		data := &Data{
			Questions: map[string]interface{}{
				"appName":      "test-app",
				"env":          "dev",
				"cluster":      "dev-cluster-1",
				"needsService": "no",
			},
		}

		result, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		if len(result.Files) != 1 {
			t.Fatalf("Expected 1 file (service disabled), got %d", len(result.Files))
		}

		file := result.Files[0]
		if !strings.Contains(file.Filename, "deployment") {
			t.Errorf("Expected deployment file to remain, got %s", file.Filename)
		}
	})
}
