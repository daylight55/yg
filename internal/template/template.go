// Package template provides template processing functionality.
package template

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Template represents a YAML template.
type Template struct {
	Path     string
	Filename string
	Content  string
}

// TemplateData holds the data for template rendering.
type TemplateData struct {
	Questions map[string]interface{}
}

// LoadTemplate loads a template file.
func LoadTemplate(templateType string) (*Template, error) {
	templatePath := filepath.Join(".yg", "_templates", templateType+".yaml")
	
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file %s: %w", templatePath, err)
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
	tmpl := &Template{Content: templateContent}
	
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

// Render renders the template with the given data.
func (t *Template) Render(data *TemplateData) (string, string, string, error) {
	funcMap := template.FuncMap{
		"questions": func() map[string]interface{} {
			return data.Questions
		},
	}

	// Render path
	pathTmpl, err := template.New("path").Funcs(funcMap).Parse(t.Path)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse path template: %w", err)
	}
	
	var pathBuf strings.Builder
	if err := pathTmpl.Execute(&pathBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to render path: %w", err)
	}
	renderedPath := pathBuf.String()

	// Render filename
	filenameTmpl, err := template.New("filename").Funcs(funcMap).Parse(t.Filename)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse filename template: %w", err)
	}
	
	var filenameBuf strings.Builder
	if err := filenameTmpl.Execute(&filenameBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to render filename: %w", err)
	}
	renderedFilename := filenameBuf.String()

	// Render content
	contentTmpl, err := template.New("content").Funcs(funcMap).Parse(t.Content)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to parse content template: %w", err)
	}
	
	var contentBuf strings.Builder
	if err := contentTmpl.Execute(&contentBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to render content: %w", err)
	}
	renderedContent := contentBuf.String()

	return renderedPath, renderedFilename, renderedContent, nil
}