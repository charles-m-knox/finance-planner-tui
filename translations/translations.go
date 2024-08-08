package translations

import (
	"embed"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

const defaultLanguage = "en_US.UTF-8"

// Load loads translations/${language}.yml and returns a map of strings for
// converted languages.
func load(allTranslations embed.FS, language string) (map[string]string, error) {
	if language == "" {
		language = defaultLanguage
	}

	t := make(map[string]string)
	file := fmt.Sprintf("translations/%v.yml", language)

	b, err := allTranslations.ReadFile(file)
	if err != nil {
		log.Printf("failed to load file %v: %v", file, err.Error())

		file := fmt.Sprintf("translations/%v.yml", defaultLanguage)

		b, err = allTranslations.ReadFile(file)
		if err != nil {
			return t, fmt.Errorf("failed to load default language file %v: %w", file, err)
		}
	}

	err = yaml.Unmarshal(b, &t)
	if err != nil {
		return t, fmt.Errorf("failed to unmarshal file %v: %w", file, err)
	}

	return t, nil
}

// Load loads translations/${language}.yml and returns a map of strings for
// converted languages. As a fallback, it loads the defaultLanguage, so that
// strings that are not yet translated will still show visible text in some
// language (instead of an empty string).
func Load(allTranslations embed.FS) (map[string]string, error) {
	t, err := load(allTranslations, defaultLanguage)
	if err != nil {
		return t, fmt.Errorf("failed to load default translations %v: %w", defaultLanguage, err)
	}

	language := os.Getenv("LANG")

	switch language {
	case "":
		fallthrough
	case defaultLanguage:
		return t, nil
	default:
		break
	}

	u, err := load(allTranslations, language)
	if err != nil {
		return t, fmt.Errorf("failed to load specified translations %v: %w", language, err)
	}

	// merge the two maps
	for k, v := range u {
		t[k] = v
	}

	return t, nil
}
