package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRepoConfig_NoFile(t *testing.T) {
	dir := t.TempDir()

	cfg, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantTitle := filepath.Base(dir)
	if cfg.Title != wantTitle {
		t.Errorf("Title = %q, want %q (directory basename)", cfg.Title, wantTitle)
	}
	if cfg.Theme != "default" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "default")
	}
}

func TestLoadRepoConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte("title: My Docs\ntheme: dracula\n")
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yml"), content, 0o644)

	cfg, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Title != "My Docs" {
		t.Errorf("Title = %q, want %q", cfg.Title, "My Docs")
	}
	if cfg.Theme != "dracula" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "dracula")
	}
}

func TestLoadRepoConfig_PartialFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte("title: Custom Title\n")
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yml"), content, 0o644)

	cfg, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Title != "Custom Title" {
		t.Errorf("Title = %q, want %q", cfg.Title, "Custom Title")
	}
	if cfg.Theme != "default" {
		t.Errorf("Theme = %q, want %q (should fallback to default)", cfg.Theme, "default")
	}
}

func TestLoadRepoConfig_InvalidTheme(t *testing.T) {
	dir := t.TempDir()
	content := []byte("title: Test\ntheme: invalid-theme\n")
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yml"), content, 0o644)

	cfg, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Theme != "default" {
		t.Errorf("Theme = %q, want %q (invalid theme should fallback)", cfg.Theme, "default")
	}
}

func TestLoadRepoConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	content := []byte("title: [unclosed bracket\n")
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yml"), content, 0o644)

	_, err := LoadRepoConfig(dir)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadRepoConfig_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yml"), []byte(""), 0o644)

	cfg, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty file = all defaults
	if cfg.Title != filepath.Base(dir) {
		t.Errorf("Title = %q, want directory basename", cfg.Title)
	}
	if cfg.Theme != "default" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "default")
	}
}

func TestLoadRepoConfig_AllThemes(t *testing.T) {
	for _, theme := range ValidThemes {
		t.Run(theme, func(t *testing.T) {
			dir := t.TempDir()
			content := []byte("theme: " + theme + "\n")
			os.WriteFile(filepath.Join(dir, ".markdown-kb.yml"), content, 0o644)

			cfg, err := LoadRepoConfig(dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Theme != theme {
				t.Errorf("Theme = %q, want %q", cfg.Theme, theme)
			}
		})
	}
}

func TestLoadRepoConfig_YamlExtension(t *testing.T) {
	dir := t.TempDir()
	content := []byte("title: YAML Ext\ntheme: nord\n")
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yaml"), content, 0o644)

	cfg, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Title != "YAML Ext" {
		t.Errorf("Title = %q, want %q", cfg.Title, "YAML Ext")
	}
	if cfg.Theme != "nord" {
		t.Errorf("Theme = %q, want %q", cfg.Theme, "nord")
	}
}

func TestLoadRepoConfig_YmlTakesPrecedence(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yml"), []byte("title: From YML\n"), 0o644)
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yaml"), []byte("title: From YAML\n"), 0o644)

	cfg, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Title != "From YML" {
		t.Errorf("Title = %q, want %q (.yml should take precedence)", cfg.Title, "From YML")
	}
}

func TestLoadRepoConfig_FontPreset(t *testing.T) {
	dir := t.TempDir()
	content := []byte("font: noto-sans\n")
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yml"), content, 0o644)

	cfg, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Font != "noto-sans" {
		t.Errorf("Font = %q, want %q", cfg.Font, "noto-sans")
	}
}

func TestLoadRepoConfig_InvalidFont(t *testing.T) {
	dir := t.TempDir()
	content := []byte("font: comic-sans\n")
	os.WriteFile(filepath.Join(dir, ".markdown-kb.yml"), content, 0o644)

	cfg, err := LoadRepoConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Font != "default" {
		t.Errorf("Font = %q, want %q (invalid font should fallback)", cfg.Font, "default")
	}
}

func TestGetFontPreset(t *testing.T) {
	p := GetFontPreset("rounded")
	if p == nil {
		t.Fatal("expected non-nil preset for 'rounded'")
	}
	if p.Label != "M PLUS Rounded 1c" {
		t.Errorf("Label = %q, want %q", p.Label, "M PLUS Rounded 1c")
	}
	if GetFontPreset("nonexistent") != nil {
		t.Error("expected nil for nonexistent font")
	}
}

func TestIsValidTheme(t *testing.T) {
	if !isValidTheme("dracula") {
		t.Error("dracula should be valid")
	}
	if isValidTheme("nonexistent") {
		t.Error("nonexistent should be invalid")
	}
}
