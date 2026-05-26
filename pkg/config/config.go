package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// JiraURL is the URL of the Jira instance
	JiraURL string `yaml:"jira-url"`
	// JiraUser is the username to authenticate with Jira
	JiraUser string `yaml:"jira-user"`
	// JiraToken is the token to authenticate with Jira
	JiraToken string `yaml:"jira-token"`

	// TempoURL is the URL of the Tempo instance
	TempoURL string `yaml:"tempo-url"`
	// TempoToken is the token to authenticate with Tempo
	TempoToken string `yaml:"tempo-token"`
	// TempoWorkAttributeKey is the key of the work attribute in Tempo
	TempoWorkAttributeKey string `yaml:"tempo-work-attribute-key"`

	// TogglToken is the token to authenticate with Toggl
	TogglToken string `yaml:"toggl-token"`
	// TogglGroupSimilarEntries determines whether similar entries should be
	// grouped. This will create a single worklog in Tempo for all similar
	// entries in Toggl in a single day.
	TogglGroupSimilarEntries bool `yaml:"toggl-group-similar-entries"`
}

var C Config

const appDirName = "toggl-tempo"
const configFileName = "config.yaml"

// configFilePath returns the path to the config file.
//
// It respects $XDG_CONFIG_HOME when set. Otherwise it falls back to
// ~/.config/toggl-tempo/config.yaml, which is the convention for CLI tools on
// both Linux and macOS.
func configFilePath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, appDirName, configFileName), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home dir: %w", err)
	}
	return filepath.Join(home, ".config", appDirName, configFileName), nil
}

// loadFile reads the YAML config file at path and unmarshals it into C. It
// returns fs.ErrNotExist when the file does not exist so callers can ignore
// missing files.
func loadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, &C); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func createEmptyConfigFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	return f.Close()
}

// Init initializes the configuration from a YAML file (if present) and then
// overlays non-empty environment variables, which always take precedence.
func Init() error {
	path, err := configFilePath()
	if err == nil {
		if loadErr := loadFile(path); loadErr != nil {
			if !errors.Is(loadErr, fs.ErrNotExist) {
				return loadErr
			}
			if err := createEmptyConfigFile(path); err != nil {
				return err
			}
		}
	}

	if v := os.Getenv("JIRA_URL"); v != "" {
		C.JiraURL = v
	}
	if v := os.Getenv("JIRA_USER"); v != "" {
		C.JiraUser = v
	}
	if v := os.Getenv("JIRA_TOKEN"); v != "" {
		C.JiraToken = v
	}
	if v := os.Getenv("TEMPO_URL"); v != "" {
		C.TempoURL = v
	}
	if v := os.Getenv("TEMPO_TOKEN"); v != "" {
		C.TempoToken = v
	}
	if v := os.Getenv("TEMPO_WORK_ATTRIBUTE_KEY"); v != "" {
		C.TempoWorkAttributeKey = v
	}
	if v := os.Getenv("TOGGL_TOKEN"); v != "" {
		C.TogglToken = v
	}
	if v := os.Getenv("TOGGL_GROUP_SIMILAR_ENTRIES"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			C.TogglGroupSimilarEntries = b
		}
	}

	return nil
}
