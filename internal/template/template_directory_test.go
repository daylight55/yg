package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadTemplateWithConfig tests template loading with config file
func TestLoadTemplateWithConfig(t *testing.T) {
	testDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	// Create config directory
	configDir := filepath.Join(testDir, ".yg")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create template directory
	templateDir := filepath.Join(testDir, ".yg", "_templates", "microservice")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	// Create config file
	configContent := `templates:
  microservice:
    type: directory
    path: microservice
  deployment:
    type: file
    path: deployment.yaml`

	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create directory template config
	dirConfigContent := `output:
  base_path: "{{.Questions.env}}/{{.Questions.cluster}}/{{.Questions.appName}}"
files:
  deployment.yaml:
    filename: "{{.Questions.appName}}-deployment.yaml"
  service.yaml:
    filename: "{{.Questions.appName}}-service.yaml"
  configmap.yaml:
    filename: "{{.Questions.appName}}-configmap.yaml"
  ingress.yaml:
    filename: "{{.Questions.appName}}-ingress.yaml"`

	dirConfigPath := filepath.Join(templateDir, ".template-config.yaml")
	if err := os.WriteFile(dirConfigPath, []byte(dirConfigContent), 0o600); err != nil {
		t.Fatalf("Failed to write directory template config: %v", err)
	}

	// Create template files
	templates := map[string]string{
		"deployment.yaml": `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}
  labels:
    app: {{.Questions.appName}}
    env: {{.Questions.env}}
    cluster: {{.Questions.cluster}}
spec:
  replicas: 3
  selector:
    matchLabels:
      app: {{.Questions.appName}}
  template:
    metadata:
      labels:
        app: {{.Questions.appName}}
    spec:
      containers:
      - name: {{.Questions.appName}}
        image: {{.Questions.appName}}:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: {{.Questions.env}}
        - name: CLUSTER
          value: {{.Questions.cluster}}`,
		"service.yaml": `apiVersion: v1
kind: Service
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}
  labels:
    app: {{.Questions.appName}}
    env: {{.Questions.env}}
spec:
  selector:
    app: {{.Questions.appName}}
  ports:
  - name: http
    port: 80
    targetPort: 8080
    protocol: TCP
  type: ClusterIP`,
		"configmap.yaml": `apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.Questions.appName}}-config
  namespace: {{.Questions.env}}
  labels:
    app: {{.Questions.appName}}
    env: {{.Questions.env}}
data:
  app.properties: |
    app.name={{.Questions.appName}}
    app.env={{.Questions.env}}
    app.cluster={{.Questions.cluster}}
    log.level=INFO`,
		"ingress.yaml": `apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{.Questions.appName}}
  namespace: {{.Questions.env}}
  labels:
    app: {{.Questions.appName}}
    env: {{.Questions.env}}
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: {{.Questions.appName}}.{{.Questions.env}}.{{.Questions.cluster}}.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: {{.Questions.appName}}
            port:
              number: 80`,
	}

	for filename, content := range templates {
		templatePath := filepath.Join(templateDir, filename)
		if err := os.WriteFile(templatePath, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to write template file %s: %v", filename, err)
		}
	}

	// Test loading microservice template
	tmpl, err := LoadTemplate("microservice")
	if err != nil {
		t.Fatalf("Failed to load microservice template: %v", err)
	}

	if tmpl.Type != TypeDirectory {
		t.Errorf("Expected template type to be directory, got %s", tmpl.Type)
	}

	if len(tmpl.Files) != 4 {
		t.Errorf("Expected 4 files in directory template, got %d", len(tmpl.Files))
	}

	expectedFiles := []string{"deployment.yaml", "service.yaml", "configmap.yaml", "ingress.yaml"}
	for _, expectedFile := range expectedFiles {
		if _, exists := tmpl.Files[expectedFile]; !exists {
			t.Errorf("Expected %s to exist in template files", expectedFile)
		}
	}

	// Test rendering the template
	data := &Data{
		Questions: map[string]interface{}{
			"appName": "sample-api-1",
			"env":     "dev",
			"cluster": "dev-cluster-1",
		},
	}

	result, err := tmpl.Render(data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if len(result.Files) != 4 {
		t.Fatalf("Expected 4 rendered files, got %d", len(result.Files))
	}

	expectedPath := "dev/dev-cluster-1/sample-api-1"
	for i, file := range result.Files {
		if file.Path != expectedPath {
			t.Errorf("File %d: Expected path %s, got %s", i, expectedPath, file.Path)
		}

		if !strings.Contains(file.Filename, "sample-api-1") {
			t.Errorf("File %d: Expected filename to contain 'sample-api-1', got %s", i, file.Filename)
		}

		if !strings.Contains(file.Content, "sample-api-1") {
			t.Errorf("File %d: Expected content to contain 'sample-api-1'", i)
		}

		if !strings.Contains(file.Content, "namespace: dev") {
			t.Errorf("File %d: Expected content to contain 'namespace: dev'", i)
		}
	}
}

// TestLoadTemplateError tests error scenarios for directory templates
func TestLoadTemplateError(t *testing.T) {
	testDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	t.Run("invalid template type in config", func(t *testing.T) {
		// Create config directory
		configDir := filepath.Join(testDir, ".yg")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create config file with invalid type
		configContent := `templates:
  invalid:
    type: unknown
    path: invalid`

		configPath := filepath.Join(configDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		_, err := LoadTemplate("invalid")
		if err == nil {
			t.Error("Expected error for invalid template type")
		}
		if !strings.Contains(err.Error(), "unsupported template type") {
			t.Errorf("Expected 'unsupported template type' error, got: %v", err)
		}
	})

	t.Run("missing directory template config", func(t *testing.T) {
		// Create config directory
		configDir := filepath.Join(testDir, ".yg")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create template directory without config
		templateDir := filepath.Join(testDir, ".yg", "_templates", "missing")
		if err := os.MkdirAll(templateDir, 0o755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		// Create config file
		configContent := `templates:
  missing:
    type: directory
    path: missing`

		configPath := filepath.Join(configDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		_, err := LoadTemplate("missing")
		if err == nil {
			t.Error("Expected error for missing directory template config")
		}
		if !strings.Contains(err.Error(), "failed to read template config") {
			t.Errorf("Expected 'failed to read template config' error, got: %v", err)
		}
	})

	t.Run("missing template file in directory", func(t *testing.T) {
		// Create config directory
		configDir := filepath.Join(testDir, ".yg")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatalf("Failed to create config directory: %v", err)
		}

		// Create template directory
		templateDir := filepath.Join(testDir, ".yg", "_templates", "incomplete")
		if err := os.MkdirAll(templateDir, 0o755); err != nil {
			t.Fatalf("Failed to create template directory: %v", err)
		}

		// Create config file
		configContent := `templates:
  incomplete:
    type: directory
    path: incomplete`

		configPath := filepath.Join(configDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Create directory template config that references missing file
		dirConfigContent := `output:
  base_path: "{{.Questions.env}}/{{.Questions.cluster}}/{{.Questions.appName}}"
files:
  missing.yaml:
    filename: "{{.Questions.appName}}-missing.yaml"`

		dirConfigPath := filepath.Join(templateDir, ".template-config.yaml")
		if err := os.WriteFile(dirConfigPath, []byte(dirConfigContent), 0o600); err != nil {
			t.Fatalf("Failed to write directory template config: %v", err)
		}

		_, err := LoadTemplate("incomplete")
		if err == nil {
			t.Error("Expected error for missing template file")
		}
		if !strings.Contains(err.Error(), "failed to read template file") {
			t.Errorf("Expected 'failed to read template file' error, got: %v", err)
		}
	})
}

// TestDirectoryTemplateWithConditionals tests directory templates with conditional files
func TestDirectoryTemplateWithConditionals(t *testing.T) {
	testDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()
	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	// Create config directory
	configDir := filepath.Join(testDir, ".yg")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create template directory
	templateDir := filepath.Join(testDir, ".yg", "_templates", "conditional")
	if err := os.MkdirAll(templateDir, 0o755); err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	// Create config file
	configContent := `templates:
  conditional:
    type: directory
    path: conditional`

	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create directory template config with conditional files
	dirConfigContent := `output:
  base_path: "{{.Questions.env}}/{{.Questions.cluster}}/{{.Questions.appName}}"
files:
  deployment.yaml:
    filename: "{{.Questions.appName}}-deployment.yaml"
  service.yaml:
    filename: "{{.Questions.appName}}-service.yaml"
    enabled: "{{if eq .Questions.needsService \"yes\"}}true{{else}}false{{end}}"
  ingress.yaml:
    filename: "{{.Questions.appName}}-ingress.yaml"
    enabled: "{{if eq .Questions.needsIngress \"yes\"}}true{{else}}false{{end}}"`

	dirConfigPath := filepath.Join(templateDir, ".template-config.yaml")
	if err := os.WriteFile(dirConfigPath, []byte(dirConfigContent), 0o600); err != nil {
		t.Fatalf("Failed to write directory template config: %v", err)
	}

	// Create template files
	templates := map[string]string{
		"deployment.yaml": "name: {{.Questions.appName}}",
		"service.yaml":    "serviceName: {{.Questions.appName}}",
		"ingress.yaml":    "ingressName: {{.Questions.appName}}",
	}

	for filename, content := range templates {
		templatePath := filepath.Join(templateDir, filename)
		if err := os.WriteFile(templatePath, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to write template file %s: %v", filename, err)
		}
	}

	// Test loading template
	tmpl, err := LoadTemplate("conditional")
	if err != nil {
		t.Fatalf("Failed to load conditional template: %v", err)
	}

	t.Run("all services enabled", func(t *testing.T) {
		data := &Data{
			Questions: map[string]interface{}{
				"appName":      "test-app",
				"env":          "dev",
				"cluster":      "dev-cluster-1",
				"needsService": "yes",
				"needsIngress": "yes",
			},
		}

		result, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		if len(result.Files) != 3 {
			t.Fatalf("Expected 3 files (all enabled), got %d", len(result.Files))
		}
	})

	t.Run("only deployment enabled", func(t *testing.T) {
		data := &Data{
			Questions: map[string]interface{}{
				"appName":      "test-app",
				"env":          "dev",
				"cluster":      "dev-cluster-1",
				"needsService": "no",
				"needsIngress": "no",
			},
		}

		result, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		if len(result.Files) != 1 {
			t.Fatalf("Expected 1 file (only deployment), got %d", len(result.Files))
		}

		file := result.Files[0]
		if !strings.Contains(file.Filename, "deployment") {
			t.Errorf("Expected deployment file, got %s", file.Filename)
		}
	})

	t.Run("partial services enabled", func(t *testing.T) {
		data := &Data{
			Questions: map[string]interface{}{
				"appName":      "test-app",
				"env":          "dev",
				"cluster":      "dev-cluster-1",
				"needsService": "yes",
				"needsIngress": "no",
			},
		}

		result, err := tmpl.Render(data)
		if err != nil {
			t.Fatalf("Failed to render template: %v", err)
		}

		if len(result.Files) != 2 {
			t.Fatalf("Expected 2 files (deployment + service), got %d", len(result.Files))
		}

		fileTypes := make(map[string]bool)
		for _, file := range result.Files {
			if strings.Contains(file.Filename, "deployment") {
				fileTypes["deployment"] = true
			} else if strings.Contains(file.Filename, "service") {
				fileTypes["service"] = true
			} else if strings.Contains(file.Filename, "ingress") {
				fileTypes["ingress"] = true
			}
		}

		if !fileTypes["deployment"] || !fileTypes["service"] || fileTypes["ingress"] {
			t.Error("Expected deployment and service files, but not ingress")
		}
	})
}
