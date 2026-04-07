// Package store implements the AssetStore interface for mcp-server-pcd.
// Assets (templates, hints, prompts) are embedded at build time via embed.FS.
// Filesystem overlays are applied at startup with last-wins precedence.
// SPDX-License-Identifier: GPL-2.0-only

package store

import (
	"embed"
	"errors"
	iofs "io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// ── Embedded asset declarations ───────────────────────────────────────────────
// These directives require internal/store/assets/{templates,hints,prompts}/
// to be populated by `make embed-assets` before compilation.

//go:embed assets/templates
var embeddedTemplates embed.FS

//go:embed assets/hints
var embeddedHints embed.FS

//go:embed assets/prompts
var embeddedPrompts embed.FS

// ── Types ─────────────────────────────────────────────────────────────────────

// TemplateRecord holds metadata and content for a deployment template.
type TemplateRecord struct {
	Name     string
	Version  string
	Language string
	Content  string
}

// ErrNotFound is returned when an asset key is not found.
var ErrNotFound = errors.New("not found")

// ── Key derivation ────────────────────────────────────────────────────────────

// assetKey strips directory prefix and suffix to produce a map key.
// suffix is one of ".template", ".hints", or "" (for prompts).
//
// Key derivation rules (per TOOLCHAIN-CONSTRAINTS):
//   templates: "cli-tool.template.md"         -> "cli-tool"
//   hints:     "cli-tool.go.milestones.hints.md" -> "cli-tool.go.milestones"
//   prompts:   "interview-prompt.md"          -> "interview"
//              "reverse-prompt.md"            -> "reverse"
//              "prompt.md"                    -> "translator"  (special mapping)
func assetKey(p, suffix string) string {
	base := path.Base(p)
	if suffix != "" {
		// templates and hints: strip "<suffix>.md"
		base = strings.TrimSuffix(base, suffix+".md")
	} else {
		// prompts: strip .md, then strip -prompt suffix if present
		base = strings.TrimSuffix(base, ".md")
		base = strings.TrimSuffix(base, "-prompt")
		// Special mapping: bare "prompt" stem -> "translator"
		if base == "prompt" {
			base = "translator"
		}
	}
	return base
}

// ── Embedded asset loader ─────────────────────────────────────────────────────

func loadEmbedded(fsys embed.FS, dir, suffix string) (map[string]string, error) {
	result := make(map[string]string)
	err := iofs.WalkDir(fsys, dir, func(p string, d iofs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		data, readErr := fsys.ReadFile(p)
		if readErr != nil {
			return readErr
		}
		result[assetKey(p, suffix)] = string(data)
		return nil
	})
	return result, err
}

// ── Overlay directories ───────────────────────────────────────────────────────

func overlayDirs(sub string) []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join("/usr/share/pcd", sub),
		filepath.Join("/etc/pcd", sub),
		filepath.Join(home, ".config", "pcd", sub),
		filepath.Join(".pcd", sub),
	}
}

func applyOverlay(base map[string]string, dir, suffix string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return // directory absent — silently skip
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		key := assetKey(e.Name(), suffix)
		base[key] = string(data)
	}
}

// ── Template metadata parser ──────────────────────────────────────────────────

// parseTemplateRecord extracts Name, Version, Language from template Markdown content.
func parseTemplateRecord(name, content string) TemplateRecord {
	rec := TemplateRecord{Name: name, Content: content}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Version:") {
			rec.Version = strings.TrimSpace(strings.TrimPrefix(line, "Version:"))
		} else if strings.HasPrefix(line, "| LANGUAGE |") || strings.Contains(line, "| LANGUAGE |") {
			// parse TEMPLATE-TABLE row: | LANGUAGE | Go | default | ...
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				key := strings.TrimSpace(parts[1])
				val := strings.TrimSpace(parts[2])
				constraint := strings.TrimSpace(parts[3])
				if strings.EqualFold(key, "LANGUAGE") && strings.EqualFold(constraint, "default") {
					rec.Language = val
				}
			}
		}
	}
	return rec
}

// ── EmbeddedLayeredStore ──────────────────────────────────────────────────────

// EmbeddedLayeredStore implements AssetStore using embedded assets as base
// and filesystem overlays applied at startup.
type EmbeddedLayeredStore struct {
	templates map[string]string // key -> content
	hints     map[string]string // key -> content
	prompts   map[string]string // key -> content
}

// NewEmbeddedLayeredStore loads embedded assets and applies filesystem overlays.
func NewEmbeddedLayeredStore() (*EmbeddedLayeredStore, error) {
	s := &EmbeddedLayeredStore{}

	var err error
	s.templates, err = loadEmbedded(embeddedTemplates, "assets/templates", ".template")
	if err != nil {
		return nil, err
	}
	s.hints, err = loadEmbedded(embeddedHints, "assets/hints", ".hints")
	if err != nil {
		return nil, err
	}
	s.prompts, err = loadEmbedded(embeddedPrompts, "assets/prompts", "")
	if err != nil {
		return nil, err
	}

	// Apply filesystem overlays (last-wins, ascending precedence)
	for _, dir := range overlayDirs("templates") {
		applyOverlay(s.templates, dir, ".template")
	}
	for _, dir := range overlayDirs("hints") {
		applyOverlay(s.hints, dir, ".hints")
	}
	for _, dir := range overlayDirs("prompts") {
		applyOverlay(s.prompts, dir, "")
	}

	return s, nil
}

// ListTemplates returns all template records (without content).
func (s *EmbeddedLayeredStore) ListTemplates() ([]TemplateRecord, error) {
	var records []TemplateRecord
	for name, content := range s.templates {
		rec := parseTemplateRecord(name, content)
		rec.Content = "" // omit content in list
		records = append(records, rec)
	}
	return records, nil
}

// GetTemplate returns the full TemplateRecord for the given name and version.
// version "latest" resolves to the only installed version.
func (s *EmbeddedLayeredStore) GetTemplate(name, version string) (TemplateRecord, error) {
	content, ok := s.templates[name]
	if !ok {
		return TemplateRecord{}, ErrNotFound
	}
	rec := parseTemplateRecord(name, content)
	return rec, nil
}

// GetHints returns the content of the hints file for the given key.
func (s *EmbeddedLayeredStore) GetHints(key string) (string, error) {
	content, ok := s.hints[key]
	if !ok {
		return "", ErrNotFound
	}
	return content, nil
}

// ListHintsKeys returns all known hints keys.
func (s *EmbeddedLayeredStore) ListHintsKeys() ([]string, error) {
	keys := make([]string, 0, len(s.hints))
	for k := range s.hints {
		keys = append(keys, k)
	}
	return keys, nil
}

// GetPrompt returns the content of the prompt for the given name.
func (s *EmbeddedLayeredStore) GetPrompt(name string) (string, error) {
	content, ok := s.prompts[name]
	if !ok {
		return "", ErrNotFound
	}
	return content, nil
}

// ListPrompts returns all known prompt names.
func (s *EmbeddedLayeredStore) ListPrompts() ([]string, error) {
	names := make([]string, 0, len(s.prompts))
	for k := range s.prompts {
		names = append(names, k)
	}
	return names, nil
}

// ── FakeStore (test double) ───────────────────────────────────────────────────

// FakeStore is an in-memory AssetStore for use in tests.
// No filesystem access. No embedded assets.
type FakeStore struct {
	Templates []TemplateRecord
	Hints     map[string]string
	Prompts   map[string]string
}

func (f *FakeStore) ListTemplates() ([]TemplateRecord, error) {
	result := make([]TemplateRecord, len(f.Templates))
	for i, t := range f.Templates {
		result[i] = TemplateRecord{
			Name:     t.Name,
			Version:  t.Version,
			Language: t.Language,
			// Content omitted in list
		}
	}
	return result, nil
}

func (f *FakeStore) GetTemplate(name, version string) (TemplateRecord, error) {
	for _, t := range f.Templates {
		if t.Name == name {
			return t, nil
		}
	}
	return TemplateRecord{}, ErrNotFound
}

func (f *FakeStore) GetHints(key string) (string, error) {
	if f.Hints == nil {
		return "", ErrNotFound
	}
	content, ok := f.Hints[key]
	if !ok {
		return "", ErrNotFound
	}
	return content, nil
}

func (f *FakeStore) ListHintsKeys() ([]string, error) {
	keys := make([]string, 0, len(f.Hints))
	for k := range f.Hints {
		keys = append(keys, k)
	}
	return keys, nil
}

func (f *FakeStore) GetPrompt(name string) (string, error) {
	if f.Prompts == nil {
		return "", ErrNotFound
	}
	content, ok := f.Prompts[name]
	if !ok {
		return "", ErrNotFound
	}
	return content, nil
}

func (f *FakeStore) ListPrompts() ([]string, error) {
	names := make([]string, 0, len(f.Prompts))
	for k := range f.Prompts {
		names = append(names, k)
	}
	return names, nil
}

// ── AssetStore interface ──────────────────────────────────────────────────────

// AssetStore is the interface satisfied by both EmbeddedLayeredStore and FakeStore.
type AssetStore interface {
	ListTemplates() ([]TemplateRecord, error)
	GetTemplate(name, version string) (TemplateRecord, error)
	GetHints(key string) (string, error)
	ListHintsKeys() ([]string, error)
	GetPrompt(name string) (string, error)
	ListPrompts() ([]string, error)
}
