package config

import "os"

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

// InitEnvConfig initializes the configuration from environment variables.
func InitEnvConfig() {
	C = Config{
		JiraURL:                  os.Getenv("JIRA_URL"),
		JiraUser:                 os.Getenv("JIRA_USER"),
		JiraToken:                os.Getenv("JIRA_TOKEN"),
		TempoURL:                 os.Getenv("TEMPO_URL"),
		TempoToken:               os.Getenv("TEMPO_TOKEN"),
		TogglToken:               os.Getenv("TOGGL_TOKEN"),
		TogglGroupSimilarEntries: true,
	}
}
