package config

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

var repoConfigFiles = []string{".markdown-kb.yml", ".markdown-kb.yaml"}

// ValidThemes lists all supported theme names.
var ValidThemes = []string{
	"default",
	"tokyo-night",
	"dracula",
	"nord",
	"solarized",
	"monokai",
	"github",
	"catppuccin",
	"gruvbox",
	"rose-pine",
}

// FontPreset defines a font configuration with Google Fonts URL and CSS family.
type FontPreset struct {
	Name   string `json:"name"`
	Label  string `json:"label"`
	URL    string `json:"url,omitempty"`
	Family string `json:"family"`
}

// ValidFonts lists all supported font presets (Japanese + English).
var ValidFonts = []FontPreset{
	{
		Name:   "default",
		Label:  "BIZ UDPGothic (Default)",
		URL:    "https://fonts.googleapis.com/css2?family=BIZ+UDPGothic:wght@400;700&display=swap",
		Family: `"BIZ UDPGothic", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`,
	},
	{
		Name:   "noto-sans",
		Label:  "Noto Sans JP",
		URL:    "https://fonts.googleapis.com/css2?family=Noto+Sans+JP:wght@400;500;700&display=swap",
		Family: `"Noto Sans JP", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`,
	},
	{
		Name:   "rounded",
		Label:  "M PLUS Rounded 1c",
		URL:    "https://fonts.googleapis.com/css2?family=M+PLUS+Rounded+1c:wght@400;500;700&display=swap",
		Family: `"M PLUS Rounded 1c", -apple-system, BlinkMacSystemFont, sans-serif`,
	},
	{
		Name:   "serif",
		Label:  "Noto Serif JP",
		URL:    "https://fonts.googleapis.com/css2?family=Noto+Serif+JP:wght@400;700&display=swap",
		Family: `"Noto Serif JP", "Georgia", "Times New Roman", serif`,
	},
	{
		Name:   "zen-kaku",
		Label:  "Zen Kaku Gothic New",
		URL:    "https://fonts.googleapis.com/css2?family=Zen+Kaku+Gothic+New:wght@400;500;700&display=swap",
		Family: `"Zen Kaku Gothic New", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`,
	},
}

// ValidFontNames returns the list of valid font preset names.
func ValidFontNames() []string {
	names := make([]string, len(ValidFonts))
	for i, f := range ValidFonts {
		names[i] = f.Name
	}
	return names
}

// GetFontPreset returns the FontPreset for the given name, or nil if not found.
func GetFontPreset(name string) *FontPreset {
	for _, f := range ValidFonts {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

// TagIcon maps a frontmatter tag to an emoji icon for the sidebar tree.
type TagIcon struct {
	Tag   string `yaml:"tag"   json:"tag"`
	Emoji string `yaml:"emoji" json:"emoji"`
}

// RepoConfig holds per-repository configuration loaded from .markdown-kb.yml.
type RepoConfig struct {
	Title    string    `yaml:"title"`
	Theme    string    `yaml:"theme"`
	Font     string    `yaml:"font"`
	TagIcons []TagIcon `yaml:"tag_icons"`
}

// LoadRepoConfig reads .markdown-kb.yml from rootDir.
// Returns a config with sensible defaults if the file doesn't exist.
func LoadRepoConfig(rootDir string) (RepoConfig, error) {
	cfg := RepoConfig{
		Title: filepath.Base(rootDir),
		Theme: "default",
		Font:  "default",
	}

	var data []byte
	for _, name := range repoConfigFiles {
		d, err := os.ReadFile(filepath.Join(rootDir, name))
		if err == nil {
			data = d
			break
		}
		if !os.IsNotExist(err) {
			return cfg, err
		}
	}
	if data == nil {
		return cfg, nil
	}

	var fileCfg RepoConfig
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return cfg, err
	}

	if fileCfg.Title != "" {
		cfg.Title = fileCfg.Title
	}
	if fileCfg.Theme != "" && isValidTheme(fileCfg.Theme) {
		cfg.Theme = fileCfg.Theme
	}
	if fileCfg.Font != "" && GetFontPreset(fileCfg.Font) != nil {
		cfg.Font = fileCfg.Font
	}
	if len(fileCfg.TagIcons) > 0 {
		cfg.TagIcons = fileCfg.TagIcons
	}

	return cfg, nil
}

func isValidTheme(name string) bool {
	for _, t := range ValidThemes {
		if t == name {
			return true
		}
	}
	return false
}
