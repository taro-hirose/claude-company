package prompts

import (
	"bytes"
	"fmt"
	"text/template"
)

// Template represents a reusable prompt template
type Template struct {
	Name     string
	Template *template.Template
}

// TemplateData represents common data structure for templates
type TemplateData struct {
	TaskID      string
	TaskDesc    string
	Context     string
	StepNumber  int
	StepDesc    string
	PreviousResult string
	AdditionalData map[string]interface{}
}

// TemplateManager manages prompt templates
type TemplateManager struct {
	templates map[string]*Template
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*Template),
	}
}

// RegisterTemplate registers a new template
func (tm *TemplateManager) RegisterTemplate(name string, templateStr string) error {
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	
	tm.templates[name] = &Template{
		Name:     name,
		Template: tmpl,
	}
	return nil
}

// ExecuteTemplate executes a template with given data
func (tm *TemplateManager) ExecuteTemplate(name string, data interface{}) (string, error) {
	tmpl, exists := tm.templates[name]
	if !exists {
		return "", fmt.Errorf("template %s not found", name)
	}
	
	var buf bytes.Buffer
	if err := tmpl.Template.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}
	
	return buf.String(), nil
}

// GetTemplate returns a template by name
func (tm *TemplateManager) GetTemplate(name string) (*Template, error) {
	tmpl, exists := tm.templates[name]
	if !exists {
		return nil, fmt.Errorf("template %s not found", name)
	}
	return tmpl, nil
}

// Common template helper functions
var templateFuncs = template.FuncMap{
	"indent": func(spaces int, text string) string {
		indentation := ""
		for i := 0; i < spaces; i++ {
			indentation += " "
		}
		
		lines := bytes.Split([]byte(text), []byte("\n"))
		var result []byte
		for i, line := range lines {
			if i > 0 {
				result = append(result, '\n')
			}
			if len(line) > 0 {
				result = append(result, []byte(indentation)...)
			}
			result = append(result, line...)
		}
		return string(result)
	},
	"truncate": func(maxLength int, text string) string {
		if len(text) <= maxLength {
			return text
		}
		return text[:maxLength-3] + "..."
	},
}

// NewTemplateWithFuncs creates a new template with helper functions
func NewTemplateWithFuncs(name string) *template.Template {
	return template.New(name).Funcs(templateFuncs)
}