package main

import (
	"embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

const defaultTheme = "standard"

// Load loads themes/${theme}.yml and returns a map of strings for
// converted themes.
func loadTheme(allThemes embed.FS, theme string) (map[string]string, error) {
	if theme == "" {
		theme = defaultTheme
	}

	t := make(map[string]string)
	file := fmt.Sprintf("themes/%v.yml", theme)

	b, err := allThemes.ReadFile(file)
	if err != nil {
		return t, fmt.Errorf("failed to load file %v: %w", file, err)
	}

	err = yaml.Unmarshal(b, &t)
	if err != nil {
		return t, fmt.Errorf("failed to unmarshal file %v: %w", file, err)
	}

	return t, nil
}

// Load loads themes/${theme}.yml and returns a map of strings of colors. As a
// fallback, it loads the defaultTheme, so that strings that are not set to a
// color will still show visible text in some sense.
func loadThemes(allThemes embed.FS, theme string) (map[string]string, error) {
	t, err := loadTheme(allThemes, defaultTheme)
	if err != nil {
		return t, fmt.Errorf("failed to load default themes %v: %w", defaultTheme, err)
	}

	switch theme {
	case "":
		fallthrough
	case defaultTheme:
		return t, nil
	default:
		break
	}

	u, err := loadTheme(allThemes, theme)
	if err != nil {
		return t, fmt.Errorf("failed to load specified themes %v: %w", theme, err)
	}

	// merge the two maps
	for k, v := range u {
		t[k] = v
	}

	return t, nil
}
