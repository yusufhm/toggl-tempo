package main

import (
	"flag"
	"strconv"

	log "github.com/sirupsen/logrus"

	"toggl-tempo/pkg/config"
	"toggl-tempo/pkg/jira"
	"toggl-tempo/pkg/tempo"
	"toggl-tempo/pkg/toggl"
)

var finalLogLevel = log.WarnLevel

func init() {
	logLevel := flag.String("log-level", "warn", "log level")
	verbose := flag.Bool("verbose", false, "verbose mode - alias for '-log-level info'")
	debug := flag.Bool("debug", false, "debug mode - alias for '-log-level debug'")
	flag.Parse()

	if *verbose {
		*logLevel = "info"
	}
	if *debug {
		*logLevel = "debug"
	}

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.WithError(err).Fatal("failed to parse log level")
	}
	finalLogLevel = level

	log.SetLevel(finalLogLevel)
	config.InitEnvConfig()
}

func main() {
	jiraAccountId := jira.MustGetAccountId()
	worklogs := tempo.MustGetCurrentWeekEntries(nil)
	log.WithField("tempo worklogs", len(worklogs)).Info("current week's entries")

	entries := toggl.MustGetCurrentWeekEntries()
	log.WithField("toggl entries", len(entries)).Info("current week's entries")
	for _, entry := range entries {
		if entry.Stop == nil {
			continue
		}

		if entry.Pid == nil {
			continue
		}

		if entry.HasTag("synced") {
			log.WithField("entry", entry).Debug("entry already synced")
			continue
		}

		project := toggl.MustGetProject(*entry.Pid, entry.Wid)
		jiraIssueKey := jira.JiraKeyFromString(project.Name)
		jiraIssueId, jiraIssueEstimate, err := jira.GetIssueIdEstimate(jiraIssueKey)
		if err != nil {
			log.Warn(err)
			continue
		}

		entryId := strconv.Itoa(entry.ID)
		input := tempo.WorklogCreateInput{
			IssueID:                  jiraIssueId,
			AuthorAccountID:          jiraAccountId,
			Description:              entry.Description,
			StartDate:                entry.Start.Format("2006-01-02"),
			StartTime:                entry.Start.Format("15:04:05"),
			TimeSpentSeconds:         int(entry.Duration),
			RemainingEstimateSeconds: jiraIssueEstimate,
		}
		tempo.MustCreateWorklog(input, entryId)
		if err := toggl.AddTimeEntryTag(entry, "synced"); err != nil {
			log.WithError(err).WithField("entry", entry).Error("failed to tag entry")
		}
	}
}
