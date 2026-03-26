package store

import (
	"fmt"
	"os"
)

// TemplateRecord represents a PCDP template with metadata and content
type TemplateRecord struct {
	Name     string `json:"name"`
	Version  string `json:"version"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

// ResourceRecord represents a PCDP resource (template, prompt, or hints)
type ResourceRecord struct {
	URI     string `json:"uri"`
	Name    string `json:"name"`
	Content string `json:"content,omitempty"`
}

// Diagnostic represents a linting diagnostic with location and severity
type Diagnostic struct {
	Severity string `json:"severity"`
	Line     int    `json:"line"`
	Section  string `json:"section"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
}

// LintResult represents the result of linting a specification
type LintResult struct {
	Valid       bool          `json:"valid"`
	Errors      int           `json:"errors"`
	Warnings    int           `json:"warnings"`
	Diagnostics []Diagnostic  `json:"diagnostics"`
}

// ============ Filesystem Interface ============

// Filesystem interface for reading files
type Filesystem interface {
	ReadFile(path string) (string, error)
}

// OSFilesystem is the production implementation using os package
type OSFilesystem struct{}

func NewOSFilesystem() Filesystem {
	return &OSFilesystem{}
}

func (fs *OSFilesystem) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FakeFilesystem is a test double for Filesystem
type FakeFilesystem struct {
	Files   map[string]string
	ReadErr map[string]error
}

func NewFakeFilesystem() *FakeFilesystem {
	return &FakeFilesystem{
		Files:   make(map[string]string),
		ReadErr: make(map[string]error),
	}
}

func (fs *FakeFilesystem) ReadFile(path string) (string, error) {
	if err, ok := fs.ReadErr[path]; ok {
		return "", err
	}
	if content, ok := fs.Files[path]; ok {
		return content, nil
	}
	return "", os.ErrNotExist
}

// ============ TemplateStore Interface ============

// TemplateStore interface for accessing templates and hints
type TemplateStore interface {
	ListTemplates() ([]TemplateRecord, error)
	GetTemplate(name, version string) (TemplateRecord, error)
	GetHints(key string) (string, error)
	ListHints() []string
}

// LayeredTemplateStore is the production implementation
// that reads from /usr/share/pcdp/templates/ with layered overrides
type LayeredTemplateStore struct {
	templates map[string]map[string]TemplateRecord
	hints     map[string]string
}

func NewLayeredTemplateStore() TemplateStore {
	// In production, this would read from the filesystem hierarchy:
	// /usr/share/pcdp/templates/, /etc/pcdp/, ~/.config/pcdp/, ./.pcdp/
	// For now, return empty store (would be populated from disk)
	return &LayeredTemplateStore{
		templates: make(map[string]map[string]TemplateRecord),
		hints:     make(map[string]string),
	}
}

func (ts *LayeredTemplateStore) ListTemplates() ([]TemplateRecord, error) {
	var results []TemplateRecord
	for _, versions := range ts.templates {
		for _, rec := range versions {
			results = append(results, rec)
		}
	}
	return results, nil
}

func (ts *LayeredTemplateStore) GetTemplate(name, version string) (TemplateRecord, error) {
	if versions, ok := ts.templates[name]; ok {
		if version == "latest" {
			// Return highest version (simplified: just return first found)
			for _, rec := range versions {
				return rec, nil
			}
		}
		if rec, ok := versions[version]; ok {
			return rec, nil
		}
		return TemplateRecord{}, fmt.Errorf("version %s not found for template %s", version, name)
	}
	return TemplateRecord{}, fmt.Errorf("unknown template: %s", name)
}

func (ts *LayeredTemplateStore) GetHints(key string) (string, error) {
	if content, ok := ts.hints[key]; ok {
		return content, nil
	}
	return "", fmt.Errorf("hints not found: %s", key)
}

func (ts *LayeredTemplateStore) ListHints() []string {
	var keys []string
	for k := range ts.hints {
		keys = append(keys, k)
	}
	return keys
}

// FakeTemplateStore is a test double for TemplateStore
type FakeTemplateStore struct {
	Templates []TemplateRecord
	Hints     map[string]string
}

func NewFakeTemplateStore() *FakeTemplateStore {
	return &FakeTemplateStore{
		Templates: []TemplateRecord{},
		Hints:     make(map[string]string),
	}
}

func (ts *FakeTemplateStore) ListTemplates() ([]TemplateRecord, error) {
	return ts.Templates, nil
}

func (ts *FakeTemplateStore) GetTemplate(name, version string) (TemplateRecord, error) {
	for _, t := range ts.Templates {
		if t.Name == name {
			if version == "latest" || version == t.Version {
				return t, nil
			}
		}
	}
	return TemplateRecord{}, fmt.Errorf("unknown template: %s", name)
}

func (ts *FakeTemplateStore) GetHints(key string) (string, error) {
	if content, ok := ts.Hints[key]; ok {
		return content, nil
	}
	return "", fmt.Errorf("hints not found: %s", key)
}

func (ts *FakeTemplateStore) ListHints() []string {
	var keys []string
	for k := range ts.Hints {
		keys = append(keys, k)
	}
	return keys
}

// ============ PromptStore Interface ============

// PromptStore interface for accessing embedded prompts
type PromptStore interface {
	GetPrompt(name string) (string, error)
	ListPrompts() []string
}

// EmbeddedPromptStore is the production implementation
// with prompts embedded as Go string constants
type EmbeddedPromptStore struct {
	prompts map[string]string
}

func NewEmbeddedPromptStore() PromptStore {
	store := &EmbeddedPromptStore{
		prompts: make(map[string]string),
	}
	// Embed prompt content from constants
	store.prompts["interview"] = promptInterview
	store.prompts["translator"] = promptTranslator
	return store
}

func (ps *EmbeddedPromptStore) GetPrompt(name string) (string, error) {
	if content, ok := ps.prompts[name]; ok {
		return content, nil
	}
	return "", fmt.Errorf("prompt not found: %s", name)
}

func (ps *EmbeddedPromptStore) ListPrompts() []string {
	return []string{"interview", "translator"}
}

// FakePromptStore is a test double for PromptStore
type FakePromptStore struct {
	Prompts map[string]string
}

func NewFakePromptStore() *FakePromptStore {
	return &FakePromptStore{
		Prompts: make(map[string]string),
	}
}

func (ps *FakePromptStore) GetPrompt(name string) (string, error) {
	if content, ok := ps.Prompts[name]; ok {
		return content, nil
	}
	return "", fmt.Errorf("prompt not found: %s", name)
}

func (ps *FakePromptStore) ListPrompts() []string {
	var names []string
	for name := range ps.Prompts {
		names = append(names, name)
	}
	return names
}
