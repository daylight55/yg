// Package template provides template processing functionality.
package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// TemplateType defines the type of template.
type TemplateType string

const (
	TemplateTypeFile      TemplateType = "file"
	TemplateTypeDirectory TemplateType = "directory"
)

// Template represents a YAML template.
type Template struct {
	Type     TemplateType // "file" or "directory"
	Path     string       // For file: template file path, For directory: base path template
	Filename string       // For file: filename template
	Content  string       // For file: content template

	// For directory templates
	Files    map[string]*FileTemplate // filename -> FileTemplate
	BasePath string                   // base path template for all files
}

// FileTemplate represents a single file within a directory template.
type FileTemplate struct {
	Filename string // filename template
	Content  string // content template
	Enabled  string // condition template (optional)
}

// DirectoryTemplateConfig represents the config for directory templates.
type DirectoryTemplateConfig struct {
	Output OutputConfig                    `yaml:"output"`
	Files  map[string]FileTemplateConfig `yaml:"files"`
}

// OutputConfig represents output configuration for directory templates.
type OutputConfig struct {
	BasePath string `yaml:"base_path"`
}

// FileTemplateConfig represents configuration for individual files.
type FileTemplateConfig struct {
	Filename string `yaml:"filename"`
	Enabled  string `yaml:"enabled,omitempty"`
}

// Data holds the data for template rendering.
type Data struct {
	Questions map[string]interface{}
}

// LoadTemplate loads either a single file or directory template.
func LoadTemplate(templateType string) (*Template, error) {
	// まず設定から template type を確認
	config, err := loadTemplateConfig()
	if err != nil {
		// 設定ファイルが存在しない場合は従来の単一ファイル読み込み
		return loadFileTemplate(templateType)
	}

	templateConfig, exists := config.Templates[templateType]
	if !exists {
		// フォールバック: 従来の単一ファイル読み込み
		return loadFileTemplate(templateType)
	}

	switch templateConfig.Type {
	case "file":
		return loadFileTemplate(templateConfig.Path)
	case "directory":
		return loadDirectoryTemplate(templateConfig.Path)
	default:
		return nil, fmt.Errorf("unsupported template type: %s", templateConfig.Type)
	}
}

// loadTemplateConfig loads the template configuration from config file.
func loadTemplateConfig() (*TemplateConfigFile, error) {
	configPath := filepath.Join(".yg", "_templates", ".yg-config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config TemplateConfigFile
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// TemplateConfigFile represents a simplified config structure for templates only.
type TemplateConfigFile struct {
	Templates map[string]TemplateConfigEntry `yaml:"templates"`
}

// TemplateConfigEntry represents template configuration entry.
type TemplateConfigEntry struct {
	Type string `yaml:"type"` // "file" or "directory"
	Path string `yaml:"path"` // path to template file or directory
}

// loadFileTemplate loads a single file template.
func loadFileTemplate(templatePath string) (*Template, error) {
	// If templatePath doesn't end with .yaml, add it (backward compatibility)
	if !strings.HasSuffix(templatePath, ".yaml") {
		templatePath = templatePath + ".yaml"
	}

	fullPath := filepath.Join(".yg", "_templates", templatePath)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", fullPath, err)
	}

	content := string(data)

	// Split the content into metadata and template content
	parts := strings.SplitN(content, "---", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid template format: missing --- separator")
	}

	metaContent := strings.TrimSpace(parts[0])
	templateContent := strings.TrimSpace(parts[1])

	// Parse metadata
	tmpl := &Template{
		Type:    TemplateTypeFile,
		Content: templateContent,
	}

	// Extract path and filename from metadata
	lines := strings.Split(metaContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "path:") {
			tmpl.Path = strings.TrimSpace(strings.TrimPrefix(line, "path:"))
		} else if strings.HasPrefix(line, "filename:") {
			tmpl.Filename = strings.TrimSpace(strings.TrimPrefix(line, "filename:"))
		}
	}

	return tmpl, nil
}

// loadDirectoryTemplate loads a directory template.
func loadDirectoryTemplate(dirName string) (*Template, error) {
	templateDir := filepath.Join(".yg", "_templates", dirName)

	// .template-config.yaml を読み込み
	configPath := filepath.Join(templateDir, ".template-config.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template config %s: %w", configPath, err)
	}

	var config DirectoryTemplateConfig
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse template config: %w", err)
	}

	// ディレクトリ内のテンプレートファイルを読み込み
	files := make(map[string]*FileTemplate)
	for filename, fileConfig := range config.Files {
		contentPath := filepath.Join(templateDir, filename)
		content, err := os.ReadFile(contentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read template file %s: %w", filename, err)
		}

		files[filename] = &FileTemplate{
			Filename: fileConfig.Filename,
			Content:  string(content),
			Enabled:  fileConfig.Enabled,
		}
	}

	return &Template{
		Type:     TemplateTypeDirectory,
		BasePath: config.Output.BasePath,
		Files:    files,
	}, nil
}

// RenderResult holds the result of template rendering.
type RenderResult struct {
	Files []RenderedFile
}

// RenderedFile represents a single rendered file.
type RenderedFile struct {
	Path     string
	Filename string
	Content  string
}

// Render renders the template and returns all generated files.
func (t *Template) Render(data *Data) (*RenderResult, error) {
	switch t.Type {
	case TemplateTypeFile:
		return t.renderSingleFile(data)
	case TemplateTypeDirectory:
		return t.renderDirectory(data)
	default:
		return nil, fmt.Errorf("unsupported template type: %s", t.Type)
	}
}

// renderSingleFile renders a single file template (backward compatibility).
func (t *Template) renderSingleFile(data *Data) (*RenderResult, error) {
	funcMap := template.FuncMap{
		"questions": func() map[string]interface{} {
			return data.Questions
		},
	}

	// Render path
	pathTmpl, err := template.New("path").Funcs(funcMap).Parse(t.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path template: %w", err)
	}

	var pathBuf strings.Builder
	if err := pathTmpl.Execute(&pathBuf, data); err != nil {
		return nil, fmt.Errorf("failed to render path: %w", err)
	}
	renderedPath := pathBuf.String()

	// Render filename
	filenameTmpl, err := template.New("filename").Funcs(funcMap).Parse(t.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filename template: %w", err)
	}

	var filenameBuf strings.Builder
	if err := filenameTmpl.Execute(&filenameBuf, data); err != nil {
		return nil, fmt.Errorf("failed to render filename: %w", err)
	}
	renderedFilename := filenameBuf.String()

	// Render content
	contentTmpl, err := template.New("content").Funcs(funcMap).Parse(t.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content template: %w", err)
	}

	var contentBuf strings.Builder
	if err := contentTmpl.Execute(&contentBuf, data); err != nil {
		return nil, fmt.Errorf("failed to render content: %w", err)
	}
	renderedContent := contentBuf.String()

	return &RenderResult{
		Files: []RenderedFile{
			{
				Path:     renderedPath,
				Filename: renderedFilename,
				Content:  renderedContent,
			},
		},
	}, nil
}

// renderDirectory renders a directory template.
func (t *Template) renderDirectory(data *Data) (*RenderResult, error) {
	result := &RenderResult{Files: []RenderedFile{}}

	// Base path をレンダリング
	basePath, err := renderTemplate("base_path", t.BasePath, data)
	if err != nil {
		return nil, fmt.Errorf("failed to render base path: %w", err)
	}

	// 各ファイルをレンダリング
	for originalName, fileTemplate := range t.Files {
		// 有効性をチェック
		if fileTemplate.Enabled != "" {
			enabled, err := renderTemplate("enabled", fileTemplate.Enabled, data)
			if err != nil {
				return nil, fmt.Errorf("failed to render enabled condition for %s: %w", originalName, err)
			}
			if enabled != "true" {
				continue // スキップ
			}
		}

		// ファイル名をレンダリング
		filename, err := renderTemplate("filename", fileTemplate.Filename, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render filename for %s: %w", originalName, err)
		}

		// コンテンツをレンダリング
		content, err := renderTemplate("content", fileTemplate.Content, data)
		if err != nil {
			return nil, fmt.Errorf("failed to render content for %s: %w", originalName, err)
		}

		result.Files = append(result.Files, RenderedFile{
			Path:     basePath,
			Filename: filename,
			Content:  content,
		})
	}

	return result, nil
}

// renderTemplate renders a template string with the given data.
func renderTemplate(name, templateStr string, data *Data) (string, error) {
	funcMap := template.FuncMap{
		"questions": func() map[string]interface{} {
			return data.Questions
		},
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
