package main

import (
	"flag"
	"fmt"
	"strconv"

	origToggl "github.com/jason0x43/go-toggl"
	log "github.com/sirupsen/logrus"

	"toggl-tempo/pkg/config"
	"toggl-tempo/pkg/jira"
	"toggl-tempo/pkg/tempo"
	"toggl-tempo/pkg/time"
	"toggl-tempo/pkg/toggl"
)

var finalLogLevel = log.WarnLevel
var lastWeek bool

func init() {
	logLevel := flag.String("log-level", "warn", "log level")
	verbose := flag.Bool("verbose", false, "verbose mode - alias for '-log-level info'")
	debug := flag.Bool("debug", false, "debug mode - alias for '-log-level debug'")
	flag.BoolVar(&lastWeek, "last-week", false, "sync last week's entries")
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

	var entries []origToggl.TimeEntry
	if lastWeek {
		entries = toggl.MustGetLastWeekEntries()
		log.WithField("toggl entries", len(entries)).Info("last week's entries")
	} else {
		entries = toggl.MustGetCurrentWeekEntries()
		log.WithField("toggl entries", len(entries)).Info("current week's entries")
	}
	if log.GetLevel() == log.DebugLevel {
		for _, e := range entries {
			log.WithField("entry", fmt.Sprintf("%+v", e)).Debug()
		}
	}

	filteredEntries, entriesToTag := toggl.FilterEntries(entries)
	if log.GetLevel() == log.DebugLevel {
		for _, e := range entries {
			log.WithField("entry", fmt.Sprintf("%+v", e)).Debug("filtered entry")
		}
	}

	for day, entries := range filteredEntries {
		log.WithField("day", day).Info("adding day's entries")
		for _, entry := range entries {
			project := toggl.MustGetProject(*entry.Pid, entry.Wid)
			jiraIssueKey := jira.JiraKeyFromString(project.Name)
			jiraIssueId, jiraIssueEstimate, err := jira.GetIssueIdEstimate(jiraIssueKey)
			if err != nil {
				log.Warn(err)
				continue
			}

			entryId := strconv.Itoa(entry.ID)
			start := entry.Start.In(time.Location())
			input := tempo.WorklogCreateInput{
				IssueID:                  jiraIssueId,
				AuthorAccountID:          jiraAccountId,
				Description:              entry.Description,
				StartDate:                start.Format("2006-01-02"),
				StartTime:                start.Format("15:04:05"),
				TimeSpentSeconds:         int(entry.Duration),
				RemainingEstimateSeconds: jiraIssueEstimate,
			}
			tempo.MustCreateWorklog(input, entryId)
		}
	}

	for _, entry := range entriesToTag {
		log.WithField("entry", entry).Trace("tagging entry as synced")
		if err := toggl.AddTimeEntryTag(entry, "synced"); err != nil {
			log.WithError(err).WithField("entry", entry).Error("failed to tag entry")
		}
	}
}
